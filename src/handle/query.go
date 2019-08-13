// Copyright 2019 HenryYee.
//
// Licensed under the AGPL, Version 3.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.gnu.org/licenses/agpl-3.0.en.html
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package handle

import (
	"Yearning-go/src/lib"
	"Yearning-go/src/modal"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type FieldInfo struct {
	gorm.Model

	Field      string  `gorm:"Column:Field";json:"field"`
	Type       string  `gorm:"Column:Type";json:"type"`
	Collation  string  `gorm:"Column:Collation";json:"collation"`
	Null       string  `gorm:"Column:Null";json:"null"`
	Key        string  `gorm:"Column:Key";json:"key"`
	Default    *string `gorm:"Column:Default";json:"default"`
	Extra      string  `gorm:"Column:Extra";json:"extra"`
	Privileges string  `gorm:"Column:Privileges";json:"privileges"`
	Comment    string  `gorm:"Column:Comment";json:"comment"`

	IsDeleted bool `gorm:"-"`
	IsNew     bool `gorm:"-"`
}

type IndexInfo struct {
	gorm.Model

	Table      string `gorm:"Column:Table"`
	NonUnique  int    `gorm:"Column:Non_unique"`
	IndexName  string `gorm:"Column:Key_name"`
	Seq        int    `gorm:"Column:Seq_in_index"`
	ColumnName string `gorm:"Column:Column_name"`
	IndexType  string `gorm:"Column:Index_type"`

	IsDeleted bool `gorm:"-"`
}

type queryOrder struct {
	IDC      string
	Source   string
	Export   uint
	Assigned string
	Text     string
	WorkId   string
}

type clearQueryOrder struct {
	WorkId []string
}

func ReferQueryOrder(c echo.Context) (err error) {
	var u modal.CoreAccount
	var t modal.CoreQueryOrder
	user, _ := lib.JwtParse(c)

	d := new(queryOrder)
	if err = c.Bind(d); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusInternalServerError, "")
	}

	state := 1

	if modal.GloOther.Query {
		state = 2
	}

	modal.DB().Select("real_name").Where("username =?", user).First(&u)

	if modal.DB().Model(modal.CoreQueryOrder{}).Where("username =? and query_per =?", user, 2).First(&t).RecordNotFound() {
		work := lib.GenWorkid()
		modal.DB().Create(&modal.CoreQueryOrder{
			WorkId:   work,
			Username: user,
			Date:     time.Now().Format("2006-01-02 15:04"),
			Text:     d.Text,
			Assigned: d.Assigned,
			Export:   d.Export,
			IDC:      d.IDC,
			Source:   d.Source,
			QueryPer: state,
			Realname: u.RealName,
			ExDate:   time.Now().Format("2006-01-02 15:04"),
		})

		lib.MessagePush(c, work, 6, "")

		return c.JSON(http.StatusOK, "查询工单已提交!")
	}
	return c.JSON(http.StatusOK, "重复提交!")
}

func FetchQueryStatus(c echo.Context) (err error) {

	user, _ := lib.JwtParse(c)

	var d modal.CoreQueryOrder

	modal.DB().Where("username =?", user).Last(&d)

	return c.JSON(http.StatusOK, map[string]interface{}{"status": d.QueryPer, "export": modal.GloOther.Export})
}

func FetchQueryDatabaseInfo(c echo.Context) (err error) {
	user, _ := lib.JwtParse(c)
	var d modal.CoreQueryOrder
	var u modal.CoreDataSource
	var sign modal.CoreGrained
	var ass modal.PermissionList

	modal.DB().Where("username =?", user).Last(&d)

	modal.DB().Where("username =?", user).First(&sign)

	if err := json.Unmarshal(sign.Permissions, &ass); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	if d.QueryPer == 1 {

		var dataBase string

		var dc [] string

		var mid [] string

		var highlist []map[string]string

		var baselist []map[string]interface{}

		modal.DB().Where("source =?", d.Source).First(&u)

		ps := lib.Decrypt(u.Password)

		db, e := gorm.Open("mysql", fmt.Sprintf("%s:%s@(%s:%d)/?charset=utf8&parseTime=True&loc=Local", u.Username, ps, u.IP, u.Port))

		defer db.Close()

		if e != nil {
			c.Logger().Error(e.Error())
			return c.JSON(http.StatusInternalServerError, "数据库实例连接失败！请检查相关配置是否正确！")
		}

		sql := "SHOW DATABASES;"
		rows, err := db.Raw(sql).Rows()
		if err != nil {
			c.Logger().Error(err.Error())
		}
		defer rows.Close()

		for rows.Next() {
			rows.Scan(&dataBase)
			dc = append(dc, dataBase)
		}

		if len(modal.GloOther.ExcludeDbList) > 0 {
			mid = lib.Intersect(dc, modal.GloOther.ExcludeDbList)
			dc = lib.NonIntersect(mid, dc)
		}

		for _, z := range dc {
			highlist = append(highlist, map[string]string{"vl": z, "meta": "库名"})
			baselist = append(baselist, map[string]interface{}{"title": z, "children": []map[string]string{{}}})
		}

		var info [] map[string]interface{}

		info = append(info, map[string]interface{}{
			"title":    d.Source,
			"expand":   "true",
			"children": baselist,
		})

		return c.JSON(http.StatusOK, map[string]interface{}{"info": info, "status": d.Export, "highlight": highlist, "sign": ass.Auditor})

	} else {
		return c.JSON(http.StatusOK, 0)
	}
}

func FetchQueryTableInfo(c echo.Context) (err error) {
	user, _ := lib.JwtParse(c)
	t := c.Param("t")
	var d modal.CoreQueryOrder
	var u modal.CoreDataSource
	modal.DB().Where("username =?", user).Last(&d)
	if d.QueryPer == 1 {

		var table string

		var highlist []map[string]string

		var tablelist []map[string]interface{}

		modal.DB().Where("source =?", d.Source).First(&u)

		ps := lib.Decrypt(u.Password)

		db, e := gorm.Open("mysql", fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", u.Username, ps, u.IP, u.Port, t))

		defer db.Close()

		if e != nil {
			c.Logger().Error(e.Error())
			return c.JSON(http.StatusInternalServerError, "数据库实例连接失败！请检查相关配置是否正确！")
		}

		sql := "show tables"
		rows, err := db.Raw(sql).Rows()
		if err != nil {
			c.Logger().Error(err.Error())
		}
		defer rows.Close()
		for rows.Next() {
			rows.Scan(&table)
			highlist = append(highlist, map[string]string{"vl": table, "meta": "表名"})
			tablelist = append(tablelist, map[string]interface{}{"title": table})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"table": tablelist, "highlight": highlist})

	} else {
		return c.JSON(http.StatusNonAuthoritativeInfo, "没有查询权限！")
	}
}

func FetchQueryTableStruct(c echo.Context) (err error) {
	t := c.Param("table")
	b := c.Param("base")
	user, _ := lib.JwtParse(c)
	var d modal.CoreQueryOrder
	var u modal.CoreDataSource
	var f []FieldInfo
	modal.DB().Where("username =?", user).Last(&d)
	modal.DB().Where("source =?", d.Source).First(&u)
	ps := lib.Decrypt(u.Password)

	db, e := gorm.Open("mysql", fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", u.Username, ps, u.IP, u.Port, b))
	if e != nil {
		c.Logger().Error(e.Error())
		return c.JSON(http.StatusInternalServerError, "数据库实例连接失败！请检查相关配置是否正确！")
	}
	defer db.Close()

	if err := db.Raw(fmt.Sprintf("SHOW FULL FIELDS FROM `%s`.`%s`", b, t)).Scan(&f).Error; err != nil {
		c.Logger().Error(err.Error())
	}

	return c.JSON(http.StatusOK, f)
}

func FetchQueryResults(c echo.Context) (err error) {

	req := new(modal.Queryresults)

	if err = c.Bind(req); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusOK, err.Error())
	}

	var d modal.CoreQueryOrder

	var u modal.CoreDataSource

	sts := false

	user, _ := lib.JwtParse(c)

	modal.DB().Where("username =? AND query_per =?", user, 1).Last(&d)

	if lib.TimeDifference(d.ExDate) {
		modal.DB().Model(modal.CoreQueryOrder{}).Update(&modal.CoreQueryOrder{QueryPer: 3})
		sts = true
	}

	modal.DB().Where("source =?", d.Source).First(&u)

	//todo 需自行实现查询SQL LIMIT限制
	// r.InsulateWordList 为脱敏字段slice 需自行实现脱敏功能
	t1 := time.Now()
	data, err := lib.QueryMethod(&u, req, r.InsulateWordList)
	queryTime := int(time.Since(t1).Seconds() * 1000)

	go func(w string, s string) {
		modal.DB().Create(&modal.CoreQueryRecord{SQL: s, WorkId: w})
	}(d.WorkId, req.Sql)

	if err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusOK, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"title": data.Field, "data": data.Data, "status": sts, "time": queryTime})
}

func AgreedQueryOrder(c echo.Context) (err error) {
	u := new(queryOrder)
	var s modal.CoreQueryOrder
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusOK, err.Error())
	}

	if modal.DB().Where("work_id=? AND query_per=?", u.WorkId, 2).Last(&s).RecordNotFound() {
		return c.JSON(http.StatusOK, "工单状态已变更！")
	}

	modal.DB().Model(modal.CoreQueryOrder{}).Where("work_id =?", u.WorkId).Update(map[string]interface{}{"query_per": 1, "ex_date": time.Now().Format("2006-01-02 15:04")})
	lib.MessagePush(c, u.WorkId, 7, "")
	return c.JSON(http.StatusOK, "该次工单查询已同意！")
}

func DisAgreedQueryOrder(c echo.Context) (err error) {
	u := new(queryOrder)
	var s modal.CoreQueryOrder
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusOK, err.Error())
	}

	if modal.DB().Where("work_id=? AND query_per=?", u.WorkId, 2).Last(&s).RecordNotFound() {
		return c.JSON(http.StatusOK, "工单状态已变更！")
	}

	modal.DB().Model(modal.CoreQueryOrder{}).Where("work_id =?", u.WorkId).Update(map[string]interface{}{"query_per": 0})
	lib.MessagePush(c, u.WorkId, 8, "")
	return c.JSON(http.StatusOK, "该次工单查询已驳回！")
}

func DelQueryOrder(c echo.Context) (err error) {
	req := new(clearQueryOrder)
	if err := c.Bind(req); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusOK, err.Error())
	}
	var d modal.CoreQueryOrder

	tx := modal.DB().Begin()
	for _, i := range req.WorkId {
		modal.DB().Where("work_id =?", i).Delete(&modal.CoreQueryOrder{})
		modal.DB().Where("work_id =?", d.WorkId).Delete(&modal.CoreQueryRecord{})
	}
	tx.Commit()

	return c.JSON(http.StatusOK, "查询工单删除!")
}

func UndoQueryOrder(c echo.Context) (err error) {
	user, _ := lib.JwtParse(c)
	modal.DB().Model(modal.CoreQueryOrder{}).Where("username =?", user).Update(map[string]interface{}{"query_per": 3})
	return c.JSON(http.StatusOK, "查询已终止！")
}

func SuperUndoQueryOrder(c echo.Context) (err error) {
	s := new(queryOrder)
	var u modal.CoreQueryOrder
	if err = c.Bind(s); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusOK, err.Error())
	}

	if !modal.DB().Where("work_id=? AND query_per=?", s.WorkId, 2).Last(&u).RecordNotFound() {
		return c.JSON(http.StatusOK, "工单状态已变更！")
	}

	modal.DB().Model(modal.CoreQueryOrder{}).Where("work_id =?", s.WorkId).Update(map[string]interface{}{"query_per": 3})
	return c.JSON(http.StatusOK, "查询已终止！")
}

func FetchQueryOrder(c echo.Context) (err error) {
	u := new(f)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return
	}

	start, end := lib.Paging(u.Page, 20)
	var pg int

	var order []modal.CoreQueryOrder

	whereField := "username LIKE ? "

	dateField := " AND date >= ? AND date <= ?"

	if u.Find.Valve {
		if u.Find.Picker[0] == "" {
			modal.DB().Where(whereField, "%"+fmt.Sprintf("%s", u.Find.User)+"%").Order("id desc").Offset(start).Limit(end).Find(&order)
			modal.DB().Model(&modal.CoreQueryOrder{}).Where(whereField, "%"+fmt.Sprintf("%s", u.Find.User)+"%").Count(&pg)
		} else {
			modal.DB().Where(whereField+dateField, "%"+fmt.Sprintf("%s", u.Find.User)+"%", u.Find.Picker[0], u.Find.Picker[1]).Order("id desc").Offset(start).Limit(end).Find(&order)
			modal.DB().Model(&modal.CoreQueryOrder{}).Where(whereField+dateField, "%"+fmt.Sprintf("%s", u.Find.User)+"%", u.Find.Picker[0], u.Find.Picker[1]).Count(&pg)
		}
	} else {
		modal.DB().Order("id desc").Offset(start).Limit(end).Find(&order)
		modal.DB().Model(&modal.CoreQueryOrder{}).Count(&pg)
	}
	return c.JSON(http.StatusOK, struct {
		Data []modal.CoreQueryOrder `json:"data"`
		Page int                    `json:"page"`
	}{
		order,
		pg,
	})
}

func QueryQuickCancel(c echo.Context) (err error) {
	modal.DB().Model(modal.CoreQueryOrder{}).Updates(&modal.CoreQueryOrder{QueryPer: 3})
	return c.JSON(http.StatusOK, "所有查询已取消！")
}

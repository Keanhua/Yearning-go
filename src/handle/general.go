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
	"Yearning-go/src/parser"
	"Yearning-go/src/soar"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type fetch struct {
	Source string
	Base   string
	Table  string
}

type cdx struct {
	F []FieldInfo `json:"f"`
	I []IndexInfo `json:"i"`
}

func GeneralIDC(c echo.Context) (err error) {

	return c.JSON(http.StatusOK, modal.GloOther.IDC)

}

func GeneralSource(c echo.Context) (err error) {
	t := c.Param("idc")
	x := c.Param("xxx")

	if t == "undefined" || x == "undefined" {
		return
	}

	var s modal.CoreGrained
	var p modal.PermissionList
	var sList []string
	var source []modal.CoreDataSource
	var inter []string
	user, _ := lib.JwtParse(c)
	modal.DB().Where("username =?", user).First(&s)
	if err := json.Unmarshal(s.Permissions, &p); err != nil {
		c.Logger().Error(err.Error())
		return err
	}

	modal.DB().Select("source").Where("id_c =?", t).Find(&source)

	if source != nil {
		for _, i := range source {
			sList = append(sList, i.Source)
		}
		if x == "dml" {
			inter = lib.Intersect(p.DMLSource, sList)
		}
		if x == "ddl" {
			inter = lib.Intersect(p.DDLSource, sList)
		}

		if x == "query" {
			inter = lib.Intersect(p.QuerySource, sList)
		}
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"assigned": p.Auditor, "source": inter, "x": x,})
}

func GeneralBase(c echo.Context) (err error) {

	t := c.Param("source")

	var s modal.CoreDataSource
	var dataBase string
	var l []string

	if t == "undefined" {
		return
	}

	modal.DB().Where("source =?", t).First(&s)
	ps := lib.Decrypt(s.Password)

	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@(%s:%s)/?charset=utf8&parseTime=True&loc=Local", s.Username, ps, s.IP, strconv.Itoa(int(s.Port))))

	defer db.Close()

	if err != nil {
		c.Logger().Error(err.Error())
		return
	}
	sql := "SHOW DATABASES;"
	rows, err := db.Raw(sql).Rows()
	if err != nil {
		c.Logger().Error(err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&dataBase)
		l = append(l, dataBase)
	}

	return c.JSON(http.StatusOK, l)
}

func GeneralTable(c echo.Context) (err error) {
	u := new(fetch)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusInternalServerError, "")
	}
	var s modal.CoreDataSource
	var table string
	var l []string

	modal.DB().Where("source =?", u.Source).First(&s)

	ps := lib.Decrypt(s.Password)

	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", s.Username, ps, s.IP, strconv.Itoa(int(s.Port)), u.Base))

	defer db.Close()

	if err != nil {
		c.Logger().Error(err.Error())
		return err
	}

	sql := "show tables"
	rows, err := db.Raw(sql).Rows()
	if err != nil {
		c.Logger().Error(err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&table)
		l = append(l, table)
	}

	return c.JSON(http.StatusOK, l)
}

func GeneralTableInfo(c echo.Context) (err error) {
	u := new(fetch)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusInternalServerError, "")
	}
	var s modal.CoreDataSource

	modal.DB().Where("source =?", u.Source).First(&s)

	ps := lib.Decrypt(s.Password)
	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", s.Username, ps, s.IP, strconv.Itoa(int(s.Port)), u.Base))
	if err != nil {
		c.Logger().Error(err.Error())
		return err
	}

	defer db.Close()

	var rows []FieldInfo
	var idx []IndexInfo

	if err := db.Raw(fmt.Sprintf("SHOW FULL FIELDS FROM `%s`.`%s`", u.Base, u.Table)).Scan(&rows).Error; err != nil {
		c.Logger().Error(err.Error())
	}

	if err := db.Raw(fmt.Sprintf("SHOW INDEX FROM `%s`.`%s`", u.Base, u.Table)).Scan(&idx).Error; err != nil {
		c.Logger().Error(err.Error())
	}
	return c.JSON(http.StatusOK, cdx{I: idx, F: rows})
}

func GeneralSQLTest(c echo.Context) (err error) {
	u := new(ddl)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusInternalServerError, "")
	}
	var s modal.CoreDataSource
	modal.DB().Where("source =?", u.Source).First(&s)
	ps := lib.Decrypt(s.Password)

	// todo 请配合juno 自行实现测试相关功能

	return c.JSON(http.StatusOK, record)
}

func GeneralOrderDetailList(c echo.Context) (err error) {
	workId := c.QueryParam("workid")
	var record []modal.CoreSqlRecord
	var count int
	start, end := lib.Paging(c.QueryParam("page"), 20)
	modal.DB().Model(&modal.CoreSqlRecord{}).Where("work_id =?", workId).Offset(start).Limit(end).Find(&record)
	modal.DB().Model(&modal.CoreSqlRecord{}).Where("work_id =?", workId).Count(&count)
	return c.JSON(http.StatusOK, struct {
		Record []modal.CoreSqlRecord `json:"record"`
		Count  int                   `json:"count"`
	}{
		Record: record,
		Count:  count,
	})
}

func GeneralOrderDetailRollSQL(c echo.Context) (err error) {
	workId := c.QueryParam("workid")
	var order modal.CoreSqlOrder
	var roll []modal.CoreRollback
	modal.DB().Where("work_id =?", workId).First(&order)
	modal.DB().Select("`sql`").Where("work_id =?", workId).Find(&roll)
	return c.JSON(http.StatusOK, struct {
		Order modal.CoreSqlOrder   `json:"order"`
		Sql   []modal.CoreRollback `json:"sql"`
	}{
		Order: order,
		Sql:   roll,
	})
}

func GeneralFetchMyOrder(c echo.Context) (err error) {
	u := new(f)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return
	}
	user, _ := lib.JwtParse(c)

	var pg int

	var order []modal.CoreSqlOrder

	queryField := "work_id, username, text, backup, date, real_name, executor, status, `data_base`, `table`,assigned,rejected"
	whereField := "username = ? AND text LIKE ? "
	dateField := " AND date >= ? AND date <= ?"

	start, end := lib.Paging(u.Page, 20)

	if u.Find.Valve {
		if u.Find.Picker[0] == "" {
			modal.DB().Select(queryField).Where(whereField, user, "%"+fmt.Sprintf("%s", u.Find.Text)+"%").Order("id desc").Offset(start).Limit(end).Find(&order)
			modal.DB().Model(&modal.CoreSqlOrder{}).Where(whereField, user, "%"+fmt.Sprintf("%s", u.Find.Text)+"%").Count(&pg)
		} else {
			modal.DB().Select(queryField).
				Where(whereField+dateField, user, "%"+fmt.Sprintf("%s", u.Find.Text)+"%", u.Find.Picker[0], u.Find.Picker[1]).Order("id desc").Offset(start).Limit(end).Find(&order)
			modal.DB().Model(&modal.CoreSqlOrder{}).Where(whereField+dateField, user, "%"+fmt.Sprintf("%s", u.Find.Text)+"%", u.Find.Picker[0], u.Find.Picker[1]).Count(&pg)
		}
	} else {
		modal.DB().Select(queryField).Where("username = ?", user).Order("id desc").Offset(start).Limit(end).Find(&order)
		modal.DB().Model(&modal.CoreSqlOrder{}).Where("username = ?", user).Count(&pg)
	}
	return c.JSON(http.StatusOK, struct {
		Data  []modal.CoreSqlOrder `json:"data"`
		Page  int                  `json:"page"`
		Multi bool                 `json:"multi"`
	}{
		order,
		pg,
		modal.GloOther.Multi,
	})
}

func GeneralFetchUndo(c echo.Context) (err error) {
	u := c.QueryParam("work_id")
	user, _ := lib.JwtParse(c)
	var undo modal.CoreSqlOrder
	if modal.DB().Where("username =? AND work_id =? AND `status` =? ", user, u, 2).First(&undo).RecordNotFound() {
		return c.JSON(http.StatusOK, "工单状态已更改！无法撤销")
	}
	modal.DB().Where("username =? AND work_id =? AND `status` =? ", user, u, 2).Delete(&modal.CoreSqlOrder{})
	return c.JSON(http.StatusOK, "工单已撤销！")
}

func GeneralQueryBeauty(c echo.Context) (err error) {
	req := new(modal.Queryresults)

	if err = c.Bind(req); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusOK, err.Error())
	}
	beauty := soar.PrettyFormat(req.Sql)
	return c.JSON(http.StatusOK, beauty)
}

func GeneralMergeDDL(c echo.Context) (err error) {
	req := new(modal.Queryresults)
	if err = c.Bind(req); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusOK, err.Error())
	}
	m, err := soar.MergeAlterTables(req.Sql)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{"err": err.Error(), "e": true})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"sols": m, "e": false})
}

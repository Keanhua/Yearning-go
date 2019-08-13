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
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"github.com/pingcap/parser"
	"log"
	"net/http"
	"time"
)

type fetchorder struct {
	Picker []string
	User   string
	Valve  bool
	Text   string
}

type f struct {
	Page int
	Find fetchorder
}

type reject struct {
	Text string
	Work string
}

type executeStr struct {
	WorkId  string
	Perform string
	Page    int
}

type referorder struct {
	Data modal.CoreSqlOrder
}

func FetchAuditOrder(c echo.Context) (err error) {
	u := new(f)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return
	}

	user, rule := lib.JwtParse(c)

	var pg int

	var order []modal.CoreSqlOrder

	queryField := "work_id, username, text, backup, date, real_name, executor, `status`, `type`, `delay`"

	whereField := "%s = ? AND username LIKE ?"

	dateField := " AND date >= ? AND date <= ?"

	start, end := lib.Paging(u.Page, 20)

	if rule == "perform" {
		if u.Find.Valve {
			whereField = fmt.Sprintf(whereField, "executor")
			if u.Find.Picker[0] == "" {
				modal.DB().Select(queryField).Where(whereField, user, "%"+fmt.Sprintf("%s", u.Find.User)+"%").Order("id desc").Offset(start).Limit(end).Find(&order)
				modal.DB().Model(&modal.CoreSqlOrder{}).Where(whereField, user, "%"+fmt.Sprintf("%s", u.Find.User)+"%").Count(&pg)
			} else {
				modal.DB().Select(queryField).Where(whereField+dateField, user, "%"+fmt.Sprintf("%s", u.Find.User)+"%", u.Find.Picker[0], u.Find.Picker[1]).Order("id desc").Offset(start).Limit(end).Find(&order)
				modal.DB().Model(&modal.CoreSqlOrder{}).Where(whereField+dateField, user, "%"+fmt.Sprintf("%s", u.Find.User)+"%", u.Find.Picker[0], u.Find.Picker[1]).Count(&pg)
			}
		} else {
			modal.DB().Select(queryField).Where("executor = ?", user).Order("id desc").Offset(start).Limit(end).Find(&order)
			modal.DB().Model(&modal.CoreSqlOrder{}).Where("executor = ?", user).Count(&pg)
		}
	} else {
		if u.Find.Valve {
			whereField = fmt.Sprintf(whereField, "assigned")
			if u.Find.Picker[0] == "" {
				modal.DB().Select(queryField).
					Where(whereField, user, "%"+fmt.Sprintf("%s", u.Find.User)+"%").Order("id desc").Offset(start).Limit(end).Find(&order)
				modal.DB().Model(&modal.CoreSqlOrder{}).Where(whereField, user, "%"+fmt.Sprintf("%s", u.Find.User)+"%").Count(&pg)
			} else {
				modal.DB().Select(queryField).
					Where(whereField+dateField, user, "%"+fmt.Sprintf("%s", u.Find.User)+"%", u.Find.Picker[0], u.Find.Picker[1]).Order("id desc").Offset(start).Limit(end).Find(&order)
				modal.DB().Model(&modal.CoreSqlOrder{}).Where(whereField+dateField, user, "%"+fmt.Sprintf("%s", u.Find.User)+"%", u.Find.Picker[0], u.Find.Picker[1]).Count(&pg)
			}
		} else {
			modal.DB().Select(queryField).Where("assigned = ?", user).Order("id desc").Offset(start).Limit(end).Find(&order)
			modal.DB().Model(&modal.CoreSqlOrder{}).Where("assigned = ?", user).Count(&pg)
		}
	}

	var ex []modal.CoreAccount

	modal.DB().Where("rule ='perform'").Find(&ex)

	call := struct {
		Multi    bool                 `json:"multi"`
		Data     []modal.CoreSqlOrder `json:"data"`
		Page     int                  `json:"page"`
		Executor []modal.CoreAccount  `json:"multi_list"`
	}{
		modal.GloOther.Multi,
		order,
		pg,
		ex,
	}
	return c.JSON(http.StatusOK, call)
}

func FetchOrderSQL(c echo.Context) (err error) {
	u := c.QueryParam("k")
	var sql modal.CoreSqlOrder
	var s []map[string]string
	modal.DB().Select(" `sql`, `delay`, `id_c`, `source`,`data_base`,`table`, `text`, `type`").Where("work_id =?", u).First(&sql)
	sqlParser := parser.New()
	stmtNodes, _, err := sqlParser.Parse(sql.SQL, "", "")
	if err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusOK, struct {
			Delay  string              `json:"delay"`
			SQL    []map[string]string `json:"sql"`
			IDC    string              `json:"idc"`
			Source string              `json:"source"`
			Base   string              `json:"base"`
			Table  string              `json:"table"`
			Text   string              `json:"text"`
			Type   uint                `json:"type"`
		}{
			sql.Delay,
			[]map[string]string{{"SQL": sql.SQL}},
			sql.IDC,
			sql.Source,
			sql.DataBase,
			sql.Table,
			sql.Text,
			sql.Type,
		})
	}
	for _, i := range stmtNodes {
		s = append(s, map[string]string{"SQL": i.Text()})
	}
	return c.JSON(http.StatusOK, struct {
		Delay  string              `json:"delay"`
		SQL    []map[string]string `json:"sql"`
		IDC    string              `json:"idc"`
		Source string              `json:"source"`
		Base   string              `json:"base"`
		Table  string              `json:"table"`
		Text   string              `json:"text"`
		Type   uint                `json:"type"`
	}{
		sql.Delay,
		s,
		sql.IDC,
		sql.Source,
		sql.DataBase,
		sql.Table,
		sql.Text,
		sql.Type,
	})
}

func RejectOrder(c echo.Context) (err error) {
	u := new(reject)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return
	}

	modal.DB().Model(&modal.CoreSqlOrder{}).Where("work_id =?", u.Work).Updates(map[string]interface{}{"rejected": u.Text, "status": 0})
	lib.MessagePush(c, u.Work, 0, u.Text)
	return c.JSON(http.StatusOK, "工单已驳回！")
}

func ExecuteOrder(c echo.Context) (err error) {
	u := new(executeStr)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return
	}
	var order modal.CoreSqlOrder
	var sor modal.CoreDataSource
	var backup bool
	modal.DB().Where("work_id =?", u.WorkId).First(&order)
	modal.DB().Where("source =?", order.Source).First(&sor)

	if order.Backup == 1 {
		backup = true
	}

	ps := lib.Decrypt(sor.Password)

	// todo  请配合juno实现相关审核执行逻辑

	return c.JSON(http.StatusOK, "工单已执行！")
}

func RollBackSQLOrder(c echo.Context) (err error) {
	u := new(referorder)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusInternalServerError, "")
	}
	w := lib.GenWorkid()
	u.Data.WorkId = w
	u.Data.Status = 2
	u.Data.Date = time.Now().Format("2006-01-02 15:04")
	modal.DB().Create(&u.Data)
	lib.MessagePush(c, w, 2, "")
	return c.JSON(http.StatusOK, "工单已提交,请等待审核人审核！")
}

type ber struct {
	U []string
}

func UndoAuditOrder(c echo.Context) (err error) {
	u := new(ber)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return
	}
	tx := modal.DB().Begin()
	for _, i := range u.U {
		tx.Where("work_id =?", i).Delete(&modal.CoreSqlOrder{})
		tx.Where("work_id =?", i).Delete(&modal.CoreRollback{})
		tx.Where("work_id =?", i).Delete(&modal.CoreSqlRecord{})
	}
	tx.Commit()
	return c.JSON(http.StatusOK, "工单已删除！")
}

func MulitAuditOrder(c echo.Context) (err error) {
	req := new(executeStr)
	if err = c.Bind(req); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusInternalServerError, "")
	}
	modal.DB().Model(&modal.CoreSqlOrder{}).Where("work_id =?", req.WorkId).Update(&modal.CoreSqlOrder{Executor: req.Perform, Status: 5})
	return c.JSON(200, "工单已提交执行人！")

}

func OscPercent(c echo.Context) (err error) {
	r := c.Param("work_id")
	var d modal.CoreSqlOrder
	modal.DB().Where("work_id =?", r).First(&d)
	return c.JSON(http.StatusOK, map[string]int{"p": d.Percent, "s": d.Current})
}

func OscKill(c echo.Context) (err error)  {
	r := c.Param("work_id")
	ser.OscIsKill[r] = true
	return c.JSON(http.StatusOK,"kill指令已发送!如工单最后显示为执行失败则生效!")
}

func DelayKill(c echo.Context) (err error)  {
	r := c.Param("work_id")
	modal.DB().Model(modal.CoreSqlOrder{}).Where("work_id =?",r).Update(&modal.CoreSqlOrder{IsKill:1})
	modal.DB().Model(&modal.CoreSqlOrder{}).Where("work_id =?", r).Updates(map[string]interface{}{"status": 4, "execute_time": time.Now().Format("2006-01-02 15:04")})
	return c.JSON(http.StatusOK,"kill指令已发送!将在到达执行时间时自动取消，状态已更改为执行失败！")
}
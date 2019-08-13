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
	"github.com/labstack/echo/v4"
	"net/http"
)

func FetchRecord(c echo.Context) (err error) {
	u := new(f)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return
	}
	start, end := lib.Paging(u.Page,20)

	var pg int

	var order []modal.CoreSqlOrder

	queryField := "work_id, username, text, execute_time, real_name, executor, `data_base`, `table`,assigned,id_c,source, `status`"
	whereField := "`status` in (?) AND text LIKE ? "
	dateField := " AND date >= ? AND date <= ?"

	if u.Find.Valve {
		if u.Find.Picker[0] == "" {
			modal.DB().Select(queryField).Where(whereField, []int{1, 4}, "%"+fmt.Sprintf("%s", u.Find.Text)+"%").Order("id desc").Offset(start).Limit(end).Find(&order)
			modal.DB().Model(&modal.CoreSqlOrder{}).Where(whereField, []int{1, 4}, "%"+fmt.Sprintf("%s", u.Find.Text)+"%").Count(&pg)
		} else {
			modal.DB().Select(queryField).
				Where(whereField+dateField, []int{1, 4}, "%"+fmt.Sprintf("%s", u.Find.Text)+"%", u.Find.Picker[0], u.Find.Picker[1]).Order("id desc").Offset(start).Limit(end).Find(&order)
			modal.DB().Model(&modal.CoreSqlOrder{}).Where(whereField+dateField, []int{1, 4}, "%"+fmt.Sprintf("%s", u.Find.Text)+"%", u.Find.Picker[0], u.Find.Picker[1]).Count(&pg)
		}
	} else {
		modal.DB().Select(queryField).Where("`status` in (?)", []int{1, 4}).Order("id desc").Offset(start).Limit(end).Find(&order)
		modal.DB().Model(&modal.CoreSqlOrder{}).Where("`status` in (?)", []int{1, 4}).Count(&pg)
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

func FetchQueryRecord(c echo.Context) (err error) {
	u := new(f)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return
	}

	start, end := lib.Paging(u.Page,20)

	var pg int

	var order []modal.CoreQueryOrder

	whereField := "`query_per` in (?) AND work_id LIKE ? "
	dateField := " AND date >= ? AND date <= ?"

	if u.Find.Valve {
		if u.Find.Picker[0] == "" {
			modal.DB().Where(whereField, []int{1, 3}, "%"+fmt.Sprintf("%s", u.Find.Text)+"%").Order("id desc").Offset(start).Limit(end).Find(&order)
			modal.DB().Model(&modal.CoreQueryOrder{}).Where(whereField, []int{1, 3}, "%"+fmt.Sprintf("%s", u.Find.Text)+"%").Count(&pg)
		} else {
			modal.DB().Where(whereField+dateField, []int{1, 3}, "%"+fmt.Sprintf("%s", u.Find.Text)+"%", u.Find.Picker[0], u.Find.Picker[1]).Order("id desc").Offset(start).Limit(end).Find(&order)
			modal.DB().Model(&modal.CoreQueryOrder{}).Where(whereField+dateField, []int{1, 3}, "%"+fmt.Sprintf("%s", u.Find.Text)+"%", u.Find.Picker[0], u.Find.Picker[1]).Count(&pg)
		}
	} else {
		modal.DB().Where("`query_per` in (?)", []int{1, 3}).Order("id desc").Offset(start).Limit(end).Find(&order)
		modal.DB().Model(&modal.CoreQueryOrder{}).Where("`query_per` in (?)", []int{1, 3}).Count(&pg)
	}
	return c.JSON(http.StatusOK, struct {
		Data []modal.CoreQueryOrder `json:"data"`
		Page int                    `json:"page"`
	}{
		order,
		pg,
	})
}

func FetchQueryRecordDetail(c echo.Context) (err error) {
	u := new(executeStr)

	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return
	}
	start, end := lib.Paging(u.Page,20)
	var detail []modal.CoreQueryRecord
	var count int
	modal.DB().Where("work_id =?", u.WorkId).Offset(start).Limit(end).Find(&detail)
	modal.DB().Model(&modal.CoreQueryRecord{}).Where("work_id =?", u.WorkId).Count(&count)
	return c.JSON(http.StatusOK, detail)
}

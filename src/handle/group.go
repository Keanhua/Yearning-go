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
	"github.com/labstack/echo/v4"
	"net/http"
)

type s struct {
	Data   []modal.CoreGrained    `json:"data"`
	Page   int                    `json:"page"`
	Source []modal.CoreDataSource `json:"source"`
	Audit  []modal.CoreAccount    `json:"audit"`
	Query  []modal.CoreDataSource `json:"query"`
}

type k struct {
	Username   string
	Permission modal.PermissionList
}

type m struct {
	Username   []string
	Permission modal.PermissionList
}

func SuperGroup(c echo.Context) (err error) {
	user, _ := lib.JwtParse(c)
	if user == "admin" {
		var f fetchdb
		var pg int
		var g []modal.CoreGrained
		var source []modal.CoreDataSource
		var query []modal.CoreDataSource
		var u []modal.CoreAccount
		con := c.QueryParam("con")
		start, end := lib.Paging(c.QueryParam("page"), 10)

		if err := json.Unmarshal([]byte(con), &f); err != nil {
			c.Logger().Error(err.Error())
		}
		if f.Valve {
			modal.DB().Where("username LIKE ?","%"+fmt.Sprintf("%s", f.Username)+"%").Offset(start).Limit(end).Find(&g)
			modal.DB().Where("username LIKE ?","%"+fmt.Sprintf("%s", f.Username)+"%").Model(modal.CoreGrained{}).Count(&pg)
		} else {
			modal.DB().Offset(start).Limit(end).Find(&g)
			modal.DB().Model(modal.CoreGrained{}).Count(&pg)
		}

		modal.DB().Select("source").Where("is_query =?", 0).Find(&source)
		modal.DB().Select("source").Where("is_query =?", 1).Find(&query)
		modal.DB().Select("username").Where("rule =?", "admin").Find(&u)
		return c.JSON(http.StatusOK, s{Data: g, Page: pg, Source: source, Audit: u, Query: query})
	}
	return c.JSON(http.StatusForbidden, "非法越权操作！")
}

func SuperGroupUpdate(c echo.Context) (err error) {
	user, _ := lib.JwtParse(c)
	if user == "admin" {
		u := new(k)
		if err = c.Bind(u); err != nil {
			c.Logger().Error(err.Error())
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		g, err := json.Marshal(u.Permission)
		if err != nil {
			c.Logger().Error(err.Error())
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		modal.DB().Model(modal.CoreGrained{}).Where("username = ?", u.Username).Update(modal.CoreGrained{Permissions: g})
		return c.JSON(http.StatusOK, fmt.Sprintf("用户:%s 权限已更新！", u.Username))
	}
	return c.JSON(http.StatusForbidden, "非法越权操作！")
}

func SuperMGroupUpdate(c echo.Context) (err error) {
	user, _ := lib.JwtParse(c)

	if user == "admin" {

		u := new(m)

		if err = c.Bind(u); err != nil {
			c.Logger().Error(err.Error())
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		g, err := json.Marshal(u.Permission)
		if err != nil {
			c.Logger().Error(err.Error())
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		for _, i := range u.Username {
			modal.DB().Model(modal.CoreGrained{}).Where("username = ?", i).Update(modal.CoreGrained{Permissions: g})
		}

		return c.JSON(http.StatusOK, "用户权限已更新！")
	}
	return c.JSON(http.StatusForbidden, "非法越权操作！")
}

func SuperDeleteGroup(c echo.Context) (err error) {
	g := c.Param("clear")
	u, err := json.Marshal(modal.InitPer)
	modal.DB().Model(modal.CoreGrained{}).Where("username =?", g).Update(modal.CoreGrained{Permissions: u})
	return c.JSON(http.StatusOK, fmt.Sprintf("用户:%s 权限已清空！", g))
}

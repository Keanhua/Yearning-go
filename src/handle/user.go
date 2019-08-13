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

type userInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type register struct {
	UserInfo map[string]string
}

type changePassword struct {
	Username string `json:"username"`
	New      string `json:"new"`
}

type changeMail struct {
	Username string
	Real     string
	Mail     string
}

type modifyUser struct {
	Username   string
	Department string
	RealName   string
	Rule       string
	Email      string
}

type fetchuser struct {
	User       string
	Department string
	Valve      bool
}
type ur struct {
	Page  int                 `json:"page"`
	Data  []modal.CoreAccount `json:"data"`
	Multi bool                `json:"multi"`
}

func UserLdapLogin(c echo.Context) (err error) {
	var account modal.CoreAccount
	u := new(userInfo)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, "")
	}

	if lib.LdapConnenct(c, &modal.GloLdap, u.Username, u.Password, false) {
		if modal.DB().Where("username = ?", u.Username).First(&account).RecordNotFound() {
			g, _ := json.Marshal(modal.InitPer)
			modal.DB().Create(&modal.CoreAccount{
				Username:   u.Username,
				RealName:   "请重置你的真实姓名",
				Password:   lib.DjangoEncrypt(lib.GenWorkid(), string(lib.GetRandom())),
				Rule:       "guest",
				Department: "",
				Email:      "",
			})
			modal.DB().Create(&modal.CoreGrained{Username: u.Username, Permissions: g, Rule: "guest"})
		}
		token, tokenErr := lib.JwtAuth(u.Username, account.Rule)
		if tokenErr != nil {
			c.Logger().Error(tokenErr.Error())
			return
		}
		dataStore := map[string]string{
			"token":       token,
			"permissions": account.Rule,
			"real_name":   account.RealName,
		}
		return c.JSON(http.StatusOK, dataStore)
	}
	return c.JSON(http.StatusUnauthorized, "")
}

func UserGeneralLogin(c echo.Context) (err error) {
	u := new(userInfo)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, "")
	}
	var account modal.CoreAccount

	if !modal.DB().Where("username = ?", u.Username).First(&account).RecordNotFound() {
		if e := lib.DjangoCheckPassword(&account, u.Password); e {
			token, tokenErr := lib.JwtAuth(u.Username, account.Rule)
			if tokenErr != nil {
				c.Logger().Error(tokenErr)
				return
			}
			dataStore := map[string]string{
				"token":       token,
				"permissions": account.Rule,
				"real_name":   account.RealName,
			}
			return c.JSON(http.StatusOK, dataStore)
		}

	}
	return c.JSON(http.StatusUnauthorized, "")

}

func UserReqSwitch(c echo.Context) (err error) {
	if modal.GloOther.Register {
		return c.JSON(http.StatusOK, 1)
	}
	return c.JSON(http.StatusOK, 0)
}

func UserRegister(c echo.Context) (err error) {

	if modal.GloOther.Register {
		u := new(register)
		if err = c.Bind(u); err != nil {
			c.Logger().Error(err)
			return c.JSON(http.StatusInternalServerError, "")
		}
		var unique modal.CoreAccount
		g, _ := json.Marshal(modal.InitPer)
		modal.DB().Where("username = ?", u.UserInfo["username"]).Select("username").First(&unique)
		if unique.Username != "" {
			return c.JSON(http.StatusOK, "用户已存在请重新注册！")
		}
		modal.DB().Create(&modal.CoreAccount{
			Username:   u.UserInfo["username"],
			RealName:   u.UserInfo["realname"],
			Password:   lib.DjangoEncrypt(u.UserInfo["password"], string(lib.GetRandom())),
			Rule:       "guest",
			Department: u.UserInfo["department"],
			Email:      u.UserInfo["email"],
		})
		modal.DB().Create(&modal.CoreGrained{Username: u.UserInfo["username"], Permissions: g, Rule: "guest"})
		return c.JSON(http.StatusOK, "注册成功！")
	}
	return c.JSON(http.StatusForbidden, "没有开启注册通道！")

}

func SuperUserRegister(c echo.Context) (err error) {

	u := new(register)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, "")
	}
	var unique modal.CoreAccount
	g, _ := json.Marshal(modal.InitPer)
	modal.DB().Where("username = ?", u.UserInfo["username"]).Select("username").First(&unique)
	if unique.Username != "" {
		return c.JSON(http.StatusOK, "用户已存在请重新注册！")
	}
	modal.DB().Create(&modal.CoreAccount{
		Username:   u.UserInfo["username"],
		RealName:   u.UserInfo["realname"],
		Password:   lib.DjangoEncrypt(u.UserInfo["password"], string(lib.GetRandom())),
		Rule:       u.UserInfo["group"],
		Department: u.UserInfo["department"],
		Email:      u.UserInfo["email"],
	})
	modal.DB().Create(&modal.CoreGrained{Username: u.UserInfo["username"], Permissions: g, Rule: u.UserInfo["group"]})
	return c.JSON(http.StatusOK, "注册成功！")
}

func ChangePassword(c echo.Context) (err error) {
	u := new(changePassword)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusInternalServerError, "")
	}
	user, _ := lib.JwtParse(c)
	modal.DB().Model(&modal.CoreAccount{}).Where("username = ?", user).Update("password", lib.DjangoEncrypt(u.New, string(lib.GetRandom())))
	return c.JSON(http.StatusOK, "密码修改成功！")
}

func ChangeMail(c echo.Context) (err error) {
	u := new(changeMail)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, "")
	}
	user, _ := lib.JwtParse(c)
	modal.DB().Model(&modal.CoreAccount{}).Where("username = ?", user).Updates(modal.CoreAccount{Email: u.Mail, RealName: u.Real})
	return c.JSON(http.StatusOK, "邮箱/真实姓名修改成功！刷新后显示最新信息!")
}

func SuperModifyUser(c echo.Context) (err error) {
	u := new(modifyUser)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusInternalServerError, "")
	}
	modal.DB().Model(modal.CoreAccount{}).Where("username = ?", u.Username).Updates(modal.CoreAccount{Email: u.Email, RealName: u.RealName, Rule: u.Rule, Department: u.Department})
	modal.DB().Model(modal.CoreGrained{}).Where("username =?", u.Username).Update(modal.CoreGrained{Rule: u.Rule})
	modal.DB().Model(modal.CoreSqlOrder{}).Where("username =?", u.Username).Update(modal.CoreSqlOrder{RealName: u.RealName})
	modal.DB().Model(modal.CoreQueryOrder{}).Where("username =?", u.Username).Update(modal.CoreQueryOrder{Realname: u.RealName})
	return c.JSON(http.StatusOK, "邮箱/真实姓名修改成功！刷新后显示最新信息!")
}

func SuperChangePassword(c echo.Context) (err error) {

	u := new(changePassword)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusInternalServerError, "")
	}
	modal.DB().Model(&modal.CoreAccount{}).Where("username = ?", u.Username).Update("password", lib.DjangoEncrypt(u.New, string(lib.GetRandom())))
	return c.JSON(http.StatusOK, "密码修改成功！")
}

func SuperFetchUser(c echo.Context) (err error) {
	var f fetchuser
	var u []modal.CoreAccount
	var pg int
	con := c.QueryParam("con")
	if err := json.Unmarshal([]byte(con), &f); err != nil {
		c.Logger().Error(err.Error())
	}
	start,end := lib.Paging(c.QueryParam("page"),10)

	if f.Valve {
		modal.DB().Where("username LIKE ? and department LIKE ?", "%"+fmt.Sprintf("%s", f.User)+"%", "%"+fmt.Sprintf("%s", f.Department)+"%").Count(&pg)
		modal.DB().Where("username LIKE ? and department LIKE ?", "%"+fmt.Sprintf("%s", f.User)+"%", "%"+fmt.Sprintf("%s", f.Department)+"%").Offset(start).Limit(end).Find(&u)
	} else {
		modal.DB().Offset(start).Limit(end).Find(&u)
		modal.DB().Model(modal.CoreAccount{}).Count(&pg)
	}

	return c.JSON(http.StatusOK, ur{Page: pg, Data: u, Multi: modal.GloOther.Multi})
}

func SuperDeleteUser(c echo.Context) (err error) {
	user := c.Param("user")

	var g []modal.CoreGrained

	modal.DB().Find(&g)

	tx := modal.DB().Begin()

	if er := tx.Where("username =?", user).Delete(&modal.CoreAccount{}).Error; er != nil {
		tx.Rollback()
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusInternalServerError, nil)
	}

	if er := tx.Where("username =?", user).Delete(&modal.CoreGrained{}).Error; er != nil {
		tx.Rollback()
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusInternalServerError, nil)
	}

	for _, i := range g {
		var p modal.PermissionList
		if err := json.Unmarshal(i.Permissions, &p); err != nil {
			c.Logger().Error(err.Error())
		}
		p.Auditor = lib.ResearchDel(p.Auditor, user)

		r, _ := json.Marshal(p)
		if err := tx.Model(&modal.CoreGrained{}).Where("id =?", i.ID).Update(modal.CoreGrained{Permissions: r}).Error; err != nil {
			tx.Rollback()
			c.Logger().Error(err.Error)
			return c.JSON(http.StatusInternalServerError, nil)
		}
	}
	tx.Commit()
	return c.JSON(http.StatusOK, fmt.Sprintf("用户: %s 已删除", user))
}

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

package lib

import (
	"Yearning-go/src/modal"
	"crypto/tls"
	"fmt"
	"github.com/labstack/echo/v4"
	"gopkg.in/gomail.v2"
	"net/http"
	"strings"
)

type UserInfo struct {
	ToUser  string
	User    string
	Pawd    string
	Smtp    string
	PubName string
}

var TemoplateTestMail = `
<html>
<body>
	<div style="text-align:center;">
		<h1>Yearning 2.0</h1>
		<h2>此邮件是测试邮件！</h2>
	</div>
</body>
</html>
`

var TmplRejectMail = `
<html>
<body>
<h1>Yearning 工单驳回通知</h1>
<br><p>工单号: %s</p>
<br><p>发起人: %s</p>
<br><p>地址: <a href="%s">%s</a></p>
<br><p>状态: 驳回</p>
<br><p>驳回说明: %s</p>
</body>
</html>
`

var TmplMail = `
<html>
<body>
<h1>Yearning 工单%s通知</h1>
<br><p>工单号: %s</p>
<br><p>发起人: %s</p>
<br><p>地址: <a href="%s">%s</a></p>
<br><p>状态: %s</p>
</body>
</html>
`

var TmplTestDing = `
# Yearning 测试！
`

var TmplReferDing = `# Yearning工单提交通知 #  \n <br>  \n  **工单编号:**  %s \n  \n **提交人员:**  <font color=\"#78beea\">%s</font><br /> \n \n **审核人员:** <font color=\"#fe8696\">%s</font><br /> \n \n**平台地址:** http://%s \n  \n **工单说明:**  %s \n \n **状态:** <font color=\"#1abefa\">已提交</font><br /> \n `

var TmplRejectDing = `# Yearning工单提交通知 #  \n <br>  \n  **工单编号:**  %s \n  \n **提交人员:**  <font color=\"#78beea\">%s</font><br /> \n \n **审核人员:** <font color=\"#fe8696\">%s</font><br /> \n \n**平台地址:** http://%s \n  \n **工单说明:**  %s \n \n **状态:** <font color=\"#df117e\">驳回</font> \n <br /> \n \n **驳回说明:**  %s `
var TmplSuccessDing = `# Yearning工单执行通知 #  \n <br>  \n  **工单编号:**  %s \n  \n **提交人员:**  <font color=\"#78beea\">%s</font><br /> \n \n **审核人员:** <font color=\"#fe8696\">%s</font><br /> \n \n**平台地址:** http://%s \n  \n **工单说明:**  %s \n \n **状态:** <font color=\"#3fd2bd\">执行成功</font>`
var TmplFailedDing = `# Yearning工单执行通知 #  \n <br>  \n  **工单编号:**  %s \n  \n **提交人员:**  <font color=\"#78beea\">%s</font><br /> \n \n **审核人员:** <font color=\"#fe8696\">%s</font><br /> \n \n**平台地址:** http://%s \n  \n **工单说明:**  %s \n \n **状态:** <font color=\"#ea2426\">执行失败</font>`
var TmplPerformDing = `# Yearning工单转交通知 #  \n <br>  \n  **工单编号:**  %s \n  \n **提交人员:**  <font color=\"#78beea\">%s</font><br /> \n \n **审核人员:** <font color=\"#fe8696\">%s</font><br /> \n \n**平台地址:** http://%s \n  \n **工单说明:**  %s \n \n **状态:** <font color=\"#de4943\">等待执行人执行</font>`
var TmplQueryRefer =  `# Yearning查询申请通知 #  \n <br>  \n  **工单编号:**  %s \n  \n **提交人员:**  <font color=\"#78beea\">%s</font><br /> \n \n **审核人员:** <font color=\"#fe8696\">%s</font><br /> \n \n**平台地址:** http://%s \n  \n **工单说明:**  %s \n \n **状态:** <font color=\"#1abefa\">已提交</font>`
var TmplSuccessQuery =  `# Yearning查询申请通知 #  \n <br>  \n  **工单编号:**  %s \n  \n **提交人员:**  <font color=\"#78beea\">%s</font><br /> \n \n **审核人员:** <font color=\"#fe8696\">%s</font><br /> \n \n**平台地址:** http://%s \n  \n **状态:** <font color=\"#3fd2bd\">同意</font>`
var TmplRejectQuery =  `# Yearning查询申请通知 #  \n <br>  \n  **工单编号:**  %s \n  \n **提交人员:**  <font color=\"#78beea\">%s</font><br /> \n \n **审核人员:** <font color=\"#fe8696\">%s</font><br /> \n \n**平台地址:** http://%s \n  \n **状态:** <font color=\"#df117e\">已驳回</font>`

var TmplGroupRefer =  `# Yearning权限申请通知 #  \n <br>  \n  **工单编号:**  %s \n  \n **提交人员:**  <font color=\"#78beea\">%s</font><br /> \n \n **审核人员:** <font color=\"#fe8696\">%s</font><br /> \n \n**平台地址:** http://%s \n  \n **状态:** <font color=\"#1abefa\">已提交</font>`
var TmplSuccessGroup =  `# Yearning权限申请通知 #  \n <br>  \n  **工单编号:**  %s \n  \n **提交人员:**  <font color=\"#78beea\">%s</font><br /> \n \n **审核人员:** <font color=\"#fe8696\">%s</font><br /> \n \n**平台地址:** http://%s \n  \n **状态:** <font color=\"#3fd2bd\">同意</font>`
var TmplRejectGroup =  `# Yearning权限申请通知 #  \n <br>  \n  **工单编号:**  %s \n  \n **提交人员:**  <font color=\"#78beea\">%s</font><br /> \n \n **审核人员:** <font color=\"#fe8696\">%s</font><br /> \n \n**平台地址:** http://%s \n  \n **状态:** <font color=\"#df117e\">已驳回</font>`

func SendMail(c echo.Context, mail modal.Message, tmpl string) {
	m := gomail.NewMessage()
	m.SetHeader("From", mail.User)
	m.SetHeader("To", mail.ToUser)
	m.SetHeader("Subject", "Yearning消息推送!")
	m.SetBody("text/html", tmpl)
	d := gomail.NewDialer(mail.Host, mail.Port, mail.User, mail.Password)
	if mail.Ssl {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		c.Logger().Error(err.Error())
		return
	}
}

func SendDingMsg(c echo.Context, msg modal.Message, sv string) {
	//请求地址模板

	//创建一个请求

	mx := fmt.Sprintf(`{"msgtype": "markdown", "markdown": {"title": "Yearning sql审计平台", "text": "%s"}}`, sv)


	req, err := http.NewRequest("POST", msg.WebHook, strings.NewReader(mx))
	if err != nil {
		c.Logger().Error(err.Error())
		return
	}


	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	//设置请求头
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	//发送请求
	resp, err := client.Do(req)

	if err != nil {
		c.Logger().Error(err.Error())
		return
	}

	//关闭请求
	defer resp.Body.Close()
}

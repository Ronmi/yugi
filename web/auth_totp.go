// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
)

func init() { addFactory(authTotpFormTestCase) }
func authTotpFormTestCase() []tmplTestCase {
	return []tmplTestCase{{"auth_totp.html", &authTotpFormPage{
		actions.FakeUser(),
	}}}
}

type authTotpFormPage struct {
	Me *actions.User
}

func (p *authTotpFormPage) Render(g *gin.Context) {
	g.HTML(200, "auth_totp.html", nil)
}

func (p *authTotpFormPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

type authTotoResultPage struct {
	Next string
}

func (p *authTotoResultPage) Render(g *gin.Context) {
	g.Redirect(302, p.Next)
}

func (p *authTotoResultPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(p))
}

func (srv) AuthTOTPForm(c Context) tmplspec {
	u := c.Session().CurrentUser()
	if !u.NZ() {
		return errParam("未登入").
			AddInfo(c).
			AddMessage("請先登入").
			Redir(config.LoginSelectPage, "登入")
	}

	return &authTotpFormPage{Me: u.Get()}
}

func (srv) AuthTOTP(c Context) tmplspec {
	type params struct {
		Code string `form:"code" binding:"required"`
	}
	var p params
	if err := c.Bind(&p); err != nil {
		return errParam("缺少驗證碼").
			AddInfo(c).
			AddMessage("請輸入驗證碼").
			Redir(config.TOTPPage, "回上一頁")
	}

	sess := c.Session()

	count := sess.OTPErrorCount()
	if count.Defaults(0) >= 3 {
		return errParam("驗證碼錯誤").
			AddInfo(c).
			AddMessage("請重試一次").
			Redir(config.TOTPPage, "回上一頁")
	}
	defer sess.Save()

	if !config.ValidateOTP(p.Code, sess.CurrentUser().Get().TOTPSecret) {
		count.Set(count.Defaults(0) + 1)
		return errParam("驗證碼錯誤").
			AddInfo(c).
			AddMessage("請重試一次").
			Redir(config.TOTPPage, "回上一頁")
	}

	count.Del()
	next := sess.
		NextMove().
		Defaults(config.URLPath(config.DashboardPage))
	sess.NextMove().Del()
	sess.Authed().Set(true)

	return &authTotoResultPage{Next: next}
}

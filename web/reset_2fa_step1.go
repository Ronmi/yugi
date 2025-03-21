// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"errors"
	"html/template"

	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func init() { addFactory(otpResetStep1TestCase) }
func otpResetStep1TestCase() []tmplTestCase {
	uri := "foo"
	qr, _ := createQR(uri)
	return []tmplTestCase{
		{Name: "reset_2fa_step1.html", Data: &otpResetStep1Page{
			Me:     actions.FakeUser(),
			URI:    uri,
			Secret: config.RandStr(8, config.AlNum),
			QRCode: template.URL(qr),
		}},
	}
}

type otpResetStep1Page struct {
	Me     *actions.User
	URI    string
	Secret string
	QRCode template.URL // datauri format
}

func (p *otpResetStep1Page) Render(c *gin.Context) {
	c.HTML(200, "reset_2fa_step1.html", p)
}

func (p *otpResetStep1Page) RenderAPI(c *gin.Context) {
	c.JSON(200, apiData(p))
}

func (srv) otpResetStep1(c Context) tmplspec {
	type params struct {
		Name string `form:"name" binding:"required"`
	}
	var p params
	if err := c.Bind(&p); err != nil {
		return errParam("缺少參數").
			AddInfo(c).
			AddMessage("請重新操作").
			Redir(config.MgrMemberListPage, "回上一頁")
	}

	me := c.Session().CurrentUser().Get()

	u, err := actions.FindUser(c.DB(), p.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errParam("使用者不存在").
				AddInfo(c).
				AddMessage("請檢查使用者名稱").
				Redir(config.MgrMemberListPage, "回上一頁")
		}
		return errInternal("內部錯誤").
			AddInfo(c).
			AddError(err).
			AddMessage("無法取得使用者資訊").
			AddMessage("請稍後再試").
			Redir(config.MgrMemberListPage, "回上一頁")
	}

	key, err := config.CreateOTPKey(u.Name)
	if err != nil {
		if err != nil {
			return errInternal("內部錯誤").
				AddInfo(c).
				AddError(err).
				AddMessage("無法建立 OTP key").
				AddMessage("請稍後再試").
				Redir(config.MgrMemberListPage, "回上一頁")
		}
	}

	err = actions.UpdateMemberOTP(c.DB(), me, u, key.Secret())
	if err != nil {
		return errInternal("內部錯誤").
			AddInfo(c).
			AddError(err).
			AddMessage("無法更新使用者 OTP key").
			AddMessage("請稍後再試").
			Redir(config.MgrMemberListPage, "回上一頁")
	}

	uri := key.URL()
	qr, err := createQR(uri)
	if err != nil {
		return errInternal("內部錯誤").
			AddInfo(c).
			AddMessage("無法建立 QR code").
			AddMessage("請稍後再試").
			Redir(config.DashboardPage, "回首頁")
	}

	return &otpResetStep1Page{
		Me:     me,
		URI:    uri,
		Secret: key.Secret(),
		QRCode: template.URL(qr),
	}
}

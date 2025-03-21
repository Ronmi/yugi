// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"bytes"
	"html/template"

	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
)

func init() { addFactory(otpStep1TestCase) }
func otpStep1TestCase() []tmplTestCase {
	uri := "foo"
	qr, _ := createQR(uri)
	return []tmplTestCase{
		{Name: "2fa_step1.html", Data: &otpStep1Page{
			Me:     actions.FakeUser(),
			URI:    uri,
			Secret: config.RandStr(8, config.AlNum),
			QRCode: template.URL(qr),
		}},
	}
}

type otpStep1Page struct {
	Me     *actions.User
	URI    string
	Secret string
	QRCode template.URL // datauri format
}

func (p *otpStep1Page) Render(c *gin.Context) {
	c.HTML(200, "2fa_step1.html", p)
}

func (p *otpStep1Page) RenderAPI(c *gin.Context) {
	c.JSON(200, apiData(p))
}

type buffer struct {
	bytes.Buffer
}

func (b *buffer) Close() error { return nil }

func (srv) otpStep1(c Context) tmplspec {
	sess := c.Session()
	defer sess.Save()
	u := sess.CurrentUser().Get()
	if u.TOTPSecret != "" {
		return errParam("已啟用 2FA").
			AddInfo(c).
			AddMessage("請勿重複操作").
			Redir(config.DashboardPage, "回首頁")
	}

	k := sess.OTPKey()
	if !k.NZ() {
		key, err := config.CreateOTPKey(u.Name)
		if err != nil {
			return errInternal("內部錯誤").
				AddInfo(c).
				AddMessage("無法建立 OTP key").
				AddMessage("請稍後再試").
				Redir(config.DashboardPage, "回首頁")
		}
		k.Set(OTPKey{key})
	}
	key := k.Get()

	uri := key.URL()
	qr, err := createQR(uri)
	if err != nil {
		return errInternal("內部錯誤").
			AddInfo(c).
			AddMessage("無法建立 QR code").
			AddMessage("請稍後再試").
			Redir(config.DashboardPage, "回首頁")
	}

	return &otpStep1Page{
		Me:     u,
		URI:    uri,
		Secret: key.Secret(),
		QRCode: template.URL(qr),
	}
}

type otpStep2Page struct{}

func (p *otpStep2Page) Render(c *gin.Context) {
	c.Redirect(302, config.URLPath(config.DashboardPage))
}

func (p *otpStep2Page) RenderAPI(c *gin.Context) {
	c.JSON(200, apiData("ok"))
}

func (srv) otpStep2(c Context) tmplspec {
	sess := c.Session()
	defer sess.Save()
	k := sess.OTPKey()
	if !k.NZ() {
		return errParam("缺少 OTP key").
			AddInfo(c).
			AddMessage("請重新操作").
			Redir(config.DashboardPage, "回首頁")
	}

	u := sess.CurrentUser().Get()

	type params struct {
		Code string `form:"code" binding:"required"`
	}
	var p params
	if err := c.Bind(&p); err != nil {
		return errParam("缺少驗證碼").
			AddInfo(c).
			AddMessage("請重新操作").
			Redir(config.Enable2FAStep1Page, "回上一頁")
	}

	if !config.ValidateOTP(p.Code, k.Get().Secret()) {
		return errParam("驗證碼錯誤").
			AddInfo(c).
			AddMessage("請檢查輸入的驗證碼").
			Redir(config.Enable2FAStep1Page, "回上一頁")
	}

	u.TOTPSecret = k.Get().Secret()
	err := actions.SaveTOTP(c.DB(), u, u.TOTPSecret)
	if err != nil {
		return errInternal("內部錯誤").
			AddInfo(c).
			AddMessage("無法儲存 TOTP 資訊").
			AddMessage("請稍後再試").
			Redir(config.Enable2FAStep1Page, "回上一頁")
	}
	k.Del()

	return &otpStep2Page{}
}

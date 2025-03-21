// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
)

type authPage struct {
	TOTP bool
	// must be validated in handler, empty if TOTP is true
	Next string
}

func (a *authPage) Render(c *gin.Context) {
	if a.TOTP {
		c.Redirect(302, config.URLPath(config.TOTPPage))
		return
	}

	if a.Next != "" {
		c.Redirect(302, a.Next)
	}

	c.Redirect(302, config.URLPath(config.DashboardPage))
}

func (a *authPage) RenderAPI(c *gin.Context) {
	c.JSON(200, apiData(a))
}

type tgAuthParam struct {
	ID        string `form:"id"`
	FirstName string `form:"first_name"`
	LastName  string `form:"last_name"`
	Username  string `form:"username"`
	PhotoURL  string `form:"photo_url"`
	AuthDate  string `form:"auth_date"`
	Hash      string `form:"hash"`
}

func (p *tgAuthParam) Validate() bool {
	arr := [][]byte{}
	arr = append(arr, []byte("auth_date="+p.AuthDate))
	arr = append(arr, []byte("first_name="+p.FirstName))
	arr = append(arr, []byte("id="+p.ID))
	arr = append(arr, []byte("last_name="+p.LastName))
	arr = append(arr, []byte("photo_url="+p.PhotoURL))
	arr = append(arr, []byte("username="+p.Username))

	secret := config.TelegramAuth.TokenHash

	hmac := hmac.New(sha256.New, secret)
	hmac.Write(bytes.Join(arr, []byte("\n")))
	expected := hex.EncodeToString(hmac.Sum(nil))

	return expected == p.Hash
}

func (srv) AuthTG(c Context) tmplspec {
	var p tgAuthParam
	if err := c.ShouldBind(&p); err != nil {
		return errParam("無效的參數").
			AddInfo(c).
			AddMessage("無法解析參數").
			Redir(config.LoginSelectPage, "重試登入")
	}

	if !p.Validate() {
		return errParam("無效的參數").
			AddInfo(c).
			AddMessage("驗證失敗").
			Redir(config.LoginSelectPage, "重試登入")
	}

	return postAuth(c, p.ID, actions.Telegram)
}

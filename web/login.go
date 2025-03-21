// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"crypto/sha1"
	"fmt"
	"math/rand/v2"

	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func init() { addFactory(loginTestCase) }
func loginTestCase() []tmplTestCase {
	return []tmplTestCase{
		{Name: "login.html", Data: &loginPage{}},
		{Name: "login.html", Data: &loginPage{
			Google:    "http://google.com",
			Facebook:  "http://facebook.com",
			Instagram: "http://instagram.com",
			Twitter:   "http://twitter.com",
			Line:      "http://line.com",
			Telegram: &config.TelegramAuthConfig{
				Bot:       "asd",
				Size:      "large",
				Radius:    10,
				Write:     true,
				TokenHash: []byte("asd"),
			},
		}},
	}
}

type loginPage struct {
	Google    string // Google oauth url
	Facebook  string
	Instagram string
	Twitter   string
	Line      string
	Telegram  *config.TelegramAuthConfig
}

func (l *loginPage) Render(c *gin.Context) {
	c.HTML(200, "login.html", l)
}

func (l *loginPage) RenderAPI(c *gin.Context) {
	c.JSON(200, apiData(l))
}

// generate random string
func genState() string {
	i64 := rand.Int64()

	b := make([]byte, 8)
	for i := 0; i < 8; i++ {
		b[i] = byte(i64 >> uint(i*8))
	}

	h := sha1.New()
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (s *srv) Login(c Context) tmplspec {
	sess := c.Session()
	state := genState()
	sess.OauthState().Set(state)
	nonce := config.RandStr(16, config.AlNum)
	sess.LoginNonce().Set(nonce)
	sess.Save()

	l := &loginPage{}
	if config.GoogleOauth != nil {
		l.Google = config.GoogleOauth.AuthCodeURL(state)
	}
	if config.FacebookOauth != nil {
		l.Facebook = config.FacebookOauth.AuthCodeURL(state)
	}
	if config.LineOauth != nil {
		l.Line = config.LineOauth.AuthCodeURL(
			state,
			oauth2.SetAuthURLParam("nonce", nonce),
		)
	}
	if config.TelegramAuth != nil {
		l.Telegram = config.TelegramAuth
	}

	return l
}

type logoutPage struct{}

func (logoutPage) Render(c *gin.Context) {
	c.Redirect(302, config.URLPath(config.LoginSelectPage))
}

func (logoutPage) RenderAPI(c *gin.Context) {
	c.JSON(200, apiData("ok"))
}

func (srv) Logout(c Context) tmplspec {
	sess := c.Session()
	sess.CurrentUser().Del()
	sess.Authed().Del()
	sess.Save()

	return &logoutPage{}
}

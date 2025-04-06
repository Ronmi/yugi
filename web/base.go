// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"context"
	"encoding/gob"
	"net/http"
	"os"
	"path"

	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-contrib/logger"
	"github.com/gin-contrib/requestid"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/raohwork/task"
	"github.com/raohwork/task/httptask"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func Wraps(g *gin.RouterGroup, as func(Handler) gin.HandlerFunc) routerSpec {
	return routerSpec{
		r:  g,
		as: as,
	}
}

type routerSpec struct {
	r  *gin.RouterGroup
	as func(Handler) gin.HandlerFunc
}

func (s routerSpec) With(middleware func(Handler) Handler) routerSpec {
	return routerSpec{
		r: s.r,
		as: func(h Handler) gin.HandlerFunc {
			return s.as(middleware(h))
		},
	}
}

func (s routerSpec) Use(m ...gin.HandlerFunc) routerSpec {
	s.r.Use(m...)
	return s
}

func (s routerSpec) Group(path string, m ...gin.HandlerFunc) routerSpec {
	return routerSpec{
		r:  s.r.Group(path, m...),
		as: s.as,
	}
}

func (s routerSpec) g(h []Handler) []gin.HandlerFunc {
	ret := make([]gin.HandlerFunc, len(h))
	for i, v := range h {
		ret[i] = s.as(v)
	}
	return ret
}

func (s routerSpec) Any(path string, h ...Handler) routerSpec {
	s.r.Any(path, s.g(h)...)
	return s
}

func (s routerSpec) GET(path string, h ...Handler) routerSpec {
	s.r.GET(path, s.g(h)...)
	return s
}

func (s routerSpec) POST(path string, h ...Handler) routerSpec {
	s.r.POST(path, s.g(h)...)
	return s
}

func (s routerSpec) PUT(path string, h ...Handler) routerSpec {
	s.r.PUT(path, s.g(h)...)
	return s
}

func (s routerSpec) DELETE(path string, h ...Handler) routerSpec {
	s.r.DELETE(path, s.g(h)...)
	return s
}

func (s routerSpec) PATCH(path string, h ...Handler) routerSpec {
	s.r.PATCH(path, s.g(h)...)
	return s
}

func (s routerSpec) OPTIONS(path string, h ...Handler) routerSpec {
	s.r.OPTIONS(path, s.g(h)...)
	return s
}

func (s routerSpec) HEAD(path string, h ...Handler) routerSpec {
	s.r.HEAD(path, s.g(h)...)
	return s
}

type srv struct {
}

func New(db *gorm.DB) (task.Task, error) {
	gob.Register(OTPKey{})

	g := gin.New()

	if err := LoadTmpl(); err != nil {
		return nil, err
	}

	if err := validateTmpl(); err != nil {
		return nil, err
	}

	g.SetHTMLTemplate(tmpl)

	g.Use(requestid.New())
	g.Use(func(c *gin.Context) {
		db := db.WithContext(context.WithValue(
			context.TODO(),
			"reqid",
			requestid.Get(c),
		))
		c.Set("db", db)
		c.Next()
	})
	g.Use(logger.SetLogger(logger.WithLogger(func(g *gin.Context, l zerolog.Logger) zerolog.Logger {
		return l.Output(config.DefaultLogWriter(os.Stderr)).
			With().
			Str("reqid", requestid.Get(g)).
			Logger()
	})))
	g.Use(sessions.Sessions("sesskey", config.SessionStore))
	srv := &srv{}

	web := Wraps(&g.RouterGroup, AsHTML)
	g.NoRoute(func(c *gin.Context) {
		err404(Context{c}).Render(c)
	})

	// 無需登入的頁面
	web.GET(config.OrgDetailPage, srv.OrgDetail)
	// 登入功能
	web.Group(config.BaseURLPath).
		GET(config.LoginSelectPage, srv.Login).
		GET(config.LogoutPage, srv.Logout).
		GET(config.GoogleCallbackPage, srv.AuthGoogle).
		GET(config.FacebookCallbackPage, srv.AuthFacebook).
		GET(config.TelegramCallbackPage, srv.AuthTG).
		GET(config.LineCallbackPage, srv.AuthLine)
	web.Group(config.BaseURLPath, MustHaveUser).
		GET(config.TOTPPage, srv.AuthTOTPForm).
		POST(config.TOTPPage, srv.AuthTOTP)
	// 登入即可用
	web.Group(config.BaseURLPath).
		With(MustAuth).
		GET(config.DashboardPage, srv.Dashboard).
		GET(config.Enable2FAStep1Page, srv.otpStep1).
		POST(config.Enable2FAStep2Page, srv.otpStep2).
		GET(config.ViewMyLogsPage, srv.ViewMyLogs)
	// 民眾專用
	web.Group(config.BaseURLPath).
		With(MustRole(actions.Member)).
		GET(config.UsrAppointmentListPage, srv.UsrAppointmentList).
		GET(config.UsrAppointmentDetailPage, srv.UsrAppointmentDetail).
		GET(config.UsrAppointmentDeletePage, srv.UsrAppointmentDelete).
		GET(config.UsrReceiptCreatePage, srv.UsrReceiptCreate).
		POST(config.UsrReceiptCreatePage, srv.UsrReceiptCreate).
		GET(config.UsrReceiptViewPage, srv.UsrReceiptView)

	// 民眾與志工共用 (建立預約)
	web.Group(config.BaseURLPath).
		With(MustRole(actions.Member, actions.Volunteer, actions.Manager)).
		GET(config.UsrScheduleListPage, srv.UsrScheduleList).
		POST(config.UsrAppointmentMakePage, srv.UsrAppointmentMake).
		GET(config.UsrAppointmentMakePage, srv.UsrAppointmentMake)

	// 幹部專用
	web.Group(config.BaseURLPath).
		With(MustRole(actions.Manager)).
		GET(config.ViewOrgLogsPage, srv.ViewOrgLogs).
		GET(config.MgrMemberListPage, srv.MgrMemberList).
		POST(config.MgrEditMemberNotePage, srv.MgrMemberNote).
		POST(config.MgrEditMemberSecretPage, srv.MgrMemberSecret).
		GET(config.MgrGrantRolePage, srv.MgrGrantRole).
		POST(config.MgrGrantRolePage, srv.MgrGrantRole).
		GET(config.MgrRevokeRolePage, srv.MgrRevokeRole).
		POST(config.MgrRevokeRolePage, srv.MgrRevokeRole).
		GET(config.MgrScheduleListPage, srv.MgrScheduleList).
		GET(config.MgrScheduleDisablePage, srv.MgrScheduleDisable).
		POST(config.MgrScheduleDisablePage, srv.MgrScheduleDisable).
		GET(config.MgrAppointmentListPage, srv.MgrAppointmentList).
		GET(config.MgrAppointmentStatusPage, srv.MgrAppointmentStatus).
		GET(config.MgrAppointmentDetailPage, srv.MgrAppointmentDetail).
		GET(config.MgrAppointmentPairSelectPage, srv.MgrAppointmentPairSelect).
		POST(config.MgrAppointmentPairSelectPage, srv.MgrAppointmentPairSelect).
		GET(config.MgrAppointmentPairPage, srv.MgrAppointmentPair).
		POST(config.MgrAppointmentPairPage, srv.MgrAppointmentPair).
		GET(config.Reset2FAStep1Page, srv.otpResetStep1).
		POST(config.Reset2FAStep1Page, srv.otpResetStep1)
	// 志工與幹部共用
	web.Group(config.BaseURLPath).
		With(MustRole(actions.Manager, actions.Volunteer, actions.Novice)).
		GET(config.VolNewScheduleFormPage, srv.VolScheduleForm).
		POST(config.VolScheduleNewPage, srv.VolScheduleNew).
		GET(config.VolScheduleListPage, srv.VolScheduleList).
		GET(config.VolScheduleDisablePage, srv.VolScheduleDisable).
		POST(config.VolScheduleDisablePage, srv.VolScheduleDisable).
		GET(config.VolAppointmentListPage, srv.VolAppointmentList).
		GET(config.VolAppointmentDetailPage, srv.VolAppointmentDetail).
		GET(config.VolAppointmentDeletePage, srv.VolAppointmentDelete).
		POST(config.VolAppointmentDeletePage, srv.VolAppointmentDelete).
		GET(config.VolAppointmentPubNotePage, srv.VolAppointmentPubNote).
		POST(config.VolAppointmentPubNotePage, srv.VolAppointmentPubNote).
		GET(config.VolAppointmentSecNotePage, srv.VolAppointmentSecNote).
		POST(config.VolAppointmentSecNotePage, srv.VolAppointmentSecNote).
		GET(config.VolAppointmentContactPage, srv.VolAppointmentContact).
		GET(config.VolAppointmentNotMatchPage, srv.VolAppointmentNotMatch).
		GET(config.VolAppointmentConfirmPage, srv.VolAppointmentConfirm).
		POST(config.VolAppointmentConfirmPage, srv.VolAppointmentConfirm).
		GET(config.VolAppointmentMissedPage, srv.VolAppointmentMissed).
		POST(config.VolAppointmentMissedPage, srv.VolAppointmentMissed).
		GET(config.VolReceiptFormPage, srv.VolReceiptForm).
		GET(config.VolReceiptCreatePage, srv.VolReceiptCreate).
		POST(config.VolReceiptCreatePage, srv.VolReceiptCreate).
		GET(config.VolReceiptSignPage, srv.VolReceiptSign).
		POST(config.VolReceiptSignPage, srv.VolReceiptSign)

	var h http.Handler = g
	if config.StaticDir != "" {
		h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := r.URL.Path
			if p[len(p)-1] == '/' {
				p += "index.html"
			}
			p = path.Join(config.StaticDir, p)
			if _, err := os.Stat(p); err == nil {
				http.ServeFile(w, r, p)
				return
			}

			g.ServeHTTP(w, r)
		})
	}
	httpsrv := &http.Server{
		Addr:    config.BindAddr,
		Handler: h,
	}

	if config.CertFile != "" && config.KeyFile != "" {
		return task.FromServer(
			func() error { return httpsrv.ListenAndServeTLS(config.CertFile, config.KeyFile) },
			func() { httpsrv.Shutdown(context.Background()) },
		), nil
	}

	return httptask.Server(httpsrv), nil
}

// only used by 2fa
func MustHaveUser(c *gin.Context) {
	sess := Context{c}.Session()
	if !sess.CurrentUser().NZ() {
		c.Redirect(302, config.URLPath(config.LoginSelectPage))
		c.Abort()
		return
	}

	c.Next()
}

func MustAuth(h Handler) Handler {
	return func(c Context) tmplspec {
		sess := c.Session()
		if !sess.CurrentUser().NZ() || !sess.Authed().NZ() {
			loc := c.Request.URL.Path
			if c.Request.URL.RawQuery != "" {
				loc += "?" + c.Request.URL.RawQuery
			}
			sess.NextMove().Set(loc)
			sess.Save()
			return errRedirPage{
				Next:    config.URLPath(config.LoginSelectPage),
				Code:    403,
				Message: "尚未登入或登入已過期",
			}
		}

		return h(c)
	}
}

func MustRole(role ...actions.Role) func(Handler) Handler {
	return func(h Handler) Handler {
		return func(c Context) tmplspec {
			sess := c.Session()
			u := sess.CurrentUser()
			if !u.NZ() {
				loc := c.Request.URL.Path
				if c.Request.URL.RawQuery != "" {
					loc += "?" + c.Request.URL.RawQuery
				}
				sess.NextMove().Set(loc)
				sess.Save()
				return errRedirPage{
					Next:    config.URLPath(config.LoginSelectPage),
					Code:    403,
					Message: "尚未登入或登入已過期",
				}
			}
			for _, r := range role {
				if u.Get().Role == r {
					return h(c)
				}
			}

			return errRedirPage{
				Next:    config.URLPath(config.DashboardPage),
				Code:    403,
				Message: "權限不足",
			}
		}
	}
}

type Handler func(Context) tmplspec

func AsHTML(h Handler) gin.HandlerFunc {
	return func(g *gin.Context) {
		c := Context{g}
		h(c).Render(g)
	}
}

func AsAPI(h Handler) gin.HandlerFunc {
	return func(g *gin.Context) {
		c := Context{g}
		h(c).RenderAPI(g)
	}
}

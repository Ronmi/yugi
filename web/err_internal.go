// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() { addFactory(errInternalTest) }
func errInternalTest() []tmplTestCase {
	return []tmplTestCase{
		{Name: "err_internal.html", Data: &errInternalPage{
			ReqID: "reqid",
			Title: "title",
			Messages: []string{
				"msg1",
				"msg2",
			},
			Next: &errParamNext{
				Path: config.DashboardPage,
				Name: "name",
			},
		}},
		{Name: "err_internal.html", Data: &errInternalPage{
			ReqID: "reqid",
		}},
		{Name: "err_internal.html", Data: &errInternalPage{
			Messages: []string{
				"msg1",
			},
		}},
		{Name: "err_internal.html", Data: &errInternalPage{
			Next: &errParamNext{
				Path: config.DashboardPage,
				Name: "name",
			},
		}},
		{Name: "err_internal.html", Data: &errInternalPage{
			Title: "title",
		}},
		{Name: "err_internal.html", Data: &errInternalPage{}},
	}
}

type extraLogInfo func(l *zerolog.Event) *zerolog.Event

func (e extraLogInfo) And(i extraLogInfo) extraLogInfo {
	return func(l *zerolog.Event) *zerolog.Event {
		return i(e(l))
	}
}

type errInternalPage struct {
	ReqID        string
	Title        string
	Messages     []string
	Next         *errParamNext
	err          error
	extraLogInfo extraLogInfo
}

func errInternal(title string) *errInternalPage {
	return &errInternalPage{
		Title: title,
	}
}

func (e *errInternalPage) With(i extraLogInfo) *errInternalPage {
	if e.extraLogInfo == nil {
		e.extraLogInfo = i
	} else {
		e.extraLogInfo = e.extraLogInfo.And(i)
	}
	return e
}

func (e *errInternalPage) AddError(err error) *errInternalPage {
	e.err = err
	return e
}

func (e *errInternalPage) AddInfo(c Context) *errInternalPage {
	e.ReqID = c.ReqID()
	return e
}

func (e *errInternalPage) AddMessage(msg string) *errInternalPage {
	e.Messages = append(e.Messages, msg)
	return e
}

func (e *errInternalPage) Redir(path, name string, queries ...string) *errInternalPage {
	e.Next = &errParamNext{
		Path:    path,
		Name:    name,
		Queries: queries,
	}
	return e
}

func (e *errInternalPage) Render(c *gin.Context) {
	l := log.Error().
		Str("req_type", "web").
		Str("req_path", c.Request.URL.Path)
	if e.err != nil {
		l = l.Err(e.err)
	}
	if e.ReqID != "" {
		l = l.Str("reqid", e.ReqID)
	}
	l.Msg(e.Title)

	c.HTML(200, "err_internal.html", e)
}

func (e *errInternalPage) RenderAPI(c *gin.Context) {
	l := log.Error().
		Str("req_type", "api").
		Str("req_path", c.Request.URL.Path)
	if e.err != nil {
		l = l.Err(e.err)
	}
	if e.ReqID != "" {
		l = l.Str("reqid", e.ReqID)
	}
	l.Msg(e.Title)

	c.JSON(500, apiError("500", e.Title, gin.H{
		"Next": e.Next,
	}))
}

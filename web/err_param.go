// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"html/template"
	"strings"

	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func init() { addFactory(errParamTest) }
func errParamTest() []tmplTestCase {
	return []tmplTestCase{
		{Name: "err_param.html", Data: &errParamPage{
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
		{Name: "err_param.html", Data: &errParamPage{
			ReqID: "reqid",
		}},
		{Name: "err_param.html", Data: &errParamPage{
			Messages: []string{
				"msg1",
			},
		}},
		{Name: "err_param.html", Data: &errParamPage{
			Next: &errParamNext{
				Path: config.DashboardPage,
				Name: "name",
			},
		}},
		{Name: "err_param.html", Data: &errParamPage{
			Title: "title",
		}},
		{Name: "err_param.html", Data: &errParamPage{}},
	}
}

type errParamNext struct {
	Path    string
	Name    string
	Queries []string
}

func (e *errParamNext) BuildQueries() template.URL {
	if len(e.Queries) == 0 {
		return ""
	}

	return template.URL("?" + strings.Join(e.Queries, "&"))
}

func (e *errParamNext) ToAPI() any {
	return gin.H{
		"Path":    config.URLPath(e.Path),
		"Name":    e.Name,
		"Queries": e.Queries,
	}
}

type errParamPage struct {
	ReqID    string
	Title    string
	Messages []string
	Next     *errParamNext
	Err      error
}

func errParam(title string) *errParamPage {
	return &errParamPage{
		Title: title,
	}
}

func (e *errParamPage) AddError(err error) *errParamPage {
	e.Err = err
	return e
}

func (e *errParamPage) AddInfo(c Context) *errParamPage {
	e.ReqID = c.ReqID()
	return e
}

func (e *errParamPage) AddMessage(msg string) *errParamPage {
	e.Messages = append(e.Messages, msg)
	return e
}

func (e *errParamPage) Redir(path, name string, queries ...string) *errParamPage {
	e.Next = &errParamNext{
		Path:    path,
		Name:    name,
		Queries: queries,
	}
	return e
}

func (e *errParamPage) Render(c *gin.Context) {
	if e.Err != nil {
		log.Error().
			Err(e.Err).
			Str("reqid", e.ReqID).
			Msg("parameter error")
	}
	c.HTML(200, "err_param.html", e)
}

func (e *errParamPage) RenderAPI(c *gin.Context) {
	if e.Err != nil {
		log.Error().
			Err(e.Err).
			Str("reqid", e.ReqID).
			Msg("parameter error")
	}
	c.JSON(400, apiError("400", e.Title, gin.H{
		"Next": e.Next.ToAPI(),
	}))
}

type err404Page struct {
	ReqID string
}

func err404(c Context) *err404Page {
	return &err404Page{
		ReqID: c.ReqID(),
	}
}

func (e *err404Page) Render(c *gin.Context) {
	c.HTML(404, "err_param.html", errParamPage{
		ReqID: e.ReqID,
		Title: "找不到頁面",
		Next: &errParamNext{
			config.DashboardPage,
			"回到首頁",
			nil,
		},
	})
}

func (e *err404Page) RenderAPI(c *gin.Context) {
	c.JSON(404, apiError("404", "找不到頁面", gin.H{
		"Next": gin.H{
			"Path": config.URLPath(config.DashboardPage),
			"Name": "回到首頁",
		},
	}))
}

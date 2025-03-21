// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var funcs = map[string]any{
	"addi": func(a, b int) int { return a + b },
	"subi": func(a, b int) int { return a - b },
	"muli": func(a, b int) int { return a * b },
	"divi": func(a, b int) int { return a / b },
	"modi": func(a, b int) int { return a % b },
	"seqi": func(a, b int) []int {
		if a > b {
			a, b = b, a
		}
		ret := make([]int, b-a+1)
		for i := range ret {
			ret[i] = a + i
		}
		return ret
	},
	"uri": config.URLPathWithCheck,
	"role": func(role actions.Role) any {
		switch role {
		case actions.Manager:
			return "幹部"
		case actions.Volunteer:
			return "志工"
		case actions.Novice:
			return "見習志工"
		case actions.Member:
			return "民眾"
		default:
			return template.HTML("<span style='color: red'>" + template.HTMLEscapeString(string(role)) + "</span>")
		}
	},
	"asRole": func(role string) (actions.Role, error) {
		switch strings.ToLower(role) {
		case "manager":
			return actions.Manager, nil
		case "volunteer":
			return actions.Volunteer, nil
		case "novice":
			return actions.Novice, nil
		case "member":
			return actions.Member, nil
		default:
			return "", errors.New("invalid role: " + role)
		}
	},
	"nl2br": func(s string) template.HTML {
		return template.HTML(strings.ReplaceAll(
			template.HTMLEscapeString(s),
			"\n",
			"<br/>",
		))
	},
	"time": func(t time.Time) string {
		return t.In(config.TZ).Format("2006-01-02 15:04:05")
	},
	"timespecfmt": func(t time.Time) string {
		return t.In(config.TZ).Format(actions.TimeSpecFmt)
	},
	"asStatus": func(str string) (actions.AppointmentStatus, error) {
		ret, ok := actions.AvaialbeAppointmentStatus[str]
		if !ok {
			return 0, errors.New("invalid appointment status: " + str)
		}
		return ret, nil
	},
	"appointmentStatus": func(s actions.AppointmentStatus) string {
		switch s {
		case actions.Pending:
			return "待確認"
		case actions.Contacting:
			return "聯絡中"
		case actions.NotMatched:
			return "幹部處理中"
		case actions.Confirmed:
			return "預約成功"
		case actions.Completed:
			return "已完成"
		case actions.Cancelled:
			return "已取消"
		case actions.Missed:
			return "未出席"
		default:
			return string(s)
		}
	},
}

var tmpl *template.Template

func LoadTmpl() (err error) {
	tmpl, err = template.New("").Funcs(funcs).ParseFS(os.DirFS(config.TmplDir), "*.html")
	if err != nil {
		return err
	}
	return validateTmpl()
}

func execTmpl(w io.Writer, name string, data any) error {
	return tmpl.ExecuteTemplate(w, name, data)
}

type tmplspec interface {
	Render(*gin.Context)
	RenderAPI(*gin.Context)
}

// shared functional tmplspec

// error with redirect
type errRedirPage struct {
	Next    string
	Code    any    `json:"code"`
	Message string `json:"message"`
}

func (r errRedirPage) Render(c *gin.Context) {
	c.Redirect(302, r.Next)
	c.Abort()
}

func (r errRedirPage) RenderAPI(c *gin.Context) {
	c.JSON(400, apiError(fmt.Sprint(r.Code), r.Message, gin.H{"Next": r.Next}))
	c.Abort()
}

// template validation

type tmplTestCase struct {
	Name string
	Data any
}

var caseFactory = []func() []tmplTestCase{}

func addCase(arr []tmplTestCase, name string, data any) []tmplTestCase {
	return append(arr, tmplTestCase{
		Name: name,
		Data: data,
	})
}

func addFactory(fac func() []tmplTestCase) {
	caseFactory = append(caseFactory, fac)
}

func validateTmpl() error {
	for _, fac := range caseFactory {
		for _, param := range fac() {
			log.Trace().
				Str("name", param.Name).
				Interface("data", param.Data).
				Msg("validating " + param.Name)
			err := execTmpl(io.Discard, param.Name, param.Data)
			if err != nil {
				return fmt.Errorf("cannot render %s with test data %+v: %w", param.Name, param.Data, err)
			}
		}
	}

	return nil
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"time"

	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
)

func init() { addFactory(scheduleFormTestCase) }
func scheduleFormTestCase() []tmplTestCase {
	return []tmplTestCase{
		{
			Name: "vol_schedule_form.html",
			Data: &mgrScheduleFormPage{
				Me: actions.FakeUser().SetOrg(actions.FakeOrg()),
			},
		},
	}
}

type mgrScheduleFormPage struct {
	Me *actions.User
}

func (p mgrScheduleFormPage) Render(g *gin.Context) {
	g.HTML(200, "vol_schedule_form.html", p)
}

// unused
func (p mgrScheduleFormPage) RenderAPI(g *gin.Context) {}

func (srv) VolScheduleForm(c Context) tmplspec {
	sess := c.Session()
	me := sess.CurrentUser().Get()

	return mgrScheduleFormPage{
		Me: me,
	}
}

func parseTime(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02T15:04", s, config.TZ)
}

type mgrScheduleNewPage struct {
	Schedule *actions.Schedule
}

func (p mgrScheduleNewPage) Render(g *gin.Context) {
	// 應該要跳轉到排程列表，但目前先跳去首頁
	g.Redirect(302, config.URLPath(config.DashboardPage))
}

func (p mgrScheduleNewPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(p.Schedule))
}

func (srv) VolScheduleNew(c Context) tmplspec {
	type params struct {
		Begin string `form:"begin" binding:"required"`
		End   string `form:"end" binding:"required"`
		Area  string `form:"area" binding:"required"`
	}
	var p params

	if err := c.Bind(&p); err != nil {
		return errParam("缺少必要資料").
			AddInfo(c).
			Redir(config.VolNewScheduleFormPage, "回上一頁")
	}

	begin, err := parseTime(p.Begin)
	if err != nil {
		return errParam("開始時間格式錯誤").
			AddInfo(c).
			Redir(config.VolNewScheduleFormPage, "回上一頁")
	}

	end, err := parseTime(p.End)
	if err != nil {
		return errParam("結束時間格式錯誤").
			AddInfo(c).
			Redir(config.VolNewScheduleFormPage, "回上一頁")
	}

	if begin.After(end) {
		return errParam("開始時間晚於結束時間").
			AddInfo(c).
			Redir(config.VolNewScheduleFormPage, "回上一頁")
	}

	me := c.Session().CurrentUser().Get()
	s, err := actions.CreateSchedule(c.DB(), me, begin, end, p.Area)
	if err != nil {
		return errInternal("無法建立排程").
			AddInfo(c).
			AddMessage("因內部錯誤而無法建立排程，請稍後重試或通知工程師").
			Redir(config.VolNewScheduleFormPage, "回上一頁")
	}

	return mgrScheduleNewPage{Schedule: s}
}

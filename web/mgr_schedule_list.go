// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"errors"

	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
)

func init() { addFactory(mgrScheduleListTestCase) }
func mgrScheduleListTestCase() []tmplTestCase {
	org := actions.FakeOrg()
	mgr := actions.FakeUser().SetOrg(org).SetRole(actions.Manager)
	vol := actions.FakeUser().SetOrg(org).SetRole(actions.Volunteer)
	data := []*actions.Schedule{
		actions.FakeSchedule(mgr),
		actions.FakeSchedule(vol),
		actions.FakeSchedule(mgr).SetDisabled(true),
		actions.FakeSchedule(vol).SetDisabled(true),
	}
	return []tmplTestCase{
		{Name: "mgr_schedule_list.html", Data: &mgrScheduleListPage{
			Me: mgr, Data: data,
		}},
		{Name: "mgr_schedule_list.html", Data: &mgrScheduleListPage{
			Me: mgr, Data: []*actions.Schedule{},
		}},
	}
}

type mgrScheduleListPage struct {
	Me       *actions.User
	Data     []*actions.Schedule
	User     string
	TimeSpec actions.TimeSpec
}

func (s *mgrScheduleListPage) Render(g *gin.Context) {
	g.HTML(200, "mgr_schedule_list.html", s)
}

func (s *mgrScheduleListPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(s.Data))
}

type scheduleListParams struct {
	User        string `form:"user"`
	BeginBefore string `form:"begin_before"`
	BeginAfter  string `form:"begin_after"`
	EndBefore   string `form:"end_before"`
	EndAfter    string `form:"end_after"`
	Include     string `form:"include"`
}

func parseScheduleListParams(c Context) (me *actions.User, u string, t actions.TimeSpec) {
	me = c.Session().CurrentUser().Get()
	var p scheduleListParams
	c.Bind(&p)
	return me, p.User, actions.ParseTimeSpec(
		p.BeginBefore, p.BeginAfter, p.EndBefore, p.EndAfter, p.Include,
	)
}

func (srv) MgrScheduleList(c Context) tmplspec {
	me, u, timeSpec := parseScheduleListParams(c)
	data, err := actions.ListOrgSchedule(c.DB(), me.Org, u, timeSpec)
	if err != nil {
		return errInternal("內部錯誤").
			AddInfo(c).
			AddMessage("請稍後重試或通知工程師").
			Redir(config.DashboardPage, "回首頁")
	}

	return &mgrScheduleListPage{
		Me:       me,
		Data:     data,
		User:     u,
		TimeSpec: timeSpec,
	}
}

type mgrScheduleDisablePage struct{}

func (s *mgrScheduleDisablePage) Render(g *gin.Context) {
	g.Redirect(302, config.URLPath(config.MgrScheduleListPage))
}

func (s *mgrScheduleDisablePage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) MgrScheduleDisable(c Context) tmplspec {
	type params struct {
		ID int64 `form:"id" binding:"required"`
	}
	var p params
	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("請檢查參數是否正確").
			Redir(config.MgrScheduleListPage, "返回列表")
	}

	me := c.Session().CurrentUser().Get()
	if err := actions.MgrDisableSchedule(c.DB(), me, p.ID); err != nil {
		if errors.Is(err, actions.ErrDisableScheduleWithConfirmed) {
			return errParam("參數錯誤").
				AddInfo(c).
				AddMessage("此行程有已經確認的預約，請先處理").
				Redir(config.MgrScheduleListPage, "返回列表")
		}
		return errInternal("內部錯誤").
			AddInfo(c).
			AddError(err).
			AddMessage("請稍後重試或通知工程師").
			Redir(config.MgrScheduleListPage, "返回列表")
	}

	return &mgrScheduleDisablePage{}
}

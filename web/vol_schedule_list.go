// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"errors"

	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func init() { addFactory(volScheduleListTestCase) }
func volScheduleListTestCase() []tmplTestCase {
	org := actions.FakeOrg()
	vol := actions.FakeUser().SetOrg(org).SetRole(actions.Volunteer)
	data := []*actions.Schedule{
		actions.FakeSchedule(vol),
		actions.FakeSchedule(vol).SetDisabled(true),
	}
	return []tmplTestCase{
		{Name: "vol_schedule_list.html", Data: &volScheduleListPage{
			Me: vol, Data: data,
		}},
		{Name: "vol_schedule_list.html", Data: &volScheduleListPage{
			Me: vol, Data: []*actions.Schedule{},
		}},
	}
}

type volScheduleListPage struct {
	Me       *actions.User
	Data     []*actions.Schedule
	User     string
	TimeSpec actions.TimeSpec
}

func (s *volScheduleListPage) Render(g *gin.Context) {
	g.HTML(200, "vol_schedule_list.html", s)
}

func (s *volScheduleListPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(s.Data))
}

func (srv) VolScheduleList(c Context) tmplspec {
	me, u, timeSpec := parseScheduleListParams(c)
	data, err := actions.ListUserSchedule(c.DB(), me, timeSpec)
	if err != nil {
		return errInternal("內部錯誤").
			AddInfo(c).
			AddMessage("請稍後重試或通知工程師").
			Redir(config.DashboardPage, "回首頁")
	}

	return &volScheduleListPage{
		Me:       me,
		Data:     data,
		User:     u,
		TimeSpec: timeSpec,
	}
}

type volScheduleDisablePage struct{}

func (s *volScheduleDisablePage) Render(g *gin.Context) {
	g.Redirect(302, config.URLPath(config.VolScheduleListPage))
}

func (s *volScheduleDisablePage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) VolScheduleDisable(c Context) tmplspec {
	type params struct {
		ID int64 `form:"id" binding:"required"`
	}
	var p params
	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("請檢查參數是否正確").
			Redir(config.VolScheduleListPage, "返回列表")
	}

	me := c.Session().CurrentUser().Get()
	err := actions.VolDisableSchedule(c.DB(), me, p.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errParam("參數錯誤").
				AddInfo(c).
				AddMessage("找不到指定的行程").
				Redir(config.VolScheduleListPage, "返回列表")
		}

		return errInternal("內部錯誤").
			AddInfo(c).
			AddMessage("請稍後重試或通知工程師").
			Redir(config.VolScheduleListPage, "返回列表")
	}

	return &volScheduleDisablePage{}
}

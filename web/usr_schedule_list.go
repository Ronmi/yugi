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

func init() { addFactory(usrScheduleListTestCase) }
func usrScheduleListTestCase() []tmplTestCase {
	org := actions.FakeOrg()
	usr := actions.FakeUser().SetOrg(org).SetRole(actions.Volunteer)
	data := []*actions.Schedule{
		actions.FakeSchedule(usr),
	}
	return []tmplTestCase{
		{Name: "usr_schedule_list.html", Data: &usrScheduleListPage{
			Me: actions.FakeUser(), Data: data,
		}},
		{Name: "usr_schedule_list.html", Data: &usrScheduleListPage{
			Me: actions.FakeUser(), Data: []*actions.Schedule{},
		}},
	}
}

type usrScheduleListPage struct {
	Me       *actions.User
	Data     []*actions.Schedule
	User     string
	TimeSpec actions.TimeSpec
}

func (s *usrScheduleListPage) Render(g *gin.Context) {
	g.HTML(200, "usr_schedule_list.html", s)
}

func (s *usrScheduleListPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(s.Data))
}

func (srv) UsrScheduleList(c Context) tmplspec {
	me, u, timeSpec := parseScheduleListParams(c)
	if timeSpec.Include == nil {
		return &usrScheduleListPage{
			Me:       me,
			User:     u,
			TimeSpec: timeSpec,
		}
	}
	data, err := actions.ListAvailableSchedule(c.DB(), time.Now(), timeSpec)
	if err != nil {
		return errInternal("內部錯誤").
			AddInfo(c).
			AddMessage("請稍後重試或通知工程師").
			Redir(config.DashboardPage, "回首頁")
	}

	return &usrScheduleListPage{
		Me:       me,
		Data:     data,
		User:     u,
		TimeSpec: timeSpec,
	}
}

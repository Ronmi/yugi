// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
)

func init() { addFactory(volAppointmentListTest) }
func volAppointmentListTest() (ret []tmplTestCase) {
	vol := actions.FakeUser().
		SetOrg(actions.FakeOrg()).
		SetRole(actions.Volunteer)
	user := actions.FakeUser()
	schedule := actions.FakeSchedule(vol)
	app := actions.FakeAppointment(user, schedule)
	app2 := actions.FakeAppointment(user, schedule)
	rcpt2 := actions.FakeReceipt(app2)
	rcpt2.Appointment = nil

	ret = addCase(ret, "vol_appointment_list.html", volAppointmentListPage{
		Me:   user,
		Data: []*actions.Appointment{app},
	})
	ret = addCase(ret, "vol_appointment_list.html", volAppointmentListPage{
		Me:   user,
		Data: []*actions.Appointment{app2.SetReceipt(rcpt2)},
	})
	ret = addCase(ret, "vol_appointment_list.html", volAppointmentListPage{
		Me: user,
	})
	return
}

type volAppointmentListPage struct {
	Me   *actions.User
	Data []*actions.Appointment
}

func (p *volAppointmentListPage) Render(g *gin.Context) {
	g.HTML(200, "vol_appointment_list.html", p)
}

func (p *volAppointmentListPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(p.Data))
}

func (srv) VolAppointmentList(c Context) tmplspec {
	me := c.Session().CurrentUser().Get()

	data, err := actions.GetVolAppointments(c.DB(), me)
	if err != nil {
		return errInternal("無法取得預約資料").
			AddInfo(c).
			Redir(config.DashboardPage, "回到首頁")
	}

	return &volAppointmentListPage{Me: me, Data: data}
}

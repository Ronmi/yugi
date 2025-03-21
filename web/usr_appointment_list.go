// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
)

func init() { addFactory(usrAppointmentListTest) }
func usrAppointmentListTest() (ret []tmplTestCase) {
	vol := actions.FakeUser().
		SetOrg(actions.FakeOrg()).
		SetRole(actions.Volunteer)
	user := actions.FakeUser()
	schedule := actions.FakeSchedule(vol)
	app := actions.FakeAppointment(user, schedule)
	app2 := actions.FakeAppointment(user, schedule)
	rcpt2 := actions.FakeReceipt(app2)
	rcpt2.Appointment = nil

	ret = addCase(ret, "usr_appointment_list.html", usrAppointmentListPage{
		Me:   user,
		Data: []*actions.Appointment{app},
	})
	ret = addCase(ret, "vol_appointment_list.html", volAppointmentListPage{
		Me:   user,
		Data: []*actions.Appointment{app2.SetReceipt(rcpt2)},
	})
	ret = addCase(ret, "usr_appointment_list.html", usrAppointmentListPage{
		Me: user,
	})
	return
}

type usrAppointmentListPage struct {
	Me   *actions.User
	Data []*actions.Appointment
}

func (p *usrAppointmentListPage) Render(g *gin.Context) {
	g.HTML(200, "usr_appointment_list.html", p)
}

func (p *usrAppointmentListPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(p.Data))
}

func (srv) UsrAppointmentList(c Context) tmplspec {
	sess := c.Session()
	me := sess.CurrentUser().Get()

	apps, err := actions.GetUserAppointments(c.DB(), me)
	if err != nil {
		return errInternal("無法取得預約資料").
			AddInfo(c).
			AddMessage("請稍後重試，或聯絡工程師").
			Redir(config.DashboardPage, "回到首頁")
	}

	return &usrAppointmentListPage{Me: me, Data: apps}
}

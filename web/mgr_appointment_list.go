// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
)

func init() { addFactory(mgrAppointmentListTest) }
func mgrAppointmentListTest() (ret []tmplTestCase) {
	mgr := actions.FakeUser().
		SetOrg(actions.FakeOrg()).
		SetRole(actions.Volunteer)
	user := actions.FakeUser()
	schedule := actions.FakeSchedule(mgr)
	app := actions.FakeAppointment(user, schedule)
	app2 := actions.FakeAppointment(user, schedule)
	rcpt2 := actions.FakeReceipt(app2)
	rcpt2.Appointment = nil

	ret = addCase(ret, "mgr_appointment_list.html", mgrAppointmentListPage{
		Me:         user,
		Data:       []*actions.Appointment{app},
		StatusList: actions.AvaialbeAppointmentStatus,
	})
	ret = addCase(ret, "mgr_appointment_list.html", mgrAppointmentListPage{
		Me:         user,
		Data:       []*actions.Appointment{app2.SetReceipt(rcpt2)},
		StatusList: actions.AvaialbeAppointmentStatus,
	})
	ret = addCase(ret, "mgr_appointment_list.html", mgrAppointmentListPage{
		Me:         user,
		StatusList: actions.AvaialbeAppointmentStatus,
	})
	return
}

type mgrAppointmentListPage struct {
	Me         *actions.User
	Data       []*actions.Appointment
	StatusList map[string]actions.AppointmentStatus
}

func (p *mgrAppointmentListPage) Render(g *gin.Context) {
	g.HTML(200, "mgr_appointment_list.html", p)
}

func (p *mgrAppointmentListPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(p.Data))
}

func (srv) MgrAppointmentList(c Context) tmplspec {
	me := c.Session().CurrentUser().Get()

	data, err := actions.GetMgrAppointments(c.DB(), me)
	if err != nil {
		return errInternal("無法取得預約資料").
			AddInfo(c).
			AddError(err).
			Redir(config.DashboardPage, "回到首頁")
	}

	return &mgrAppointmentListPage{
		Me:         me,
		Data:       data,
		StatusList: actions.AvaialbeAppointmentStatus,
	}
}

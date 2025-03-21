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

func init() { addFactory(mgrAppointmentDetailTest) }
func mgrAppointmentDetailTest() (ret []tmplTestCase) {
	mgr := actions.FakeUser().
		SetOrg(actions.FakeOrg()).
		SetRole(actions.Volunteer)
	user := actions.FakeUser()
	schedule := actions.FakeSchedule(mgr)
	app := actions.FakeAppointment(user, schedule)
	rcpt := actions.FakeReceipt(app)
	app2 := actions.FakeAppointment(user, schedule)
	rcpt2 := actions.FakeReceipt(app2).
		Receive("中五", 1)
	app3 := actions.FakeAppointment(user, schedule)
	rcpt3 := actions.FakeReceipt(app3).
		Receive("中五", 1).
		Receive("中四", 1)
	app4 := actions.FakeAppointment(user, schedule).SetStatus(actions.Confirmed)

	ret = addCase(ret, "mgr_appointment_detail.html", mgrAppointmentDetailPage{
		Me:         mgr,
		Data:       app,
		StatusList: actions.AvaialbeAppointmentStatus,
	})
	ret = addCase(ret, "mgr_appointment_detail.html", mgrAppointmentDetailPage{
		Me:         mgr,
		Data:       app.SetReceipt(rcpt),
		StatusList: actions.AvaialbeAppointmentStatus,
	})
	ret = addCase(ret, "mgr_appointment_detail.html", mgrAppointmentDetailPage{
		Me:         mgr,
		Data:       app2.SetReceipt(rcpt2),
		StatusList: actions.AvaialbeAppointmentStatus,
	})
	ret = addCase(ret, "mgr_appointment_detail.html", mgrAppointmentDetailPage{
		Me:         mgr,
		Data:       app3.SetReceipt(rcpt3),
		StatusList: actions.AvaialbeAppointmentStatus,
	})
	ret = addCase(ret, "mgr_appointment_detail.html", mgrAppointmentDetailPage{
		Me:         mgr,
		Data:       app4,
		StatusList: actions.AvaialbeAppointmentStatus,
	})
	return
}

type mgrAppointmentDetailPage struct {
	Me         *actions.User
	Data       *actions.Appointment
	StatusList map[string]actions.AppointmentStatus
}

func (p *mgrAppointmentDetailPage) Render(g *gin.Context) {
	g.HTML(200, "mgr_appointment_detail.html", p)
}

func (p *mgrAppointmentDetailPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(p.Data))
}

func (srv) MgrAppointmentDetail(c Context) tmplspec {
	type params struct {
		ID int64 `form:"id" binding:"required"`
	}
	var p params

	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("無法取得預約 ID").
			Redir(config.MgrAppointmentListPage, "回到預約列表")
	}

	me := c.Session().CurrentUser().Get()

	data, err := actions.MgrGetAppointment(c.DB(), me, p.ID)
	if err != nil {
		if errors.Is(err, actions.ErrNotMyAppointment) {
			return errParam("參數錯誤").
				AddInfo(c).
				AddMessage("這不是您的預約").
				Redir(config.MgrAppointmentListPage, "回到預約列表")
		}

		return errInternal("無法取得預約").
			AddInfo(c).
			AddMessage("請稍後再試").
			AddError(err).
			Redir(config.MgrAppointmentListPage, "回到預約列表")
	}

	return &mgrAppointmentDetailPage{
		Me:         me,
		Data:       data,
		StatusList: actions.AvaialbeAppointmentStatus,
	}
}

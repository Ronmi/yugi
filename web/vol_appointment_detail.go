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

func init() { addFactory(volAppointmentDetailTest) }
func volAppointmentDetailTest() (ret []tmplTestCase) {
	vol := actions.FakeUser().
		SetOrg(actions.FakeOrg()).
		SetRole(actions.Volunteer)
	user := actions.FakeUser()
	schedule := actions.FakeSchedule(vol)
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

	ret = addCase(ret, "vol_appointment_detail.html", volAppointmentDetailPage{
		Me:   vol,
		Data: app,
	})
	ret = addCase(ret, "vol_appointment_detail.html", volAppointmentDetailPage{
		Me:   vol,
		Data: app.SetReceipt(rcpt),
	})
	ret = addCase(ret, "vol_appointment_detail.html", volAppointmentDetailPage{
		Me:   vol,
		Data: app2.SetReceipt(rcpt2),
	})
	ret = addCase(ret, "vol_appointment_detail.html", volAppointmentDetailPage{
		Me:   vol,
		Data: app3.SetReceipt(rcpt3),
	})
	ret = addCase(ret, "vol_appointment_detail.html", volAppointmentDetailPage{
		Me:   vol,
		Data: app4,
	})
	return
}

type volAppointmentDetailPage struct {
	Me   *actions.User
	Data *actions.Appointment
}

func (p *volAppointmentDetailPage) Render(g *gin.Context) {
	g.HTML(200, "vol_appointment_detail.html", p)
}

func (p *volAppointmentDetailPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(p.Data))
}

func (srv) VolAppointmentDetail(c Context) tmplspec {
	type params struct {
		ID int64 `form:"id" binding:"required"`
	}
	var p params

	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("無法取得預約 ID").
			Redir(config.VolAppointmentListPage, "回到預約列表")
	}

	me := c.Session().CurrentUser().Get()

	data, err := actions.VolGetAppointment(c.DB(), me, p.ID)
	if err != nil {
		if errors.Is(err, actions.ErrNotMyAppointment) {
			return errParam("參數錯誤").
				AddInfo(c).
				AddMessage("這不是您的預約").
				Redir(config.VolAppointmentListPage, "回到預約列表")
		}

		return errInternal("無法取得預約").
			AddInfo(c).
			AddMessage("請稍後再試").
			AddError(err).
			Redir(config.VolAppointmentListPage, "回到預約列表")
	}

	return &volAppointmentDetailPage{
		Me:   me,
		Data: data,
	}
}

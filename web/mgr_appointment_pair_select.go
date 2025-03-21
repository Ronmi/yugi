// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"errors"
	"strconv"

	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func init() { addFactory(mgrAppointmentPairSelectTest) }
func mgrAppointmentPairSelectTest() (ret []tmplTestCase) {
	mgr := actions.FakeUser().
		SetOrg(actions.FakeOrg()).
		SetRole(actions.Volunteer)
	user := actions.FakeUser()
	schedule := actions.FakeSchedule(mgr)
	app := actions.FakeAppointment(user, schedule)
	rcpt := actions.FakeReceipt(app)
	app = app.SetReceipt(rcpt)
	app2 := actions.FakeAppointment(user, schedule)
	ret = addCase(ret, "mgr_appointment_pair_select.html", &mgrAppointmentPairSelectPage{
		Me:          mgr,
		Appointment: app,
	})
	ret = addCase(ret, "mgr_appointment_pair_select.html", &mgrAppointmentPairSelectPage{
		Me:          mgr,
		Appointment: app,
		Schedules:   []*actions.Schedule{schedule},
	})
	ret = addCase(ret, "mgr_appointment_pair_select.html", &mgrAppointmentPairSelectPage{
		Me:          mgr,
		Appointment: app2,
	})
	ret = addCase(ret, "mgr_appointment_pair_select.html", &mgrAppointmentPairSelectPage{
		Me:          mgr,
		Appointment: app2,
		Schedules:   []*actions.Schedule{schedule},
	})
	return
}

type mgrAppointmentPairSelectPage struct {
	Me          *actions.User
	Appointment *actions.Appointment
	Schedules   []*actions.Schedule
}

func (p *mgrAppointmentPairSelectPage) Render(g *gin.Context) {
	g.HTML(200, "mgr_appointment_pair_select.html", p)
}

func (p *mgrAppointmentPairSelectPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(gin.H{
		"Appointment": p.Appointment,
		"Schedules":   p.Schedules,
	}))
}

func (srv) MgrAppointmentPairSelect(c Context) tmplspec {
	me, _, spec := parseScheduleListParams(c)

	type params struct {
		ID int64 `form:"id" binding:"required"`
	}
	var p params

	if err := c.Bind(&p); err != nil || spec.Include == nil {
		return errParam("無法解析參數").
			AddInfo(c).
			AddError(err).
			AddMessage("請稍後再試或通知工程師").
			Redir(config.MgrAppointmentListPage, "回到預約列表")
	}

	schedules, err := actions.ListOrgSchedule(c.DB(), me.Org, "", spec)
	if err != nil {
		return errInternal("無法取得行程").
			AddInfo(c).
			AddError(err).
			AddMessage("請稍後再試或通知工程師").
			Redir(
				config.MgrAppointmentDetailPage,
				"回到預約明細",
				"id="+strconv.FormatInt(p.ID, 10),
			)
	}

	app, err := actions.MgrGetAppointment(c.DB(), me, p.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errParam("找不到預約").
				AddInfo(c).
				AddMessage("請重新確認").
				Redir(config.MgrAppointmentListPage, "回到預約列表")
		}
		if errors.Is(err, actions.ErrNotMyAppointment) {
			return errParam("無法取得預約").
				AddInfo(c).
				AddMessage("無此權限").
				Redir(config.MgrAppointmentListPage, "回到預約列表")
		}
		return errInternal("無法取得預約").
			AddInfo(c).
			AddError(err).
			AddMessage("請稍後再試或通知工程師").
			Redir(config.MgrAppointmentListPage, "回到預約列表")
	}

	return &mgrAppointmentPairSelectPage{
		Me:          me,
		Appointment: app,
		Schedules:   schedules,
	}
}

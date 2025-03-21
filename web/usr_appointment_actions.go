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

type usrAppointmentMakePage struct {
	Me   *actions.User
	Data *actions.Appointment
}

func (s *usrAppointmentMakePage) Render(g *gin.Context) {
	g.Redirect(302, config.URLPath(config.UsrAppointmentListPage))
}

func (s *usrAppointmentMakePage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(s.Data))
}

func (srv) UsrAppointmentMake(c Context) tmplspec {
	type params struct {
		ScheduleID int64  `form:"schedule_id" binding:"required"`
		Name       string `form:"name" binding:"required"`
		Prefer     string `form:"prefer" binding:"required"`
		Phone      string `form:"phone" binding:"required"`
	}
	var p params

	if err := c.Bind(&p); err != nil {
		return errParam("資料錯誤").
			AddInfo(c).
			AddMessage("請檢查輸入的資料是否正確").
			Redir(config.UsrScheduleListPage, "回到行程列表")
	}

	me := c.Session().CurrentUser().Get()

	a, err := actions.CreateAppointment(c.DB(), p.ScheduleID, me, actions.ContactMethod{
		Name:   p.Name,
		Prefer: p.Prefer,
		Phone:  p.Phone,
	})
	if err != nil {
		if errors.Is(err, actions.ErrScheduleDisabled) {
			return errParam("行程已停用").
				AddInfo(c).
				AddMessage("此行程已停用，無法預約").
				Redir(config.UsrScheduleListPage, "回到行程列表")
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errParam("重複預約").
				AddInfo(c).
				AddMessage("您已經有一份預約了").
				Redir(config.UsrScheduleListPage, "回到行程列表")
		}

		return errInternal("無法建立預約").
			AddInfo(c).
			AddMessage("請稍後再試").
			Redir(config.UsrScheduleListPage, "回到行程列表")
	}

	return &usrAppointmentMakePage{Me: me, Data: a}
}

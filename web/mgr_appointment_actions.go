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

type mgrAppointmentStatusPage struct{}

func (s *mgrAppointmentStatusPage) Render(g *gin.Context) {
	g.Redirect(304, config.URLPath(config.MgrAppointmentListPage))
}

func (s *mgrAppointmentStatusPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) MgrAppointmentStatus(c Context) tmplspec {
	type params struct {
		ID     int64 `form:"id" binding:"required"`
		Status int   `form:"status" binding:"required"`
	}
	var p params
	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("請重新確認後再試").
			Redir(config.MgrAppointmentListPage, "返回列表")
	}

	if p.Status < 0 || p.Status >= int(actions.InvalidStatus) {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("請重新確認後再試").
			Redir(config.MgrAppointmentListPage, "返回列表")
	}

	me := c.Session().CurrentUser().Get()

	err := actions.MgrSetAppointmentStatus(
		c.DB(), me, p.ID, actions.AppointmentStatus(p.Status),
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errParam("參數錯誤").
				AddInfo(c).
				AddMessage("請重新確認後再試").
				Redir(config.MgrAppointmentListPage, "返回列表")
		}
		if errors.Is(err, actions.ErrNotMyAppointment) {
			return errParam("沒有權限").
				AddInfo(c).
				AddMessage("請重新確認後再試").
				Redir(config.MgrAppointmentListPage, "返回列表")
		}
		return errInternal("無法更新預約狀態").
			AddInfo(c).
			AddMessage("請稍後重試或通知工程師").
			Redir(config.MgrAppointmentListPage, "返回列表")
	}

	return &mgrAppointmentStatusPage{}
}

type mgrAppointmentPairPage struct {
	ID int64
}

func (p *mgrAppointmentPairPage) Render(g *gin.Context) {
	g.Redirect(302, config.URLPath(config.MgrAppointmentDetailPage)+"?id="+strconv.FormatInt(p.ID, 10))
}

func (p *mgrAppointmentPairPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) MgrAppointmentPair(c Context) tmplspec {
	type params struct {
		AID int64 `form:"aid" binding:"required"`
		SID int64 `form:"sid" binding:"required"`
	}
	var p params
	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("請重新確認後再試").
			Redir(config.MgrAppointmentListPage, "返回列表")
	}

	me := c.Session().CurrentUser().Get()

	err := actions.MgrPairAppointment(c.DB(), me, p.AID, p.SID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errParam("參數錯誤").
				AddInfo(c).
				AddMessage("請重新確認後再試").
				Redir(config.MgrAppointmentListPage, "返回列表")
		}
		if errors.Is(err, actions.ErrNotMyAppointment) {
			return errParam("沒有權限").
				AddInfo(c).
				AddMessage("請重新確認後再試").
				Redir(config.MgrAppointmentListPage, "返回列表")
		}
		return errInternal("無法更新預約狀態").
			AddInfo(c).
			AddError(err).
			AddMessage("請稍後重試或通知工程師").
			Redir(
				config.MgrAppointmentDetailPage,
				"返回預約明細",
				"id="+strconv.FormatInt(p.AID, 10),
			)
	}

	return &mgrAppointmentPairPage{ID: p.AID}
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
)

type usrAppointmentDeletePage struct{}

func (s *usrAppointmentDeletePage) Render(g *gin.Context) {
	g.Redirect(302, config.URLPath(config.UsrAppointmentListPage))
}

func (s *usrAppointmentDeletePage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) UsrAppointmentDelete(c Context) tmplspec {
	type params struct {
		ID int64 `form:"id" binding:"required"`
	}
	var p params
	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("無法取得預約 ID").
			Redir(config.UsrAppointmentListPage, "返回預約列表")
	}

	me := c.Session().CurrentUser().Get()

	err := actions.UserDeleteAppointment(c.DB(), p.ID, me)
	if err != nil {
		return errInternal("無法刪除預約").
			AddInfo(c).
			AddMessage(err.Error()).
			Redir(config.UsrAppointmentListPage, "返回預約列表")
	}

	return &usrAppointmentDeletePage{}
}

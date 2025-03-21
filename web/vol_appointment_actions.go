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
)

type volAppointmentDeletePage struct{}

func (p *volAppointmentDeletePage) Render(g *gin.Context) {
	g.Redirect(302, config.URLPath(config.VolAppointmentListPage))
}

func (p *volAppointmentDeletePage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) VolAppointmentDelete(c Context) tmplspec {
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
	err := actions.VolDeleteAppointment(c.DB(), p.ID, me)
	if err != nil {
		if errors.Is(err, actions.ErrNotMyAppointment) {
			return errParam("參數錯誤").
				AddInfo(c).
				AddMessage("這不是您的預約").
				Redir(config.VolAppointmentListPage, "回到預約列表")
		}

		return errInternal("無法取消預約").
			AddInfo(c).
			AddMessage("請稍後再試").
			Redir(config.VolAppointmentListPage, "回到預約列表")
	}

	return &volAppointmentDeletePage{}
}

type volAppointmentPubNotePage struct {
	ID         int64
	FromDetail bool
}

func (p *volAppointmentPubNotePage) Render(g *gin.Context) {
	if p.FromDetail {
		g.Redirect(302, config.URLPath(config.VolAppointmentDetailPage)+"?id="+strconv.FormatInt(p.ID, 10))
		return
	}
	g.Redirect(302, config.URLPath(config.VolAppointmentListPage))
}

func (p *volAppointmentPubNotePage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) VolAppointmentPubNote(c Context) tmplspec {
	type params struct {
		ID         int64  `form:"id" binding:"required"`
		Note       string `form:"note" binding:"required"`
		FromDetail bool   `form:"from_detail"`
	}
	var p params

	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("無法取得預約 ID 或備註").
			Redir(config.VolAppointmentListPage, "回到預約列表")
	}

	me := c.Session().CurrentUser().Get()

	err := actions.VolEditAppointmentPublicNote(
		c.DB(), me, p.ID, p.Note,
	)
	if err != nil {
		return errInternal("無法修改備註").
			AddInfo(c).
			AddMessage("請稍後再試").
			Redir(config.VolAppointmentListPage, "回到預約列表")
	}

	return &volAppointmentPubNotePage{p.ID, p.FromDetail}
}

type volAppointmentSecNotePage struct{}

func (p *volAppointmentSecNotePage) Render(g *gin.Context) {
	g.Redirect(302, config.URLPath(config.VolAppointmentListPage))
}

func (p *volAppointmentSecNotePage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) VolAppointmentSecNote(c Context) tmplspec {
	type params struct {
		ID   int64  `form:"id" binding:"required"`
		Note string `form:"note" binding:"required"`
	}
	var p params

	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("無法取得預約 ID 或備註").
			Redir(config.VolAppointmentListPage, "回到預約列表")
	}

	me := c.Session().CurrentUser().Get()

	err := actions.VolEditAppointmentSecretNote(
		c.DB(), me, p.ID, p.Note,
	)
	if err != nil {
		return errInternal("無法修改備註").
			AddInfo(c).
			AddMessage("請稍後再試").
			Redir(config.VolAppointmentListPage, "回到預約列表")
	}

	return &volAppointmentSecNotePage{}
}

type volAppointmentContactPage struct{}

func (p *volAppointmentContactPage) Render(g *gin.Context) {
	g.Redirect(302, config.URLPath(config.VolAppointmentListPage))
}

func (p *volAppointmentContactPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) VolAppointmentContact(c Context) tmplspec {
	type params struct {
		ID int64 `form:"id" binding:"required"`
	}
	var p params

	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("無法取得預約 ID 或聯絡方式").
			Redir(config.VolAppointmentListPage, "回到預約列表")
	}

	me := c.Session().CurrentUser().Get()

	err := actions.VolSetAppointmentStatus(
		c.DB(), me, p.ID, actions.Contacting,
		actions.Pending,
	)
	if err != nil {
		return errInternal("無法把狀態改為聯絡中").
			AddInfo(c).
			AddMessage("請稍後再試").
			AddError(err).
			Redir(config.VolAppointmentListPage, "回到預約列表")
	}

	return &volAppointmentContactPage{}
}

type volAppointmentNotMatchPage struct{}

func (p *volAppointmentNotMatchPage) Render(g *gin.Context) {
	g.Redirect(302, config.URLPath(config.VolAppointmentListPage))
}

func (p *volAppointmentNotMatchPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) VolAppointmentNotMatch(c Context) tmplspec {
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

	err := actions.VolSetAppointmentStatus(
		c.DB(), me, p.ID, actions.NotMatched,
		actions.Contacting,
	)
	if err != nil {
		return errInternal("無法把狀態改為重新配對").
			AddInfo(c).
			AddMessage("請稍後再試").
			Redir(config.VolAppointmentListPage, "回到預約列表")
	}

	return &volAppointmentNotMatchPage{}
}

type volAppointmentConfirmPage struct {
	ID         int64
	FromDetail bool
}

func (p *volAppointmentConfirmPage) Render(g *gin.Context) {
	if p.FromDetail {
		g.Redirect(302, config.URLPath(config.VolAppointmentDetailPage)+"?id="+strconv.FormatInt(p.ID, 10))
		return
	}
	g.Redirect(302, config.URLPath(config.VolAppointmentListPage))
}

func (p *volAppointmentConfirmPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) VolAppointmentConfirm(c Context) tmplspec {
	type params struct {
		ID         int64 `form:"id" binding:"required"`
		FromDetail bool  `form:"from_detail"`
	}
	var p params

	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("無法取得預約 ID").
			AddError(err).
			Redir(config.VolAppointmentListPage, "回到預約列表")
	}

	me := c.Session().CurrentUser().Get()

	err := actions.VolSetAppointmentStatus(
		c.DB(), me, p.ID, actions.Confirmed,
	)
	if err != nil {
		return errInternal("無法把狀態改為確認").
			AddInfo(c).
			AddMessage("請稍後再試").
			AddError(err).
			Redir(config.VolAppointmentListPage, "回到預約列表")
	}

	return &volAppointmentConfirmPage{p.ID, p.FromDetail}
}

type volAppointmentMissedPage struct {
	ID         int64
	FromDetail bool
}

func (p *volAppointmentMissedPage) Render(g *gin.Context) {
	if p.FromDetail {
		g.Redirect(302, config.URLPath(config.VolAppointmentDetailPage)+"?id="+strconv.FormatInt(p.ID, 10))
		return
	}
	g.Redirect(302, config.URLPath(config.VolAppointmentListPage))
}

func (p *volAppointmentMissedPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) VolAppointmentMissed(c Context) tmplspec {
	type params struct {
		ID         int64 `form:"id" binding:"required"`
		FromDetail bool  `form:"from_detail"`
	}
	var p params

	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("無法取得預約 ID").
			AddError(err).
			Redir(config.VolAppointmentListPage, "回到預約列表")
	}

	me := c.Session().CurrentUser().Get()

	err := actions.VolSetAppointmentStatus(
		c.DB(), me, p.ID, actions.Missed,
		actions.Confirmed,
	)
	if err != nil {
		return errInternal("無法把狀態改為未出席").
			AddInfo(c).
			AddMessage("請稍後再試").
			AddError(err).
			Redir(config.VolAppointmentListPage, "回到預約列表")
	}

	return &volAppointmentMissedPage{p.ID, p.FromDetail}
}

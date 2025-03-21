// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
)

type mgrGrantRolePage struct{}

func (mgrGrantRolePage) Render(g *gin.Context) {
	g.Redirect(302, config.URLPath(config.MgrMemberListPage))
}

func (mgrGrantRolePage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) MgrGrantRole(c Context) tmplspec {
	sess := c.Session()
	u := sess.CurrentUser().Get()

	if u.Org == nil {
		return errInternal("no org").
			AddInfo(c).
			AddMessage("請通知管理員").
			With(sess.LogCurrentUser())
	}

	type params struct {
		Name string `form:"name" binding:"required"`
		Role string `form:"role" binding:"required,oneof=novice volunteer manager"`
	}
	var p params
	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("請重新檢查後重試").
			Redir(config.MgrMemberListPage, "回成員列表")
	}

	victim, err := actions.FindUser(c.DB(), p.Name)
	if err != nil {
		return errParam("找不到使用者").
			AddInfo(c).
			AddMessage("請重新檢查後重試").
			Redir(config.MgrMemberListPage, "回成員列表")
	}
	if victim.TOTPSecret == "" {
		return errParam("使用者尚未設定二步驟驗證").
			AddInfo(c).
			AddMessage("請通知使用者設定後再授予權限").
			Redir(config.MgrMemberListPage, "回成員列表")
	}

	err = actions.OrgGrantRole(c.DB(), u, p.Name, actions.Role(p.Role))
	if err != nil {
		return errInternal("無法授予權限").
			AddError(err).
			AddInfo(c).
			AddMessage("請稍後重試一次").
			With(sess.LogCurrentUser())
	}

	return mgrGrantRolePage{}
}

func (srv) MgrRevokeRole(c Context) tmplspec {
	sess := c.Session()
	u := sess.CurrentUser().Get()

	if u.Org == nil {
		return errInternal("no org").
			AddInfo(c).
			AddMessage("請通知管理員").
			With(sess.LogCurrentUser())
	}

	type params struct {
		Name string `form:"name" binding:"required"`
	}
	var p params
	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("請重新檢查後重試").
			Redir(config.MgrMemberListPage, "回成員列表")
	}

	err := actions.OrgRevokeRole(c.DB(), u, p.Name)
	if err != nil {
		return errInternal("無法退出團隊").
			AddError(err).
			AddInfo(c).
			AddMessage("請稍後重試一次").
			With(sess.LogCurrentUser())
	}

	return mgrGrantRolePage{}
}

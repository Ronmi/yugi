// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func init() { addFactory(memberListTest) }
func memberListTest() (ret []tmplTestCase) {
	user := actions.FakeUser().
		SetOrg(actions.FakeOrg()).
		SetRole(actions.Manager)
	ret = addCase(ret, "mgr_member_list.html", mgrMemberListPage{
		Me: user,
	})
	ret = addCase(ret, "mgr_member_list.html", mgrMemberListPage{
		Data: []*actions.User{
			actions.FakeUser().SetOrg(user.Org).SetRole(actions.Volunteer),
		},
		Me: user,
	})
	ret = addCase(ret, "mgr_member_list.html", mgrMemberListPage{
		Data: []*actions.User{
			actions.FakeUser().SetOrg(user.Org).SetRole(actions.Manager),
			actions.FakeUser().SetOrg(user.Org).SetRole(actions.Volunteer),
			actions.FakeUser().SetOrg(user.Org).SetRole(actions.Manager),
		},
		Me: user,
	})
	return
}

type mgrMemberListPage struct {
	Data []*actions.User
	Me   *actions.User
}

func (p *mgrMemberListPage) Render(g *gin.Context) {
	g.HTML(200, "mgr_member_list.html", p)
}

func (p *mgrMemberListPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(p.Data))
}

func (srv) MgrMemberList(c Context) tmplspec {
	sess := c.Session()
	u := sess.CurrentUser().Get()

	if u.Org == nil {
		log.Error().Interface("user", u).Msg("no org")
		return errInternal("no org").AddInfo(c)
	}

	users, err := actions.ListMembers(c.DB(), u.Org.ID)
	if err != nil {
		return errInternal("伺服器錯誤").
			AddInfo(c).
			AddMessage("無法取得成員列表").
			AddMessage("請稍後再試或與系統管理員聯繫").
			Redir(config.DashboardPage, "回首頁")
	}

	return &mgrMemberListPage{Data: users, Me: u}
}

type memberNoteResultPage struct{}

func (_ memberNoteResultPage) Render(g *gin.Context) {
	g.Redirect(302, config.URLPath(config.MgrMemberListPage))
}

func (_ memberNoteResultPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) MgrMemberNote(c Context) tmplspec {
	type params struct {
		ID   string `form:"id" binding:"required"`
		Note string `form:"note"` // 可以是空白
	}
	var p params

	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("請檢查您的輸入").
			Redir(config.MgrMemberListPage, "回成員列表")
	}

	victim, err := actions.FindUserByID(c.DB(), p.ID)
	if err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("無法找到指定的使用者").
			Redir(config.MgrMemberListPage, "回成員列表")
	}

	u := c.Session().CurrentUser().Get()
	if u.Org.ID != victim.Org.ID {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("無法找到指定的使用者").
			Redir(config.MgrMemberListPage, "回成員列表")
	}

	err = actions.UpdateMemberNote(c.DB(), u, victim, p.Note)
	if err != nil {
		return errInternal("伺服器錯誤").
			AddInfo(c).
			AddMessage("無法更新備註").
			Redir(config.MgrMemberListPage, "回成員列表")
	}

	return memberNoteResultPage{}
}

type mgrMemberSecretResultPage struct{}

func (_ mgrMemberSecretResultPage) Render(g *gin.Context) {
	g.Redirect(302, config.URLPath(config.MgrMemberListPage))
}

func (_ mgrMemberSecretResultPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) MgrMemberSecret(c Context) tmplspec {
	type params struct {
		ID     string `form:"id" binding:"required"`
		Secret string `form:"secret"` // 可以是空白
	}
	var p params

	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("請檢查您的輸入").
			Redir(config.MgrMemberListPage, "回成員列表")
	}

	victim, err := actions.FindUserByID(c.DB(), p.ID)
	if err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("無法找到指定的使用者").
			Redir(config.MgrMemberListPage, "回成員列表")
	}

	u := c.Session().CurrentUser().Get()
	if u.Org.ID != victim.Org.ID {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("無法找到指定的使用者").
			Redir(config.MgrMemberListPage, "回成員列表")
	}

	err = actions.UpdateMemberSecret(c.DB(), u, victim, p.Secret)
	if err != nil {
		return errInternal("伺服器錯誤").
			AddInfo(c).
			AddMessage("無法更新秘密備註").
			Redir(config.MgrMemberListPage, "回成員列表")
	}

	return mgrMemberSecretResultPage{}
}

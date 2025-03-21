// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
)

func init() { addFactory(viewLogsTest) }
func viewLogsTest() (ret []tmplTestCase) {
	user := actions.FakeUser()
	ret = addCase(ret, "view_logs.html", viewLogsPage{
		Me: user,
	})
	ret = addCase(ret, "view_logs.html", viewLogsPage{
		Data: []*actions.Log{
			actions.FakeLog(user),
		},
		Me: user,
	})
	return
}

type viewLogsPage struct {
	Me   *actions.User
	Data []*actions.Log
}

func (p *viewLogsPage) Render(g *gin.Context) {
	g.HTML(200, "view_logs.html", p)
}

func (p *viewLogsPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(p.Data))
}

func (srv) ViewMyLogs(c Context) tmplspec {
	sess := c.Session()
	me := sess.CurrentUser().Get()

	data, err := actions.MyLogs(c.DB(), me)
	if err != nil {
		return errInternal("無法取得紀錄").
			AddInfo(c).
			AddMessage("請稍後再試").
			AddError(err).
			Redir(config.DashboardPage, "回到首頁")
	}

	return &viewLogsPage{
		Me:   me,
		Data: data,
	}
}

func (srv) ViewOrgLogs(c Context) tmplspec {
	sess := c.Session()
	me := sess.CurrentUser().Get()

	if me.Role != actions.Manager || me.Org == nil {
		return errInternal("權限不足").
			AddInfo(c).
			AddMessage("您不是幹部").
			Redir(config.DashboardPage, "回到首頁")
	}

	data, err := actions.OrgLogs(c.DB(), me.Org)
	if err != nil {
		return errInternal("無法取得紀錄").
			AddInfo(c).
			AddMessage("請稍後再試").
			AddError(err).
			Redir(config.DashboardPage, "回到首頁")
	}

	return &viewLogsPage{
		Me:   me,
		Data: data,
	}
}

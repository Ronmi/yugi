// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"github.com/Ronmi/yugi/actions"
	"github.com/gin-gonic/gin"
)

func init() { addFactory(dashboardTestCase) }
func dashboardTestCase() []tmplTestCase {
	return []tmplTestCase{
		{Name: "dash.html", Data: &dashboardPage{
			Me: actions.FakeUser().
				SetOrg(actions.FakeOrg()).
				SetRole(actions.Member),
		}},
		{Name: "dash.html", Data: &dashboardPage{
			Me: actions.FakeUser().
				SetOrg(actions.FakeOrg()).
				SetRole(actions.Novice),
		}},
		{Name: "dash.html", Data: &dashboardPage{
			Me: actions.FakeUser().
				SetOrg(actions.FakeOrg()).
				SetRole(actions.Volunteer),
		}},
		{Name: "dash.html", Data: &dashboardPage{
			Me: actions.FakeUser().
				SetOrg(actions.FakeOrg()).
				SetRole(actions.Manager),
		}},
	}
}

type dashboardPage struct {
	Me *actions.User
}

func (u *dashboardPage) Render(g *gin.Context) {
	g.HTML(200, "dash.html", u)
}

func (u *dashboardPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(gin.H{
		"ID":         u.Me.ID,
		"Name":       u.Me.Name,
		"Provider":   u.Me.OauthProvider,
		"Org":        u.Me.Org,
		"Role":       u.Me.Role,
		"2FAEnabled": u.Me.TOTPSecret != "",
	}))
}

func (s *srv) Dashboard(c Context) tmplspec {
	sess := c.Session()
	u := sess.CurrentUser().Get()

	return &dashboardPage{u}
}

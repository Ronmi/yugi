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

func init() { addFactory(orgDetailTest) }
func orgDetailTest() (ret []tmplTestCase) {
	user := actions.FakeUser()
	org := actions.FakeOrg()

	ret = addCase(ret, "org_detail.html", orgDetailPage{
		Data: org,
	})
	ret = addCase(ret, "org_detail.html", orgDetailPage{
		Me:   user,
		Data: org,
	})
	ret = addCase(ret, "org_detail.html", orgDetailPage{
		Me:    user,
		Data:  org,
		Count: 10,
	})
	return
}

type orgDetailPage struct {
	Me    *actions.User
	Data  *actions.Org
	Count int64
}

func (p *orgDetailPage) Render(g *gin.Context) {
	g.HTML(200, "org_detail.html", p)
}

func (p *orgDetailPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(p.Data))
}

func (srv) OrgDetail(c Context) tmplspec {
	type params struct {
		Name string `form:"name" binding:"required"`
	}
	var p params
	if err := c.Bind(&p); err != nil {
		return errParam("找不到這個罷免團體").
			AddInfo(c).
			AddMessage("請重新確認").
			Redir(config.DashboardPage, "回到首頁")
	}
	me := c.Session().CurrentUser().Get()

	org, cnt, err := actions.GetOrg(c.DB(), p.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errParam("找不到這個罷免團體").
				AddInfo(c).
				AddMessage("請重新確認").
				Redir(config.DashboardPage, "回到首頁")
		}
		return errInternal("無法取得罷免團體資料").
			AddInfo(c).
			AddError(err).
			Redir(config.DashboardPage, "回到首頁")
	}

	return &orgDetailPage{
		Me:    me,
		Data:  org,
		Count: cnt,
	}
}

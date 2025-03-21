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

func init() { addFactory(volReceiptFormTest) }
func volReceiptFormTest() (ret []tmplTestCase) {
	vol := actions.FakeUser().
		SetOrg(actions.FakeOrg()).
		SetRole(actions.Volunteer)
	user := actions.FakeUser()
	schedule := actions.FakeSchedule(vol)
	app := actions.FakeAppointment(user, schedule)
	recp := actions.FakeReceipt(app)
	ret = addCase(ret, "vol_receipt_form.html", volReceiptFormPage{
		Me:   vol,
		Data: recp,
	})
	return
}

type volReceiptFormPage struct {
	Me   *actions.User
	Data *actions.Receipt
}

func (p *volReceiptFormPage) Render(g *gin.Context) {
	g.HTML(200, "vol_receipt_form.html", p)
}

func (p *volReceiptFormPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData(p.Data))
}

func (srv) VolReceiptForm(c Context) tmplspec {
	type params struct {
		ID int64 `form:"id" binding:"required"`
	}
	var p params
	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("請檢查網址是否正確").
			Redir(config.DashboardPage, "回首頁")
	}

	me := c.Session().CurrentUser().Get()

	rcpt, err := actions.VolGetReceipt(c.DB(), me, p.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errParam("找不到收據").
				AddInfo(c).
				AddMessage("請檢查網址是否正確").
				Redir(config.DashboardPage, "回首頁")
		}

		if errors.Is(err, actions.ErrNotMyAppointment) {
			return errParam("找不到收據").
				AddInfo(c).
				AddMessage("請檢查網址是否正確").
				Redir(config.DashboardPage, "回首頁")
		}

		return errInternal("無法取得收據").
			AddInfo(c).
			AddError(err).
			AddMessage("請稍後再試").
			Redir(config.DashboardPage, "回首頁")
	}

	return &volReceiptFormPage{
		Me:   me,
		Data: rcpt,
	}
}

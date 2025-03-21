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

type volReceiptCreatePage struct {
	ID int64
}

func (v *volReceiptCreatePage) Render(g *gin.Context) {
	g.Redirect(302, config.URLPath(config.VolReceiptFormPage))
}

func (v *volReceiptCreatePage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) VolReceiptCreate(c Context) tmplspec {
	type params struct {
		ID     int64  `form:"id" binding:"required"`
		Secret string `form:"secret" binding:"required"`
	}
	var p params
	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("請檢查網址是否正確").
			Redir(config.DashboardPage, "回首頁")
	}

	me := c.Session().CurrentUser().Get()

	rcpt, err := actions.VolCreateReceipt(c.DB(), me, p.ID, p.Secret)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errParam("收據已存在").
				AddInfo(c).
				AddMessage("請重新確認").
				Redir(config.DashboardPage, "回首頁")
		}
		if errors.Is(err, actions.ErrNotMyAppointment) {
			return errParam("請不是你的單").
				AddInfo(c).
				AddMessage("請重新確認").
				Redir(config.DashboardPage, "回首頁")
		}
		return errInternal("無法建立收據").
			AddInfo(c).
			AddError(err).
			AddMessage("請稍後重試或回報工程師").
			Redir(config.DashboardPage, "回首頁")
	}

	return &volReceiptCreatePage{ID: rcpt.ID}
}

type volReceiptSignPage struct {
	ID int64
}

func (v *volReceiptSignPage) Render(g *gin.Context) {
	g.Redirect(302, config.URLPath(config.VolAppointmentDetailPage)+"?id="+strconv.FormatInt(v.ID, 10))
}

func (v *volReceiptSignPage) RenderAPI(g *gin.Context) {
	g.JSON(200, apiData("ok"))
}

func (srv) VolReceiptSign(c Context) tmplspec {
	type params struct {
		ID     int64    `form:"id" binding:"required"`
		Note   string   `form:"note"`
		Area   []string `form:"area"`
		Number []int    `form:"number"`
	}
	var p params
	if err := c.Bind(&p); err != nil {
		return errParam("參數錯誤").
			AddInfo(c).
			AddMessage("請檢查網址是否正確").
			Redir(config.DashboardPage, "回首頁")
	}

	me := c.Session().CurrentUser().Get()

	receives := map[string]int{}
	l := len(p.Area)
	if x := len(p.Number); x < l {
		l = x
	}
	for i := 0; i < l; i++ {
		if p.Number[i] > 0 {
			if _, ok := receives[p.Area[i]]; ok {
				return errParam("參數錯誤").
					AddInfo(c).
					AddMessage("請重新確認連署書份數有沒有填錯").
					Redir(
						config.VolAppointmentDetailPage,
						"回到到預約資訊",
						"id="+strconv.FormatInt(p.ID, 10),
					)
			}
			receives[p.Area[i]] = p.Number[i]
		}
	}

	err := actions.VolSignReceipt(c.DB(), me, p.ID, p.Note, receives)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errParam("收據已簽名").
				AddInfo(c).
				AddMessage("請重新確認").
				Redir(config.DashboardPage, "回首頁")
		}
		if errors.Is(err, actions.ErrIncompatibleStatus) {
			return errParam("狀態不正確").
				AddInfo(c).
				AddMessage("預約狀態不正確").
				Redir(config.DashboardPage, "回首頁")
		}
		if errors.Is(err, actions.ErrNotMyAppointment) {
			return errParam("這不是你的收據").
				AddInfo(c).
				AddMessage("請重新確認").
				Redir(config.DashboardPage, "回首頁")
		}
		return errInternal("無法簽名收據").
			AddInfo(c).
			AddError(err).
			AddMessage("請稍後重試或回報工程師").
			Redir(config.DashboardPage, "回首頁")
	}

	return &volReceiptSignPage{ID: p.ID}
}

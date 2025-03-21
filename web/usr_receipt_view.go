// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"errors"
	"html/template"
	"strconv"

	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func init() { addFactory(usrReceiptviewTestCase) }
func usrReceiptviewTestCase() []tmplTestCase {
	uri := "foo"
	qr, _ := createQR(uri)
	me := actions.FakeUser()
	vol := actions.FakeUser().SetOrg(actions.FakeOrg())
	schedule := actions.FakeSchedule(vol)
	app := actions.FakeAppointment(me, schedule)
	recp := actions.FakeReceipt(app)
	return []tmplTestCase{
		{Name: "usr_receipt_view.html", Data: &usrReceiptViewPage{
			Me:      me,
			Receipt: recp,
			URI:     uri,
			QRCode:  template.URL(qr),
		}},
	}
}

type usrReceiptViewPage struct {
	Me      *actions.User
	Receipt *actions.Receipt
	URI     string
	QRCode  template.URL // datauri format
}

func (p *usrReceiptViewPage) Render(c *gin.Context) {
	c.HTML(200, "usr_receipt_view.html", p)
}

func (p *usrReceiptViewPage) RenderAPI(c *gin.Context) {
	c.JSON(200, apiData(gin.H{
		"Receipt": p.Receipt,
		"QRCode":  p.QRCode,
	}))
}

func (srv) UsrReceiptView(c Context) tmplspec {
	type params struct {
		ID int64 `form:"id" binding:"required"`
	}
	var p params
	if err := c.Bind(&p); err != nil {
		return errParam("資料錯誤").
			AddInfo(c).
			AddMessage("請檢查輸入的資料是否正確").
			Redir(config.UsrAppointmentListPage, "回到預約列表")
	}

	me := c.Session().CurrentUser().Get()
	recp, err := actions.UsrGetReceipt(c.DB(), me, p.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errParam("資料錯誤").
				AddInfo(c).
				AddMessage("請檢查輸入的資料是否正確").
				Redir(config.UsrAppointmentListPage, "回到預約列表")
		}

		return errInternal("無法取得簽收條碼").
			AddInfo(c).
			AddError(err).
			AddMessage("請稍後再試").
			Redir(config.UsrAppointmentListPage, "回到預約列表")
	}

	uri := config.FullURL(config.VolReceiptFormPage) + "?id=" + strconv.FormatInt(recp.ID, 10)
	qr, err := createQR(uri)
	if err != nil {
		return errInternal("無法產生簽收條碼").
			AddInfo(c).
			AddError(err).
			AddMessage("請稍後再試").
			Redir(config.UsrAppointmentListPage, "回到預約列表")
	}

	return &usrReceiptViewPage{
		Me:      me,
		Receipt: recp,
		URI:     uri,
		QRCode:  template.URL(qr),
	}
}

type usrReceiptCreatePage struct {
	Receipt *actions.Receipt
}

func (p usrReceiptCreatePage) Render(c *gin.Context) {
	c.Redirect(302, config.URLPath(config.UsrReceiptViewPage)+"?id="+strconv.FormatInt(p.Receipt.ID, 10))
}

func (p usrReceiptCreatePage) RenderAPI(c *gin.Context) {
	c.JSON(200, apiData(p.Receipt))
}

func (srv) UsrReceiptCreate(c Context) tmplspec {
	type params struct {
		ID     int64  `form:"id" binding:"required"`
		Secret string `form:"secret" binding:"required"`
	}
	var p params
	if err := c.Bind(&p); err != nil {
		return errParam("資料錯誤").
			AddInfo(c).
			AddMessage("請檢查輸入的資料是否正確").
			Redir(config.UsrAppointmentListPage, "回到預約列表")
	}

	me := c.Session().CurrentUser().Get()

	ret, err := actions.UsrCreateReceipt(
		c.DB(), me, p.ID, p.Secret,
	)
	if err != nil {
		return errInternal("無法建立簽收條碼").
			AddInfo(c).
			AddError(err).
			AddMessage("請稍後再試").
			Redir(config.UsrAppointmentListPage, "回到預約列表")
	}

	return &usrReceiptCreatePage{
		Receipt: ret,
	}
}

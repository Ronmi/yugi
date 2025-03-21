// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"context"
	"errors"
	"net/url"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	BaseURL     string
	BaseURLPath string
)

func initURL(_ context.Context) error {
	viper.SetDefault("web.baseUrl", "http://localhost:8080")
	BaseURL = viper.GetString("web.baseUrl")
	if !strings.HasSuffix(BaseURL, "/") {
		BaseURL += "/"
	}

	u, err := url.Parse(BaseURL)
	if err != nil {
		return err
	}

	BaseURLPath = u.Path
	log.Debug().Str("base_url", BaseURL).Str("base_url_path", BaseURLPath).Msg("URL config")
	return nil
}

const URLConfig = "url"

func init() {
	runner.Add(URLConfig, initURL, logConfig)
}

// path of pages
const (
	// 登入相關
	LoginSelectPage      = "login"
	LogoutPage           = "logout"
	GoogleCallbackPage   = "auth/google"
	FacebookCallbackPage = "auth/facebook"
	LineCallbackPage     = "auth/line"
	TelegramCallbackPage = "auth/telegram"
	TOTPPage             = "auth/totp" // 驗證 TOTP

	// 所有人都能用的功能
	DashboardPage      = "dashboard"        // 首頁
	Enable2FAStep1Page = "2fa/enable/step1" // 掃 QR code
	Enable2FAStep2Page = "2fa/enable/step2" // 確認驗證碼
	Reset2FAStep1Page  = "2fa/reset/step1"  // 掃 QR code

	// 不同角色看到的內容會不同的功能
	ViewMyLogsPage  = "logs/mine"
	ViewOrgLogsPage = "logs/all"

	// 民眾專用的功能
	UsrScheduleListPage      = "usr/schedule/list"      // 列出所有可預約的項目
	UsrAppointmentMakePage   = "usr/appointment/make"   // 預約
	UsrAppointmentListPage   = "usr/appointment/list"   // 列出自己的所有預約
	UsrAppointmentDetailPage = "usr/appointment/detail" // 預約詳細資料
	UsrAppointmentDeletePage = "usr/appointment/delete" // 取消預約 (限未成立的預約)
	UsrReceiptCreatePage     = "usr/receipt/create"     // 產生簽收條碼
	UsrReceiptViewPage       = "usr/receipt/view"       // 查看簽收條碼

	// 志工用的功能
	VolScheduleListPage        = "vol/schedule/list"         // 列出所有可預約的項目
	VolScheduleNewPage         = "vol/schedule/new"          // 新增可預約的項目
	VolScheduleDisablePage     = "vol/schedule/disable"      // 停用可預約的項目
	VolAppointmentListPage     = "vol/appointment/list"      // 列出跟自己有關的所有預約
	VolAppointmentDetailPage   = "vol/appointment/detail"    // 預約詳細資料
	VolAppointmentDeletePage   = "vol/appointment/delete"    // 取消預約
	VolAppointmentPubNotePage  = "vol/appointment/note"      // 編輯預約公開備註
	VolAppointmentSecNotePage  = "vol/appointment/secret"    // 編輯預約秘密備註
	VolAppointmentContactPage  = "vol/appointment/contact"   // 聯絡中
	VolAppointmentNotMatchPage = "vol/appointment/not-match" // 移交幹部
	VolAppointmentConfirmPage  = "vol/appointment/confirm"   // 確認預約
	VolAppointmentMissedPage   = "vol/appointment/missed"    // 錯過預約
	VolReceiptFormPage         = "vol/receipt/form"          // 簽收頁面
	VolReceiptCreatePage       = "vol/receipt/create"        // 產生簽收條碼
	VolReceiptSignPage         = "vol/receipt/sign"          // 完成簽收

	// 幹部用的功能
	MgrScheduleListPage          = "mgr/schedule/list"           // 列出所有可預約的項目
	MgrScheduleDisablePage       = "mgr/schedule/disable"        // 停用可預約的項目
	MgrMemberListPage            = "mgr/member/list"             // 列出所有成員
	MgrAppointmentListPage       = "mgr/appointment/list"        // 預約詳細資料
	MgrAppointmentStatusPage     = "mgr/appointment/status"      // 修改預約狀態
	MgrAppointmentDetailPage     = "mgr/appointment/detail"      // 預約詳細資料
	MgrAppointmentPairSelectPage = "mgr/appointment/pair/select" // 選擇要重新配對的行程
	MgrAppointmentPairPage       = "mgr/appointment/pair"        // 重新配對預約
	MgrEditMemberNotePage        = "mgr/member/note/edit"        // 編輯成員公開備註
	MgrEditMemberSecretPage      = "mgr/member/secret/edit"      // 編輯成員秘密備註
	MgrGrantRolePage             = "mgr/role/grant"              // 授予角色
	MgrRevokeRolePage            = "mgr/role/revoke"             // 撤銷角色

	// 只在 HTML 模式出現的位址，如錯誤頁面
	VolNewScheduleFormPage = "vol/schedule/form"
)

var validPages = map[string]bool{
	LoginSelectPage:      true,
	LogoutPage:           true,
	GoogleCallbackPage:   true,
	FacebookCallbackPage: true,
	LineCallbackPage:     true,
	TelegramCallbackPage: true,
	TOTPPage:             true,

	DashboardPage:      true,
	Enable2FAStep1Page: true,
	Enable2FAStep2Page: true,
	Reset2FAStep1Page:  true,

	ViewMyLogsPage:  true,
	ViewOrgLogsPage: true,

	UsrScheduleListPage:      true,
	UsrAppointmentMakePage:   true,
	UsrAppointmentListPage:   true,
	UsrAppointmentDetailPage: true,
	UsrAppointmentDeletePage: true,
	UsrReceiptCreatePage:     true,
	UsrReceiptViewPage:       true,

	VolScheduleListPage:        true,
	VolScheduleNewPage:         true,
	VolScheduleDisablePage:     true,
	VolAppointmentListPage:     true,
	VolAppointmentDetailPage:   true,
	VolAppointmentDeletePage:   true,
	VolAppointmentPubNotePage:  true,
	VolAppointmentSecNotePage:  true,
	VolAppointmentContactPage:  true,
	VolAppointmentNotMatchPage: true,
	VolAppointmentConfirmPage:  true,
	VolAppointmentMissedPage:   true,
	VolReceiptFormPage:         true,
	VolReceiptCreatePage:       true,
	VolReceiptSignPage:         true,

	MgrScheduleListPage:          true,
	MgrScheduleDisablePage:       true,
	MgrMemberListPage:            true,
	MgrAppointmentListPage:       true,
	MgrAppointmentStatusPage:     true,
	MgrAppointmentDetailPage:     true,
	MgrAppointmentPairSelectPage: true,
	MgrAppointmentPairPage:       true,
	MgrEditMemberNotePage:        true,
	MgrEditMemberSecretPage:      true,
	MgrGrantRolePage:             true,
	MgrRevokeRolePage:            true,

	VolNewScheduleFormPage: true,
}

func IsValidPage(page string) bool {
	return validPages[page]
}

// page 必須用 config 裡的常數以免錯字
func FullURL(page string) string {
	return BaseURL + page
}

func FullURLWithCheck(page string) (string, error) {
	if !validPages[page] {
		return "", errors.New("invalid page: " + page)
	}

	return FullURL(page), nil
}

// page 必須用 config 裡的常數以免錯字
func URLPath(page string) string {
	return path.Join(BaseURLPath, page)
}

func URLPathWithCheck(page string) (string, error) {
	if !validPages[page] {
		return "", errors.New("invalid page: " + page)
	}

	return URLPath(page), nil
}

func URLPathWithPrefix(prefix, page string) string {
	return path.Join(BaseURLPath, prefix, page)
}

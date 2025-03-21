// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
)

func postAuth(c Context, userID string, provider actions.OauthProvider) tmplspec {
	user, err := actions.Login(c.DB(), userID, provider)
	if err != nil {
		return errInternal("無法登入").
			AddInfo(c).
			AddMessage("伺服器內部錯誤").
			AddMessage("請與志工聯繫").
			Redir(config.LoginSelectPage, "重試登入")
	}

	sess := c.Session()
	defer sess.Save()
	sess.OauthState().Del()
	sess.CurrentUser().Set(user)

	if user.TOTPSecret != "" {
		sess.Authed().Set(false)
		return &authPage{TOTP: true}
	}

	sess.Authed().Set(true)
	return &authPage{
		Next: sess.NextMove().IfZero(config.URLPath(config.DashboardPage)),
	}
}

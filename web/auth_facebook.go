// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"encoding/json"

	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
)

func (srv) AuthFacebook(c Context) tmplspec {
	type params struct {
		Code  string `form:"code" binding:"required"`
		State string `form:"state" binding:"required"`
	}
	var p params

	if err := c.Bind(&p); err != nil {
		return errParam("無效的參數").
			AddInfo(c).
			AddMessage("登入失敗或 Facebook 伺服器錯誤").
			Redir(config.LoginSelectPage, "重試登入")
	}

	// 驗證 state
	state := c.Session().OauthState()
	if !state.NZ() {
		return errParam("內部驗證失敗").
			AddInfo(c).
			AddMessage("無法取得內部驗證碼，請重試").
			Redir(config.LoginSelectPage, "重試登入")
	}

	if state.Get() != p.State {
		return errParam("驗證失敗").
			AddInfo(c).
			AddMessage("錯誤的內部驗證碼，請重試").
			Redir(config.LoginSelectPage, "重試登入")
	}

	token, err := config.FacebookOauth.Exchange(c, p.Code)
	if err != nil {
		return errInternal("無法取得 Facebook 認證金鑰").
			AddInfo(c).
			AddError(err).
			AddMessage("我們無法連線到 Facebook 伺服器").
			Redir(config.LoginSelectPage, "重試登入")
	}

	cl := config.FacebookOauth.Client(c, token)
	res, err := cl.Get("https://graph.facebook.com/v22.0/me?access_token=" + token.AccessToken)
	if err != nil {
		return errInternal("無法取得 Facebook 使用者資訊").
			AddInfo(c).
			AddError(err).
			AddMessage("我們無法連線到 Facebook 伺服器").
			Redir(config.LoginSelectPage, "重試登入")
	}
	defer res.Body.Close()
	type respjson struct {
		ID string `json:"id"`
	}
	var resp respjson
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return errInternal("無法解析 Facebook 使用者資訊").
			AddInfo(c).
			AddError(err).
			AddMessage("我們無法解析 Facebook 伺服器的回應").
			Redir(config.LoginSelectPage, "重試登入")
	}

	return postAuth(c, resp.ID, actions.Facebook)
}

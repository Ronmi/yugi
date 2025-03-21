// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"context"
	"time"

	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

func (srv) AuthLine(c Context) tmplspec {
	nonce := c.Session().LoginNonce().Get()
	defer c.Session().LoginNonce().Del()

	type params struct {
		Code  string `form:"code" binding:"required"`
		State string `form:"state" binding:"required"`
	}
	var p params

	if err := c.Bind(&p); err != nil {
		return errParam("無效的參數").
			AddInfo(c).
			AddMessage("登入失敗或 Line 伺服器錯誤").
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

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()
	accToken, err := config.LineOauth.Exchange(ctx, p.Code)
	if err != nil {
		return errInternal("無法取得 Line 資訊").
			AddInfo(c).
			AddMessage("我們無法連線到 Line 伺服器").
			Redir(config.LoginSelectPage, "重試登入")
	}

	log.Trace().
		Str("nonce", nonce).
		Interface("id_token", accToken.Extra("id_token")).
		Msg("dump info")
	idjwt, ok := accToken.Extra("id_token").(string)
	if !ok {
		return errInternal("無法取得 Line 資訊").
			AddInfo(c).
			AddMessage("Line 伺服器沒有回傳你的資料").
			Redir(config.LoginSelectPage, "重試登入")
	}

	type myClaims struct {
		Nonce string `json:"nonce"`
		jwt.RegisteredClaims
	}

	idToken, err := jwt.ParseWithClaims(idjwt, &myClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(config.LineOauth.ClientSecret), nil
	},
		jwt.WithIssuer("https://access.line.me"),
		jwt.WithAudience(config.LineOauth.ClientID),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return errInternal("無法取得 Line 資訊").
			AddInfo(c).
			AddMessage("Line 伺服器回傳的資料無法驗證").
			Redir(config.LoginSelectPage, "重試登入")
	}
	log.Trace().
		Interface("claims", idToken.Claims).
		Msg("claims")
	claims, ok := idToken.Claims.(*myClaims)
	if !ok || claims.Nonce != nonce {
		return errInternal("無法取得 Line 資訊").
			AddInfo(c).
			AddMessage("Line 伺服器回傳的資料中，驗證碼不相符").
			Redir(config.LoginSelectPage, "重試登入")
	}

	return postAuth(c, claims.Subject, actions.Line)
}

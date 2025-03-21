// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"context"
	"crypto/sha256"
	"errors"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

var (
	GoogleOauth   *oauth2.Config
	FacebookOauth *oauth2.Config
	LineOauth     *oauth2.Config
	TelegramAuth  *TelegramAuthConfig
)

type TelegramAuthConfig struct {
	Bot    string
	Size   string
	Radius int
	Write  bool

	// cache fields
	TokenHash []byte
}

func getOauthConfig(e error, key string, ep oauth2.Endpoint, cb string, scope ...string) (*oauth2.Config, error) {
	if e != nil {
		return nil, e
	}
	client := viper.GetString("auth." + key + ".client")
	secret := viper.GetString("auth." + key + ".secret")
	if client == "" || secret == "" {
		return nil, nil
	}

	cb, err := FullURLWithCheck(cb)
	if err != nil {
		return nil, err
	}

	return &oauth2.Config{
		ClientID:     client,
		ClientSecret: secret,
		Endpoint:     ep,
		RedirectURL:  cb,
		Scopes:       scope,
	}, nil
}

func initTelegramAuth() {
	viper.SetDefault("auth.telegram.size", "large")
	viper.SetDefault("auth.telegram.radius", 0)
	viper.SetDefault("auth.telegram.write", false)
	bot := viper.GetString("auth.telegram.bot")
	token := viper.GetString("auth.telegram.token")
	size := viper.GetString("auth.telegram.size")
	radius := viper.GetInt("auth.telegram.radius")
	write := viper.GetBool("auth.telegram.write")

	if bot == "" || token == "" {
		return
	}

	h := sha256.New()
	h.Write([]byte(token))
	TelegramAuth = &TelegramAuthConfig{
		Bot:       bot,
		Size:      size,
		Radius:    radius,
		Write:     write,
		TokenHash: h.Sum(nil),
	}
}

func initOauth(_ context.Context) (err error) {
	GoogleOauth, err = getOauthConfig(err, "google", google.Endpoint, GoogleCallbackPage, "profile", "email")
	FacebookOauth, err = getOauthConfig(err, "facebook", facebook.Endpoint, FacebookCallbackPage, "email")
	LineOauth, err = getOauthConfig(err, "line", oauth2.Endpoint{
		AuthURL:  "https://access.line.me/oauth2/v2.1/authorize",
		TokenURL: "https://api.line.me/oauth2/v2.1/token",
	}, LineCallbackPage, "openid")
	initTelegramAuth()

	if GoogleOauth == nil && FacebookOauth == nil && LineOauth == nil && TelegramAuth == nil {
		return errors.New("no oauth config")
	}

	return
}

const authConfig = "auth"

func init() {
	runner.Add(authConfig, initOauth, URLConfig)
}

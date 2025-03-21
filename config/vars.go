// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	TmplDir   string
	StaticDir string
	BindAddr  string
	CertFile  string
	KeyFile   string
)

func initVars(_ context.Context) error {
	viper.SetDefault("web.bindAddr", ":8080")
	TmplDir = viper.GetString("web.tmplDir")
	StaticDir = viper.GetString("web.staticDir")
	BindAddr = viper.GetString("web.bindAddr")
	CertFile = viper.GetString("web.certFile")
	KeyFile = viper.GetString("web.keyFile")

	log.Info().
		Str("tmpl_dir", TmplDir).
		Str("bind_addr", BindAddr).
		Str("static_dir", StaticDir).
		Str("cert_file", CertFile).
		Str("key_file", KeyFile).
		Msg("config vars")
	return nil
}

const VarsConfig = "vars"

func init() {
	runner.Add(VarsConfig, initVars, logConfig, randNameConfig)
}

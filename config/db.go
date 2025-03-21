// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"context"
	"net/url"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/raohwork/gorm0log"
	"github.com/raohwork/task"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var DB *gorm.DB

func initDB() (err error) {
	viper.SetDefault("db.dsn", "sqlite3://:memory:")
	dsn := viper.GetString("db.dsn")
	dsnURL, err := url.Parse(dsn)
	if err != nil {
		return err
	}
	dsnURL.Scheme = "sqlite3"
	params := dsnURL.Query()
	params.Add("_pragma", "journal_mode=WAL")
	params.Add("_pragma", "encoding='UTF-8'")
	dsn = dsnURL.Host + "?" + params.Encode()

	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger: &gorm0log.Logger{
			Logger: log.Logger,
			Config: gorm0log.Config{
				SlowThreshold:        time.Second,
				ErrorLevel:           gorm0log.DebugCommonErr,
				ParameterizedQueries: false,
				Customize: func(c context.Context, l *zerolog.Event) {
					x := c.Value("reqid")
					if x != nil {
						l.Interface("reqid", x)
					}
				},
			},
		},
	})
	if err != nil {
		return
	}

	log.Debug().Str("dsn", dsn).Msg("database connected")
	return
}

const DBConfig = "db"

func init() {
	runner.Add(DBConfig, task.NoCtx(initDB), logConfig)
}

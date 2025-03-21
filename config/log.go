// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func DefaultLogWriter(out io.Writer) zerolog.ConsoleWriter {
	w := zerolog.NewConsoleWriter()
	w.Out = out
	w.NoColor = !viper.GetBool("log.color")
	w.TimeFormat = "2006-01-02 15:04:05.000"
	w.TimeLocation = TZ
	return w
}

func newLogger(out io.Writer) zerolog.Logger {
	return zerolog.New(DefaultLogWriter(out)).With().Timestamp().Logger()
}

func initLog(_ context.Context) (err error) {
	viper.SetDefault("log.level", "error")
	viper.SetDefault("log.color", false)

	level, err := zerolog.ParseLevel(viper.GetString("log.level"))
	if err != nil {
		return
	}

	log.Logger = newLogger(os.Stderr)
	zerolog.SetGlobalLevel(level)

	log.Debug().Str("config", viper.ConfigFileUsed()).Msg("config file loaded")
	return
}

const logConfig = "log"

func init() {
	runner.Add(logConfig, initLog)
}

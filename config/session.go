// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"crypto/rand"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/raohwork/task"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var SessionStore sessions.Store

func genSessKey(size int) ([]byte, error) {
	key := make([]byte, size)
	_, err := rand.Read(key)
	return key, err
}

func initSess() (err error) {
	authKey := []byte(viper.GetString("session.authKey"))
	encKey := []byte(viper.GetString("session.encKey"))
	if len(authKey) == 0 {
		authKey, err = genSessKey(64)
		if err != nil {
			return
		}
	}
	if len(encKey) == 0 {
		encKey, err = genSessKey(32)
		if err != nil {
			return
		}
	}
	SessionStore = memstore.NewStore(authKey, encKey)

	log.Info().Msg("session store initialized")
	return nil
}

const sessConfig = "session"

func init() {
	runner.Add(sessConfig, task.NoCtx(initSess), logConfig)
}

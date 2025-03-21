// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"fmt"
	"math/rand"

	"github.com/raohwork/task"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var charGroups [4][]rune = [4][]rune{}

func initRandName() (err error) {
	groups := [4]string{
		viper.GetString("randName.group1"),
		viper.GetString("randName.group2"),
		viper.GetString("randName.group3"),
		viper.GetString("randName.group4"),
	}
	cnt := int64(1)
	for i, group := range groups {
		charGroups[i] = []rune(group)
		l := len(charGroups[i])
		if l < 30 {
			return fmt.Errorf("randName.group%d 太短: %d", i+1, l)
		}
		cnt *= int64(l)
	}

	log.Info().Int64("count", cnt).Msg("randName 可能的组合数")

	return nil
}

func RandName() string {
	var name string
	for i := 0; i < 4; i++ {
		name += string(charGroups[i][rand.Intn(len(charGroups[i]))])
	}
	return name
}

const (
	AlphaLower = "abcdefghijklmnopqrstuvwxyz"
	AlphaUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Alpha      = AlphaLower + AlphaUpper
	Digits     = "0123456789"
	Symbols    = "!@#$%^&*()-_=+[]{};:'\",.<>/?\\|`~"
	AlNum      = Alpha + Digits
	AlNumSym   = AlNum + Symbols
)

func RandStr(size int, chars string) string {
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = chars[rand.Intn(len(chars))]
	}
	return string(buf)
}

const randNameConfig = "rand_name"

func init() {
	runner.Add(randNameConfig, task.NoCtx(initRandName))
}

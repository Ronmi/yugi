// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"errors"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/raohwork/task"
	"github.com/spf13/viper"
)

var otpIssuer string

const otpConfig = "otp"

func init() {
	runner.Add(otpConfig, task.NoCtx(func() error {
		otpIssuer = viper.GetString("otp.issuer")
		if otpIssuer == "" {
			return errors.New("otp.issuer is required")
		}

		return nil
	}), logConfig)
}

func CreateOTPKey(acct string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      otpIssuer,
		AccountName: acct,
	})
}

func ValidateOTP(code, secret string) bool {
	return totp.Validate(code, secret)
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"context"
	"time"

	"github.com/raohwork/task/deptask"
)

func Init(ctx context.Context) error {
	return runner.RunSomeSync(ctx, runner.ListDeps()...)
}

func RunSome(tasks ...string) error {
	return runner.RunSomeSync(context.Background(), tasks...)
}

var runner = deptask.New()

var TZ = time.FixedZone("TW", 8*3600)

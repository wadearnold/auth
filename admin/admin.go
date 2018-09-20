// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package admin

import (
	"runtime"
)

// Init is the entrypoint into the admin package. It will
// configure the runtime.
func Init() error {
	if pprofProfileEnabled("block", true) {
		runtime.SetBlockProfileRate(1)
	}
	if pprofProfileEnabled("mutex", true) {
		runtime.SetMutexProfileFraction(1)
	}
	return nil
}

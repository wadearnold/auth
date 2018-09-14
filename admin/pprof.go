// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package admin

import (
	"fmt"
	"os"
	"strings"
)

// pprofHandlers is a map which holds configuration for
// serving various pprof handlers on the admin servlet.
//
// We only want to expose on the admin servlet because these
// profiles/dumps can contain sensitive info (i.e. password hash, email)
// or alter the app performance.
//
// These values can be overridden by setting the appropriate PPROF_*
// environment variable.
//
// TODO(adam): generate readme.md env vars from this
var pprofHandlers = map[string]bool{
	"allocs":       true,
	"block":        true,
	"cmdline":      true,
	"goroutine":    true,
	"heap":         true,
	"mutex":        true,
	"profile":      true,
	"threadcreate": true,
	"trace":        true,
}

// pprofProfileEnabled calls out to os.Getenv with the following pattern:
//  PPROF_$name where $name is uppercase
//
// A string of "yes" returns true, and "no" returns false, otherwise
// zero is returned. Empty strings return zero.
func pprofProfileEnabled(name string, zero bool) bool {
	v := os.Getenv(fmt.Sprintf("PPROF_%s", strings.ToUpper(name)))
	switch strings.ToLower(v) {
	case "yes":
		return true
	case "no":
		return false
	}
	return zero
}

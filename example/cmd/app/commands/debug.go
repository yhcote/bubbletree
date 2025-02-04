// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

//go:build !release

package commands

import (
	"net/http"
	_ "net/http/pprof"
)

const releaseBuild = false

func init() {
	if !releaseBuild {
		// benchmarking and profiling
		go func() {
			log.Error(http.ListenAndServe("localhost:6060", nil).Error())
		}()
	}
}

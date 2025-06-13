// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

//go:build !release

package commands

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/yhcote/bubbletree/logger"
)

const (
	releaseBuild    = false
	pprofServerAddr = "localhost:6161"
)

func init() {
	if !releaseBuild {
		// benchmarking and profiling
		logger.Log().Info("Starting go profiling server", "addr", pprofServerAddr)
		go func() {
			logger.Log().Error(http.ListenAndServe(pprofServerAddr, nil).Error())
		}()
	}
}

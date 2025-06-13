// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package app

import (
	"errors"

	"github.com/spf13/viper"
	"github.com/yhcote/bubbletree/logger"
)

var (
	ErrViperToLocalConfig = errors.New("could not translate viper to local app settings")
)

// Set some reasonable defaults when possible
var (
	DefaultPlaceholder = "-PLACEHOLDER-"
)

// ViperToLocalConfig translates configurations loaded into Viper into a
// locally structured config instance.
func ViperToLocalConfig(vpr *viper.Viper) (config Config, err error) {
	if err = vpr.UnmarshalExact(&config); err != nil {
		logger.Log().Error("unmarshal error", "error", err)
		return config, ErrViperToLocalConfig
	}
	return
}

// Config is the local viper unmarshaled application config
type Config struct {
	Placeholder string
}

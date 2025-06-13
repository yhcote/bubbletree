// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package configurator

import (
	"example/internal/app"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

// Msg/Cmd's

type (
	// ConfigReadyMsg is a model-global message sent when the model is
	// done loading or creating the system config.
	ConfigReadyMsg struct {
		Config app.Config
	}

	// ConfigMissingMsg is a model-global message sent when no configuration
	// was found at default locations. In this case we prepare a user input
	// form.
	ConfigMissingMsg struct{}
)

// getConfigCmd is responsible for loading an existing application
// configuration file, if available, or to create a new one otherwise. The
// function calls 'isComplete()' returning whether a config file has all
// required settings. This allows the model to capture missing settings even
// when a config file is found but incomplete.
func getConfigCmd(viper *viper.Viper) tea.Cmd {
	return func() tea.Msg {
		if config, complete := isComplete(viper); complete {
			return ConfigReadyMsg{Config: config}
		} else {
			return ConfigMissingMsg{}
		}
	}
}

// configReadyCmd returns a model-global message when the system configuration
// is ready.
func configReadyCmd(config app.Config) tea.Cmd {
	return func() tea.Msg {
		return ConfigReadyMsg{Config: config}
	}
}

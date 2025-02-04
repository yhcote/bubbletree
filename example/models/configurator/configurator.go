// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package configurator

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

// Msg/Cmd's

type (
	// ConfigReadyMsg is a model-global message sent when the model is
	// done loading or creating the system config.
	ConfigReadyMsg struct{}

	// ConfigMissingMsg is a model-global message sent when no configuration
	// was found at default locations. In this case we prepare a user input
	// form.
	ConfigMissingMsg struct{}
)

// getConfigCmd is responsible for loading an existing application configuration
// file, if available, or to create a new one otherwise. The function calls
// 'incompleteConfig()' returning if a config file is missing required
// settings. This allows the model to capture missing settings even when a
// config file is found but incomplete.
func getConfigCmd(viper *viper.Viper) tea.Cmd {
	return func() tea.Msg {
		if incompleteConfig(viper) {
			return ConfigMissingMsg{}
		}
		return ConfigReadyMsg{}
	}
}

// configReadyCmd returns a model-global message when the system configuration
// is ready. modelFinishedCmd is also batched in because this marks the end of
// the model's execution.
func configReadyCmd() tea.Msg {
	return ConfigReadyMsg{}
}

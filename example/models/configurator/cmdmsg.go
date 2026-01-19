// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package configurator

import (
	"example/internal/app"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
	"github.com/yhcote/bubbletree"
)

// Msg/Cmd's

type (
	// ConfigReadyMsg is a model-global message sent when the model is
	// done loading or creating the system config. The loaded/created
	// config is sent along with the message.
	ConfigReadyMsg struct {
		Config app.Config
	}

	// ConfigMissingMsg is a model-global message sent when no configuration
	// was found at default locations. In this case we prepare a user input
	// form.
	ConfigMissingMsg struct{}

	// ConfigCancelMsg is a model-global message sent when a configuration
	// session is cancelled by the user. The model disables the form.
	ConfigCancelMsg struct{}
)

// GetConfigCmd is responsible for loading an existing application
// configuration file, if available, or to create a new one otherwise. The
// function calls 'isComplete()' returning whether a config file has all
// required settings. This allows the model to capture missing settings even
// when a config file is found but incomplete.
func GetConfigCmd(viper *viper.Viper, reconf bool) tea.Cmd {
	return func() tea.Msg {
		config, complete, err := isComplete(viper)
		if err != nil {
			return bubbletree.ErrMsg{Err: fmt.Errorf("GetConfigCmd: unexpected error while checking config: %w", err)}
		}
		if complete && !reconf {
			return ConfigReadyMsg{Config: config}
		}
		return ConfigMissingMsg{}
	}
}

// CancelConfigCmd sends a configuration session cancellation.
func CancelConfigCmd() tea.Cmd {
	return func() tea.Msg {
		return ConfigCancelMsg{}
	}
}

// configReadyCmd returns a model-global message when the system configuration
// is ready.
func configReadyCmd(config app.Config) tea.Cmd {
	return func() tea.Msg {
		return ConfigReadyMsg{Config: config}
	}
}

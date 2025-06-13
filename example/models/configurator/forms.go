// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package configurator

import (
	"encoding/json"
	"fmt"
	"os"

	"example/internal/app"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
	"github.com/yhcote/bubbletree"
)

// Form input config string variables. Those buffers the new configurations
// entered by the user before the form is completed, validated, and added to
// viper.
var (
	inputPlaceholder string
)

// newForm creates the input form needed to collect application required
// configuration settings if those are missing.
func newForm(vpr *viper.Viper) (form *huh.Form) {
	config, err := app.ViperToLocalConfig(vpr)
	if err != nil {
		return
	}

	// When a config file was loaded and values exist use them, otherwise
	// use defaults when available.
	inputPlaceholder = config.Placeholder

	// Otherwise use defaults when available.
	if inputPlaceholder == "" {
		inputPlaceholder = app.DefaultPlaceholder
	}

	form = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Empty Configuration"),
			huh.NewInput().
				Title("General Config • Placeholder").
				Description("An example config placeholder").
				Value(&inputPlaceholder).
				Validate(stringFieldValidation),
			huh.NewConfirm().
				Title("Save new config?").
				Validate(func(v bool) error {
					if !v {
						return fmt.Errorf("Continue with changes...") //nolint:staticcheck
					}
					return nil
				}).
				Affirmative("Yes").Negative("No"),
		),
	).WithShowHelp(false).WithShowErrors(false)
	return
}

// stringFieldValidation checks that a text field has been filled.
func stringFieldValidation(v string) error {
	if v == "" {
		return fmt.Errorf("Field cannot be empty") //nolint:staticcheck
	}
	return nil
}

// updateForm is called within the model Update() routine to process messages
// of an active form.
func (m Model) updateForm(msg tea.Msg) (bubbletree.LeafModel, tea.Cmd) {
	var cmds []tea.Cmd

	// Process the current form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		// We're done here, save form fields to viper configs and write down a new
		// config version to disk.
		config := app.Config{
			Placeholder: inputPlaceholder,
		}

		jsonConfig, err := json.MarshalIndent(config, "", "    ")
		if err != nil {
			err = fmt.Errorf("while marshaling config to JSON: %w", err)
			cmds = append(cmds, bubbletree.ErrCmd(err))
		} else {
			m.Logger.Debug("JSON Config", "config", string(jsonConfig))
			if err = os.WriteFile(m.Viper.ConfigFileUsed(), jsonConfig, 0644); err != nil {
				err = fmt.Errorf("while writing JSON config to file: %w", err)
				cmds = append(cmds, bubbletree.ErrCmd(err))
			} else {
				cmds = append(cmds, configReadyCmd(config))
			}
		}
	}
	return m, tea.Batch(cmds...)
}

// incompleteConfig returns whether the current config (new or read from
// file), is missing application required values.
func isComplete(vpr *viper.Viper) (config app.Config, complete bool) {
	config, err := app.ViperToLocalConfig(vpr)
	if err != nil {
		return
	}
	if config.Placeholder == "" {
		return
	}
	return config, true
}

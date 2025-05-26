// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package configurator

import (
	"fmt"

	"example/internal/app"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
	"github.com/yhcote/bubbletree"
)

// newForm creates the input form needed to collect application required
// configuration settings if those are missing.
func newForm(viper *viper.Viper) *huh.Form {
	if viper == nil {
		log.Error("passed viper configs is 'nil'")
		return nil
	}

	// Get viper config values from existing config file.
	name := viper.GetString(app.EmployeeName)
	dob := viper.GetString(app.EmployeeDOB)
	role := viper.GetString(app.EmployeeRole)
	salary := viper.GetString(app.EmployeeSalary)
	hiringDate := viper.GetString(app.EmployeeHiringDate)
	passwd := viper.GetString(app.EmployeeAcctPasswd)

	// Apply defaults when no values came from the existing config file.
	if role == "" {
		role = app.DefaultRole
	}
	if passwd == "" {
		passwd = app.DefaultPasswd
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Employee Configuration").
				Description("Enter required fields below."),
			huh.NewInput().Key(app.EmployeeName).
				Title("Employee • Name").
				Description("Name of the employee").
				Value(&name).
				Validate(fieldValidation),
			huh.NewInput().Key(app.EmployeeDOB).
				Title("Employee • DOB").
				Description("Date Of Birth of the employee").
				Value(&dob).
				Validate(fieldValidation),
			huh.NewInput().Key(app.EmployeeRole).
				Title("Employee • Role").
				Description("Role of the employee").
				Value(&role).
				Validate(fieldValidation),
		),
		huh.NewGroup(
			huh.NewNote().
				Title("Employee Configuration").
				Description("Enter required fields below, (Cont.)."),
			huh.NewInput().Key(app.EmployeeSalary).
				Title("Employee • Salary").
				Description("Salary of the employee").
				Value(&salary).
				Validate(fieldValidation),
			huh.NewInput().Key(app.EmployeeHiringDate).
				Title("Employee • Hiring Date").
				Description("Hiring date of the employee").
				Value(&hiringDate).
				Validate(fieldValidation),
			huh.NewInput().Key(app.EmployeeAcctPasswd).
				Title("Employee • Account Password").
				Description("Initial password of the employee's account").
				Value(&passwd).
				Validate(fieldValidation).
				EchoMode(huh.EchoModePassword),
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
}

// fieldValidation checks that a text field has been filled.
func fieldValidation(v string) error {
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
		for _, key := range app.AllConfigKeys {
			m.Viper.Set(key, m.form.GetString(key))
		}

		if err := m.Viper.WriteConfig(); err != nil {
			err = fmt.Errorf("while writing viper config file: %v", err)
			cmds = append(cmds, bubbletree.ErrCmd(err))
		} else {
			cmds = append(cmds, configReadyCmd)
		}
		log.Debug("Config", "all settings", spew.Sdump(m.Viper.AllSettings()))
	}
	return m, tea.Batch(cmds...)
}

// incompleteConfig returns whether the current config (new or read from
// file), is missing application required values.
func incompleteConfig(viper *viper.Viper) bool {
	for _, key := range app.AllConfigKeys {
		if viper.Get(key) == nil {
			log.Warn("required setting not set", "setting", key)
			return true
		}
	}
	return false
}

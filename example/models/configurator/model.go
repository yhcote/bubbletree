// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package configurator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
	"github.com/yhcote/bubbletree"
)

const (
	// short model name used for identification.
	modelName = "configurator"
)

var (
	// unique model instance id based on 'modelName'.
	lastID atomic.Int64
)

// New creates and initializes a new model ready to be used.
func New(opts ...Option) bubbletree.LeafModel {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Model{
		DefaultLeafModel: bubbletree.DefaultLeafModel{
			DefaultCommonModel: bubbletree.DefaultCommonModel{
				ID:     fmt.Sprintf("%s-%d", modelName, lastID.Add(1)),
				Ctx:    ctx,
				Cancel: cancel,
			},
		},
	}
	for _, opt := range opts {
		opt(m)
	}

	m.form = newForm(m.Viper)

	m.Logger.Info("New model created", "ModelID", m.ID)
	return m
}

// Model is the definition of the configurator model.
type Model struct {
	// Include fields and default methods of bubbletree.DefaultLeafModel.
	bubbletree.DefaultLeafModel

	// Force reconfiguration, even when complete configuration detected.
	reconf bool

	// The current charm form used to input missing required program settings.
	form *huh.Form

	// Whether the form was completed.
	formCompleted bool
}

// Init sends a kick-off tea command when needed.
func (m Model) Init() tea.Cmd {
	if m.form == nil {
		return bubbletree.ErrCmd(errors.New("m.form is nil, an initialization error occured"))
	}
	return nil
}

// Update is responsible for accepting a tea message passed down from the
// parent model and update the model data when appropriate.
func (m Model) Update(msg tea.Msg) (bubbletree.LeafModel, tea.Cmd) {
	var (
		cmds []tea.Cmd
		ok   bool
	)
	if m.IsDisabled() {
		return m, nil
	}

	switch msg := msg.(type) {
	// When focus requested, activate the configurator.
	case bubbletree.SetFocusMsg:
		if msg.IsRecipient(m.GetModelID()) {
			if !m.IsFocused() {
				if m.form == nil {
					// `Huh` form failed to initialize, wait for ErrMsg to come down.
					return m, nil
				}
				m.State = bubbletree.ActiveState
				m.LogStateChange(msg)

				cmds = append(cmds, getConfigCmd(m.Viper, m.reconf))
				m.LogAction(msg, "Requesting configuration")
			}
		}

	// The configuration file could not be found, read from user input form.
	case ConfigMissingMsg:
		if m.IsActive() {
			cmds = append(cmds, m.form.Init())
			m.LogAction(msg, "Requesting config form initialization")
		}

	// When the system is configured, call shutdown, we're done.
	case ConfigReadyMsg:
		if m.IsActive() {
			m.formCompleted = true
			m.LogNotice(msg, "Form completed")

			cmds = append(cmds, bubbletree.ShutDownCmd([]string{m.GetModelID()}))
			m.LogAction(msg, "Requesting model shutdown")
		}
	}

	// Run the default message handlers from bubbletree.
	leafModel, cmd := m.DefaultLeafModel.Update(msg)
	if m.DefaultLeafModel, ok = leafModel.(bubbletree.DefaultLeafModel); !ok {
		panic("DefaultLeafModel.Update didn't returned 'leadModel' as expected 'bubbletree.DefaultLeafModel' type")
	}
	cmds = append(cmds, cmd)

	// Run Huh Forms until we're done capturing config.
	if !m.formCompleted && m.IsActive() && m.IsFocused() {
		model, cmd := m.updateForm(msg)
		cmds = append(cmds, cmd)
		return model, tea.Batch(cmds...)
	}

	return m, tea.Batch(cmds...)
}

// View is the model's rendering routine that creates the output reflecting
// the current state of the model data. The rendered string is passed back up
// to the root model for final window composition.
func (m Model) View(w, h int) string {
	if m.IsFocused() && m.IsActive() {
		m.form.WithWidth(w - 2)
		return viewStyle.Render(strings.TrimSuffix(m.form.View(), "\n\n"))
	}
	return ""
}

// The general style of the finished view, before returning uptree.
var viewStyle = lipgloss.NewStyle().Margin(1)

// GetViewHeader returns the model's header view string.
func (m Model) GetViewHeader() string {
	return "Acquiring System Configuration"
}

// GetViewFooter returns the model's footer view string.
func (m Model) GetViewFooter() string {
	var s string
	if m.form == nil {
		return s
	}
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	if s == "" {
		s = m.form.Help().ShortHelpView(m.form.KeyBinds())
	}
	return s
}

// Options

// Option is used to set options for the new model at creation.
type Option func(*Model)

func WithLogger(logger *slog.Logger) Option {
	return func(m *Model) {
		m.Logger = logger
	}
}

func WithViper(viper *viper.Viper) Option {
	return func(m *Model) {
		m.Viper = viper
	}
}

func WithDisabled() Option {
	return func(m *Model) {
		m.Properties |= bubbletree.Disabled
	}
}

func WithReconfigure(force bool) Option {
	return func(m *Model) {
		m.reconf = force
	}
}

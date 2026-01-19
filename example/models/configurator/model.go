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

	m.Logger.Info("New model created", "ModelID", m.ID)
	return m
}

// Model is the global top-level definition of the current model.
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
	// The configuration file could not be found, read from user input form.
	case ConfigMissingMsg:
		if m.IsInactive() {
			m.State = bubbletree.ActiveState
			m.LogStateChange(msg)
		}

		m.formCompleted = false
		m.form = newForm(m.Viper)
		if m.form == nil {
			cmds = append(cmds, bubbletree.ErrCmd(errors.New("m.form is nil, an initialization error occured")))
			break
		}

		cmds = append(cmds, m.form.Init())
		m.LogAction(msg, "Requesting config form initialization")

	// When the system is configured, call shutdown, we're done.
	case ConfigReadyMsg:
		if m.IsActive() {
			m.formCompleted = true
			m.LogNotice(msg, "Form completed")
		}

	// When a configuration session is cancelled.
	case ConfigCancelMsg:
		if m.IsActive() {
			m.formCompleted = true
			m.LogNotice(msg, "Form cancelled")
		}
	}

	// Run the default message handlers from bubbletree.
	leafModel, cmd := m.DefaultLeafModel.Update(msg)
	if m.DefaultLeafModel, ok = leafModel.(bubbletree.DefaultLeafModel); !ok {
		panic("DefaultLeafModel.Update didn't returned 'leadModel' as expected 'bubbletree.DefaultLeafModel' type")
	}
	cmds = append(cmds, cmd)

	// Run Huh Forms until we're done capturing config.
	if m.form != nil && !m.formCompleted && m.IsActive() {
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
	if m.IsActive() {
		if m.form != nil && !m.formCompleted {
			m.form.WithWidth(w - 6) // tab + card + form margins.
			formView := strings.TrimSuffix(m.form.View(), "\n\n")

			if optProvider, ok := m.Theme.(bubbletree.OptionalStyleProvider); ok {
				return optProvider.GetCardStyle().Render(formView)
			}
			return m.Theme.GetBaseStyle().Render(formView)
		}
		return m.Theme.RenderNormalText("Settings Saved!")
	}
	return ""
}

// GetViewHeader returns the model's header view string.
func (m Model) GetViewHeader(w, h int) string {
	return m.Theme.RenderNormalText("Acquiring System Configuration")
}

// GetViewFooter returns the model's footer view string.
func (m Model) GetViewFooter(w, h int) string {
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

func WithTheme(theme bubbletree.Themer) Option {
	return func(m *Model) {
		m.Theme = theme
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

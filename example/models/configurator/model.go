// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package configurator

import (
	"context"
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

// const ()

var (
	// slog logger to use throughout the model.
	log *slog.Logger

	// unique model instance id.
	lastID atomic.Int64
)

// New creates and initializes a new model ready to be used.
func New(opts ...Option) bubbletree.LeafModel {
	ctx, cancel := context.WithCancel(context.Background())
	m := Model{
		Common: bubbletree.Common{
			ID:     nextID(), // Model's unique instance identifier
			Ctx:    ctx,      // Model's context used to control long running Cmd's
			Cancel: cancel,   // Model's context cancel func associated with Ctx
		},
	}

	for _, opt := range opts {
		opt(&m)
	}

	// Conveniently alias m.Logger to log.
	if m.Logger != nil {
		log = m.Logger
	} else {
		log = slog.Default() // Best effort
	}

	m.form = newForm(m.Viper)

	log.Info("New model created", "ModelID", m.ID)
	return m
}

func nextID() string {
	return fmt.Sprintf("configurator%d", lastID.Add(1))
}

// Model is the definition of the configurator model.
type Model struct {
	// Include all common fields of a CommonModel from the 'bubbletree' package.
	bubbletree.Common

	// The current charm form used to input missing required program settings.
	form *huh.Form

	// Whether the form was completed.
	formCompleted bool
}

// Init sends a kick-off tea command when needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update is responsible for accepting a tea message passed down from the
// parent model and update the model data when appropriate.
func (m Model) Update(msg tea.Msg) (bubbletree.LeafModel, tea.Cmd) {
	var cmds []tea.Cmd

	if m.IsDisabled() {
		return m, nil
	}

	switch msg := msg.(type) {
	// The configuration file could not be found, read from user input form.
	case ConfigMissingMsg:
		if m.IsInactive() {
			m.State = bubbletree.ActiveState
			cmds = append(cmds, m.form.Init())
			log.Info("model state change on message", "Msg", fmt.Sprintf("%T%+v", msg, msg), "ModelID", m.GetModelID(), "NewState", m.State)
			log.Info("model action on message", "Msg", fmt.Sprintf("%T%+v", msg, msg), "ModelID", m.GetModelID(), "Action", "Initialize config form")
		}
	// When the system is configured, activate the rest of the application.
	case ConfigReadyMsg:
		m.formCompleted = true
		cmds = append(cmds, bubbletree.ShutDownCmd([]string{m.GetModelID()}))
		log.Info("model change on message", "Msg", fmt.Sprintf("%T%+v", msg, msg), "ModelID", m.GetModelID(), "Form", "Completed")

	// General Model Msgs section

	// When disabled requested, accept and mark this model as disabled.
	case bubbletree.SetDisabledMsg:
		if msg.IsRecipient(m.GetModelID()) {
			if !m.IsDisabled() {
				m.Properties |= bubbletree.Disabled
				log.Info("property change on message", "Msg", fmt.Sprintf("%T%+v", msg, msg), "Adding Disabled", m.GetModelID())
			}
		} else {
			if m.IsDisabled() {
				m.Properties &= ^bubbletree.Disabled
				log.Info("property change on message", "Msg", fmt.Sprintf("%T%+v", msg, msg), "Removing Disabled", m.GetModelID())
			}
		}
	// When focus requested, accept and mark this model as focused.
	case bubbletree.SetFocusMsg:
		if msg.IsRecipient(m.GetModelID()) {
			if !m.IsFocused() {
				m.Properties |= bubbletree.Focused
				cmds = append(cmds, getConfigCmd(m.Viper))
				log.Info("property change on message", "Msg", fmt.Sprintf("%T%+v", msg, msg), "Adding Focus", m.GetModelID())
				log.Info("model action on message", "Msg", fmt.Sprintf("%T%+v", msg, msg), "ModelID", m.GetModelID(), "Action", "Getting config")
			}
		} else {
			if m.IsFocused() {
				m.Properties &= ^bubbletree.Focused
				log.Info("property change on message", "Msg", fmt.Sprintf("%T%+v", msg, msg), "Losing Focus", m.GetModelID())
			}
		}
	// ShuttingDownMsg means that the application is terminating: cleanup and inactivate.
	case bubbletree.ShutDownMsg:
		if msg.IsRecipient(m.GetModelID()) && !m.IsShuttingDown() {
			m.State = bubbletree.ShuttingDownState
			m.CancelContext()
			cmds = append(cmds, bubbletree.ModelFinishedCmd(m.GetModelID()))
			log.Info("model state change on message", "Msg", fmt.Sprintf("%T%+v", msg, msg), "ModelID", m.GetModelID(), "NewState", m.State)
		}
	// ModelFinishedMsg marks the end-of-life for the model instance.
	case bubbletree.ModelFinishedMsg:
		if msg.IsRecipient(m.GetModelID()) && !m.IsFinished() {
			m.State = bubbletree.FinishedState
			log.Info("model state change on message", "Msg", fmt.Sprintf("%T%+v", msg, msg), "ModelID", m.GetModelID(), "NewState", m.State)
		}
	}

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
		return lipgloss.NewStyle().Margin(0, 1).
			Render(strings.TrimSuffix(m.form.View(), "\n\n"))
	}
	return ""
}

// GetViewHeader returns the model's header view string.
func (m Model) GetViewHeader() string {
	return "Acquiring System Configuration"
}

// GetViewFooter returns the model's footer view string.
func (m Model) GetViewFooter() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	if s == "" {
		s = m.form.Help().ShortHelpView(m.form.KeyBinds())
	}
	return s
}

// CancelContext calls the cancel function on the model's context.
func (m Model) CancelContext() {
	log.Info("Cancelling model's context")
	if m.Cancel != nil {
		m.Cancel()
	} else {
		log.Warn("The model's cancel function is unexpectedly nil, context cannot be cancelled.")
	}
}

// GetModelID returns the name or ID of the model's instance. It is used by
// the parent models to organize and keep track of child model instances and
// to identify instances currently active, among other things..
func (m Model) GetModelID() string {
	return m.ID
}

// GetState returns the current model's instance state.
func (m Model) GetState() bubbletree.State {
	return m.State
}

// IsActive returns whether the current model is in Active state.
func (m Model) IsActive() bool {
	return m.State == bubbletree.ActiveState
}

// IsInactive returns whether the current model is in Inactive state.
func (m Model) IsInactive() bool {
	return m.State == bubbletree.InactiveState
}

// IsShuttingDown returns whether the current model is still shutting down
// or not.
func (m Model) IsShuttingDown() bool {
	return m.State == bubbletree.ShuttingDownState
}

// IsFinished returns whether the current model has completed its shutdown
// sequence: its goroutines returned, and its resources are fully freed.
func (m Model) IsFinished() bool {
	return m.State == bubbletree.FinishedState
}

// GetProperties returns the current model's instance properties.
func (m Model) GetProperties() bubbletree.Properties {
	return m.Properties
}

// IsDisabled returns whether the current model is in Disabled state.
func (m Model) IsDisabled() bool {
	return m.Properties&bubbletree.Disabled != 0
}

// IsFocused returns whether the current model is in focus.
func (m Model) IsFocused() bool {
	return m.Properties&bubbletree.Focused != 0
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

// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package root

import (
	"fmt"
	"log/slog"
	"time"

	"example/models/configurator"
	"example/models/coreapp"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
	"github.com/yhcote/bubbletree"
)

const (
	// The maximum amount of time the combination of running Update() +
	// View() before emitting a warning. If the Update() + View() cycle
	// becomes too long, the system stop being responsive and screen
	// updates may start appearing choppy. Try to keep each iteration
	// faster than 60fps (16.666 mSec)
	maxIterationTime = 16 * time.Millisecond
)

var (
	// slog logger to use throughout the model.
	log *slog.Logger
)

// New creates the root model of a bubbletea program. It's usually
// called just before tea.NewProgram. Options can be passed in with opts, to
// initialize the program with command line arguments, and possibly configs,
// communicated to the rest of the program and child models.
func New(opts ...Option) bubbletree.RootModel {
	var m Model

	for _, opt := range opts {
		opt(&m)
	}

	// Conveniently alias m.Logger to log.
	if m.logger != nil {
		log = m.logger
	} else {
		log = slog.Default() // Best effort
	}

	// When no optional configuration state has been requested,
	// use a default one.
	if m.spewcfg == nil {
		m.spewcfg = spew.NewDefaultConfig()
	}

	// Create all child models
	m.configuratorModel = configurator.New(configurator.WithLogger(m.logger), configurator.WithViper(m.viper))
	m.coreappModel = coreapp.New(coreapp.WithLogger(m.logger), coreapp.WithViper(m.viper))

	return m
}

// Model is the application's root Model definition. Child models are embeded
// here so that the entire application's information is available from this
// base structure.
type Model struct {
	// Last program recorded error.
	err error

	// Terminal window size is set and we can start generating views.
	ready bool

	// Ending program cleanup happened, bubble tea is quitting.
	quitting bool

	// Options passed in by the main program.
	BaseOpts

	// The currently focused child model id or "" for base model.
	focused string

	// The terminal window information, dimensions.
	screen

	// The update and view methods execution timing.
	timing

	// Direct child models.
	configuratorModel bubbletree.LeafModel
	coreappModel      bubbletree.BranchModel
}

// screen holds the current state details of the terminal window.
type screen struct {
	width  int
	height int
}

// timing holds information about the update and view execution performance
// in terms of elapsed time.
type timing struct {
	updateRuntime time.Duration
	viewRuntime   time.Duration
}

// Init sends a kick-off tea command when needed.
func (m Model) Init() tea.Cmd {
	// Start the application in configuration state
	return tea.Batch(
		m.configuratorModel.Init(),
		m.coreappModel.Init(),
		bubbletree.SetFocusCmd(m.configuratorModel.GetModelID()),
	)
}

// Update gathers messages for base and child models, and routes them down the
// model tree where they will be handled by appropriate component(s).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	t1 := time.Now()
	switch msg := msg.(type) {
	// Check base model interpreted pressed keys
	case tea.KeyMsg:
		log.Debug("Update()", "message", m.spewcfg.Sprintf("%#+v", msg))
		switch msg.String() {
		// Key pressed meaning exit.
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	// Root model sent a set focus message, record it
	case bubbletree.SetFocusMsg:
		m.focused = msg.ModelID
	// The terminal window has been changed.
	case tea.WindowSizeMsg:
		log.Debug("Update()", "message", m.spewcfg.Sprintf("%#+v", msg))
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.ready = true
		}
	// Configurator workflow success, the app is now setup and configured.
	case configurator.ConfigReadyMsg:
		log.Debug("Update()", "message", m.spewcfg.Sprintf("%#+v", msg))
		cmds = append(cmds, bubbletree.SetFocusCmd(m.coreappModel.GetModelID()))
	// CoreApp workflow success, the app is now loaded and enters its main workflow.
	case bubbletree.ModelFinishedMsg:
		log.Debug("Update()", "message", m.spewcfg.Sprintf("%#+v", msg))
		if msg.IsRecipient(m.coreappModel.GetModelID()) {
			cmds = append(cmds, bubbletree.SetFocusCmd(""))
		}
	// A model encountered an error, decode and treat the error
	case bubbletree.ErrMsg:
		log.Error("an error was returned by a child model", "err", fmt.Sprintf("%+v", msg.Err))
		m.quitting = true
		return m, tea.Quit
	}

	// Propagate current message to child models, so they call Update(msg).
	rm, cmd := m.UpdateNodeModels(msg)
	m = rm.(Model)
	cmds = append(cmds, cmd)

	// Record event loop elapsed time. Keep the main loop snappy.
	m.updateRuntime = time.Since(t1)

	// Send message to all child components, then return new batched commands.
	return m, tea.Batch(cmds...)
}

// View reaches all model children with the updated model and gathers each
// component's renderered subviews. Those are then composed into a base view
// that will be the final application view being displayed.
func (m Model) View() string {
	var view string

	// Waiting for minimal setup (e.g.: initial term window resize)
	if !m.ready {
		return ""
	}

	if m.quitting {
		return renderQuittingView(m.err, m.width, m.height)
	}

	t1 := time.Now()
	switch m.focused {
	// Show the config form view for settings input
	case m.configuratorModel.GetModelID():
		view = renderConfigView(m.configuratorModel, m.width, m.height)
	// Show the main application view
	case m.coreappModel.GetModelID():
		view = renderCoreAppView(m.coreappModel, m.width, m.height)
	// Show a base placeholder when no other models are in focus
	default:
		view = renderBaseView(m.width, m.height)
	}

	// Record main view rendering elapsed time. Keep the rendering fast.
	m.viewRuntime = time.Since(t1)
	if m.updateRuntime+m.viewRuntime > maxIterationTime {
		log.Warn("update/view iteration took too long",
			"update", m.updateRuntime.Round(time.Microsecond).String(),
			"view", m.viewRuntime.Round(time.Microsecond).String(),
		)
	}
	m.updateRuntime = 0
	m.viewRuntime = 0

	return view
}

// UpdateNodeModels routes a tea.Msg to all registered child component
// models. Each component's Update() routine is called with 'msg'.
func (m Model) UpdateNodeModels(msg tea.Msg) (bubbletree.RootModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.configuratorModel, cmd = m.configuratorModel.Update(msg)
	cmds = append(cmds, cmd)

	m.coreappModel, cmd = m.coreappModel.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// Msg/Cmd's

// Options

// Option is used to set options on the base model of the application.
type Option func(*Model)

// BaseOpts describes general configurations or states of the base model
// for the application.
type BaseOpts struct {
	logger  *slog.Logger
	viper   *viper.Viper
	spewcfg *spew.ConfigState
}

// WithLogger sets the logger to use for model logging.
func WithLogger(logger *slog.Logger) Option {
	return func(m *Model) {
		m.logger = logger
	}
}

// WithViper sets the active viper config to use in the model.
func WithViper(viper *viper.Viper) Option {
	return func(m *Model) {
		m.viper = viper
	}
}

// WithSpewConfigState sets spew utility configuration state instance to
// a specified one.
func WithSpewConfigState(spewcfg *spew.ConfigState) Option {
	return func(m *Model) {
		m.spewcfg = spewcfg
	}
}

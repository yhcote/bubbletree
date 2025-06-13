// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package root

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"example/internal/app"
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

// New creates the root model of a bubbletea program. It's usually
// called just before tea.NewProgram. Options can be passed in with opts, to
// initialize the program with command line arguments, and possibly configs,
// communicated to the rest of the program and child models.
func New(opts ...Option) (bubbletree.RootModel, error) {
	var m Model
	for _, opt := range opts {
		opt(&m)
	}
	if m.logger == nil {
		m.logger = slog.Default()
	}
	if m.configViper == nil {
		return nil, fmt.Errorf("configuration through 'viper' is expected, pass the 'WithConfigViper' option")
	}
	if m.spewcfg == nil {
		m.spewcfg = spew.NewDefaultConfig()
	}

	// Create and link all descendant models used in the application.
	m.Models = new(sync.Map)
	var model bubbletree.CommonModel

	model = configurator.New(configurator.WithLogger(m.logger), configurator.WithViper(m.configViper))
	m.LinkNewModel(model, &m.modelConfigID)

	model = coreapp.New(coreapp.WithLogger(m.logger), coreapp.WithViper(m.configViper))
	m.LinkNewModel(model, &m.modelCoreappID)

	return m, nil
}

// Model is the application's root Model definition. Child models are embeded
// here so that the entire application's information is available from this
// base structure.
type Model struct {
	// Include fields and default methods of bubbletree.DefaultRootModel.
	bubbletree.DefaultRootModel

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

	// Direct child models saved IDs for direct and easy access.
	modelConfigID  string
	modelCoreappID string
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
	cmds := []tea.Cmd{
		// Set the terminal window title.
		tea.SetWindowTitle(fmt.Sprintf("%s  ver: %s", app.ProgramName, app.ProgramVersion)),

		// Start the application in configuration state
		bubbletree.SetFocusCmd(m.modelConfigID),
	}
	// Run all descendant's Init() routine and collect their returned tea Cmds.
	m.Models.Range(func(key, value any) bool {
		if model, ok := value.(bubbletree.CommonModel); ok {
			cmds = append(cmds, model.Init())
		} else {
			panic("stored model isn't of bubbletree.CommonModel")
		}
		return true
	})
	return tea.Batch(cmds...)
}

// Update gathers messages for base and child models, and routes them down the
// model tree where they will be handled by appropriate component(s).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	t1 := time.Now()
	switch msg := msg.(type) {
	// Check base model interpreted pressed keys
	case tea.KeyMsg:
		m.logger.Debug("Update()", "message", m.spewcfg.Sprintf("%#+v", msg))
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
		m.logger.Debug("Update()", "message", m.spewcfg.Sprintf("%#+v", msg))
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.ready = true
		}

	// Configurator workflow success, the app is now setup and configured.
	case configurator.ConfigReadyMsg:
		cmds = append(cmds, bubbletree.SetFocusCmd(m.modelCoreappID))

	// CoreApp workflow success, the app is now loaded and enters its main workflow.
	case bubbletree.ModelFinishedMsg:
		if msg.IsRecipient(m.modelCoreappID) {
			cmds = append(cmds, bubbletree.SetFocusCmd(""))
		}

	// A model encountered an error, decode and treat the error
	case bubbletree.ErrMsg:
		m.Err = fmt.Errorf("received model error message: %w", msg.Err)
		m.logger.Error("terminating root model", "error", m.Err)
		m.quitting = true
		return m, tea.Quit
	}

	// Propagate current message to child models, so they call Update(msg).
	cmds = append(cmds, m.UpdateNodeModels(msg))

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
		return renderQuittingView(m.Err, m.width, m.height)
	}

	t1 := time.Now()
	switch m.focused {
	// Show the config form view for settings input
	case m.modelConfigID:
		view = renderConfigView(m.MustGetModel(m.modelConfigID), m.width, m.height)

	// Show the main application view
	case m.modelCoreappID:
		view = renderCoreAppView(m.MustGetModel(m.modelCoreappID), m.width, m.height)

	// Show a base placeholder when no other models are in focus
	default:
		view = renderBaseView(m.width, m.height)
	}

	// Record main view rendering elapsed time. Keep the rendering fast.
	m.viewRuntime = time.Since(t1)
	if m.updateRuntime+m.viewRuntime > maxIterationTime {
		m.logger.Warn("update/view iteration took too long",
			"update", m.updateRuntime.Round(time.Microsecond).String(),
			"view", m.viewRuntime.Round(time.Microsecond).String(),
		)
	}
	m.updateRuntime = 0
	m.viewRuntime = 0

	return view
}

// Msg/Cmd's

// Options

// Option is used to set options on the base model of the application.
type Option func(*Model)

// BaseOpts describes general configurations or states of the base model
// for the application.
type BaseOpts struct {
	logger      *slog.Logger
	configViper *viper.Viper
	spewcfg     *spew.ConfigState
}

// WithLogger sets the logger to use for model logging.
func WithLogger(logger *slog.Logger) Option {
	return func(m *Model) {
		m.logger = logger
	}
}

// WithConfigViper sets the active viper config to use in the model.
func WithConfigViper(viper *viper.Viper) Option {
	return func(m *Model) {
		m.configViper = viper
	}
}

// WithSpewConfigState sets spew utility configuration state instance to
// a specified one.
func WithSpewConfigState(spewcfg *spew.ConfigState) Option {
	return func(m *Model) {
		m.spewcfg = spewcfg
	}
}

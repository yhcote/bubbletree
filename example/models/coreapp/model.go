// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package coreapp

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"example/models/configurator"
	"example/ui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davecgh/go-spew/spew"
	"github.com/yhcote/bubbletree"
	"github.com/yhcote/bubbletree/logger"
)

const (
	// short model name used for identification.
	modelName = "coreapp"
)

var (
	// unique model instance id based on 'modelName'.
	lastID atomic.Int64
)

// Run creates and initializes a new model ready to be used.
func Run(opts ...bubbletree.AppOption) error {
	ctx, cancel := context.WithCancel(context.Background())
	m := Model{
		DefaultAppModel: bubbletree.DefaultAppModel{
			DefaultBranchModel: bubbletree.DefaultBranchModel{
				DefaultCommonModel: bubbletree.DefaultCommonModel{
					ID:     fmt.Sprintf("%s-%d", modelName, lastID.Add(1)),
					Ctx:    ctx,
					Cancel: cancel,
				},
			},
		},
	}
	for _, opt := range opts {
		opt(&m.DefaultAppModel)
	}
	if m.OptProgname == "" {
		m.OptProgname = "nameUnset"
	}
	if m.OptProgver == "" {
		m.OptProgver = "versionUnset"
	}
	if m.OptLogger == nil {
		m.OptLogger = logger.Log()
	}
	if m.OptConfigViper == nil {
		return fmt.Errorf("configuration through 'viper' is expected, pass the 'WithConfigViper' option")
	}
	if m.OptSpewcfg == nil {
		m.OptSpewcfg = spew.NewDefaultConfig()
	}
	if m.OptTheme == nil {
		m.OptTheme = bubbletree.DefaultMinimalTheme()
	}
	m.Logger = m.OptLogger
	m.Viper = m.OptConfigViper
	m.Theme = m.OptTheme

	// Create and link all descendant models used in the application.
	m.Models = new(sync.Map)
	var model bubbletree.CommonModel

	// Add the Configurator Model.
	model = configurator.New(
		configurator.WithLogger(m.Logger),
		configurator.WithViper(m.Viper),
		configurator.WithTheme(m.Theme),
		configurator.WithReconfigure(m.OptReconf),
	)
	m.LinkNewModel(model, &m.modelConfigID)

	// focusedID to self to cover View invocations during bootstrap time.
	m.focusedID = m.ID

	m.Logger.Info("New model created", "ModelID", m.ID)

	// Create a root model shim to the bubble tea framework and start the
	// event loop engine.
	if teaProgram, err := tea.NewProgram(bubbletree.New(m), tea.WithAltScreen()).Run(); err != nil {
		return err
	} else if teaProgram.(bubbletree.RootModel).LastError() != nil {
		return teaProgram.(bubbletree.RootModel).LastError()
	}

	return nil
}

// Model is the definition of the coreapp model.
type Model struct {
	// Include fields and default methods of bubbletree.DefaultAppModel.
	bubbletree.DefaultAppModel

	// Direct child models saved IDs for direct and easy access.
	modelConfigID string

	// Focused Model, receives console input events (keyboard/mouse)
	focusedID string

	// UI related variables.
	topbar    *components.Winbar
	tabber    *components.Tabs
	bottombar *components.Winbar
}

// Init sends a kick-off tea command when needed.
func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	cmd := tea.SetWindowTitle(fmt.Sprintf("%s  ver: %s", m.OptProgname, m.OptProgver))
	cmds = append(cmds, cmd)

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

// Update is responsible for accepting a tea message passed down from the
// parent model and update the model data when appropriate.
func (m Model) Update(msg tea.Msg) (bubbletree.BranchModel, tea.Cmd) {
	var (
		cmds []tea.Cmd
		ok   bool
	)
	if m.IsDisabled() {
		return m, nil
	}

	t1 := time.Now()
	switch msg := msg.(type) {
	// Console input is usually passed down to the model it as identified
	// as in focus. Special keys to quit or Function Keys to switch model
	// tabs are handled here and not passed down to descendant models.
	case tea.KeyMsg:
		m.Logger.Debug("Message", "tea.KeyMsg", m.OptSpewcfg.Sprintf("%#+v", msg))
		switch msg.String() {
		case "ctrl+c", "esc":
			m.LogAction(msg, "Requesting Model Finished")
			return m, bubbletree.ModelFinishedCmd(m.ID)
		case "f1":
			if m.focusedID != m.ID {
				// If we starting a config session (prior f2), end it.
				if m.focusedID == m.modelConfigID {
					cmds = append(cmds,
						configurator.CancelConfigCmd(),
					)
					m.LogAction(msg, "Requesting configuration cancellation")
				}

				m.tabber.SetActiveTab(0)
				m.focusedID = m.ID
			}
		case "f2":
			if m.focusedID != m.modelConfigID {
				m.tabber.SetActiveTab(1)
				m.focusedID = m.modelConfigID
				cmds = append(cmds,
					configurator.GetConfigCmd(m.Viper, true),
				)
				m.LogAction(msg, "Requesting (forced) configuration")
			}
		case "f3":
			if m.focusedID != m.ID {
				// If we starting a config session (prior f2), end it.
				if m.focusedID == m.modelConfigID {
					cmds = append(cmds,
						configurator.CancelConfigCmd(),
					)
					m.LogAction(msg, "Requesting configuration cancellation")
				}

				m.tabber.SetActiveTab(2)
				m.focusedID = m.ID // Set to self (coreapp), for yet to be handled tabs.
			}
		default:
			// If NOT Self (Coreapp)...
			if m.focusedID != m.ID {
				// Get the focused model, give it access to input events.
				cmds = append(cmds, m.UpdateNodeModel(m.MustGetModel(m.focusedID), msg))
				return m, tea.Batch(cmds...)
			}
		}

	// Tea always sends at least one WindowSizeMsg at startup, use this
	// signal to activate the core application and initialize the initial
	// UI and system configuration.
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		m.Logger.Debug("Message", "tea.WindowSizeMsg", m.OptSpewcfg.Sprintf("%#+v", msg))
		if m.IsInactive() {
			m.State = bubbletree.ActiveState
			m.LogStateChange(msg)

			// Initialize the window top bar.
			m.topbar = components.NewWinbar(m.Theme, true, 0, 0)

			// Initialize the root tabber pane.
			m.tabber = components.NewTabs(m.Theme, []components.TabTitle{
				{Name: "Dashboard", ShortcutKey: "f1"},
				{Name: "Settings", ShortcutKey: "f2"},
				{Name: "Logs", ShortcutKey: "f3"},
			}, 0, 0)

			// Initialize the window bottom bar.
			m.bottombar = components.NewWinbar(m.Theme, false, 0, 0)

			// Switch to main view
			m.focusedID = m.ID

			// Get the system config.
			cmds = append(cmds, configurator.GetConfigCmd(m.Viper, m.OptReconf))
			m.LogAction(msg, "Requesting configuration")
		}

	// Configurator needs input from the user for a missing configuration.
	case configurator.ConfigMissingMsg:
		if m.IsActive() {
			m.tabber.SetActiveTab(1)
			m.focusedID = m.modelConfigID

			m.LogAction(msg, "Requesting switch to configuration tab")
		}
	}

	// Run the default message handlers from bubbletree.
	branchModel, cmd := m.DefaultAppModel.Update(msg)
	if m.DefaultBranchModel, ok = branchModel.(bubbletree.DefaultBranchModel); !ok {
		panic("DefaultBranchModel.Update didn't returned 'branchModel' as expected 'bubbletree.DefaultBranchModel' type")
	}
	cmds = append(cmds, cmd)

	// Propagate current message to child models, so they call Update(msg).
	cmds = append(cmds, m.UpdateNodeModels(msg))

	// Save event loop elapsed time. Keep the main loop snappy.
	m.UpdateRuntime = time.Since(t1)

	return m, tea.Batch(cmds...)
}

// AppView implements the appnode interface method allowing the rootnode to
// call the application's view method, passing whether the application is
// quitting or not. When quitting an optional program error is also passed
// along.
func (m Model) AppView(quitting bool, err error) string {
	if quitting {
		if err != nil {
			m.Logger.Error("terminating program", "error", err)
		}
		return m.QuittingView(err)
	}
	return m.View(m.Width, m.Height)
}

const (
	// The maximum amount of time the combination of running Update() +
	// View() before emitting a warning. If the Update() + View() cycle
	// becomes too long, the system stop being responsive and screen
	// updates may start appearing choppy. Try to keep each iteration
	// faster than 60fps (16.666 mSec)
	maxIterationTime = 16 * time.Millisecond
)

// View is the model's rendering routine that creates the output reflecting
// the current state of the model data. The rendered string is passed back up
// to the root model for sending to the display.
func (m Model) View(maxWidth, maxHeight int) string {
	if m.IsInactive() {
		return ""
	}

	t1 := time.Now()
	view := m.renderNormalWindow(m.Width, m.Height)
	m.ViewRuntime = time.Since(t1)

	// One cycle passed, check Update+View timing for adequate performance.
	if m.UpdateRuntime+m.ViewRuntime > maxIterationTime {
		m.Logger.Warn("update/view iteration took too long",
			"update", m.UpdateRuntime.Round(time.Microsecond).String(),
			"view", m.ViewRuntime.Round(time.Microsecond).String(),
		)
	}
	m.UpdateRuntime = 0
	m.ViewRuntime = 0

	return view
}

// QuittingView displays the last window, whether the application exits
// normally or with errors.
func (m Model) QuittingView(err error) string {
	if m.IsInactive() {
		return ""
	}
	return m.renderQuittingWindow(err, m.Width, m.Height)
}

// Copyright 2023-2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package bubbletree

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// RootModel defines the root node of the bubbletree framework. Only one
// model should implement this interface in a program. Only this model has
// its Init, Update, and View methods called directly by bubble tea. In
// bubbletree, the root model instantiates and registers its direct
// descendants and is responsible for calling Init/Update/View as appropriate.
// The idea is that tea.Msgs are propagated down the model tree. Those
// messages will generate new tea.Cmds from all the tree components, which
// will be flowing back up to the root model. Those tree gathered commands
// are returned to bubble tea for processing. Similarly, at View execution
// time, the root model will have all the nodes from the tree run their
// View() method and the result will be aggregated by the root model View()
// method for final UI composition.
type RootModel interface {
	// tea.Model is the Charmbracelet bubble tea main Init/Update/View interface
	// that connects the program to the bubble tea framework.
	tea.Model

	// LastError returns the last error saved in the model instance.
	LastError() error
}

// New returns a new DefaultRootModel instance.
func New(app AppModel) RootModel {
	return &DefaultRootModel{
		CoreApp: app,
	}
}

// DefaultRootModel implements default methods for the RootModel interface.
// It can be used as a base implementation when no specific code is required.
// The implementor of the RootModel interface may overwrite the default
// behavior of the model by reimplementing desired methods.
type DefaultRootModel struct {
	// The directly linked main application model.
	CoreApp AppModel

	// Ending program cleanup happened, bubble tea is quitting.
	Quitting bool

	// The last error recorded in the model.
	Err error
}

// Init is the default implementation of the RootModel interface.
func (m DefaultRootModel) Init() tea.Cmd {
	return m.CoreApp.Init()
}

// Update is the default implementation of the RootModel interface.
func (m DefaultRootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Check if the core application sends a program exit signal.
	case ModelFinishedMsg:
		if msg.IsRecipient(m.CoreApp.GetModelID()) {
			m.Quitting = true
			return m, tea.Quit
		}

	// A model encountered an error, decode and treat the error
	case ErrMsg:
		m.Err = fmt.Errorf("error message received from model tree: %w", msg.Err)
		m.Quitting = true
		return m, tea.Quit
	}

	// Propagate current message to CoreApp's Update(msg).
	branchModel, cmd := m.CoreApp.Update(msg)
	m.CoreApp = branchModel.(AppModel)

	// Return model tree gathered new commands from descendant models.
	return m, cmd
}

// View is the default implementation of the RootModel interface.
func (m DefaultRootModel) View() string {
	return m.CoreApp.AppView(m.Quitting, m.Err)
}

// LastError returns the last error recorded by the root model.
func (m DefaultRootModel) LastError() error {
	return m.Err
}

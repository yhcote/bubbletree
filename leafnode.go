// Copyright 2023-2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package bubbletree

import (
	tea "github.com/charmbracelet/bubbletea"
)

// LeafModel defines a leaf (outer) node in the bubbletree framework. These
// models don't have child nodes, implementing all the logic within the
// bounds of a single model. Its Init/Update/View methods are called by its
// parent branch model. LeafModels act like a regular bubble tea "bubble".
type LeafModel interface {
	// CommonModel embeds the common model interface that a leaf model should
	// implement first.
	CommonModel

	// Update is responsible for accepting a tea message passed down from the
	// parent model and updating the model data when appropriate.
	Update(msg tea.Msg) (LeafModel, tea.Cmd)
}

// DefaultLeafModel implements default methods for the LeafModel interface.
// It can be used as a base implementation when no specific code is required.
// The implementor of the LeafModel interface may overwrite the default
// behavior of the model by reimplementing desired methods.
type DefaultLeafModel struct {
	// Include all common fields and methods of the default CommonModel
	// interface implementation.
	DefaultCommonModel
}

// Update is the default implementation of the LeafModel interface. It is the
// update logic in reaction to the new message passed as parameter. The
// updated copy of the model is returned along with optional new tea commands.
// Note that a model can implement its own handling of specific messages and
// then call this general session message handling code, in combination, via
// an embeded DefaultLeafModel instead of completely overwriting the method.
func (m DefaultLeafModel) Update(msg tea.Msg) (LeafModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// When disabled requested, accept and mark this model as disabled.
	case SetDisabledMsg:
		if msg.IsRecipient(m.GetModelID()) {
			if !m.IsDisabled() {
				old, new := m.Properties.SetDisabled()
				m.LogPropertyChange(msg, old, new)
			}
		} else {
			if m.IsDisabled() {
				old, new := m.Properties.UnsetDisabled()
				m.LogPropertyChange(msg, old, new)
			}
		}

	// When focus requested, accept and mark this model as focused.
	case SetFocusMsg:
		if msg.IsRecipient(m.GetModelID()) {
			if !m.IsFocused() {
				old, new := m.Properties.SetFocused()
				m.LogPropertyChange(msg, old, new)
			}
		} else {
			if m.IsFocused() {
				old, new := m.Properties.UnsetFocused()
				m.LogPropertyChange(msg, old, new)
			}
		}

	// ShuttingDownMsg means that the application is terminating: cleanup and inactivate.
	case ShutDownMsg:
		if msg.IsRecipient(m.GetModelID()) && !m.IsShuttingDown() {
			m.State = ShuttingDownState
			m.LogStateChange(msg)

			m.CancelContext()
			cmds = append(cmds, ModelFinishedCmd(m.GetModelID()))
			m.LogAction(msg, "Requesting model finished")
		}

	// ModelFinishedMsg marks the end-of-life for the model instance.
	case ModelFinishedMsg:
		if msg.IsRecipient(m.GetModelID()) && !m.IsFinished() {
			m.State = FinishedState
			m.LogStateChange(msg)
		}
	}

	return m, tea.Batch(cmds...)
}

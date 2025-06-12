// Copyright 2023-2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package bubbletree

import (
	"fmt"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// BranchModel defines a branch (inner) node in the bubbletree framework.
// These models have child models that are used to divide the logic of the
// model. This is useful for the implementation of large modules or for
// models defining user interfaces comprising of many components. A branch
// model node has tasks similar to the root model although it doesn't
// communicate directly with the bubble tea framework. It is responsible to
// instantiate and register its direct children and run the Init/Update/View
// methods of its children and passing the results back up the tree.
type BranchModel interface {
	// CommonModel embeds the common model interface that a branch model should
	// implement first.
	CommonModel

	// Update is responsible for accepting a tea message passed down from the
	// parent model and updating the model data when appropriate.
	Update(msg tea.Msg) (BranchModel, tea.Cmd)

	// UpdateNodeModels is responsible to call the Update method on all child models
	// that the current model has registered. It also gathers the child models new
	// commands generated and returns them alongside the updated branch model.
	UpdateNodeModels(msg tea.Msg) tea.Cmd
}

// DefaultBranchModel implements default methods for the BranchModel interface.
// It can be used as a base implementation when no specific code is required.
// The implementor of the BranchModel interface may overwrite the default
// behavior of the model by reimplementing desired methods.
type DefaultBranchModel struct {
	// Include all common fields and methods of the default CommonModel
	// interface implementation.
	DefaultCommonModel

	// The <ModelID, *Model> map of all registered descendant models.
	Models *sync.Map
}

// Update is the default implementation of the BranchModel interface. It is the
// update logic in reaction to the new message passed as parameter. The
// updated copy of the model is returned along with optional new tea commands.
// Note that a model can implement its own handling of specific messages and
// then call this general session message handling code, in combination, via
// an embeded DefaultBranchModel instead of completely overwriting the method.
func (m DefaultBranchModel) Update(msg tea.Msg) (BranchModel, tea.Cmd) {
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

// UpdateNodeModels is the default implementation of the BranchModel interface.
func (m DefaultBranchModel) UpdateNodeModels(msg tea.Msg) tea.Cmd {
	var (
		cmds  []tea.Cmd
		wg    sync.WaitGroup
		cchan = make(chan tea.Cmd)
	)

	m.Models.Range(func(key, value any) bool {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var cmd tea.Cmd
			if model, ok := value.(BranchModel); ok {
				model, cmd = model.Update(msg)
				m.Models.Store(model.GetModelID(), model)
				cchan <- cmd
			} else if model, ok := value.(LeafModel); ok {
				model, cmd = model.Update(msg)
				m.Models.Store(model.GetModelID(), model)
				cchan <- cmd
			} else {
				panic("current model doesn't implement a branch or a leaf model")
			}
		}()
		return true
	})

	// Wait for all goroutines, then close all channels while the main function,
	// just below, collects data.
	go func() {
		wg.Wait()
		close(cchan)
	}()

	for cmd := range cchan {
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// LinkNewModel takes a new descendant model and updates the model ID saved
// by the model for later reference in addition to adding that new model to a
// map of descendant models.
func (m DefaultBranchModel) LinkNewModel(model CommonModel, modelID *string) {
	*modelID = model.GetModelID()
	m.Models.Store(model.GetModelID(), model)
}

// MustGetModel returns a model identified by the parameter modelID from the
// linked model sync.Map. The function will panic if the model cannot be found
// or if the found object doesn't implement the 'CommonModel' interface.
func (m DefaultBranchModel) MustGetModel(modelID string) CommonModel {
	if model, ok := m.Models.Load(modelID); !ok {
		panic(fmt.Sprintf("cannot load '%s' model from map", modelID))
	} else {
		if model, ok := any(model).(CommonModel); !ok {
			panic(fmt.Sprintf("loaded '%s' model doesn't implement CommonModel", modelID))
		} else {
			return model
		}
	}
}

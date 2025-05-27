// Copyright 2023-2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package bubbletree

import (
	"sync"

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

	// UpdateNodeModels is responsible to call the Update method on all child models
	// that the root model has registered. It also gathers the child models new
	// commands generated and returns them.
	UpdateNodeModels(msg tea.Msg) tea.Cmd
}

// DefaultRootModel implements default methods for the RootModel interface.
// It can be used as a base implementation when no specific code is required.
// The implementor of the RootModel interface may overwrite the default
// behavior of the model by reimplementing desired methods.
type DefaultRootModel struct {
	// The <ModelID, *Model> map of all registered descendant models.
	Models *sync.Map
}

// UpdateNodeModels is the default implementation of the RootModel interface.
// It runs the update method of all descendants models found in the 'Models'
// sync.Map (decribed above), concurrently.
func (m DefaultRootModel) UpdateNodeModels(msg tea.Msg) tea.Cmd {
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

// Init is the default implementation of the RootModel interface.
func (m DefaultRootModel) Init() tea.Cmd {
	return nil
}

// Update is the default implementation of the RootModel interface.
func (m DefaultRootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return nil, nil
}

// View is the default implementation of the RootModel interface.
func (m DefaultRootModel) View() string {
	return ""
}

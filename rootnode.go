// Copyright 2023-2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package bubbletree

import (
	"fmt"
	"log/slog"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
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

	// LastError returns the last error saved in the model instance.
	LastError() error
}

// DefaultRootModel implements default methods for the RootModel interface.
// It can be used as a base implementation when no specific code is required.
// The implementor of the RootModel interface may overwrite the default
// behavior of the model by reimplementing desired methods.
type DefaultRootModel struct {
	// Options passed in by the main program.
	RootOpts

	// The last error recorded in the model.
	Err error

	// The <ModelID, *Model> map of all registered descendant models.
	Models *sync.Map
}

// RootOpts describes general configurations or states of the root model
// for the application.
type RootOpts struct {
	Logger      *slog.Logger
	ConfigViper *viper.Viper
	Spewcfg     *spew.ConfigState
	Reconf      bool
}

// RootOption is used to set options on the base model of the application.
type RootOption func(*DefaultRootModel)

// WithLogger sets the logger to use for model logging.
func WithLogger(logger *slog.Logger) RootOption {
	return func(m *DefaultRootModel) {
		m.Logger = logger
	}
}

// WithConfigViper sets the active viper config to use in the model.
func WithConfigViper(viper *viper.Viper) RootOption {
	return func(m *DefaultRootModel) {
		m.ConfigViper = viper
	}
}

// WithSpewConfigState sets spew utility configuration state instance to
// a specified one.
func WithSpewConfigState(spewcfg *spew.ConfigState) RootOption {
	return func(m *DefaultRootModel) {
		m.Spewcfg = spewcfg
	}
}

// WithReconfigure ignores a complete config file and brings up the user input
// form, enabling configuration changes.
func WithReconfigure(force bool) RootOption {
	return func(m *DefaultRootModel) {
		m.Reconf = force
	}
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

// LastError returns the last error recorded by the root model.
func (m DefaultRootModel) LastError() error {
	return m.Err
}

// LinkNewModel takes a new descendant model and updates the model ID saved
// by the model for later reference in addition to adding that new model to a
// map of descendant models.
func (m DefaultRootModel) LinkNewModel(model CommonModel, modelID *string) {
	*modelID = model.GetModelID()
	m.Models.Store(model.GetModelID(), model)
}

// MustGetModel returns a model identified by the parameter modelID from the
// linked model sync.Map. The function will panic if the model cannot be found
// or if the found object doesn't implement the 'CommonModel' interface.
func (m DefaultRootModel) MustGetModel(modelID string) CommonModel {
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

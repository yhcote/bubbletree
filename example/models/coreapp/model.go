// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package coreapp

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"example/internal/app"
	"example/models/configurator"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
	"github.com/yhcote/bubbletree"
)

const (
	// short model name used for identification.
	modelName = "coreapp"
)

var (
	// unique model instance id based on 'modelName'.
	lastID atomic.Int64
)

// New creates and initializes a new model ready to be used.
func New(opts ...Option) bubbletree.AppModel {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Model{
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
		opt(m)
	}

	// Create and link all descendant models used in the application.
	// NOTE: No descendants at the moment (movers will be implemented).
	m.Models = new(sync.Map)

	m.Logger.Info("New model created", "ModelID", m.ID)
	return m
}

// Model is the definition of the coreapp model.
type Model struct {
	// Include fields and default methods of bubbletree.DefaultAppModel.
	bubbletree.DefaultAppModel
}

// Init sends a kick-off tea command when needed.
func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd

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

	switch msg := msg.(type) {
	// When the system is configured, activate main model logic.
	case configurator.ConfigReadyMsg:
		if m.IsInactive() {
			m.State = bubbletree.ActiveState
			m.LogStateChange(msg)
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

	return m, tea.Batch(cmds...)
}

// View is the model's rendering routine that creates the output reflecting
// the current state of the model data. The rendered string is passed back up
// to the root model for final window composition.
func (m Model) View(w, h int) string {
	return ""
}

// GetViewHeader returns the model's header view string.
func (m Model) GetViewHeader(w, h int) string {
	return fmt.Sprintf("%v %v / model: %v", app.ProgramName, app.ProgramVersion, m.ID)
}

// Options

// Option is used to set options for the new model at creation.
type Option func(*Model)

func WithLogger(logger *slog.Logger) Option {
	return func(m *Model) {
		m.OptLogger = logger
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

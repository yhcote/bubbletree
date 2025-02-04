// Copyright 2023-2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package bubbletree

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/spf13/viper"

	tea "github.com/charmbracelet/bubbletea"
)

// RootModel defines the root node of the bubbletree framework. Only one
// model should implement this interface in a program. Only this model has
// its Init, Update, and View methods called directly by bubble tea. In
// bubbletree, the root model instantiates and registers its direct children
// and is responsible for calling Init/Update/View as appropriate. The idea
// is that tea.Msgs are propagated down the model tree. Those messages will
// generate new tea.Cmds from all the tree components, which will be flowing
// back up to the root model. Those tree gathered commands are returned to
// bubble tea for processing. Similarly, at View execution time, the root
// model will have all the nodes from the tree run their View() method and
// the result will be aggregated by the root model View() method for final
// UI composition.
type RootModel interface {
	// tea.Model is the Charmbracelet bubble tea main Init/Update/View interface
	// that connects the program to the bubble tea framework.
	tea.Model

	// UpdateNodeModels is responsible to call the Update method on all child models
	// that the root model has registered. It also gathers the child models new
	// commands generated and returns them alongside the updated root model.
	UpdateNodeModels(msg tea.Msg) (RootModel, tea.Cmd)
}

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
	UpdateNodeModels(msg tea.Msg) (BranchModel, tea.Cmd)
}

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

// CommonModel is an interface defining the routine requirements for each
// models used in the bubbletree framework. With each components (bubbles in
// bubbletea framework parlance) implementing the interface, it is easy and
// predictable to define how all parts of the bubbletree system work together
// with the other model types defined above. Most models will not only
// implement this iterface but will rather choose between a leaf, a branch or
// the one root model.
type CommonModel interface {
	// Init sends a kick-off tea command for the model if appropriate.
	Init() tea.Cmd

	// View is the model rendering routine that creates the output reflecting
	// the current state of the model data. This rendered string is passed back
	// up the tree to the root model for final window composition. The w and h
	// parameters sets the model's virtual view window size. This could be the
	// same as the actual terminal window size if the root model communicates it as
	// being the case.
	View(w, h int) string

	// GetViewHeader returns the view portion that should be displayed in the app's
	// header section of the final composed UI window. This should be a short string
	// or a unicode icon.
	GetViewHeader() string

	// GetViewFooter returns the view portion that should be displayed in the app's
	// footer section of the final composed UI window. This should be a short string
	// or a unicode icon.
	GetViewFooter() string

	// CancelContext is used to explicitely cancel the model's instance context.
	CancelContext()

	// GetModelID is an identification helper routine. Calling it will return the
	// name of the model as identified internally. This ID can be combined with
	// an arbitrary prefix to uniquely identify a specific model instance out of
	// others created from the same model.
	GetModelID() string

	// GetState returns the model's current state. For now, it informs whether the
	// model is doing work (ActiveState), whether it hasn't started yet
	// (InactiveState), that it is shutting down (ShuttingDownState) or terminated
	// with all resources freed (FinishedState).
	GetState() State

	// IsActive returns whether the  model's instance is currently in "Active" state.
	IsActive() bool

	// IsInactive returns whether the model's instance is currently in "Inactive"
	// state.
	IsInactive() bool

	// IsShuttingDown returns whether the model's instance has acknowledged the
	// ShuttingDownMsg. The model should have initiated it own shutdown actions by
	// cleaning up resources and canceling contexts when required.
	IsShuttingDown() bool

	// IsFinished returns whether the model's instance has completed its shutdown
	// actions and all long running goroutines, if any, have returned from the
	// receiving the context cancellation signal.
	IsFinished() bool

	// GetProperties returns the model's instance current property set.
	GetProperties() Properties

	// IsDisabled returns whether the model's instance has the "Disabled" property
	// set. Disabled models will not respond to loop events and is a useful state
	// to debug the rest of the system without the disabled model(s).
	IsDisabled() bool

	// IsFocused returns whether the model's instance has the "Focused" property set.
	// The focused model instance may use this as a signal to activate itself and
	// generate non-empty views.
	IsFocused() bool
}

// Common defines data that are common use amongst all instantiated models.
type Common struct {
	// The unique ID for this instance: e.g., 'config00', 'config01'...
	ID string

	// The current model's instance State.
	State State

	// Whether or not this model is currently in focus/active.
	Properties Properties

	// The initial viper Viper passed along.
	Viper *viper.Viper

	// slog Logger to use throughout the model.
	Logger *slog.Logger

	// The model context needed to notify long running goroutines.
	Ctx context.Context

	// The cancel func associated with the above context.
	Cancel context.CancelFunc
}

// The possible bubble tree model states
const (
	InactiveState State = iota
	ActiveState
	ShuttingDownState
	FinishedState
)

type State int

func (s State) String() string {
	switch s {
	case InactiveState:
		return "INACTIVE"
	case ActiveState:
		return "ACTIVE"
	case ShuttingDownState:
		return "SHUTTINGDOWN"
	case FinishedState:
		return "FINISHED"
	default:
		return "UNKNOWN"
	}
}

// The possible bubble tree model properties
const (
	Disabled Properties = 1 << iota
	Focused
)

type Properties int

func (p Properties) String() string {
	var props []string
	if p&Disabled != 0 {
		props = append(props, "DISABLED")
	}
	if p&Focused != 0 {
		props = append(props, "FOCUSED")
	}
	return strings.Join(props, "|")
}

// Msg/Cmd's

type (
	// ShutDownMsg is a model-global message used to request that the list of
	// specified model instances enter their shutdown sequence. This allows time for
	// graceful cleanup before tea.Quit get called at a later time or when a model
	// instance self terminate.
	ShutDownMsg struct{ ModelIDs []string }

	// ModelFinishedMsg is a model-global message sent when the model has completed
	// its shutdown sequence: its goroutines returned, and its resources are fully
	// freed.
	ModelFinishedMsg struct{ ModelID string }

	// SetFocusMsg is a model-global message sent to request a change of model
	// instance focus. The message includes the Model ID that needs focus. Target
	// models should set the Focused property on the model and act approprietely.
	SetFocusMsg struct{ ModelID string }

	// SetDisabledMsg is a model-global message sent to request that models instances
	// listed have the "Disabled" properties set.
	SetDisabledMsg struct{ ModelIDs []string }

	// ErrMsg is a model-global message sent when an error occured while running the
	// model.
	ErrMsg struct{ Err error }
)

// IsRecipient returns whether the message is destined to the specified model
// instance.
func (msg ShutDownMsg) IsRecipient(id string) bool {
	return slices.Contains(msg.ModelIDs, id)
}

// IsRecipient returns whether the message is destined to the specified model
// instance.
func (msg ModelFinishedMsg) IsRecipient(id string) bool {
	return msg.ModelID == id
}

// IsRecipient returns whether the message is destined to the specified model
// instance.
func (msg SetFocusMsg) IsRecipient(id string) bool {
	return msg.ModelID == id
}

// IsRecipient returns whether the message is destined to the specified model
// instance.
func (msg SetDisabledMsg) IsRecipient(id string) bool {
	return slices.Contains(msg.ModelIDs, id)
}

// ShutDownCmd returns a model-global message when the system is requested
// to shut down. This is typically sent by the root model following an fatal
// error or when the user wants to close the application. Model instances should
// cleanup and move into an FinishedState upon receiving this message.
func ShutDownCmd(ids []string) tea.Cmd {
	return func() tea.Msg {
		return ShutDownMsg{ModelIDs: ids}
	}
}

// ModelFinishedCmd returns a model-global message when the model has completed
// its its shutdown sequence.
func ModelFinishedCmd(id string) tea.Cmd {
	return func() tea.Msg {
		return ModelFinishedMsg{ModelID: id}
	}
}

// SetFocusCmd returns a model-global message when component focus needs to
// change to a specific model instance.
func SetFocusCmd(id string) tea.Cmd {
	return func() tea.Msg {
		return SetFocusMsg{ModelID: id}
	}
}

// SetDisabledCmd returns a model-global message to set the targeted models
// instances as disabled.
func SetDisabledCmd(ids []string) tea.Cmd {
	return func() tea.Msg {
		return SetDisabledMsg{ModelIDs: ids}
	}
}

// ErrCmd returns a model-global message when an error occured.
func ErrCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrMsg{err}
	}
}

func (e ErrMsg) Error() string { return e.Err.Error() }

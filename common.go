// Copyright 2023-2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package bubbletree

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/spf13/viper"

	tea "github.com/charmbracelet/bubbletea"
	charmlog "github.com/charmbracelet/log"
)

// CommonModel is an interface defining the routine requirements for each
// models used in the bubbletree framework. With each components (bubbles in
// bubbletea framework parlance) implementing the interface, it is easy and
// predictable to define how all parts of the bubbletree system work together
// with the other model types defined as: root, branch and leaf. Most models
// implementations will not only satisfy this iterface but will also implement
// either a leaf, a branch or the one-per-application root model.
type CommonModel interface {
	Init() tea.Cmd
	View(w, h int) string
	GetViewHeader(w, h int) string
	GetViewFooter(w, h int) string
	CancelContext()
	GetModelID() string
	GetState() State
	IsActive() bool
	IsInactive() bool
	IsShuttingDown() bool
	IsFinished() bool
	GetProperties() Properties
	IsDisabled() bool
	IsFocused() bool
}

// DefaultCommonModel defines data that are common use amongst all instantiated models.
type DefaultCommonModel struct {
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

	// Optional theme for UI styling. Can be nil if application doesn't use theming.
	Theme Themer
}

// Init is the default implementation of the CommonModel interface. It sends
// a kick-off tea command for the model when needed.
func (m DefaultCommonModel) Init() tea.Cmd {
	return nil
}

// View is the default implementation of the CommonModel interface. It is the
// model rendering routine that creates the output reflecting the current
// state of the model data. This rendered string is passed back up the tree
// to the root model for final window composition. The w and h parameters
// sets the model's virtual view window size. This could be the same as the
// actual terminal window size if the root model communicates it as being the
// case.
func (m DefaultCommonModel) View(w, h int) string {
	if m.IsActive() {
		return "Application Running..."
	}
	return ""
}

// GetViewHeader is the default implementation of the CommonModel interface.
// It returns the view portion that should be displayed in the app's header
// section of the final composed UI window. This should be a short string or
// a unicode icon.
func (m DefaultCommonModel) GetViewHeader(w, h int) string {
	if m.IsActive() {
		return "Application's header notes / menus"
	}
	return ""
}

// GetViewFooter is the default implementation of the CommonModel interface.
// It returns the view portion that should be displayed in the app's footer
// section of the final composed UI window. This should be a short string or
// a unicode icon.
func (m DefaultCommonModel) GetViewFooter(w, h int) string {
	if m.IsActive() {
		return "Application's footer notes / menus"
	}
	return ""
}

// CancelContext is the default implementation of the CommonModel interface.
// It is used to explicitely cancel the model's instance context.
func (m DefaultCommonModel) CancelContext() {
	if m.Cancel != nil {
		m.Logger.Info("Cancelling model's context")
		m.Cancel()
	} else {
		m.Logger.Warn("The model's cancel function is unexpectedly nil, context cannot be cancelled.")
	}
}

// GetModelID is the default implementation of the CommonModel interface. It
// is an identification helper routine. Calling it will return the name of
// the model as identified internally. This ID can be combined with an
// arbitrary prefix to uniquely identify a specific model instance out of
// others created from the same model.
func (m DefaultCommonModel) GetModelID() string {
	return m.ID
}

// GetState is the default implementation of the CommonModel interface. It
// returns the model's current state. For now, it informs whether the model
// is doing work (ActiveState), whether it hasn't started yet (InactiveState),
// that it is shutting down (ShuttingDownState) or terminated with all
// resources freed (FinishedState).
func (m DefaultCommonModel) GetState() State {
	return m.State
}

// IsActive is the default implementation of the CommonModel interface. It
// returns whether the  model's instance is currently in "Active" state.
func (m DefaultCommonModel) IsActive() bool {
	return m.State == ActiveState
}

// IsInactive is the default implementation of the CommonModel interface. It
// returns whether the model's instance is currently in "Inactive" state.
func (m DefaultCommonModel) IsInactive() bool {
	return m.State == InactiveState
}

// IsShuttingDown is the default implementation of the CommonModel interface.
// It returns whether the model's instance has acknowledged the
// ShuttingDownMsg. The model should have initiated it own shutdown actions
// by cleaning up resources and canceling contexts when required.
func (m DefaultCommonModel) IsShuttingDown() bool {
	return m.State == ShuttingDownState
}

// IsFinished is the default implementation of the CommonModel interface. It
// returns whether the model's instance has completed its shutdown actions
// and all long running goroutines, if any, have returned from the receiving
// the context cancellation signal.
func (m DefaultCommonModel) IsFinished() bool {
	return m.State == FinishedState
}

// GetProperties is the default implementation of the CommonModel interface.
// It returns the model's instance current property set.
func (m DefaultCommonModel) GetProperties() Properties {
	return m.Properties
}

// IsDisabled is the default implementation of the CommonModel interface. It
// returns whether the model's instance has the "Disabled" property set.
// Disabled models will not respond to loop events and is a useful state to
// debug the rest of the system without the disabled model(s).
func (m DefaultCommonModel) IsDisabled() bool {
	return m.Properties&Disabled != 0
}

// IsFocused is the default implementation of the CommonModel interface. It
// returns whether the model's instance has the "Focused" property set. The
// focused model instance may use this as a signal to activate itself and
// generate non-empty views.
func (m DefaultCommonModel) IsFocused() bool {
	return m.Properties&Focused != 0
}

// GetTheme returns the theme if set, otherwise returns the default minimal theme.
// This ensures models always have access to a valid theme for rendering.
func (m DefaultCommonModel) GetTheme() Themer {
	if m.Theme != nil {
		return m.Theme
	}
	return DefaultMinimalTheme()
}

// LogStateChange sends a standardized log entry after a model state change.
func (m DefaultCommonModel) LogStateChange(msg any) {
	handler, ok := m.Logger.Handler().(*charmlog.Logger)
	if !ok {
		panic("logger's handler concrete type is not *charmlog.Logger")
	}
	handler.Helper()
	handler.Info("Model State Change",
		"ModelID", m.GetModelID(),
		"NewState", m.State,
		"OnMsg", fmt.Sprintf("%T", msg))
}

// LogAction sends a standardized log entry after a new queued tea.Cmd.
func (m DefaultCommonModel) LogAction(msg any, action string) {
	handler, ok := m.Logger.Handler().(*charmlog.Logger)
	if !ok {
		panic("logger's handler concrete type is not *charmlog.Logger")
	}
	handler.Helper()
	handler.Info("Model Action",
		"ModelID", m.GetModelID(),
		"Action", action,
		"OnMsg", fmt.Sprintf("%T", msg))
}

// LogPropertyChange sends a standardized log entry after a model property
// change.
func (m DefaultCommonModel) LogPropertyChange(msg any, old, new Properties) {
	handler, ok := m.Logger.Handler().(*charmlog.Logger)
	if !ok {
		panic("logger's handler concrete type is not *charmlog.Logger")
	}
	handler.Helper()
	handler.Info("Model Property Change",
		"ModelID", m.GetModelID(),
		"Properties", fmt.Sprintf("%v -> %v", old, new),
		"OnMsg", fmt.Sprintf("%T", msg))
}

// LogNotice sends a standardized log entry of a notification.
func (m DefaultCommonModel) LogNotice(msg any, notice string) {
	handler, ok := m.Logger.Handler().(*charmlog.Logger)
	if !ok {
		panic("logger's handler concrete type is not *charmlog.Logger")
	}
	handler.Helper()
	handler.Info("Model Notification",
		"ModelID", m.GetModelID(),
		"Notice", notice,
		"OnMsg", fmt.Sprintf("%T", msg))
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
	if p == 0 {
		return "NONE"
	}
	if p&Disabled != 0 {
		props = append(props, "DISABLED")
	}
	if p&Focused != 0 {
		props = append(props, "FOCUSED")
	}
	return strings.Join(props, "|")
}

// SetDisabled sets the Disabled model property.
func (p *Properties) SetDisabled() (old, new Properties) {
	old = *p
	*p |= Disabled
	return old, *p
}

// UnsetDisabled unsets the Disabled model property.
func (p *Properties) UnsetDisabled() (old, new Properties) {
	old = *p
	*p &= ^Disabled
	return old, *p
}

// SetFocused sets the Focused model property.
func (p *Properties) SetFocused() (old, new Properties) {
	old = *p
	*p |= Focused
	return old, *p
}

// UnsetFocused unsets the Focused model property.
func (p *Properties) UnsetFocused() (old, new Properties) {
	old = *p
	*p &= ^Focused
	return old, *p
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

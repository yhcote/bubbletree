// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package bubbletree

import (
	"log/slog"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
)

// AppModel defines a Core Application model in the bubbletree framework. This
// single instance model is the only one being directly connected to the
// RootModel instance of the system (defined to work directly in the bubble tea
// context, rather than bubbletree). It essentially is a BranchModel with the
// addition of the AppView method which facilitates the program quitting event.
// When the root model detects that the program is quitting (with errors or
// not), is passes this information to AppView() so that the Core Application
// may display the last window before quitting.
type AppModel interface {
	// BranchModel embeds the branch model interface that an App Model should
	// implement first.
	BranchModel

	// AppView is a App Model special routine allowing the bridge between the
	// RootNode instance, implementing the bubble tea Model, and the bubbletree
	// Core Application (App Model). Particularly, when the root model is aware
	// of a descendant model producing an error, it is able to pass down the
	// error information so that the Core Application may display the last View
	// (with the passed down error) before the program exits. In other words,
	// the AppView method isn't simply calling the regular View method only when
	// the application is quitting.
	AppView(quitting bool, err error) string

	// QuittingView is a App Model special routine that is allowed to display
	// the last window before the programs exits.
	QuittingView(err error) string
}

// DefaultAppModel implements default methods for the AppModel interface.
// It can be used as a base implementation when no specific code is required.
// The implementor of the AppModel interface may overwrite the default
// behavior of the model by reimplementing desired methods.
type DefaultAppModel struct {
	// Include all common fields and methods of the default BranchModel
	// interface implementation.
	DefaultBranchModel

	// Options passed in by the main program.
	AppOpts

	// Real-time screen dimensions (updated by tea.WindowSizeMsg).
	Screen

	// Update + View execution performance.
	Timing
}

// Screen holds the current state details of the terminal window.
type Screen struct {
	Width  int
	Height int
}

// Timing holds information about the update and view execution performance
// in terms of elapsed time.
type Timing struct {
	UpdateRuntime time.Duration
	ViewRuntime   time.Duration
}

// AppOpts describes general configurations or states of the root model
// for the application.
type AppOpts struct {
	OptProgname    string
	OptProgver     string
	OptLogger      *slog.Logger
	OptConfigViper *viper.Viper
	OptSpewcfg     *spew.ConfigState
	OptReconf      bool
	OptTheme       Themer // Optional theme for UI styling
}

// AppOption is used to set options on the app model.
type AppOption func(*DefaultAppModel)

// WithProgname passes the base program name.
func WithProgname(progname string) AppOption {
	return func(m *DefaultAppModel) {
		m.OptProgname = progname
	}
}

// WithProgver passes the base program version.
func WithProgver(progver string) AppOption {
	return func(m *DefaultAppModel) {
		m.OptProgver = progver
	}
}

// WithLogger sets the logger to use for model logging.
func WithLogger(logger *slog.Logger) AppOption {
	return func(m *DefaultAppModel) {
		m.OptLogger = logger
	}
}

// WithConfigViper sets the active viper config to use in the model.
func WithConfigViper(viper *viper.Viper) AppOption {
	return func(m *DefaultAppModel) {
		m.OptConfigViper = viper
	}
}

// WithSpewConfigState sets spew utility configuration state instance to
// a specified one.
func WithSpewConfigState(spewcfg *spew.ConfigState) AppOption {
	return func(m *DefaultAppModel) {
		m.OptSpewcfg = spewcfg
	}
}

// WithReconfigure ignores a complete config file and brings up the user input
// form, enabling configuration changes.
func WithReconfigure(force bool) AppOption {
	return func(m *DefaultAppModel) {
		m.OptReconf = force
	}
}

// WithTheme sets the theme for the entire application tree.
func WithTheme(theme Themer) AppOption {
	return func(m *DefaultAppModel) {
		m.OptTheme = theme
	}
}

// AppView is the default implementation of the AppModel interface.
func (m DefaultAppModel) AppView(quitting bool, err error) string {
	if quitting {
		if err != nil {
			m.Logger.Error("terminating program", "error", err)
		}
		return m.QuittingView(err)
	}
	return m.View(m.Width, m.Height)
}

// QuittingView is the default implementation of the AppModel interface.
func (m DefaultAppModel) QuittingView(err error) string {
	return ""
}

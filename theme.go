// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package bubbletree

import "github.com/charmbracelet/lipgloss"

// Themer is the minimal interface that any theme implementation must satisfy.
// Applications using bubbletree can provide their own implementations with
// additional methods and styling capabilities beyond this base interface.
type Themer interface {
	// Core color accessors - every theme must provide these base colors
	GetPrimaryColor() lipgloss.Color
	GetSecondaryColor() lipgloss.Color
	GetSuccessColor() lipgloss.Color
	GetErrorColor() lipgloss.Color
	GetTextColor() lipgloss.Color       // Text color for content on main background
	GetBackgroundColor() lipgloss.Color // Main background color
	GetAccentTextColor() lipgloss.Color // Text color for content on colored surfaces (buttons, badges, etc.)

	// Base style accessors - minimal set for framework components
	GetBaseStyle() lipgloss.Style   // Default text style
	GetHeaderStyle() lipgloss.Style // For headers/titles
	GetErrorStyle() lipgloss.Style  // For error messages

	// Render helpers - convenience methods
	RenderText(string) string
	RenderHeader(string) string
	RenderError(string) string
}

// OptionalStyleProvider is an optional interface for themes that provide
// additional component styles. Applications can type-assert to check if
// their theme supports these extensions.
type OptionalStyleProvider interface {
	GetButtonStyle() lipgloss.Style
	GetCardStyle() lipgloss.Style
	GetTabStyle() lipgloss.Style
	GetWinbarStyle() lipgloss.Style
}

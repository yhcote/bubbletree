// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package bubbletree

import "github.com/charmbracelet/lipgloss"

// minimalTheme is bubbletree's built-in fallback theme with no dependencies.
// It uses ANSI color codes for maximum compatibility.
type minimalTheme struct {
	baseStyle   lipgloss.Style
	headerStyle lipgloss.Style
	errorStyle  lipgloss.Style
}

// DefaultMinimalTheme returns a minimal theme implementation using basic ANSI colors.
// This theme is used as a fallback when no theme is provided to the application.
func DefaultMinimalTheme() Themer {
	return &minimalTheme{
		baseStyle:   lipgloss.NewStyle(),
		headerStyle: lipgloss.NewStyle().Bold(true),
		errorStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("9")),
	}
}

// Core color accessors using ANSI color codes
func (t *minimalTheme) GetPrimaryColor() lipgloss.Color    { return lipgloss.Color("12") } // Bright blue
func (t *minimalTheme) GetSecondaryColor() lipgloss.Color  { return lipgloss.Color("8") }  // Gray
func (t *minimalTheme) GetSuccessColor() lipgloss.Color    { return lipgloss.Color("10") } // Bright green
func (t *minimalTheme) GetErrorColor() lipgloss.Color      { return lipgloss.Color("9") }  // Bright red
func (t *minimalTheme) GetTextColor() lipgloss.Color       { return lipgloss.Color("7") }  // White
func (t *minimalTheme) GetBackgroundColor() lipgloss.Color { return lipgloss.Color("0") }  // Black
func (t *minimalTheme) GetAccentTextColor() lipgloss.Color { return lipgloss.Color("0") }  // Black text on colored backgrounds

// Base style accessors
func (t *minimalTheme) GetBaseStyle() lipgloss.Style   { return t.baseStyle }
func (t *minimalTheme) GetHeaderStyle() lipgloss.Style { return t.headerStyle }
func (t *minimalTheme) GetErrorStyle() lipgloss.Style  { return t.errorStyle }

// Render helpers
func (t *minimalTheme) RenderText(s string) string   { return t.baseStyle.Render(s) }
func (t *minimalTheme) RenderHeader(s string) string { return t.headerStyle.Render(s) }
func (t *minimalTheme) RenderError(s string) string  { return t.errorStyle.Render(s) }

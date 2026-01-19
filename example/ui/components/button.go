// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/yhcote/bubbletree"
)

// ButtonPrimary represents a primary action button.
// Optimized for 60 FPS rendering by caching computed styles.
type ButtonPrimary struct {
	theme bubbletree.Themer
	text  string

	// Style cache - computed once, used for every render
	cachedStyle lipgloss.Style
	styleDirty  bool // true when style needs recomputation
}

// ButtonSecondary represents a secondary action button.
type ButtonSecondary struct {
	theme bubbletree.Themer
	text  string

	cachedStyle lipgloss.Style
	styleDirty  bool
}

// ButtonSuccess represents a success/confirmation button.
type ButtonSuccess struct {
	theme bubbletree.Themer
	text  string

	cachedStyle lipgloss.Style
	styleDirty  bool
}

// ButtonError represents a destructive/error button.
type ButtonError struct {
	theme bubbletree.Themer
	text  string

	cachedStyle lipgloss.Style
	styleDirty  bool
}

func NewPrimaryButton(theme bubbletree.Themer, text string) *ButtonPrimary {
	return &ButtonPrimary{
		theme:      theme,
		text:       text,
		styleDirty: true, // Compute on first render
	}
}

func NewSecondaryButton(theme bubbletree.Themer, text string) *ButtonSecondary {
	return &ButtonSecondary{
		theme:      theme,
		text:       text,
		styleDirty: true,
	}
}

func NewSuccessButton(theme bubbletree.Themer, text string) *ButtonSuccess {
	return &ButtonSuccess{
		theme:      theme,
		text:       text,
		styleDirty: true,
	}
}

func NewErrorButton(theme bubbletree.Themer, text string) *ButtonError {
	return &ButtonError{
		theme:      theme,
		text:       text,
		styleDirty: true,
	}
}

// Render returns the styled button string. Optimized for 60 FPS by using
// cached styles - only recomputes when theme changes.
func (b *ButtonPrimary) Render() string {
	if b.styleDirty {
		// Compute style once, cache result
		var base lipgloss.Style
		if rtTheme, ok := b.theme.(bubbletree.OptionalStyleProvider); ok {
			base = rtTheme.GetButtonStyle()
		} else {
			base = b.theme.GetBaseStyle()
		}
		b.cachedStyle = base.Background(b.theme.GetPrimaryColor())
		b.styleDirty = false
	}

	// Use cached style - zero allocations on subsequent renders
	return b.cachedStyle.Render(b.text)
}

func (b *ButtonSecondary) Render() string {
	if b.styleDirty {
		var base lipgloss.Style
		if rtTheme, ok := b.theme.(bubbletree.OptionalStyleProvider); ok {
			base = rtTheme.GetButtonStyle()
		} else {
			base = b.theme.GetBaseStyle()
		}
		b.cachedStyle = base.Background(b.theme.GetSecondaryColor())
		b.styleDirty = false
	}
	return b.cachedStyle.Render(b.text)
}

func (b *ButtonSuccess) Render() string {
	if b.styleDirty {
		var base lipgloss.Style
		if rtTheme, ok := b.theme.(bubbletree.OptionalStyleProvider); ok {
			base = rtTheme.GetButtonStyle()
		} else {
			base = b.theme.GetBaseStyle()
		}
		b.cachedStyle = base.Background(b.theme.GetSuccessColor())
		b.styleDirty = false
	}
	return b.cachedStyle.Render(b.text)
}

func (b *ButtonError) Render() string {
	if b.styleDirty {
		var base lipgloss.Style
		if rtTheme, ok := b.theme.(bubbletree.OptionalStyleProvider); ok {
			base = rtTheme.GetButtonStyle()
		} else {
			base = b.theme.GetBaseStyle()
		}
		b.cachedStyle = base.Background(b.theme.GetErrorColor())
		b.styleDirty = false
	}
	return b.cachedStyle.Render(b.text)
}

// SetTheme updates the button's theme and invalidates the style cache.
func (b *ButtonPrimary) SetTheme(theme bubbletree.Themer) {
	b.theme = theme
	b.styleDirty = true
}

func (b *ButtonSecondary) SetTheme(theme bubbletree.Themer) {
	b.theme = theme
	b.styleDirty = true
}

func (b *ButtonSuccess) SetTheme(theme bubbletree.Themer) {
	b.theme = theme
	b.styleDirty = true
}

func (b *ButtonError) SetTheme(theme bubbletree.Themer) {
	b.theme = theme
	b.styleDirty = true
}

// SetText updates the button text without invalidating the style cache.
// This is efficient since text content doesn't affect styling.
func (b *ButtonPrimary) SetText(text string) {
	b.text = text
}

func (b *ButtonSecondary) SetText(text string) {
	b.text = text
}

func (b *ButtonSuccess) SetText(text string) {
	b.text = text
}

func (b *ButtonError) SetText(text string) {
	b.text = text
}

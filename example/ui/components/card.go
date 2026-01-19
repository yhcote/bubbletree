// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/yhcote/bubbletree"
)

// Card represents a styled content container.
// Optimized for 60 FPS rendering by caching computed styles.
type Card struct {
	theme   bubbletree.Themer
	content string

	// Style cache - computed once, used for every render
	cachedStyle lipgloss.Style
	styleDirty  bool // true when style needs recomputation
}

func NewCard(theme bubbletree.Themer, content string) *Card {
	return &Card{
		theme:      theme,
		content:    content,
		styleDirty: true, // Compute on first render
	}
}

// SetContent updates the card's content without invalidating the style cache.
// This is efficient since content changes don't affect styling.
func (c *Card) SetContent(content string) {
	c.content = content
}

// GetContent returns the current content
func (c *Card) GetContent() string {
	return c.content
}

// AppendContent adds content to the existing content
func (c *Card) AppendContent(content string) {
	c.content += content
}

// PrependContent adds content to the beginning of existing content
func (c *Card) PrependContent(content string) {
	c.content = content + c.content
}

// ClearContent clears all content
func (c *Card) ClearContent() {
	c.content = ""
}

// Render efficiently renders the card with current content.
// Optimized for 60 FPS by using cached styles - only recomputes when theme changes.
func (c *Card) Render() string {
	if c.styleDirty {
		// Compute style once, cache result
		if rtTheme, ok := c.theme.(bubbletree.OptionalStyleProvider); ok {
			c.cachedStyle = rtTheme.GetCardStyle()
		} else {
			c.cachedStyle = c.theme.GetBaseStyle()
		}
		c.styleDirty = false
	}

	// Use cached style - zero allocations on subsequent renders
	return c.cachedStyle.Render(c.content)
}

// SetTheme updates the card's theme and invalidates the style cache.
func (c *Card) SetTheme(theme bubbletree.Themer) {
	c.theme = theme
	c.styleDirty = true
}

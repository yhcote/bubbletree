// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/yhcote/bubbletree"
)

type Winbar struct {
	theme        bubbletree.Themer
	isTopBar     bool
	maxWidth     int
	maxHeight    int
	content      string
	contentStyle lipgloss.Style
}

func NewWinbar(theme bubbletree.Themer, isTopBar bool, width, height int) *Winbar {
	w := &Winbar{
		theme:     theme,
		isTopBar:  isTopBar,
		maxWidth:  width,
		maxHeight: height,
		content:   "",
	}

	// Cache styles (expensive to create)
	w.updateStyles()

	return w
}

// updateStyles rebuilds cached styles when theme or dimensions change.
func (w *Winbar) updateStyles() {
	// Use interface methods to get colors
	secondaryColor := w.theme.GetSecondaryColor()

	// Try to get Winbar style from OptionalStyleProvider, fallback to base style
	var winbarBaseStyle lipgloss.Style
	if rtTheme, ok := w.theme.(bubbletree.OptionalStyleProvider); ok {
		winbarBaseStyle = rtTheme.GetWinbarStyle()
	} else {
		winbarBaseStyle = w.theme.GetBaseStyle()
	}

	w.contentStyle = winbarBaseStyle.
		Width(w.maxWidth). // Stretch the bar across the window.
		MaxWidth(w.maxWidth).
		MaxHeight(w.maxHeight).
		BorderForeground(secondaryColor)
	if w.isTopBar {
		w.contentStyle = w.contentStyle.Border(lipgloss.InnerHalfBlockBorder(), false, false, true, false)
	} else {
		w.contentStyle = w.contentStyle.Border(lipgloss.InnerHalfBlockBorder(), true, false, false, false)
	}
}

// SetContent sets the content to display in the window bar body.
func (w *Winbar) SetContent(content string) {
	w.content = content
}

// GetContent returns the current content.
func (w *Winbar) GetContent() string {
	return w.content
}

// SetMaxWidth updates the window bar maximum width and refreshes cached styles.
func (w *Winbar) SetMaxWidth(width int) {
	if w.maxWidth != width {
		w.maxWidth = width
		w.updateStyles() // Refresh cached styles
	}
}

// SetMaxHeight updates the window bar maximum height and refreshes cached styles.
func (w *Winbar) SetMaxHeight(height int) {
	if w.maxHeight != height {
		w.maxHeight = height
		w.updateStyles() // Refresh cached styles
	}
}

// SetTheme updates the theme and refreshes cached styles.
func (w *Winbar) SetTheme(theme bubbletree.Themer) {
	w.theme = theme
	w.updateStyles() // Refresh cached styles
}

// GetContentWidth returns the actual available content width inside the window
// bar. This accounts for the tab's own padding, margins, and borders.
func (w *Winbar) GetContentWidth() int {
	return w.contentStyle.GetWidth() - w.contentStyle.GetHorizontalFrameSize()
}

// GetContentHeight returns the actual available content height inside the window
// bar. This accounts for the tab's own padding, margins, and borders.
func (w *Winbar) GetContentHeight() int {
	return w.contentStyle.GetHeight() - w.contentStyle.GetVerticalFrameSize()
}

// Render renders the complete window bar interface.
func (w *Winbar) Render(width, height int) string {
	w.SetMaxWidth(width)
	w.SetMaxHeight(height)
	return w.contentStyle.Render(w.content)
}

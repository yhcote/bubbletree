// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yhcote/bubbletree"
)

type Tabs struct {
	theme     bubbletree.Themer
	tabs      []TabTitle
	activeTab int
	maxWidth  int
	maxHeight int
	content   string

	// Cached styles for performance (avoid recreating on every render)
	contentStyle         lipgloss.Style
	activeTabStyle       lipgloss.Style
	inactiveTabStyle     lipgloss.Style
	activeTabNameStyle   lipgloss.Style
	inactiveTabNameStyle lipgloss.Style
	shortcutKeyStyle     lipgloss.Style
	gapStyle             lipgloss.Style

	// Cached border objects
	activeTabBorder   lipgloss.Border
	inactiveTabBorder lipgloss.Border

	// Pre-allocated for string building
	renderedTabs []string
}

type TabTitle struct {
	Name        string
	ShortcutKey string
}

func NewTabs(theme bubbletree.Themer, tabs []TabTitle, width, height int) *Tabs {
	t := &Tabs{
		theme:        theme,
		tabs:         tabs,
		activeTab:    0,
		maxWidth:     width,
		maxHeight:    height,
		content:      "",
		renderedTabs: make([]string, len(tabs)), // Pre-allocate slice
	}

	// Cache border objects (expensive to create)
	t.activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ", // Space for active tab (pulls it forward)
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	t.inactiveTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─", // Full bottom border (shadowed look)
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	// Cache styles (expensive to create)
	t.updateStyles()

	return t
}

// updateStyles rebuilds cached styles when theme or dimensions change
func (t *Tabs) updateStyles() {
	// Use interface methods to get colors
	primaryColor := t.theme.GetPrimaryColor()
	secondaryColor := t.theme.GetSecondaryColor()
	textColor := t.theme.GetTextColor()

	// Try to get Tab style from OptionalStyleProvider, fallback to base style
	var tabBaseStyle lipgloss.Style
	if rtTheme, ok := t.theme.(bubbletree.OptionalStyleProvider); ok {
		tabBaseStyle = rtTheme.GetTabStyle()
	} else {
		tabBaseStyle = t.theme.GetBaseStyle()
	}

	t.contentStyle = tabBaseStyle.
		Width(t.maxWidth).
		Height(t.maxHeight)

	t.inactiveTabStyle = tabBaseStyle.
		Border(t.inactiveTabBorder, true).
		BorderForeground(secondaryColor).
		Padding(0, 1)

	t.activeTabStyle = tabBaseStyle.
		Border(t.activeTabBorder, true).
		BorderForeground(secondaryColor).
		Padding(0, 1)

	t.inactiveTabNameStyle = tabBaseStyle.
		Foreground(textColor)

	t.activeTabNameStyle = tabBaseStyle.
		Foreground(primaryColor).
		Bold(true)

	t.shortcutKeyStyle = tabBaseStyle.
		Foreground(secondaryColor).MarginLeft(1)

	t.gapStyle = tabBaseStyle.
		Foreground(secondaryColor)
}

// SetActiveTab changes the active tab
func (t *Tabs) SetActiveTab(index int) {
	if index >= 0 && index < len(t.tabs) {
		t.activeTab = index
	}
}

// GetActiveTab returns the current active tab index
func (t *Tabs) GetActiveTab() int {
	return t.activeTab
}

// SetContent sets the content to display in the tab body
func (t *Tabs) SetContent(content string) {
	t.content = content
}

// GetContent returns the current content
func (t *Tabs) GetContent() string {
	return t.content
}

// SetMaxWidth updates the maximum tab width and refreshes cached styles.
func (t *Tabs) SetMaxWidth(width int) {
	if t.maxWidth != width {
		t.maxWidth = width
		t.updateStyles() // Refresh cached styles
	}
}

// SetMaxHeight updates the maximum tab height and refreshes cached styles.
func (t *Tabs) SetMaxHeight(height int) {
	if t.maxHeight != height {
		t.maxHeight = height
		t.updateStyles() // Refresh cached styles
	}
}

// SetTheme updates the theme and refreshes cached styles
func (t *Tabs) SetTheme(theme bubbletree.Themer) {
	t.theme = theme
	t.updateStyles() // Refresh cached styles
}

// GetContentWidth returns the actual available content width inside the tabs.
// This accounts for the tab's own padding, margins, and borders.
func (t *Tabs) GetContentWidth() int {
	return t.contentStyle.GetWidth() - t.contentStyle.GetHorizontalFrameSize()
}

// GetContentHeight returns the actual available content height inside the tabs.
// This accounts for the tab's own padding, margins, and borders.
func (t *Tabs) GetContentHeight() int {
	return t.contentStyle.GetHeight() - t.contentStyle.GetVerticalFrameSize()
}

// Render renders the complete tabs interface
func (t *Tabs) Render(width, height int) string {
	var output strings.Builder

	t.SetMaxWidth(width)

	// Render tab header
	header := t.renderTabs()
	output.WriteString(header + "\n")

	// Calculate remaining height for content (subtract header)
	headerHeight := lipgloss.Height(header)
	t.SetMaxHeight(height - headerHeight) // Now this is the actual content area height

	// Render tab content
	output.WriteString(t.renderContent())

	return output.String()
}

// renderContent renders the content area below tabs
func (t *Tabs) renderContent() string {
	return t.contentStyle.Render(t.content)
}

// renderTabs creates the tab header row (optimized for 60FPS)
func (t *Tabs) renderTabs() string {
	// Use cached styles and pre-allocated slice - zero allocations in render path
	for i, tabTitle := range t.tabs {
		if i == t.activeTab {
			t.renderedTabs[i] = t.activeTabStyle.Render(t.activeTabNameStyle.Render(tabTitle.Name) +
				t.shortcutKeyStyle.Render(tabTitle.ShortcutKey))
		} else {
			t.renderedTabs[i] = t.inactiveTabStyle.Render(t.inactiveTabNameStyle.Render(tabTitle.Name) +
				t.shortcutKeyStyle.Render(tabTitle.ShortcutKey))
		}
	}

	// Join tabs horizontally using pre-allocated slice
	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, t.renderedTabs...)

	// Calculate remaining width and fill gap efficiently
	remainingWidth := t.maxWidth - lipgloss.Width(tabRow)
	if remainingWidth > 0 {
		// Pre-build gap content to avoid repeated string operations
		var gapBuilder strings.Builder
		gapBuilder.Grow(remainingWidth*3 + 2) // Pre-allocate capacity

		// Row 1: spaces
		for range remainingWidth {
			gapBuilder.WriteByte(' ')
		}
		gapBuilder.WriteByte('\n')

		// Row 2: spaces
		for range remainingWidth {
			gapBuilder.WriteByte(' ')
		}
		gapBuilder.WriteByte('\n')

		// Row 3: border continuation
		for range remainingWidth {
			gapBuilder.WriteString("─")
		}

		// Use cached gap style - no style creation
		tabGap := t.gapStyle.Render(gapBuilder.String())
		tabRow = lipgloss.JoinHorizontal(lipgloss.Top, tabRow, tabGap)
	}

	return tabRow
}

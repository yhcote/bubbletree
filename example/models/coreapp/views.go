// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package coreapp

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/yhcote/bubbletree/logger"
)

const (
	topBarMaxHeight    = 2
	bottomBarMaxHeight = 2
)

// renderNormalWindow renders the complete window view for the normally
// executing program.
func (m Model) renderNormalWindow(maxWidth, maxHeight int) string {
	header := m.renderHeader(maxWidth, topBarMaxHeight)
	header = sureFit(header, maxWidth, topBarMaxHeight)

	footer := m.renderFooter(maxWidth, bottomBarMaxHeight)
	footer = sureFit(footer, maxWidth, bottomBarMaxHeight)

	content := m.renderContent(maxWidth, maxHeight-(lipgloss.Height(header)+lipgloss.Height(footer)))
	content = sureFit(content, maxWidth, maxHeight-(lipgloss.Height(header)+lipgloss.Height(footer)))

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

// renderHeader renders the top bar of the window. This is the status or info
// bar. It's also the top zone of the application window.
func (m Model) renderHeader(maxWidth, maxHeight int) string {
	if m.IsInactive() {
		return ""
	}

	s := m.Theme.RenderSecondaryText(m.OptProgname + " / " + m.focusedID)
	if m.focusedID != m.ID {
		// Get the focused model and generate its current view header.
		header := m.MustGetModel(m.focusedID).GetViewHeader(maxWidth, maxHeight)
		if header != "" {
			s += m.Theme.RenderPrimaryText(" • ") + header
		}
	}
	m.topbar.SetContent(s)

	return m.topbar.Render(maxWidth, maxHeight)
}

// renderContent renders the body of the window, the main content of the
// application output. It's also the middle zone of the application window.
func (m Model) renderContent(maxWidth, maxHeight int) string {
	// Self (coreapp) is in focus, deal with general possible views.
	if m.focusedID == m.ID {
		switch m.tabber.GetActiveTab() {
		case 0:
			m.tabber.SetContent(m.Theme.RenderNormalText("Program initializing"))
		case 2:
			m.tabber.SetContent(
				m.Theme.RenderNormalText("\nLog Window Unimplemented, Check ") +
					m.Theme.RenderPrimaryText(logger.GetLoggerOutputName()),
			)
		default:
			m.tabber.SetContent(m.Theme.RenderNormalText("unknown active tab index"))
		}
	} else {
		// Get the focused model and generate its current state view.
		content := m.MustGetModel(m.focusedID).View(maxWidth, maxHeight)
		m.tabber.SetContent(content)
	}

	return m.tabber.Render(maxWidth, maxHeight)
}

// renderFooter renders the top bar of the window. This is the status or info
// bar. It's also the top zone of the application window.
func (m Model) renderFooter(maxWidth, maxHeight int) string {
	if m.IsInactive() {
		return m.Theme.RenderSecondaryText("Program Initializing...")
	}

	s := m.Theme.RenderSecondaryText("Press ") +
		m.Theme.RenderPrimaryText("ESC") +
		m.Theme.RenderSecondaryText(" to quit")
	if m.focusedID != m.ID {
		// Get the focused model and generate its current view footer.
		footer := m.MustGetModel(m.focusedID).GetViewFooter(maxWidth, maxHeight)
		if footer != "" {
			s += m.Theme.RenderPrimaryText(" • ") + footer
		}
	}
	m.bottombar.SetContent(s)

	return m.bottombar.Render(maxWidth, maxHeight)
}

// renderQuittingWindow renders the complete window view for the quitting
// program. It should clearly display the error that caused the exit in case
// of abnormal exit.
func (m Model) renderQuittingWindow(err error, maxWidth, maxHeight int) string {
	header := m.renderHeader(maxWidth, topBarMaxHeight)
	header = sureFit(header, maxWidth, topBarMaxHeight)

	footer := ""
	if err != nil {
		footer = m.Theme.RenderErrorText(fmt.Sprintf("ERROR: %v", err))
		footer = sureFit(footer, maxWidth, bottomBarMaxHeight)
	}

	content := m.renderContent(maxWidth, maxHeight-(lipgloss.Height(header)+lipgloss.Height(footer)))
	content = sureFit(content, maxWidth, maxHeight-(lipgloss.Height(header)+lipgloss.Height(footer)))

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

// sureFit makes sure that view strings built by decendant models respected
// maxWidth and maxHeight. In case reviewed strings don't fit in their zones,
// this function will force truncate them to fit.
func sureFit(str string, maxWidth, maxHeight int) string {
	// When remaining space has no columns or rows left, return an empty
	// string.
	if maxWidth <= 0 || maxHeight <= 0 {
		logger.Log().Error("string element will not fit (0 length width or height)",
			"width", maxWidth, "height", maxHeight, "strWidth", lipgloss.Width(str),
			"strHeight", lipgloss.Height(str))
		return ""
	}

	// When remaining space cannot fit the size of the string completely,
	// shoehorn (truncate) that string into available space.
	if maxHeight < lipgloss.Height(str) || maxWidth < lipgloss.Width(str) {
		logger.Log().Error("string element doesn't fit", "width", maxWidth, "height", maxHeight,
			"strWidth", lipgloss.Width(str), "strHeight", lipgloss.Height(str))
		return lipgloss.NewStyle().MaxWidth(maxWidth).MaxHeight(maxHeight).Render(str)
	}

	return str
}

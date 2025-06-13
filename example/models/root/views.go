// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package root

import (
	"fmt"

	"example/internal/app"

	"github.com/charmbracelet/lipgloss"
	"github.com/yhcote/bubbletree"
)

// This file includes the helper code for the base model. We want to keep a
// clean model.go file with its core (Model Init() Update() View())
// definitions.
var (
	// Styles
	headerStyle  = lipgloss.NewStyle()
	footerStyle  = lipgloss.NewStyle()
	contentStyle = lipgloss.NewStyle()
)

func renderBaseView(w, h int) string {
	version := fmt.Sprintf("%v %v", app.ProgramName, app.ProgramVersion)
	header := renderShareHeight(headerStyle, w, &h, "No Active Workflows")
	footer := renderShareHeight(footerStyle, w, &h, version)
	content := renderFillHeight(contentStyle, w, &h, "")
	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

func renderConfigView(model bubbletree.CommonModel, w, h int) string {
	header := renderShareHeight(headerStyle, w, &h, model.GetViewHeader())
	footer := renderShareHeight(footerStyle, w, &h, model.GetViewFooter())
	content := renderFillHeight(contentStyle.Align(lipgloss.Center, lipgloss.Center), w, &h, model.View(areaSize(contentStyle, w, &h)))
	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

func renderCoreAppView(model bubbletree.CommonModel, w, h int) string {
	header := renderShareHeight(headerStyle, w, &h, model.GetViewHeader())
	footer := renderShareHeight(footerStyle, w, &h, model.GetViewFooter())
	content := renderFillHeight(contentStyle, w, &h, model.View(areaSize(contentStyle, w, &h)))
	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

func renderQuittingView(err error, w, h int) string {
	header := renderShareHeight(headerStyle, w, &h, "Quitting Application")
	footer := renderShareHeight(footerStyle, w, &h, "")
	var content string
	if err != nil {
		content = renderFillHeight(contentStyle, w, &h, fmt.Sprintf("application error: %v", err))
	} else {
		content = renderFillHeight(contentStyle, w, &h, "\n  See you later!")
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

func renderShareHeight(s lipgloss.Style, w int, h *int, strs ...string) string {
	_, vframe := s.GetFrameSize()
	out := s.MaxWidth(w).Width(w - vframe).Render(strs...)

	sh := lipgloss.Height(out)
	*h -= sh
	return out
}

func renderFillHeight(s lipgloss.Style, w int, h *int, strs ...string) string {
	hframe, vframe := s.GetFrameSize()
	out := s.MaxWidth(w).Width(w - vframe).MaxHeight(*h).Height(*h - hframe).Render(strs...)
	*h = 0
	return out
}

func areaSize(s lipgloss.Style, w int, h *int) (int, int) {
	hframe, vframe := s.GetFrameSize()
	width := w - vframe
	height := *h - hframe
	return width, height
}

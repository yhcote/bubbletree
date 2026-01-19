// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package themes

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/yhcote/bubbletree"
)

// Ensure PunchyTheme implements bubbletree.Themer and OptionalStyleProvider
var (
	_ bubbletree.Themer                = (*PunchyTheme)(nil)
	_ bubbletree.OptionalStyleProvider = (*PunchyTheme)(nil)
)

// PunchyTheme extends the base Themer interface with application-specific
// styles and methods. It also implements OptionalStyleProvider.
type PunchyTheme struct {
	Colors Colors
	Styles Styles
}

type Colors struct {
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Success    lipgloss.Color
	Error      lipgloss.Color
	Text       lipgloss.Color
	AccentText lipgloss.Color // Text color for colored surfaces (buttons, badges, etc.)
	Background lipgloss.Color
}

type Styles struct {
	Base   lipgloss.Style
	Header lipgloss.Style
	Error  lipgloss.Style
	Button lipgloss.Style
	Card   lipgloss.Style
	Tab    lipgloss.Style
	Winbar lipgloss.Style
}

// Implement bubbletree.Themer interface
func (t *PunchyTheme) GetPrimaryColor() lipgloss.Color    { return t.Colors.Primary }
func (t *PunchyTheme) GetSecondaryColor() lipgloss.Color  { return t.Colors.Secondary }
func (t *PunchyTheme) GetSuccessColor() lipgloss.Color    { return t.Colors.Success }
func (t *PunchyTheme) GetErrorColor() lipgloss.Color      { return t.Colors.Error }
func (t *PunchyTheme) GetTextColor() lipgloss.Color       { return t.Colors.Text }
func (t *PunchyTheme) GetBackgroundColor() lipgloss.Color { return t.Colors.Background }
func (t *PunchyTheme) GetAccentTextColor() lipgloss.Color { return t.Colors.AccentText }

func (t *PunchyTheme) GetBaseStyle() lipgloss.Style   { return t.Styles.Base }
func (t *PunchyTheme) GetHeaderStyle() lipgloss.Style { return t.Styles.Header }
func (t *PunchyTheme) GetErrorStyle() lipgloss.Style  { return t.Styles.Error }

func (t *PunchyTheme) RenderNormalText(s string) string { return t.Styles.Base.Render(s) }
func (t *PunchyTheme) RenderHeaderText(s string) string { return t.Styles.Header.Render(s) }
func (t *PunchyTheme) RenderErrorText(s string) string  { return t.Styles.Error.Render(s) }

// Expand bubbletree.Themer interface
func (t *PunchyTheme) RenderPrimaryText(s string) string {
	return t.Styles.Base.Foreground(t.GetPrimaryColor()).Render(s)
}
func (t *PunchyTheme) RenderSecondaryText(s string) string {
	return t.Styles.Base.Foreground(t.GetSecondaryColor()).Render(s)
}

// Implement bubbletree.OptionalStyleProvider interface
func (t *PunchyTheme) GetButtonStyle() lipgloss.Style { return t.Styles.Button }
func (t *PunchyTheme) GetCardStyle() lipgloss.Style   { return t.Styles.Card }
func (t *PunchyTheme) GetTabStyle() lipgloss.Style    { return t.Styles.Tab }
func (t *PunchyTheme) GetWinbarStyle() lipgloss.Style { return t.Styles.Winbar }

// Default theme - defaults to Dark theme
func Default() bubbletree.Themer {
	return Dark()
}

func Light() *PunchyTheme {
	colors := Colors{
		Primary:    lipgloss.Color("#7C3AED"), // Purple
		Secondary:  lipgloss.Color("#64748B"), // Slate
		Success:    lipgloss.Color("#10B981"), // Emerald
		Error:      lipgloss.Color("#EF4444"), // Red
		Text:       lipgloss.Color("#374151"), // Dark gray text
		AccentText: lipgloss.Color("#FFFFFF"), // White text on buttons
		Background: lipgloss.Color("#F9FAFB"), // Light background
	}
	return NewColorTheme(colors)
}

func Dark() *PunchyTheme {
	colors := Colors{
		Primary:    lipgloss.Color("#7e8de2"), // Lighter purple for dark mode
		Secondary:  lipgloss.Color("#494949"), // Lighter slate
		Success:    lipgloss.Color("#34D399"), // Lighter emerald
		Error:      lipgloss.Color("#F87171"), // Lighter red
		Text:       lipgloss.Color("#bbbbbb"), // Light gray text
		AccentText: lipgloss.Color("#000000"), // Dark text on buttons
		Background: lipgloss.Color("#1b1b1b"), // Dark background
	}
	return NewColorTheme(colors)
}

func NewColorTheme(colors Colors) *PunchyTheme {
	return &PunchyTheme{
		Colors: colors,
		Styles: Styles{
			Base: lipgloss.NewStyle().
				Foreground(colors.Text),
			Header: lipgloss.NewStyle().
				Foreground(colors.Primary).
				Bold(true).
				Margin(1, 0).
				Align(lipgloss.Center),
			Error: lipgloss.NewStyle().
				Foreground(colors.Error),
			Button: lipgloss.NewStyle().
				Foreground(colors.AccentText).
				Background(colors.Primary). // Default to primary, will be overridden
				Padding(0, 2).
				Bold(true),
			Card: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colors.Secondary).
				Padding(0, 1).
				Margin(0, 1),
			Tab: lipgloss.NewStyle(),
			Winbar: lipgloss.NewStyle().
				Padding(0, 1),
		},
	}
}

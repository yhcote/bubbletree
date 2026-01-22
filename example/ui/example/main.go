// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package main

import (
	"fmt"
	"strings"

	"example/ui/components"
	"example/ui/themes"

	"github.com/charmbracelet/lipgloss"
)

func main() {
	// Demo both themes
	fmt.Println("=== LIGHT THEME ===")
	lightTheme := themes.Light()
	demoTheme(lightTheme)

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	fmt.Println("=== DARK THEME ===")
	darkTheme := themes.Dark()
	demoTheme(darkTheme)
}

func demoTheme(theme *themes.PunchyTheme) {
	// Create the root tabs component with all other components as leaf nodes
	fmt.Println(demoTabsAsRoot(theme))
}

func demoTabsAsRoot(theme *themes.PunchyTheme) string {
	var output strings.Builder

	output.WriteString("🗂️  Complete Application Demo (Tabs as Root Component):\n\n")

	// Create tabs as the root component
	tabs := components.NewTabs(theme, []components.TabTitle{
		{Name: "Components", ShortcutKey: "f1"},
		{Name: "Dashboard", ShortcutKey: "f2"},
		{Name: "Settings", ShortcutKey: "f3"},
		{Name: "About", ShortcutKey: "f4"},
	}, 128, 60)

	// Tab 1: Components Demo (buttons, cards, etc.)
	tabs.SetActiveTab(0)
	tabs.SetContent(createComponentsTabContent(theme))
	output.WriteString("=== COMPONENTS TAB ===\n")
	output.WriteString(tabs.Render(128, 60) + "\n\n")

	// Tab 2: Dashboard Demo (simulated app content) - pass tabs instance
	tabs.SetActiveTab(1)
	tabs.SetContent(createDashboardTabContent(theme, tabs))
	output.WriteString("=== DASHBOARD TAB ===\n")
	output.WriteString(tabs.Render(128, 60) + "\n\n")

	// Tab 3: Settings Demo (forms and configuration)
	tabs.SetActiveTab(2)
	tabs.SetContent(createSettingsTabContent(theme))
	output.WriteString("=== SETTINGS TAB ===\n")
	output.WriteString(tabs.Render(128, 60) + "\n\n")

	// Tab 4: About Demo (text and layout)
	tabs.SetActiveTab(3)
	tabs.SetContent(createAboutTabContent(theme))
	output.WriteString("=== ABOUT TAB ===\n")
	output.WriteString(tabs.Render(128, 60) + "\n")

	return output.String()
}

// Components tab content - showcases all UI components
func createComponentsTabContent(theme *themes.PunchyTheme) string {
	var content strings.Builder

	// Header
	header := theme.Styles.Header.Render("🎨 UI Components Showcase")
	content.WriteString(header + "\n\n")

	// Button components section
	primaryButton := components.NewPrimaryButton(theme, "Primary")
	secondaryButton := components.NewSecondaryButton(theme, "Secondary")
	successButton := components.NewSuccessButton(theme, "Success")
	errorButton := components.NewErrorButton(theme, "Error")

	spacer := lipgloss.NewStyle().Background(theme.GetBackgroundColor()).Render("  ")
	buttonRow := lipgloss.JoinHorizontal(lipgloss.Top,
		primaryButton.Render(), spacer,
		secondaryButton.Render(), spacer,
		successButton.Render(), spacer,
		errorButton.Render(),
	)

	buttonCard := components.NewCard(theme, "Button Components:\n\n"+buttonRow)
	content.WriteString(buttonCard.Render() + "\n\n")

	// Text styles section
	textContent := "Text Styles:\n\n" +
		theme.RenderNormalText("• Default text style") + "\n" +
		theme.Styles.Header.MarginTop(1).MarginBottom(0).Render("• Header Style") + "\n" +
		lipgloss.NewStyle().Foreground(theme.Colors.Success).Render("• ✓ Success message") + "\n" +
		lipgloss.NewStyle().Foreground(theme.Colors.Error).Render("• ✗ Error message")

	textCard := components.NewCard(theme, textContent)
	content.WriteString(textCard.Render() + "\n\n")

	// Color palette section
	colorCard := components.NewCard(theme, createColorPalette(theme))
	content.WriteString(colorCard.Render() + "\n\n")

	// Dynamic card demo
	dynamicCard := components.NewCard(theme, "Dynamic Content Demo:\n\n• Content can be updated at runtime\n• Optimized for 60FPS updates\n• Perfect for real-time data")
	content.WriteString(dynamicCard.Render())

	return content.String()
}

// Dashboard tab content - simulated application dashboard
func createDashboardTabContent(theme *themes.PunchyTheme, tabs *components.Tabs) string {
	var content strings.Builder

	// Dashboard header
	header := theme.Styles.Header.Render("📊 Application Dashboard")
	content.WriteString(header + "\n\n")

	// Create 3x3 grid using actual content width
	content.WriteString(createDashboardGrid(theme, tabs))

	return content.String()
}

// Create a 3x3 grid of dashboard cards
func createDashboardGrid(theme *themes.PunchyTheme, tabs *components.Tabs) string {
	// Get the actual available content width from the tabs component
	// This accounts for tab padding, margins, and borders automatically
	contentWidth := tabs.GetContentWidth()

	// Row 1: User metrics - compact content
	card1 := components.NewCard(theme,
		"👥 Active Users\n\n"+
			"Current: 1,247\n"+
			"↑ 12% this week\n"+
			"Peak: 1,456\n\n"+
			"Trend: ████████░░")

	card2 := components.NewCard(theme,
		"📈 Registrations\n\n"+
			"Today: 89 new\n"+
			"Week: 623 total\n"+
			"Rate: 3.2%\n\n"+
			"Score: ████████░░")

	card3 := components.NewCard(theme,
		"🌐 Online Now\n\n"+
			"Active: 456\n"+
			"Session: 24min\n"+
			"Bounce: 23%\n\n"+
			"Engage: █████████░")

	// Row 2: Financial metrics - compact format
	card4 := components.NewCard(theme,
		"💰 Revenue\n\n"+
			"Today: $2,431\n"+
			"Month: $12,483\n"+
			"↑ 8.3% growth\n\n"+
			"Target: ██████░░░░")

	card5 := components.NewCard(theme,
		"🛒 Orders\n\n"+
			"Today: 234\n"+
			"Average: $52.87\n"+
			"Time: 2.3min\n\n"+
			"Rate: █████████░")

	card6 := components.NewCard(theme,
		"💳 Payments\n\n"+
			"Success: 97.8%\n"+
			"Failed: 5\n"+
			"Refunds: $247\n\n"+
			"Health: ██████████")

	// Row 3: System metrics - compact indicators
	card7 := components.NewCard(theme,
		"🖥️  CPU\n\n"+
			"Usage: 67%\n"+
			"Peak: 89%\n"+
			"Load: 2.34\n\n"+
			"Perf: ███████░░░")

	card8 := components.NewCard(theme,
		"💾 Memory\n\n"+
			"Used: 4.2GB/8GB\n"+
			"Cache: 1.8GB\n"+
			"Free: 2.0GB\n\n"+
			"Eff: ████████░░")

	card9 := components.NewCard(theme,
		"📦 Storage\n\n"+
			"Used: 89%\n"+
			"Free: 55GB\n"+
			"Backup: 99.2%\n\n"+
			"Health: █████████░")

	// Build rows WITHOUT manual spacers - margins handle spacing automatically
	row1 := lipgloss.JoinHorizontal(lipgloss.Top,
		card1.Render(),
		card2.Render(),
		card3.Render(),
	)

	row2 := lipgloss.JoinHorizontal(lipgloss.Top,
		card4.Render(),
		card5.Render(),
		card6.Render(),
	)

	row3 := lipgloss.JoinHorizontal(lipgloss.Top,
		card7.Render(),
		card8.Render(),
		card9.Render(),
	)

	// Apply container style using actual content width from tabs
	containerStyle := lipgloss.NewStyle().
		Background(theme.Colors.Background).
		Width(contentWidth). // Use calculated content width, not hardcoded 128
		Align(lipgloss.Left)

	// Render each row with container styling
	styledRow1 := containerStyle.Render(row1)
	styledRow2 := containerStyle.Render(row2)
	styledRow3 := containerStyle.Render(row3)

	// Join styled rows vertically - now all have consistent width/background
	grid := lipgloss.JoinVertical(lipgloss.Left,
		styledRow1,
		styledRow2,
		styledRow3,
	)

	return grid
}

// Settings tab content - configuration interface
func createSettingsTabContent(theme *themes.PunchyTheme) string {
	var content strings.Builder

	header := theme.Styles.Header.Render("⚙️ Application Settings")
	content.WriteString(header + "\n\n")

	// Theme settings
	themeCard := components.NewCard(theme,
		"Theme Configuration:\n\n"+
			"• Current theme: "+(func() string {
			if theme == themes.Light() {
				return "Light Mode"
			}
			return "Dark Mode"
		})()+"\n"+
			"• Background: Themed\n"+
			"• Contrast: High\n"+
			"• Font: System Default")
	content.WriteString(themeCard.Render() + "\n\n")

	// Action buttons
	saveButton := components.NewSuccessButton(theme, "Save Settings")
	resetButton := components.NewSecondaryButton(theme, "Reset to Default")
	cancelButton := components.NewErrorButton(theme, "Cancel")

	spacer := lipgloss.NewStyle().Background(theme.GetBackgroundColor()).Render("  ")
	buttonRow := lipgloss.JoinHorizontal(lipgloss.Top,
		saveButton.Render(), spacer,
		resetButton.Render(), spacer,
		cancelButton.Render(),
	)

	actionsCard := components.NewCard(theme, "Actions:\n\n"+buttonRow)
	content.WriteString(actionsCard.Render())

	return content.String()
}

// About tab content - application information
func createAboutTabContent(theme *themes.PunchyTheme) string {
	var content strings.Builder

	header := theme.Styles.Header.Render("ℹ️ About This Application")
	content.WriteString(header + "\n\n")

	aboutCard := components.NewCard(theme,
		"Lipgloss UI Demo Application\n\n"+
			"Version: 1.0.0\n"+
			"Built with: Go 1.24 + Lipgloss v1.1.0\n\n"+
			"Features:\n"+
			"• Modern terminal UI components\n"+
			"• Light and Dark theme support\n"+
			"• 60FPS optimized rendering\n"+
			"• Component-based architecture\n"+
			"• Responsive layouts\n\n"+
			"This demo showcases a complete component tree with tabs as the\n"+
			"root component and various UI elements as leaf nodes.")

	content.WriteString(aboutCard.Render())

	return content.String()
}

// Helper function for color palette (now used within components tab)
func createColorPalette(theme *themes.PunchyTheme) string {
	var palette strings.Builder

	palette.WriteString("Color Palette:\n\n")

	// Create color swatches
	primarySwatch := lipgloss.NewStyle().
		Background(theme.Colors.Primary).
		Foreground(theme.Colors.AccentText).
		Padding(0, 2).
		Render("Primary")

	secondarySwatch := lipgloss.NewStyle().
		Background(theme.Colors.Secondary).
		Foreground(theme.Colors.AccentText).
		Padding(0, 2).
		Render("Secondary")

	successSwatch := lipgloss.NewStyle().
		Background(theme.Colors.Success).
		Foreground(theme.Colors.AccentText).
		Padding(0, 2).
		Render("Success")

	errorSwatch := lipgloss.NewStyle().
		Background(theme.Colors.Error).
		Foreground(theme.Colors.AccentText).
		Padding(0, 2).
		Render("Error")

	// Use themed spacer for consistency
	spacer := lipgloss.NewStyle().Background(theme.GetBackgroundColor()).Render("  ")
	swatches := lipgloss.JoinHorizontal(lipgloss.Top,
		primarySwatch, spacer,
		secondarySwatch, spacer,
		successSwatch, spacer,
		errorSwatch,
	)

	palette.WriteString(swatches)

	return palette.String()
}

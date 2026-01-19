// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package components

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/yhcote/bubbletree"
)

// Table provides an enhanced table component with proper alignment and theming.
// Optimized for 60 FPS rendering by caching computed styles.
type Table struct {
	theme   bubbletree.Themer
	headers []string
	rows    [][]string
	widths  []int
	width   int

	// Style cache - computed once, used for every render
	headerStyle       lipgloss.Style
	rowStyle          lipgloss.Style
	cellStyle         lipgloss.Style
	alternateRowStyle lipgloss.Style // For odd rows
	borderStyle       lipgloss.Style
	styleDirty        bool // true when styles need recomputation
}

// NewTable creates a new table component
func NewTable(theme bubbletree.Themer, headers []string) *Table {
	t := &Table{
		theme:      theme,
		headers:    headers,
		rows:       [][]string{},
		widths:     make([]int, len(headers)),
		styleDirty: true, // Compute on first render
	}

	t.updateStyles()
	t.calculateMinWidths()

	return t
}

// updateStyles creates cached styles for the table.
// Only called when styleDirty is true (theme change or first render).
func (t *Table) updateStyles() {
	if !t.styleDirty {
		return
	}

	primaryColor := t.theme.GetPrimaryColor()
	secondaryColor := t.theme.GetSecondaryColor()
	textColor := t.theme.GetTextColor()
	bgColor := t.theme.GetBackgroundColor()

	t.headerStyle = lipgloss.NewStyle().
		Background(bgColor).
		Foreground(primaryColor).
		Bold(true).
		Padding(0, 1)

	t.rowStyle = lipgloss.NewStyle().
		Background(bgColor).
		Foreground(textColor)

	t.cellStyle = lipgloss.NewStyle().
		Background(bgColor).
		Foreground(textColor).
		Padding(0, 1)

	// Compute alternate row background (slightly different shade)
	// Use 25% adjustment for good contrast on dark backgrounds
	alternateRowBg := adjustBrightness(bgColor, 0.25)
	t.alternateRowStyle = lipgloss.NewStyle().
		Background(alternateRowBg).
		Foreground(textColor).
		Padding(0, 1)

	t.borderStyle = lipgloss.NewStyle().
		Foreground(secondaryColor).
		Background(bgColor)

	t.styleDirty = false
}

// AddRow adds a row to the table
func (t *Table) AddRow(row []string) {
	if len(row) != len(t.headers) {
		// Pad or truncate to match header count
		paddedRow := make([]string, len(t.headers))
		for i := 0; i < len(t.headers); i++ {
			if i < len(row) {
				paddedRow[i] = row[i]
			} else {
				paddedRow[i] = ""
			}
		}
		row = paddedRow
	}

	t.rows = append(t.rows, row)
	t.updateColumnWidths(row)
}

// SetWidth sets the total table width and redistributes column widths
func (t *Table) SetWidth(width int) {
	t.width = width
	t.redistributeWidths()
}

// calculateMinWidths calculates minimum required widths for each column
func (t *Table) calculateMinWidths() {
	// Initialize with header widths
	for i, header := range t.headers {
		t.widths[i] = utf8.RuneCountInString(header)
	}
}

// updateColumnWidths updates column widths based on content
func (t *Table) updateColumnWidths(row []string) {
	for i, cell := range row {
		if i < len(t.widths) {
			// Strip ANSI codes before measuring width
			cleaned := stripANSI(cell)
			cellWidth := utf8.RuneCountInString(cleaned)
			if cellWidth > t.widths[i] {
				t.widths[i] = cellWidth
			}
		}
	}
}

// redistributeWidths redistributes column widths to fit within total width
func (t *Table) redistributeWidths() {
	if t.width <= 0 {
		return
	}

	// Calculate total width needed for borders and padding
	totalOverhead := len(t.headers)*3 + 1 // |_cell_|_cell_|
	availableWidth := t.width - totalOverhead

	if availableWidth <= 0 {
		return
	}

	// Calculate current total content width
	totalContentWidth := 0
	for _, width := range t.widths {
		totalContentWidth += width
	}

	// If content fits naturally, use natural widths
	if totalContentWidth <= availableWidth {
		return
	}

	// Otherwise, redistribute proportionally
	for i := range t.widths {
		proportion := float64(t.widths[i]) / float64(totalContentWidth)
		t.widths[i] = int(proportion * float64(availableWidth))

		// Ensure minimum width of 3 characters
		if t.widths[i] < 3 {
			t.widths[i] = 3
		}
	}
}

// Render renders the complete table
func (t *Table) Render() string {
	if len(t.headers) == 0 {
		return ""
	}

	var output strings.Builder

	// Render top border
	output.WriteString(t.renderTopBorder() + "\n")

	// Render header
	output.WriteString(t.renderHeader() + "\n")

	// Render separator
	output.WriteString(t.renderSeparator() + "\n")

	// Render rows
	for i, row := range t.rows {
		output.WriteString(t.renderRow(row, i) + "\n")
	}

	// Render bottom border
	output.WriteString(t.renderBottomBorder() + "\n")

	return output.String()
}

// renderTopBorder renders the top border of the table
func (t *Table) renderTopBorder() string {
	var parts []string

	parts = append(parts, "┌")
	for i, width := range t.widths {
		parts = append(parts, strings.Repeat("─", width+2)) // +2 for padding
		if i < len(t.widths)-1 {
			parts = append(parts, "┬")
		}
	}
	parts = append(parts, "┐")

	return t.borderStyle.Render(strings.Join(parts, ""))
}

// renderHeader renders the table header
func (t *Table) renderHeader() string {
	var cells []string

	for i, header := range t.headers {
		cellContent := t.truncateOrPad(header, t.widths[i])
		cell := t.headerStyle.Width(t.widths[i] + 2).Render(cellContent) // +2 for padding
		cells = append(cells, cell)
	}

	return t.borderStyle.Render("│") + strings.Join(cells, t.borderStyle.Render("│")) + t.borderStyle.Render("│")
}

// renderSeparator renders the separator line between header and rows
func (t *Table) renderSeparator() string {
	var parts []string

	parts = append(parts, "├")
	for i, width := range t.widths {
		parts = append(parts, strings.Repeat("─", width+2)) // +2 for padding
		if i < len(t.widths)-1 {
			parts = append(parts, "┼")
		}
	}
	parts = append(parts, "┤")

	return t.borderStyle.Render(strings.Join(parts, ""))
}

// renderRow renders a single table row
func (t *Table) renderRow(row []string, rowIndex int) string {
	var cells []string

	for i, cell := range row {
		if i < len(t.widths) {
			cellContent := t.truncateOrPad(cell, t.widths[i])

			// Alternate row styling for better readability - use cached style
			style := t.cellStyle
			if rowIndex%2 == 1 {
				style = t.alternateRowStyle
			}

			styledCell := style.Width(t.widths[i] + 2).Render(cellContent)
			cells = append(cells, styledCell)
		}
	}

	return t.borderStyle.Render("│") + strings.Join(cells, t.borderStyle.Render("│")) + t.borderStyle.Render("│")
}

// renderBottomBorder renders the bottom border of the table
func (t *Table) renderBottomBorder() string {
	var parts []string

	parts = append(parts, "└")
	for i, width := range t.widths {
		parts = append(parts, strings.Repeat("─", width+2)) // +2 for padding
		if i < len(t.widths)-1 {
			parts = append(parts, "┴")
		}
	}
	parts = append(parts, "┘")

	return t.borderStyle.Render(strings.Join(parts, ""))
}

// truncateOrPad truncates or pads content to fit the specified width
func (t *Table) truncateOrPad(content string, width int) string {
	// Remove any existing ANSI codes for width calculation
	cleaned := stripANSI(content)
	contentWidth := utf8.RuneCountInString(cleaned)

	if contentWidth > width {
		// Truncate and add ellipsis
		if width >= 3 {
			runes := []rune(cleaned)
			truncated := string(runes[:width-3]) + "..."
			return truncated
		}
		return string([]rune(cleaned)[:width])
	}

	// Pad with spaces
	padding := width - contentWidth
	return content + strings.Repeat(" ", padding)
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(s string) string {
	// Simple ANSI stripping - could be more comprehensive
	var result strings.Builder
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}

	return result.String()
}

// GetPreferredWidth returns the preferred width for the table
func (t *Table) GetPreferredWidth() int {
	totalWidth := 0
	for _, width := range t.widths {
		totalWidth += width
	}
	// Add overhead for borders and padding
	return totalWidth + len(t.headers)*3 + 1
}

// Clear removes all rows but keeps headers and structure
func (t *Table) Clear() {
	t.rows = [][]string{}
	t.calculateMinWidths()
}

// SetTheme updates the table's theme and invalidates the style cache.
func (t *Table) SetTheme(theme bubbletree.Themer) {
	t.theme = theme
	t.styleDirty = true
}

// adjustBrightness adjusts a color's brightness by the given factor.
// For dark backgrounds, it lightens; for light backgrounds, it darkens.
// factor is typically 0.05 (5% adjustment).
func adjustBrightness(color lipgloss.Color, factor float64) lipgloss.Color {
	colorStr := string(color)

	// Handle hex colors (#RRGGBB or #RGB)
	if strings.HasPrefix(colorStr, "#") {
		hex := strings.TrimPrefix(colorStr, "#")

		// Expand short form (#RGB -> #RRGGBB)
		if len(hex) == 3 {
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		}

		if len(hex) == 6 {
			r, _ := strconv.ParseInt(hex[0:2], 16, 64)
			g, _ := strconv.ParseInt(hex[2:4], 16, 64)
			b, _ := strconv.ParseInt(hex[4:6], 16, 64)

			// Calculate perceived brightness (0-255)
			brightness := (r*299 + g*587 + b*114) / 1000

			// For dark colors (< 128), lighten; for light colors, darken
			adjustment := factor
			if brightness >= 128 {
				adjustment = -factor
			}

			// Apply adjustment
			r = int64(clamp(int(float64(r) * (1 + adjustment))))
			g = int64(clamp(int(float64(g) * (1 + adjustment))))
			b = int64(clamp(int(float64(b) * (1 + adjustment))))

			return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
		}
	}

	// For ANSI colors or unsupported formats, return original
	return color
}

// clamp restricts a value to 0-255 range
func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

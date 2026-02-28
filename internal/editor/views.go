package editor

import (
	"fmt"
	"strings"

	"d2r-multiboxing/internal/modfile"
)

// View implements tea.Model.
func (m Model) View() string {
	switch m.mode {
	case viewEdit:
		return m.viewEditScreen()
	default:
		return m.viewListScreen()
	}
}

// viewListScreen renders the main list/browse view.
func (m Model) viewListScreen() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render(fmt.Sprintf(" D2R Mod Editor — %s ", m.mod.Name))
	b.WriteString(title)
	b.WriteString("\n")

	// Tabs
	b.WriteString(m.renderTabs())
	b.WriteString("\n")

	// Search bar
	if m.searching {
		b.WriteString(m.searchInput.View())
	} else if m.searchQuery != "" {
		b.WriteString(searchPromptStyle.Render(fmt.Sprintf("🔍 Filter: %q  ", m.searchQuery)))
		b.WriteString(helpDescStyle.Render("(/ to edit, Esc to clear)"))
	} else {
		b.WriteString(helpDescStyle.Render("  / search  ↑↓ navigate  Enter edit  Tab switch tab  Ctrl+R restart D2R  q quit"))
	}
	b.WriteString("\n")

	// Entry count
	b.WriteString(helpDescStyle.Render(fmt.Sprintf("  %d entries", len(m.filtered))))
	b.WriteString("\n")

	// List entries
	end := m.scrollTop + m.viewHeight
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := m.scrollTop; i < end; i++ {
		ref := m.filtered[i]
		isSelected := i == m.cursor

		key := keyStyle.Render(ref.Entry.Key)
		preview := modfile.RenderForTerminal(truncate(ref.Entry.EnUS, 60))

		line := fmt.Sprintf(" %s %s", key, preview)

		// Append file tag if on "All" tab
		if m.activeTab == 0 {
			line += " " + fileTagStyle.Render("["+ref.File.Name+"]")
		}

		if isSelected {
			line = selectedItemStyle.Render("▸" + line)
		} else {
			line = normalItemStyle.Render(" " + line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Pad remaining lines
	for i := end - m.scrollTop; i < m.viewHeight; i++ {
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString(m.renderStatusBar())

	return b.String()
}

// viewEditScreen renders the entry edit view.
func (m Model) viewEditScreen() string {
	if m.editRef == nil {
		return "No entry selected"
	}

	var b strings.Builder

	title := titleStyle.Render(fmt.Sprintf(" Editing: %s ", m.editRef.Entry.Key))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Entry info
	b.WriteString(labelStyle.Render("  File: "))
	b.WriteString(fileTagStyle.Render(m.editRef.File.Name+".json"))
	b.WriteString("\n")
	b.WriteString(labelStyle.Render("  ID:   "))
	b.WriteString(fmt.Sprintf("%d", m.editRef.Entry.ID))
	b.WriteString("\n")
	b.WriteString(labelStyle.Render("  Key:  "))
	b.WriteString(m.editRef.Entry.Key)
	b.WriteString("\n\n")

	// Edit fields
	for i, input := range m.editInputs {
		if i == m.editField {
			b.WriteString("  ▸ ")
		} else {
			b.WriteString("    ")
		}
		b.WriteString(input.View())
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Live preview
	b.WriteString(labelStyle.Render("  Preview (enUS):"))
	b.WriteString("\n")
	preview := modfile.RenderForTerminal(m.editInputs[0].Value())
	b.WriteString(previewBorderStyle.Render("  " + preview))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("  Preview (zhTW):"))
	b.WriteString("\n")
	previewZh := modfile.RenderForTerminal(m.editInputs[1].Value())
	b.WriteString(previewBorderStyle.Render("  " + previewZh))
	b.WriteString("\n\n")

	// Color code reference
	b.WriteString(helpDescStyle.Render("  Color codes: "))
	b.WriteString(modfile.RenderForTerminal("\u00ffc0wht "))
	b.WriteString(modfile.RenderForTerminal("\u00ffc1red "))
	b.WriteString(modfile.RenderForTerminal("\u00ffc2grn "))
	b.WriteString(modfile.RenderForTerminal("\u00ffc3blu "))
	b.WriteString(modfile.RenderForTerminal("\u00ffc4gld "))
	b.WriteString(modfile.RenderForTerminal("\u00ffc5gry "))
	b.WriteString(modfile.RenderForTerminal("\u00ffc8org "))
	b.WriteString(modfile.RenderForTerminal("\u00ffc9yel "))
	b.WriteString(modfile.RenderForTerminal("\u00ffc;pur "))
	b.WriteString("\n\n")

	// Help & status
	b.WriteString(helpDescStyle.Render("  Tab switch field  Ctrl+S save  Esc cancel"))
	b.WriteString("\n")

	if m.statusMsg != "" {
		if m.statusIsOK {
			b.WriteString("  " + successStyle.Render(m.statusMsg))
		} else {
			b.WriteString("  " + errorStyle.Render(m.statusMsg))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderTabs renders the tab bar.
func (m Model) renderTabs() string {
	var tabs []string

	// "All" tab
	label := fmt.Sprintf(" All (%d) ", m.countAllEntries())
	if m.activeTab == 0 {
		tabs = append(tabs, activeTabStyle.Render(label))
	} else {
		tabs = append(tabs, inactiveTabStyle.Render(label))
	}

	for i, f := range m.mod.Files {
		label := fmt.Sprintf(" %s (%d) ", f.Name, len(f.Entries))
		if m.activeTab == i+1 {
			tabs = append(tabs, activeTabStyle.Render(label))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(label))
		}
	}

	return strings.Join(tabs, " ")
}

func (m Model) countAllEntries() int {
	total := 0
	for _, f := range m.mod.Files {
		total += len(f.Entries)
	}
	return total
}

// renderStatusBar renders the bottom status bar.
func (m Model) renderStatusBar() string {
	left := fmt.Sprintf(" %d/%d ", m.cursor+1, len(m.filtered))
	if m.statusMsg != "" {
		if m.statusIsOK {
			left += successStyle.Render(m.statusMsg)
		} else {
			left += errorStyle.Render(m.statusMsg)
		}
	}

	bar := statusBarStyle.Width(m.width).Render(left)
	return bar
}

// truncate shortens a string to maxLen runes, appending "…" if truncated.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

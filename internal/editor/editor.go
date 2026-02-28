package editor

import (
	"strings"

	"d2r-multiboxing/internal/modfile"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// viewMode represents the current screen of the TUI.
type viewMode int

const (
	viewList viewMode = iota
	viewEdit
)

// Model is the top-level bubbletea model for the mod editor.
type Model struct {
	mod *modfile.Mod

	// View state
	mode       viewMode
	activeTab  int // index into mod.Files
	cursor     int // index into filtered entries
	scrollTop  int
	viewHeight int // visible rows for the list

	// Search
	searchInput textinput.Model
	searching   bool
	searchQuery string
	filtered    []modfile.EntryRef // currently displayed entries

	// Edit state
	editField  int // 0 = enUS, 1 = zhTW
	editInputs [2]textinput.Model
	editRef    *modfile.EntryRef

	// Status
	statusMsg  string
	statusIsOK bool

	// Terminal size
	width  int
	height int

	// RestartRequested is set to true when the user requests a D2R restart (Ctrl+R in list view).
	// The caller should check this after the TUI exits.
	RestartRequested bool
}

// New creates a new editor model for the given mod.
func New(mod *modfile.Mod) Model {
	si := textinput.New()
	si.Placeholder = "Type to search Key / enUS / zhTW..."
	si.Prompt = "🔍 "
	si.CharLimit = 100

	var editInputs [2]textinput.Model
	for i := range editInputs {
		editInputs[i] = textinput.New()
		editInputs[i].CharLimit = 2000
	}
	editInputs[0].Prompt = "enUS: "
	editInputs[1].Prompt = "zhTW: "

	m := Model{
		mod:         mod,
		searchInput: si,
		editInputs:  editInputs,
		viewHeight:  20,
	}

	m.applyFilter()
	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.SetWindowTitle("D2R Mod Editor — " + m.mod.Name)
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Reserve lines for: title(1) + tabs(1) + search(1) + header(1) + status(1) + help(1)
		m.viewHeight = m.height - 6
		if m.viewHeight < 3 {
			m.viewHeight = 3
		}
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case viewList:
			return m.updateList(msg)
		case viewEdit:
			return m.updateEdit(msg)
		}
	}

	return m, nil
}

// updateList handles key events in list view.
func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If searching, delegate to search input first
	if m.searching {
		switch msg.String() {
		case "esc":
			m.searching = false
			m.searchInput.Blur()
			return m, nil
		case "enter":
			m.searching = false
			m.searchInput.Blur()
			m.searchQuery = m.searchInput.Value()
			m.applyFilter()
			m.cursor = 0
			m.scrollTop = 0
			return m, nil
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			// Live filter
			m.searchQuery = m.searchInput.Value()
			m.applyFilter()
			if m.cursor >= len(m.filtered) {
				m.cursor = max(0, len(m.filtered)-1)
			}
			m.scrollTop = 0
			return m, cmd
		}
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "tab", "right":
		if len(m.mod.Files) > 0 {
			m.activeTab = (m.activeTab + 1) % (len(m.mod.Files) + 1) // +1 for "All" tab
			m.applyFilter()
			m.cursor = 0
			m.scrollTop = 0
		}

	case "shift+tab", "left":
		if len(m.mod.Files) > 0 {
			total := len(m.mod.Files) + 1
			m.activeTab = (m.activeTab - 1 + total) % total
			m.applyFilter()
			m.cursor = 0
			m.scrollTop = 0
		}

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.scrollTop {
				m.scrollTop = m.cursor
			}
		}

	case "down", "j":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
			if m.cursor >= m.scrollTop+m.viewHeight {
				m.scrollTop = m.cursor - m.viewHeight + 1
			}
		}

	case "pgup":
		m.cursor -= m.viewHeight
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.scrollTop = m.cursor

	case "pgdown":
		m.cursor += m.viewHeight
		if m.cursor >= len(m.filtered) {
			m.cursor = max(0, len(m.filtered)-1)
		}
		if m.cursor >= m.scrollTop+m.viewHeight {
			m.scrollTop = m.cursor - m.viewHeight + 1
		}

	case "home":
		m.cursor = 0
		m.scrollTop = 0

	case "end":
		m.cursor = max(0, len(m.filtered)-1)
		if m.cursor >= m.scrollTop+m.viewHeight {
			m.scrollTop = m.cursor - m.viewHeight + 1
		}

	case "/":
		m.searching = true
		return m, m.searchInput.Focus()

	case "enter":
		if len(m.filtered) > 0 {
			return m.enterEditMode()
		}

	case "ctrl+r":
		m.RestartRequested = true
		return m, tea.Quit
	}

	return m, nil
}

// updateEdit handles key events in edit view.
func (m Model) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = viewList
		m.statusMsg = ""
		return m, nil

	case "ctrl+s":
		return m.saveEdit()

	case "tab", "shift+tab":
		m.editField = (m.editField + 1) % 2
		m.editInputs[(m.editField+1)%2].Blur()
		cmd := m.editInputs[m.editField].Focus()
		return m, cmd
	}

	var cmd tea.Cmd
	m.editInputs[m.editField], cmd = m.editInputs[m.editField].Update(msg)
	return m, cmd
}

// enterEditMode switches to edit view for the currently selected entry.
func (m Model) enterEditMode() (tea.Model, tea.Cmd) {
	ref := m.filtered[m.cursor]
	m.editRef = &ref
	m.mode = viewEdit
	m.editField = 0
	m.statusMsg = ""

	m.editInputs[0].SetValue(ref.Entry.EnUS)
	m.editInputs[1].SetValue(ref.Entry.ZhTW)
	m.editInputs[1].Blur()

	return m, m.editInputs[0].Focus()
}

// saveEdit writes the edited values back to the entry and saves the file.
func (m Model) saveEdit() (tea.Model, tea.Cmd) {
	if m.editRef == nil {
		return m, nil
	}

	m.editRef.Entry.EnUS = m.editInputs[0].Value()
	m.editRef.Entry.ZhTW = m.editInputs[1].Value()

	if err := m.editRef.File.Save(); err != nil {
		m.statusMsg = "✗ " + err.Error()
		m.statusIsOK = false
	} else {
		m.statusMsg = "✔ Saved " + m.editRef.File.Name + ".json"
		m.statusIsOK = true
	}

	return m, nil
}

// applyFilter rebuilds the filtered entry list based on active tab and search query.
func (m *Model) applyFilter() {
	var source []modfile.EntryRef
	if m.activeTab == 0 {
		// "All" tab
		source = m.mod.AllEntries()
	} else {
		fileIdx := m.activeTab - 1
		if fileIdx < len(m.mod.Files) {
			sf := m.mod.Files[fileIdx]
			for i := range sf.Entries {
				source = append(source, modfile.EntryRef{
					File:  sf,
					Index: i,
					Entry: &sf.Entries[i],
				})
			}
		}
	}

	if m.searchQuery == "" {
		m.filtered = source
		return
	}

	query := strings.ToLower(m.searchQuery)
	m.filtered = nil
	for _, ref := range source {
		if strings.Contains(strings.ToLower(ref.Entry.Key), query) ||
			strings.Contains(strings.ToLower(modfile.StripColorCodes(ref.Entry.EnUS)), query) ||
			strings.Contains(strings.ToLower(modfile.StripColorCodes(ref.Entry.ZhTW)), query) {
			m.filtered = append(m.filtered, ref)
		}
	}
}

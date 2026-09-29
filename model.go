package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type modalState int

const (
	modalNone modalState = iota
	modalConnecting
	modalRenaming
	modalDeleting
	modalHelp
	modalHookConfirm
	modalImporting
)

type model struct {
	width  int
	height int

	// Config list
	configs []string
	cursor  int

	// Connection state
	activeVPNs []string
	vpnStatus  map[string]VPNStatus

	// Modal state
	modal       modalState
	renameInput textinput.Model
	renameOrig  string
	importPath  string   // pending source file when renaming as part of import
	importName  string   // target name while confirming a config with hooks
	importHooks []string // hook lines shown in the confirm overlay
	connectName string
	spinner     spinner.Model
	filePicker  filepicker.Model

	// Help
	keys keyMap
	help help.Model

	// Feedback
	message    string
	messageExp time.Time // when to clear the message
}

// listLen returns the number of rows in the left panel list.
func (m model) listLen() int {
	return len(m.configs)
}

// selectedConfig returns the WireGuard config under the cursor, or "" if the
// list is empty.
func (m model) selectedConfig() string {
	if m.cursor < 0 || m.cursor >= len(m.configs) {
		return ""
	}
	return m.configs[m.cursor]
}

// isActive reports whether the named config is a currently active tunnel.
func (m model) isActive(name string) bool {
	return name != "" && slices.Contains(m.activeVPNs, name)
}

// refreshVPNState re-queries the active tunnel list and per-tunnel stats.
func (m *model) refreshVPNState() {
	m.activeVPNs = GetActiveVPNs()
	statuses := make(map[string]VPNStatus, len(m.activeVPNs))
	for _, name := range m.activeVPNs {
		if s, err := GetVPNStatus(name); err == nil {
			statuses[name] = s
		}
	}
	m.vpnStatus = statuses
}

func (m *model) restoreSelection(name string) {
	if name != "" {
		if i := slices.Index(m.configs, name); i >= 0 {
			m.cursor = i
			return
		}
	}
	if n := m.listLen(); m.cursor >= n {
		m.cursor = max(0, n-1)
	}
}

// Messages

type connectDoneMsg struct {
	name string
	err  error
}

// handshakeCheckMsg fires handshakeWait after a successful connect to
// check that the server actually answered.
type handshakeCheckMsg struct{ name string }

// handshakeWait covers WireGuard's 5-second handshake retry twice over.
const handshakeWait = 12 * time.Second

type disconnectDoneMsg struct {
	name string
	err  error
}

type importDoneMsg struct {
	name string
	err  error
}

type renameDoneMsg struct {
	oldName string
	newName string
	err     error
}

type deleteDoneMsg struct {
	name string
	err  error
}

type statusTickMsg time.Time

func statusTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return statusTickMsg(t)
	})
}

func newFilePicker() filepicker.Model {
	home, _ := os.UserHomeDir()
	fp := filepicker.New()
	fp.CurrentDirectory = home
	fp.AllowedTypes = []string{".conf", ".wg"}
	fp.KeyMap.Back = key.NewBinding(key.WithKeys("h", "backspace", "left"), key.WithHelp("h", "back"))
	return fp
}

func initialModel() model {
	ti := textinput.New()
	ti.Prompt = ""
	s := textinput.DefaultDarkStyles()
	s.Focused.Text = lipgloss.NewStyle().Foreground(green)
	s.Cursor.Color = accent
	ti.SetStyles(s)
	ti.CharLimit = 64

	sp := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(spinnerStyle),
	)

	m := model{
		renameInput: ti,
		spinner:     sp,
		filePicker:  newFilePicker(),
		keys:        newKeyMap(),
		help:        newHelp(),
	}
	m.refreshVPNState()
	m.configs = ListConfigs()
	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(statusTick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.SetWidth(msg.Width)
		return m, nil

	case statusTickMsg:
		sel := m.selectedConfig()
		oldLen := m.listLen()
		m.refreshVPNState()
		configs := ListConfigs()
		if len(configs) != len(m.configs) {
			m.configs = configs
		}
		if m.listLen() != oldLen {
			m.restoreSelection(sel)
		}
		return m, statusTick()

	case connectDoneMsg:
		m.modal = modalNone
		m.refreshVPNState()
		if msg.err != nil {
			m.setMessage(errorStyle.Render("  " + msg.err.Error()))
		} else {
			m.setMessage(connectedStyle.Render("  Connected to " + msg.name))
			m.configs = ListConfigs()
			name := msg.name
			return m, tea.Tick(handshakeWait, func(time.Time) tea.Msg { return handshakeCheckMsg{name} })
		}
		m.configs = ListConfigs()
		return m, nil

	case handshakeCheckMsg:
		if !m.isActive(msg.name) {
			return m, nil
		}
		if s, err := GetVPNStatus(msg.name); err == nil && s.Handshake == "" {
			logger.Warn("handshake", "name", msg.name, "result", "none after "+handshakeWait.String())
			m.setMessage(warnStyle.Render("  ⚠ " + msg.name + ": no reply from the server (check its endpoint and that it has this device's key)"))
		} else if err == nil {
			logger.Info("handshake", "name", msg.name, "latest", s.Handshake)
		}
		return m, nil

	case disconnectDoneMsg:
		m.refreshVPNState()
		if msg.err != nil {
			m.setMessage(errorStyle.Render("  Disconnect failed: " + msg.err.Error()))
		} else {
			m.setMessage(dimStyle.Render("  Disconnected " + msg.name))
		}
		return m, nil

	case importDoneMsg:
		m.importPath = ""
		if msg.err != nil {
			m.setMessage(errorStyle.Render("  Import failed: " + msg.err.Error()))
		} else {
			m.setMessage(connectedStyle.Render("  Imported " + msg.name))
			m.configs = ListConfigs()
			m.restoreSelection(msg.name)
		}
		return m, nil

	case renameDoneMsg:
		m.modal = modalNone
		if msg.err != nil {
			m.setMessage(errorStyle.Render("  Rename failed: " + msg.err.Error()))
		} else {
			m.setMessage(connectedStyle.Render("  Renamed " + msg.oldName + " → " + msg.newName))
			m.configs = ListConfigs()
		}
		return m, nil

	case deleteDoneMsg:
		m.modal = modalNone
		if msg.err != nil {
			m.setMessage(errorStyle.Render("  Delete failed: " + msg.err.Error()))
		} else {
			m.setMessage(dimStyle.Render("  Deleted " + msg.name))
			m.configs = ListConfigs()
			if m.cursor >= len(m.configs) && m.cursor > 0 {
				m.cursor--
			}
		}
		return m, nil

	case spinner.TickMsg:
		if m.modal == modalConnecting {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyPressMsg:
		// Clear message on any keypress
		m.message = ""

		// Global quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Help overlay dismisses on any key
		if m.modal == modalHelp {
			m.modal = modalNone
			m.help.ShowAll = false
			return m, nil
		}

		// Route to modal handlers
		switch m.modal {
		case modalRenaming:
			return m.updateRename(msg)
		case modalDeleting:
			return m.updateDelete(msg)
		case modalHookConfirm:
			return m.updateHookConfirm(msg)
		case modalConnecting:
			return m, nil // block input while connecting
		case modalImporting:
			return m.updateImport(msg)
		}

		// Normal mode keybindings
		return m.updateNormal(msg)
	}

	// Pass non-key messages to filepicker when importing (readDir responses)
	if m.modal == modalImporting {
		return m.updateImportMsg(msg)
	}

	return m, nil
}

func (m *model) updateNormal(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Down):
		if m.listLen() > 0 {
			m.cursor++
			if m.cursor >= m.listLen() {
				m.cursor = 0
			}
		}

	case key.Matches(msg, m.keys.Up):
		if m.listLen() > 0 {
			m.cursor--
			if m.cursor < 0 {
				m.cursor = m.listLen() - 1
			}
		}

	case key.Matches(msg, m.keys.Toggle):
		return m.toggleSelected()

	case key.Matches(msg, m.keys.Import):
		m.modal = modalImporting
		m.filePicker = newFilePicker()
		m.filePicker, _ = m.filePicker.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		return m, m.filePicker.Init()

	case key.Matches(msg, m.keys.Rename):
		selected := m.selectedConfig()
		if selected == "" {
			break
		}
		if m.isActive(selected) {
			m.setMessage(warnStyle.Render("  Disconnect before renaming"))
			break
		}
		m.importPath = ""
		return m, m.beginRename(selected)

	case key.Matches(msg, m.keys.Delete):
		selected := m.selectedConfig()
		if selected == "" {
			break
		}
		if m.isActive(selected) {
			m.setMessage(warnStyle.Render("  Disconnect before deleting"))
			break
		}
		m.modal = modalDeleting

	case key.Matches(msg, m.keys.Help):
		m.help.ShowAll = true
		m.modal = modalHelp
	}

	return m, nil
}

func (m *model) toggleSelected() (tea.Model, tea.Cmd) {
	selected := m.selectedConfig()
	if selected == "" {
		return m, nil
	}
	if m.isActive(selected) {
		return m, func() tea.Msg {
			err := DisconnectVPN(selected)
			return disconnectDoneMsg{name: selected, err: err}
		}
	}
	if err := interfaceNameError(selected); err != nil {
		m.setMessage(warnStyle.Render("  " + err.Error() + " — rename before connecting"))
		m.importPath = ""
		return m, m.beginRename(selected)
	}
	m.modal = modalConnecting
	m.connectName = selected
	active := slices.Clone(m.activeVPNs)
	return m, tea.Batch(m.spinner.Tick, func() tea.Msg {
		// Only tunnels whose AllowedIPs overlap the new config need to
		// come down first; disjoint tunnels stay connected alongside it.
		for _, name := range conflictingVPNs(selected, active) {
			if err := DisconnectVPN(name); err != nil {
				return connectDoneMsg{name: selected, err: err}
			}
		}
		err := ConnectVPN(selected)
		return connectDoneMsg{name: selected, err: err}
	})
}

func (m *model) updateImport(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.modal = modalNone
		return m, nil
	}

	// Let filepicker handle the key
	var cmd tea.Cmd
	m.filePicker, cmd = m.filePicker.Update(msg)

	// Check if a file was selected
	if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
		m.modal = modalNone
		return m, m.handleFileSelected(path)
	}

	// Check if user tried to select a disabled file
	if didSelect, _ := m.filePicker.DidSelectDisabledFile(msg); didSelect {
		m.setMessage(errorStyle.Render("  Only .conf and .wg files supported"))
		return m, cmd
	}

	return m, cmd
}

func (m *model) updateImportMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.filePicker, cmd = m.filePicker.Update(msg)
	return m, cmd
}

func (m *model) handleFileSelected(path string) tea.Cmd {
	if !ValidateConfig(path) {
		m.setMessage(errorStyle.Render("  Invalid config: missing [Interface]"))
		return nil
	}

	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	name = strings.ReplaceAll(name, " ", "-")
	name = sanitizeName(name)
	if err := interfaceNameError(name); err != nil {
		m.importPath = path
		m.setMessage(warnStyle.Render("  " + err.Error() + " — enter a new name"))
		return m.beginRename(name)
	}

	return m.startImport(path, name)
}

// startImport imports a config under a validated name. Configs with root
// hook commands stop at a confirm overlay that lists them; everything else
// goes straight to the passwordless helper.
func (m *model) startImport(path, name string) tea.Cmd {
	if slices.Contains(m.configs, name) {
		logger.Info("import", "name", name, "result", "refused: name exists")
		m.setMessage(errorStyle.Render("  '" + name + "' already exists — delete or rename it first"))
		return nil
	}
	if hooks := ConfigHooks(path); len(hooks) > 0 {
		m.modal = modalHookConfirm
		m.importPath, m.importName, m.importHooks = path, name, hooks
		return nil
	}
	return func() tea.Msg {
		err := ImportConfig(path, name)
		return importDoneMsg{name: name, err: err}
	}
}

func (m *model) updateHookConfirm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	path, name := m.importPath, m.importName
	m.modal = modalNone
	m.importPath, m.importName, m.importHooks = "", "", nil
	if msg.String() != "y" && msg.String() != "Y" {
		logger.Info("import", "name", name, "result", "declined: has root hooks")
		m.setMessage(dimStyle.Render("  Import cancelled"))
		return m, nil
	}
	// Suspends the TUI so sudo can ask for the password in the terminal.
	return m, tea.ExecProcess(installWithHooksCmd(path, name), func(err error) tea.Msg {
		logAction("import (with hooks, authenticated)", name, err)
		return importDoneMsg{name: name, err: err}
	})
}

func (m *model) beginRename(orig string) tea.Cmd {
	m.modal = modalRenaming
	m.renameOrig = orig
	m.renameInput.SetValue(suggestInterfaceName(orig))
	m.renameInput.Focus()
	m.renameInput.CursorEnd()
	return textinput.Blink
}

func (m *model) updateRename(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.modal = modalNone
		m.renameInput.Blur()
		m.importPath = ""
		return m, nil

	case "enter":
		newName := sanitizeName(m.renameInput.Value())
		if newName == "" {
			m.setMessage(errorStyle.Render("  Name cannot be empty"))
			return m, nil
		}
		if err := interfaceNameError(newName); err != nil {
			m.setMessage(errorStyle.Render("  " + err.Error()))
			return m, nil
		}
		if m.importPath == "" && newName == m.renameOrig {
			m.modal = modalNone
			m.renameInput.Blur()
			return m, nil
		}
		for _, cfg := range m.configs {
			if cfg == newName {
				m.setMessage(errorStyle.Render("  '" + newName + "' already exists"))
				return m, nil
			}
		}
		m.renameInput.Blur()
		if path := m.importPath; path != "" {
			m.modal = modalNone
			m.importPath = ""
			return m, m.startImport(path, newName)
		}
		oldName := m.renameOrig
		return m, func() tea.Msg {
			err := RenameConfig(oldName, newName)
			return renameDoneMsg{oldName: oldName, newName: newName, err: err}
		}
	}

	// Pass other keys to textinput
	var cmd tea.Cmd
	m.renameInput, cmd = m.renameInput.Update(msg)
	return m, cmd
}

func (m *model) updateDelete(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		name := m.selectedConfig()
		if name == "" {
			m.modal = modalNone
			return m, nil
		}
		m.modal = modalNone
		return m, func() tea.Msg {
			err := RemoveConfig(name)
			return deleteDoneMsg{name: name, err: err}
		}
	default:
		m.modal = modalNone
	}
	return m, nil
}

func (m *model) setMessage(msg string) {
	m.message = msg
	m.messageExp = time.Now().Add(3 * time.Second)
}

func sanitizeName(name string) string {
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	result := b.String()
	if result == "" {
		result = "imported"
	}
	return result
}

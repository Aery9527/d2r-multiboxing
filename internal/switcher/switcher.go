package switcher

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"d2rhl/internal/common/config"
	"d2rhl/internal/common/d2r"
	"d2rhl/internal/common/process"

	"golang.org/x/sys/windows"
)

// Windows API
var (
	user32 = windows.NewLazySystemDLL("user32.dll")

	procRegisterHotKey      = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey    = user32.NewProc("UnregisterHotKey")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procPostThreadMessageW  = user32.NewProc("PostThreadMessageW")
	procSetWindowsHookExW   = user32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")
)

// Windows constants
const (
	wmHotkey      = 0x0312
	wmQuit        = 0x0012
	wmKeyDown     = 0x0100
	wmSysKeyDown  = 0x0104
	wmXButtonDown = 0x020B
	wmAppSwitch   = 0x8001 // custom message: mouse hook → message pump

	whKeyboardLL = 13
	whMouseLL    = 14

	hotkeyID = 1
)

// msg is the Windows MSG structure.
type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

// kbdllHookStruct is the Windows KBDLLHOOKSTRUCT structure.
type kbdllHookStruct struct {
	VkCode    uint32
	ScanCode  uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

// msllHookStruct is the Windows MSLLHOOKSTRUCT structure.
type msllHookStruct struct {
	Pt        struct{ X, Y int32 }
	MouseData uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

var (
	mu sync.Mutex

	// Monitor state (outer): whether the switcher service is running.
	monitorRunning bool
	monitorStop    chan struct{}

	// Inner state: the active hotkey/hook, set by startHotkey/startMouseHook/startGamepadPoll.
	innerStopFn  func()
	innerRunning bool

	excludedAccounts []string // display names excluded from the switch cycle
)

// UpdateExcludedAccounts updates the set of account display names that should be
// skipped during window cycling. Safe to call while the switcher is running.
func UpdateExcludedAccounts(names []string) {
	mu.Lock()
	excludedAccounts = names
	mu.Unlock()
}

// IsRunning returns whether the switcher service is active (i.e., Start has been called).
// The underlying hotkey/hook may not be registered if no D2R windows are currently open.
func IsRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return monitorRunning
}

// Start begins the switcher service. Instead of registering the hotkey immediately, it
// launches a background monitor that only registers the hotkey/hook when D2R windows are
// detected. This prevents the hotkey from consuming keypresses in the CLI when no game
// windows are open.
func Start(cfg *config.SwitcherConfig) error {
	if cfg == nil || !cfg.Enabled || cfg.Key == "" {
		return nil
	}

	mu.Lock()
	if monitorRunning {
		mu.Unlock()
		return fmt.Errorf("switcher already running")
	}
	stop := make(chan struct{})
	monitorStop = stop
	monitorRunning = true
	mu.Unlock()

	go runMonitor(cfg, stop)
	return nil
}

// Stop shuts down the switcher service and releases any active hotkey/hook.
func Stop() {
	mu.Lock()
	if !monitorRunning {
		mu.Unlock()
		return
	}
	close(monitorStop)
	monitorRunning = false
	mu.Unlock()
}

// runMonitor polls for D2R windows every 500 ms and dynamically registers or unregisters
// the hotkey/hook based on whether any game windows are present.
func runMonitor(cfg *config.SwitcherConfig, stop <-chan struct{}) {
	checkAndUpdate := func() {
		hasWindows := len(process.FindWindowsByTitlePrefix(d2r.WindowTitlePrefix)) > 0
		mu.Lock()
		active := innerRunning
		mu.Unlock()

		if hasWindows && !active {
			startInner(cfg)
		} else if !hasWindows && active {
			stopInner()
		}
	}

	// Check immediately so the hotkey is active without waiting for the first tick.
	checkAndUpdate()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			stopInner()
			return
		case <-ticker.C:
			checkAndUpdate()
		}
	}
}

// startInner registers the configured hotkey/hook. Errors are silently ignored because
// registration happens asynchronously relative to the caller of Start.
func startInner(cfg *config.SwitcherConfig) {
	var err error
	switch {
	case IsMouseButton(cfg.Key):
		err = startMouseHook(MouseButtonID(cfg.Key), switchToNext)
	case IsGamepadButton(cfg.Key):
		var gamepadMods []string
		for _, m := range cfg.Modifiers {
			if IsGamepadButton(m) {
				gamepadMods = append(gamepadMods, m)
			}
		}
		err = startGamepadPoll(cfg.GamepadIndex, gamepadMods, cfg.Key, switchToNext)
	default:
		vk, ok := KeyToVK(cfg.Key)
		if !ok {
			return
		}
		modFlags := ModifiersToFlags(cfg.Modifiers)
		err = startHotkey(modFlags, vk, switchToNext)
	}
	_ = err
}

// stopInner stops the active hotkey/hook if one is running.
func stopInner() {
	mu.Lock()
	defer mu.Unlock()
	if innerStopFn != nil {
		innerStopFn()
		innerStopFn = nil
		innerRunning = false
	}
}

// switchToNext finds all D2R windows and switches focus to the next one,
// skipping accounts listed in excludedAccounts.
func switchToNext() {
	hwnds := process.FindWindowsByTitlePrefix(d2r.WindowTitlePrefix)

	mu.Lock()
	excluded := make(map[string]struct{}, len(excludedAccounts))
	for _, name := range excludedAccounts {
		excluded[d2r.WindowTitle(name)] = struct{}{}
	}
	mu.Unlock()

	if len(excluded) > 0 {
		filtered := hwnds[:0]
		for _, hwnd := range hwnds {
			if _, skip := excluded[process.GetWindowTitle(hwnd)]; !skip {
				filtered = append(filtered, hwnd)
			}
		}
		hwnds = filtered
	}

	if len(hwnds) == 0 {
		return
	}

	if len(hwnds) == 1 {
		_ = process.SwitchToWindow(hwnds[0])
		return
	}

	fg := process.GetForegroundHwnd()
	nextIdx := 0
	for i, hwnd := range hwnds {
		if hwnd == fg {
			nextIdx = (i + 1) % len(hwnds)
			break
		}
	}

	_ = process.SwitchToWindow(hwnds[nextIdx])
}

// getMessage wraps the GetMessageW call.
func getMessage(m *msg) uintptr {
	ret, _, _ := procGetMessageW.Call(
		uintptr(unsafe.Pointer(m)), 0, 0, 0,
	)
	return ret
}

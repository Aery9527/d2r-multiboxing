package switcher

import (
	"fmt"
	"sync"
	"unsafe"

	"d2r-multiboxing/internal/config"
	"d2r-multiboxing/internal/d2r"
	"d2r-multiboxing/internal/process"

	"golang.org/x/sys/windows"
)

// Windows API
var (
	user32 = windows.NewLazySystemDLL("user32.dll")

	procRegisterHotKey     = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey   = user32.NewProc("UnregisterHotKey")
	procGetMessageW        = user32.NewProc("GetMessageW")
	procPostThreadMessageW = user32.NewProc("PostThreadMessageW")
	procSetWindowsHookExW  = user32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procCallNextHookEx     = user32.NewProc("CallNextHookEx")
	procGetAsyncKeyState   = user32.NewProc("GetAsyncKeyState")
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
	mu       sync.Mutex
	stopFunc func()
	running  bool
)

// IsRunning returns whether the switcher is currently active.
func IsRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return running
}

// Start begins listening for the configured hotkey/mouse button to switch D2R windows.
func Start(cfg *config.SwitcherConfig) error {
	if cfg == nil || !cfg.Enabled || cfg.Key == "" {
		return nil
	}

	mu.Lock()
	if running {
		mu.Unlock()
		return fmt.Errorf("switcher already running")
	}
	mu.Unlock()

	if IsMouseButton(cfg.Key) {
		buttonID := MouseButtonID(cfg.Key)
		return startMouseHook(buttonID, switchToNext)
	}

	if IsGamepadButton(cfg.Key) {
		// 從 Modifiers 中篩出搖桿修飾鍵
		var gamepadMods []string
		for _, m := range cfg.Modifiers {
			if IsGamepadButton(m) {
				gamepadMods = append(gamepadMods, m)
			}
		}
		return startGamepadPoll(cfg.GamepadIndex, gamepadMods, cfg.Key, switchToNext)
	}

	vk, ok := KeyToVK(cfg.Key)
	if !ok {
		return fmt.Errorf("unknown key: %s", cfg.Key)
	}
	modFlags := ModifiersToFlags(cfg.Modifiers)
	return startHotkey(modFlags, vk, switchToNext)
}

// Stop stops the switcher and releases resources.
func Stop() {
	mu.Lock()
	defer mu.Unlock()
	if stopFunc != nil {
		stopFunc()
		stopFunc = nil
		running = false
	}
}

// switchToNext finds all D2R windows and switches focus to the next one.
func switchToNext() {
	hwnds := process.FindWindowsByTitlePrefix(d2r.WindowTitlePrefix)
	if len(hwnds) < 2 {
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

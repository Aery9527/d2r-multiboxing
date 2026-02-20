package process

import (
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                    = windows.NewLazySystemDLL("user32.dll")
	procEnumWindows           = user32.NewProc("EnumWindows")
	procGetWindowThreadProcID = user32.NewProc("GetWindowThreadProcessId")
	procGetWindowTextW        = user32.NewProc("GetWindowTextW")
	procSetWindowTextW        = user32.NewProc("SetWindowTextW")
	procIsWindowVisible       = user32.NewProc("IsWindowVisible")
	procGetForegroundWindow   = user32.NewProc("GetForegroundWindow")
	procSetForegroundWindow   = user32.NewProc("SetForegroundWindow")
	procKeybdEvent            = user32.NewProc("keybd_event")
)

// FindWindowByTitle returns true if a visible top-level window with the exact title exists.
func FindWindowByTitle(title string) bool {
	found := false
	buf := make([]uint16, 512)

	cb := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		visible, _, _ := procIsWindowVisible.Call(hwnd)
		if visible == 0 {
			return 1
		}
		n, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		if n > 0 && syscall.UTF16ToString(buf[:n]) == title {
			found = true
			return 0
		}
		return 1
	})

	procEnumWindows.Call(cb, 0)
	return found
}

// It retries up to maxRetries times with retryInterval between attempts,
// because D2R window may not be immediately available after launch.
func RenameWindow(pid uint32, newTitle string, maxRetries int, retryInterval time.Duration) error {
	for attempt := 0; attempt <= maxRetries; attempt++ {
		hwnd, err := findWindowByPID(pid)
		if err == nil && hwnd != 0 {
			return setWindowText(hwnd, newTitle)
		}
		if attempt < maxRetries {
			time.Sleep(retryInterval)
		}
	}
	return fmt.Errorf("window not found for PID %d after %d retries", pid, maxRetries)
}

// findWindowByPID enumerates top-level windows to find one owned by the given PID.
func findWindowByPID(targetPID uint32) (windows.Handle, error) {
	var foundHwnd windows.Handle

	// EnumWindows callback: return 1 to continue, 0 to stop
	cb := syscall.NewCallback(func(hwnd uintptr, lParam uintptr) uintptr {
		visible, _, _ := procIsWindowVisible.Call(hwnd)
		if visible == 0 {
			return 1 // 跳過不可見視窗
		}

		var pid uint32
		procGetWindowThreadProcID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
		if pid == targetPID {
			foundHwnd = windows.Handle(hwnd)
			return 0 // 找到了，停止列舉
		}
		return 1
	})

	procEnumWindows.Call(cb, 0)

	if foundHwnd == 0 {
		return 0, fmt.Errorf("no visible window found for PID %d", targetPID)
	}

	return foundHwnd, nil
}

// setWindowText sets the title of a window.
func setWindowText(hwnd windows.Handle, text string) error {
	textPtr, err := syscall.UTF16PtrFromString(text)
	if err != nil {
		return fmt.Errorf("failed to convert text: %w", err)
	}

	ret, _, err := procSetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(textPtr)))
	if ret == 0 {
		return fmt.Errorf("SetWindowTextW failed: %w", err)
	}

	return nil
}

// FindWindowsByTitlePrefix returns handles of all visible windows whose title starts with prefix.
func FindWindowsByTitlePrefix(prefix string) []windows.Handle {
	var handles []windows.Handle
	buf := make([]uint16, 512)

	cb := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		visible, _, _ := procIsWindowVisible.Call(hwnd)
		if visible == 0 {
			return 1
		}
		n, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		if n > 0 && strings.HasPrefix(syscall.UTF16ToString(buf[:n]), prefix) {
			handles = append(handles, windows.Handle(hwnd))
		}
		return 1
	})

	procEnumWindows.Call(cb, 0)
	return handles
}

// GetForegroundHwnd returns the handle of the current foreground window.
func GetForegroundHwnd() windows.Handle {
	hwnd, _, _ := procGetForegroundWindow.Call()
	return windows.Handle(hwnd)
}

const (
	vkMenu         = 0x12
	keyeventfKeyup = 0x0002
)

// SwitchToWindow sets the given window as the foreground window.
// Falls back to simulating an Alt key press if the direct call fails.
func SwitchToWindow(hwnd windows.Handle) error {
	ret, _, _ := procSetForegroundWindow.Call(uintptr(hwnd))
	if ret != 0 {
		return nil
	}

	// Fallback: simulate Alt key press to unlock foreground switching
	procKeybdEvent.Call(vkMenu, 0, 0, 0)
	procKeybdEvent.Call(vkMenu, 0, keyeventfKeyup, 0)

	ret, _, _ = procSetForegroundWindow.Call(uintptr(hwnd))
	if ret != 0 {
		return nil
	}

	return fmt.Errorf("SetForegroundWindow failed for hwnd %v", hwnd)
}

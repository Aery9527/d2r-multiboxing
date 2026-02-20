package switcher

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// DetectKeyPress waits for the user to press a key combination (keyboard or mouse side button).
// Returns the modifiers and key name. If the user presses Escape, returns empty key (cancellation).
func DetectKeyPress() (modifiers []string, key string, err error) {
	type result struct {
		modifiers []string
		key       string
	}

	resultCh := make(chan result, 1)
	errCh := make(chan error, 1)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		threadID := windows.GetCurrentThreadId()
		var detected result

		// Keyboard hook callback
		kbCallback := syscall.NewCallback(func(nCode uintptr, wParam uintptr, lParam uintptr) uintptr {
			if int32(nCode) >= 0 && (wParam == wmKeyDown || wParam == wmSysKeyDown) {
				data := (*kbdllHookStruct)(unsafe.Pointer(lParam))
				vk := data.VkCode

				// Escape cancels detection
				if vk == vkEscape {
					procPostThreadMessageW.Call(uintptr(threadID), wmQuit, 0, 0)
					ret, _, _ := procCallNextHookEx.Call(0, nCode, wParam, lParam)
					return ret
				}

				// Skip modifier-only keys
				if !isModifierKey(vk) {
					if isKeyPressed(vkControl) {
						detected.modifiers = append(detected.modifiers, "ctrl")
					}
					if isKeyPressed(vkShift) {
						detected.modifiers = append(detected.modifiers, "shift")
					}
					if isKeyPressed(vkMenu) {
						detected.modifiers = append(detected.modifiers, "alt")
					}

					if name, ok := VKToKeyName(vk); ok {
						detected.key = name
					} else {
						detected.key = fmt.Sprintf("VK_0x%02X", vk)
					}

					procPostThreadMessageW.Call(uintptr(threadID), wmQuit, 0, 0)
				}
			}
			ret, _, _ := procCallNextHookEx.Call(0, nCode, wParam, lParam)
			return ret
		})

		// Mouse hook callback
		msCallback := syscall.NewCallback(func(nCode uintptr, wParam uintptr, lParam uintptr) uintptr {
			if int32(nCode) >= 0 && wParam == wmXButtonDown {
				data := (*msllHookStruct)(unsafe.Pointer(lParam))
				button := uint16(data.MouseData >> 16)
				if button == xButton1 {
					detected.key = "XButton1"
					procPostThreadMessageW.Call(uintptr(threadID), wmQuit, 0, 0)
				} else if button == xButton2 {
					detected.key = "XButton2"
					procPostThreadMessageW.Call(uintptr(threadID), wmQuit, 0, 0)
				}
			}
			ret, _, _ := procCallNextHookEx.Call(0, nCode, wParam, lParam)
			return ret
		})

		// Install hooks
		kbHook, _, kbErr := procSetWindowsHookExW.Call(whKeyboardLL, kbCallback, 0, 0)
		if kbHook == 0 {
			errCh <- fmt.Errorf("install keyboard hook failed: %v", kbErr)
			return
		}
		defer procUnhookWindowsHookEx.Call(kbHook)

		msHook, _, msErr := procSetWindowsHookExW.Call(whMouseLL, msCallback, 0, 0)
		if msHook == 0 {
			errCh <- fmt.Errorf("install mouse hook failed: %v", msErr)
			return
		}
		defer procUnhookWindowsHookEx.Call(msHook)

		errCh <- nil // hooks installed successfully

		// Message pump — runs until WM_QUIT is posted
		var m msg
		for {
			ret := getMessage(&m)
			if ret == 0 || ret == ^uintptr(0) {
				break
			}
		}

		resultCh <- detected
	}()

	if err := <-errCh; err != nil {
		return nil, "", err
	}

	r := <-resultCh
	return r.modifiers, r.key, nil
}

// isKeyPressed checks if a key is currently pressed using GetAsyncKeyState.
func isKeyPressed(vk uint32) bool {
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return ret&0x8000 != 0
}

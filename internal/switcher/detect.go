package switcher

import (
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// DetectKeyPress waits for the user to press a key combination (keyboard, mouse side button, or gamepad button).
// Returns the modifiers, key name, and gamepad controller index.
// If the user presses Escape, returns empty key (cancellation).
func DetectKeyPress() (modifiers []string, key string, gamepadIndex int, err error) {
	type result struct {
		modifiers    []string
		key          string
		gamepadIndex int
	}

	resultCh := make(chan result, 1)
	errCh := make(chan error, 1)
	stopGamepad := make(chan struct{})
	var once sync.Once
	var hookThreadID uint32

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		hookThreadID = windows.GetCurrentThreadId()

		// Keyboard hook callback
		kbCallback := syscall.NewCallback(func(nCode uintptr, wParam uintptr, lParam uintptr) uintptr {
			if int32(nCode) >= 0 && (wParam == wmKeyDown || wParam == wmSysKeyDown) {
				data := (*kbdllHookStruct)(unsafe.Pointer(lParam))
				vk := data.VkCode

				// Escape cancels detection
				if vk == vkEscape {
					once.Do(func() { resultCh <- result{} })
					procPostThreadMessageW.Call(uintptr(hookThreadID), wmQuit, 0, 0)
					ret, _, _ := procCallNextHookEx.Call(0, nCode, wParam, lParam)
					return ret
				}

				// Skip modifier-only keys
				if !isModifierKey(vk) {
					var r result
					if isKeyPressed(vkControl) {
						r.modifiers = append(r.modifiers, "ctrl")
					}
					if isKeyPressed(vkShift) {
						r.modifiers = append(r.modifiers, "shift")
					}
					if isKeyPressed(vkMenu) {
						r.modifiers = append(r.modifiers, "alt")
					}

					if name, ok := VKToKeyName(vk); ok {
						r.key = name
					} else {
						r.key = fmt.Sprintf("VK_0x%02X", vk)
					}

					once.Do(func() { resultCh <- r })
					procPostThreadMessageW.Call(uintptr(hookThreadID), wmQuit, 0, 0)
					// 吞掉該按鍵，避免字元殘留到 stdin
					return 1
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
				var r result
				if button == xButton1 {
					r.key = "XButton1"
				} else if button == xButton2 {
					r.key = "XButton2"
				}
				if r.key != "" {
					once.Do(func() { resultCh <- r })
					procPostThreadMessageW.Call(uintptr(hookThreadID), wmQuit, 0, 0)
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
	}()

	if err := <-errCh; err != nil {
		return nil, "", 0, err
	}

	// 搖桿偵測（與鍵盤/滑鼠同時進行）
	go func() {
		idx, mods, btn := detectGamepadButtonPress(stopGamepad)
		if btn != "" {
			once.Do(func() {
				resultCh <- result{modifiers: mods, key: btn, gamepadIndex: idx}
			})
			procPostThreadMessageW.Call(uintptr(hookThreadID), wmQuit, 0, 0)
		}
	}()

	r := <-resultCh
	close(stopGamepad)
	return r.modifiers, r.key, r.gamepadIndex, nil
}

// isKeyPressed checks if a key is currently pressed using GetAsyncKeyState.
func isKeyPressed(vk uint32) bool {
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return ret&0x8000 != 0
}

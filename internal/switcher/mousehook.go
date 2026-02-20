package switcher

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// startMouseHook installs a low-level mouse hook to listen for a specific XButton.
func startMouseHook(targetButton uint16, onTrigger func()) error {
	errCh := make(chan error, 1)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		threadID := windows.GetCurrentThreadId()

		callback := syscall.NewCallback(func(nCode uintptr, wParam uintptr, lParam uintptr) uintptr {
			if int32(nCode) >= 0 && wParam == wmXButtonDown {
				data := (*msllHookStruct)(unsafe.Pointer(lParam))
				button := uint16(data.MouseData >> 16) // HIWORD
				if button == targetButton {
					procPostThreadMessageW.Call(uintptr(threadID), wmAppSwitch, 0, 0)
				}
			}
			ret, _, _ := procCallNextHookEx.Call(0, nCode, wParam, lParam)
			return ret
		})

		hook, _, err := procSetWindowsHookExW.Call(whMouseLL, callback, 0, 0)
		if hook == 0 {
			errCh <- fmt.Errorf("SetWindowsHookExW(WH_MOUSE_LL) failed: %v", err)
			return
		}

		mu.Lock()
		stopFunc = func() {
			procPostThreadMessageW.Call(uintptr(threadID), wmQuit, 0, 0)
		}
		running = true
		mu.Unlock()
		errCh <- nil

		var m msg
		for {
			ret := getMessage(&m)
			if ret == 0 || ret == ^uintptr(0) {
				break
			}
			if m.Message == wmAppSwitch {
				onTrigger()
			}
		}

		procUnhookWindowsHookEx.Call(hook)
	}()

	return <-errCh
}

package switcher

import (
	"fmt"
	"runtime"

	"golang.org/x/sys/windows"
)

// startHotkey registers a global hotkey and listens for it in a message loop.
func startHotkey(modifiers uint32, vk uint32, onTrigger func()) error {
	errCh := make(chan error, 1)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		threadID := windows.GetCurrentThreadId()

		ret, _, err := procRegisterHotKey.Call(0, hotkeyID, uintptr(modifiers|modNoRepeat), uintptr(vk))
		if ret == 0 {
			errCh <- fmt.Errorf("RegisterHotKey failed (key may be in use by another program): %v", err)
			return
		}

		stopFunc = func() {
			procPostThreadMessageW.Call(uintptr(threadID), wmQuit, 0, 0)
		}
		running = true
		errCh <- nil

		var m msg
		for {
			ret := getMessage(&m)
			if ret == 0 || ret == ^uintptr(0) {
				break
			}
			if m.Message == wmHotkey {
				onTrigger()
			}
		}

		procUnregisterHotKey.Call(0, hotkeyID)
	}()

	return <-errCh
}

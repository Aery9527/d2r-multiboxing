package switcher

import (
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// XInput DLL and functions
var (
	xinput             = windows.NewLazySystemDLL("xinput1_4.dll")
	procXInputGetState = xinput.NewProc("XInputGetState")
)

// XInput constants
const (
	xinputMaxControllers = 4
	triggerThreshold     = 128
)

// XInput button masks
const (
	xinputGamepadDPadUp        uint16 = 0x0001
	xinputGamepadDPadDown      uint16 = 0x0002
	xinputGamepadDPadLeft      uint16 = 0x0004
	xinputGamepadDPadRight     uint16 = 0x0008
	xinputGamepadStart         uint16 = 0x0010
	xinputGamepadBack          uint16 = 0x0020
	xinputGamepadLeftThumb     uint16 = 0x0040
	xinputGamepadRightThumb    uint16 = 0x0080
	xinputGamepadLeftShoulder  uint16 = 0x0100
	xinputGamepadRightShoulder uint16 = 0x0200
	xinputGamepadA             uint16 = 0x1000
	xinputGamepadB             uint16 = 0x2000
	xinputGamepadX             uint16 = 0x4000
	xinputGamepadY             uint16 = 0x8000
)

// xinputGamepadState is the XINPUT_GAMEPAD structure.
type xinputGamepadState struct {
	Buttons      uint16
	LeftTrigger  byte
	RightTrigger byte
	ThumbLX      int16
	ThumbLY      int16
	ThumbRX      int16
	ThumbRY      int16
}

// xinputState is the XINPUT_STATE structure.
type xinputState struct {
	PacketNumber uint32
	Gamepad      xinputGamepadState
}

// gamepadButtonMasks maps button names to XInput button mask values.
var gamepadButtonMasks = map[string]uint16{
	"Gamepad_A":         xinputGamepadA,
	"Gamepad_B":         xinputGamepadB,
	"Gamepad_X":         xinputGamepadX,
	"Gamepad_Y":         xinputGamepadY,
	"Gamepad_LB":        xinputGamepadLeftShoulder,
	"Gamepad_RB":        xinputGamepadRightShoulder,
	"Gamepad_Back":      xinputGamepadBack,
	"Gamepad_Start":     xinputGamepadStart,
	"Gamepad_LS":        xinputGamepadLeftThumb,
	"Gamepad_RS":        xinputGamepadRightThumb,
	"Gamepad_DPadUp":    xinputGamepadDPadUp,
	"Gamepad_DPadDown":  xinputGamepadDPadDown,
	"Gamepad_DPadLeft":  xinputGamepadDPadLeft,
	"Gamepad_DPadRight": xinputGamepadDPadRight,
}

// gamepadMaskToName maps XInput button masks to button names (for detection).
var gamepadMaskToName map[uint16]string

func init() {
	gamepadMaskToName = make(map[uint16]string, len(gamepadButtonMasks))
	for name, mask := range gamepadButtonMasks {
		gamepadMaskToName[mask] = name
	}
}

// XInputAvailable checks if the XInput DLL can be loaded.
func XInputAvailable() bool {
	return xinput.Load() == nil
}

// getXInputState retrieves the state of an XInput controller.
func getXInputState(index int) (*xinputState, bool) {
	var state xinputState
	ret, _, _ := procXInputGetState.Call(
		uintptr(index),
		uintptr(unsafe.Pointer(&state)),
	)
	return &state, ret == 0
}

// GamepadButtonMask returns the XInput button mask for a gamepad key name.
func GamepadButtonMask(key string) uint16 {
	return gamepadButtonMasks[key]
}

// detectGamepadButtonPress polls all XInput controllers for a new button press.
// Returns controller index and button name when detected.
// Returns (-1, "") when stop channel is closed or XInput is unavailable.
func detectGamepadButtonPress(stop <-chan struct{}) (int, string) {
	if !XInputAvailable() {
		<-stop
		return -1, ""
	}

	type controllerSnapshot struct {
		connected    bool
		buttons      uint16
		leftTrigger  bool
		rightTrigger bool
	}

	// 讀取初始狀態，避免偵測到已按住的按鈕
	var prev [xinputMaxControllers]controllerSnapshot
	for i := 0; i < xinputMaxControllers; i++ {
		if state, ok := getXInputState(i); ok {
			prev[i] = controllerSnapshot{
				connected:    true,
				buttons:      state.Gamepad.Buttons,
				leftTrigger:  state.Gamepad.LeftTrigger >= triggerThreshold,
				rightTrigger: state.Gamepad.RightTrigger >= triggerThreshold,
			}
		}
	}

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return -1, ""
		case <-ticker.C:
			for i := 0; i < xinputMaxControllers; i++ {
				state, ok := getXInputState(i)
				if !ok {
					prev[i].connected = false
					continue
				}

				if !prev[i].connected {
					// 剛連接，設定基準值
					prev[i] = controllerSnapshot{
						connected:    true,
						buttons:      state.Gamepad.Buttons,
						leftTrigger:  state.Gamepad.LeftTrigger >= triggerThreshold,
						rightTrigger: state.Gamepad.RightTrigger >= triggerThreshold,
					}
					continue
				}

				// 偵測新按下的按鈕（邊緣觸發）
				newButtons := state.Gamepad.Buttons & ^prev[i].buttons
				if newButtons != 0 {
					for mask, name := range gamepadMaskToName {
						if newButtons&mask != 0 {
							return i, name
						}
					}
				}

				// 偵測左扳機
				ltPressed := state.Gamepad.LeftTrigger >= triggerThreshold
				if ltPressed && !prev[i].leftTrigger {
					return i, "Gamepad_LT"
				}

				// 偵測右扳機
				rtPressed := state.Gamepad.RightTrigger >= triggerThreshold
				if rtPressed && !prev[i].rightTrigger {
					return i, "Gamepad_RT"
				}

				prev[i].buttons = state.Gamepad.Buttons
				prev[i].leftTrigger = ltPressed
				prev[i].rightTrigger = rtPressed
			}
		}
	}
}

// startGamepadPoll starts polling a specific controller button and calls onTrigger on each press.
func startGamepadPoll(controllerIndex int, key string, onTrigger func()) error {
	if !XInputAvailable() {
		return fmt.Errorf("XInput 不可用（缺少 xinput1_4.dll）")
	}

	isLT := key == "Gamepad_LT"
	isRT := key == "Gamepad_RT"
	buttonMask := GamepadButtonMask(key)

	if !isLT && !isRT && buttonMask == 0 {
		return fmt.Errorf("unknown gamepad button: %s", key)
	}

	errCh := make(chan error, 1)

	go func() {
		stopCh := make(chan struct{})

		stopFunc = func() {
			close(stopCh)
		}
		running = true
		errCh <- nil

		var wasPressed bool
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				state, ok := getXInputState(controllerIndex)
				if !ok {
					wasPressed = false
					continue
				}

				var pressed bool
				if isLT {
					pressed = state.Gamepad.LeftTrigger >= triggerThreshold
				} else if isRT {
					pressed = state.Gamepad.RightTrigger >= triggerThreshold
				} else {
					pressed = state.Gamepad.Buttons&buttonMask != 0
				}

				if pressed && !wasPressed {
					onTrigger()
				}
				wasPressed = pressed
			}
		}
	}()

	return <-errCh
}

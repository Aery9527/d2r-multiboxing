package switcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyToVK(t *testing.T) {
	tests := []struct {
		name string
		key  string
		vk   uint32
		ok   bool
	}{
		{"Tab", "Tab", 0x09, true},
		{"F1", "F1", 0x70, true},
		{"F12", "F12", 0x7B, true},
		{"A", "A", 0x41, true},
		{"Z", "Z", 0x5A, true},
		{"0", "0", 0x30, true},
		{"9", "9", 0x39, true},
		{"Backtick", "`", 0xC0, true},
		{"Space", "Space", 0x20, true},
		{"Unknown", "XButton1", 0, false},
		{"Numpad0", "Num0", 0x60, true},
		{"Numpad9", "Num9", 0x69, true},
		{"Left", "Left", 0x25, true},
		{"Delete", "Delete", 0x2E, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vk, ok := KeyToVK(tt.key)
			assert.Equal(t, tt.ok, ok)
			if ok {
				assert.Equal(t, tt.vk, vk)
			}
		})
	}
}

func TestVKToKeyName(t *testing.T) {
	name, ok := VKToKeyName(0x09)
	assert.True(t, ok)
	assert.Equal(t, "Tab", name)

	_, ok = VKToKeyName(0xFFFF)
	assert.False(t, ok)
}

func TestModifiersToFlags(t *testing.T) {
	assert.Equal(t, uint32(0), ModifiersToFlags(nil))
	assert.Equal(t, modControl, ModifiersToFlags([]string{"ctrl"}))
	assert.Equal(t, modControl|modAlt, ModifiersToFlags([]string{"ctrl", "alt"}))
	assert.Equal(t, modControl|modAlt|modShift, ModifiersToFlags([]string{"ctrl", "alt", "shift"}))
	assert.Equal(t, modShift, ModifiersToFlags([]string{"Shift"})) // case insensitive
}

func TestIsMouseButton(t *testing.T) {
	assert.True(t, IsMouseButton("XButton1"))
	assert.True(t, IsMouseButton("XButton2"))
	assert.False(t, IsMouseButton("Tab"))
	assert.False(t, IsMouseButton("F1"))
}

func TestMouseButtonID(t *testing.T) {
	assert.Equal(t, xButton1, MouseButtonID("XButton1"))
	assert.Equal(t, xButton2, MouseButtonID("XButton2"))
	assert.Equal(t, uint16(0), MouseButtonID("Tab"))
}

func TestFormatHotkey(t *testing.T) {
	assert.Equal(t, "Tab（Tab 鍵）", FormatHotkey(nil, "Tab"))
	assert.Equal(t, "Ctrl+Tab（Tab 鍵）", FormatHotkey([]string{"ctrl"}, "Tab"))
	assert.Equal(t, "Ctrl+Alt+F1", FormatHotkey([]string{"ctrl", "alt"}, "F1"))
	assert.Equal(t, "XButton1（滑鼠側鍵：後）", FormatHotkey(nil, "XButton1"))
	assert.Equal(t, "XButton2（滑鼠側鍵：前）", FormatHotkey(nil, "XButton2"))
	assert.Equal(t, "Num0（數字鍵盤 0）", FormatHotkey(nil, "Num0"))
	assert.Equal(t, "A", FormatHotkey(nil, "A"))                                        // 字母鍵不加描述
	assert.Equal(t, "VK_0xFF（未知按鍵 VK_0xFF）", FormatHotkey(nil, "VK_0xFF"))
}

func TestIsModifierKey(t *testing.T) {
	assert.True(t, isModifierKey(vkShift))
	assert.True(t, isModifierKey(vkControl))
	assert.True(t, isModifierKey(vkMenu))
	assert.True(t, isModifierKey(vkLShift))
	assert.True(t, isModifierKey(vkRControl))
	assert.False(t, isModifierKey(vkTab))
	assert.False(t, isModifierKey(0x41)) // A
}

func TestIsGamepadButton(t *testing.T) {
	assert.True(t, IsGamepadButton("Gamepad_A"))
	assert.True(t, IsGamepadButton("Gamepad_LT"))
	assert.True(t, IsGamepadButton("Gamepad_DPadUp"))
	assert.False(t, IsGamepadButton("XButton1"))
	assert.False(t, IsGamepadButton("Tab"))
	assert.False(t, IsGamepadButton("A"))
}

func TestFormatSwitcherDisplay(t *testing.T) {
	// 鍵盤：委派給 FormatHotkey
	assert.Equal(t, "Ctrl+Tab（Tab 鍵）", FormatSwitcherDisplay([]string{"ctrl"}, "Tab", 0))
	assert.Equal(t, "XButton1（滑鼠側鍵：後）", FormatSwitcherDisplay(nil, "XButton1", 0))

	// 搖桿
	assert.Equal(t, "搖桿 #1 A 按鈕", FormatSwitcherDisplay(nil, "Gamepad_A", 0))
	assert.Equal(t, "搖桿 #2 LB（左肩鍵）", FormatSwitcherDisplay(nil, "Gamepad_LB", 1))
	assert.Equal(t, "搖桿 #3 LT（左扳機）", FormatSwitcherDisplay(nil, "Gamepad_LT", 2))
	assert.Equal(t, "搖桿 #1 十字鍵 ↑", FormatSwitcherDisplay(nil, "Gamepad_DPadUp", 0))
}

func TestGamepadButtonMask(t *testing.T) {
	assert.Equal(t, uint16(0x1000), GamepadButtonMask("Gamepad_A"))
	assert.Equal(t, uint16(0x2000), GamepadButtonMask("Gamepad_B"))
	assert.Equal(t, uint16(0x0100), GamepadButtonMask("Gamepad_LB"))
	assert.Equal(t, uint16(0), GamepadButtonMask("Unknown"))
	assert.Equal(t, uint16(0), GamepadButtonMask("Gamepad_LT")) // 扳機不在 button mask 中
}

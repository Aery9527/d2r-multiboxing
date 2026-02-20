// Package switcher provides window switching functionality for D2R multiboxing.
package switcher

import "strings"

// Windows virtual key codes
const (
	vkTab      uint32 = 0x09
	vkReturn   uint32 = 0x0D
	vkShift    uint32 = 0x10
	vkControl  uint32 = 0x11
	vkMenu     uint32 = 0x12 // Alt
	vkEscape   uint32 = 0x1B
	vkSpace    uint32 = 0x20
	vkLShift   uint32 = 0xA0
	vkRShift   uint32 = 0xA1
	vkLControl uint32 = 0xA2
	vkRControl uint32 = 0xA3
	vkLMenu    uint32 = 0xA4
	vkRMenu    uint32 = 0xA5
	vkOEM3     uint32 = 0xC0 // ` ~
)

// Modifier flags for RegisterHotKey
const (
	modAlt      uint32 = 0x0001
	modControl  uint32 = 0x0002
	modShift    uint32 = 0x0004
	modNoRepeat uint32 = 0x4000
)

// Mouse button IDs (HIWORD of mouseData in MSLLHOOKSTRUCT)
const (
	xButton1 uint16 = 1
	xButton2 uint16 = 2
)

var (
	keyNameToVK map[string]uint32
	vkToKeyName map[uint32]string
)

func init() {
	keyNameToVK = map[string]uint32{
		"Backspace": 0x08,
		"Tab":       vkTab,
		"Enter":     vkReturn,
		"Escape":    vkEscape,
		"Space":     vkSpace,
		"`":         vkOEM3,

		// 方向鍵
		"Left": 0x25, "Up": 0x26, "Right": 0x27, "Down": 0x28,

		// 功能鍵
		"Insert": 0x2D, "Delete": 0x2E,
		"Home": 0x24, "End": 0x23,
		"PageUp": 0x21, "PageDown": 0x22,

		// 數字鍵盤
		"Num0": 0x60, "Num1": 0x61, "Num2": 0x62, "Num3": 0x63,
		"Num4": 0x64, "Num5": 0x65, "Num6": 0x66, "Num7": 0x67,
		"Num8": 0x68, "Num9": 0x69,
		"Num*": 0x6A, "Num+": 0x6B, "Num-": 0x6D, "Num.": 0x6E, "Num/": 0x6F,

		// 符號鍵
		"-":  0xBD, // OEM_MINUS
		"=":  0xBB, // OEM_PLUS
		"[":  0xDB, // OEM_4
		"]":  0xDD, // OEM_6
		"\\": 0xDC, // OEM_5
		";":  0xBA, // OEM_1
		"'":  0xDE, // OEM_7
		",":  0xBC, // OEM_COMMA
		".":  0xBE, // OEM_PERIOD
		"/":  0xBF, // OEM_2
	}

	// F1-F12
	fKeys := []string{"F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10", "F11", "F12"}
	for i, name := range fKeys {
		keyNameToVK[name] = 0x70 + uint32(i)
	}

	// A-Z
	for c := byte('A'); c <= 'Z'; c++ {
		keyNameToVK[string(c)] = uint32(c)
	}

	// 0-9
	for c := byte('0'); c <= '9'; c++ {
		keyNameToVK[string(c)] = uint32(c)
	}

	// Build reverse map
	vkToKeyName = make(map[uint32]string, len(keyNameToVK))
	for name, vk := range keyNameToVK {
		vkToKeyName[vk] = name
	}
}

// KeyToVK converts a key name to its Windows virtual key code.
func KeyToVK(name string) (uint32, bool) {
	vk, ok := keyNameToVK[name]
	return vk, ok
}

// VKToKeyName converts a Windows virtual key code to a key name.
func VKToKeyName(vk uint32) (string, bool) {
	name, ok := vkToKeyName[vk]
	return name, ok
}

// ModifiersToFlags converts modifier name list to RegisterHotKey flags.
func ModifiersToFlags(mods []string) uint32 {
	var flags uint32
	for _, mod := range mods {
		switch strings.ToLower(mod) {
		case "ctrl":
			flags |= modControl
		case "alt":
			flags |= modAlt
		case "shift":
			flags |= modShift
		}
	}
	return flags
}

// IsMouseButton returns true if the key name represents a mouse side button.
func IsMouseButton(key string) bool {
	return key == "XButton1" || key == "XButton2"
}

// MouseButtonID returns the XBUTTON ID for a mouse button key name.
func MouseButtonID(key string) uint16 {
	switch key {
	case "XButton1":
		return xButton1
	case "XButton2":
		return xButton2
	default:
		return 0
	}
}

// FormatHotkey formats a hotkey combination for display with a friendly description.
// Example: "Ctrl+Tab（Ctrl + Tab 鍵）" or "XButton1（滑鼠側鍵：後）"
func FormatHotkey(modifiers []string, key string) string {
	parts := make([]string, 0, len(modifiers)+1)
	for _, mod := range modifiers {
		switch strings.ToLower(mod) {
		case "ctrl":
			parts = append(parts, "Ctrl")
		case "alt":
			parts = append(parts, "Alt")
		case "shift":
			parts = append(parts, "Shift")
		default:
			parts = append(parts, mod)
		}
	}
	parts = append(parts, key)
	combo := strings.Join(parts, "+")

	if desc := keyDescription(key); desc != "" {
		return combo + "（" + desc + "）"
	}
	return combo
}

// keyDisplayName maps key names to user-friendly Chinese descriptions.
var keyDisplayName = map[string]string{
	"XButton1":  "滑鼠側鍵：後",
	"XButton2":  "滑鼠側鍵：前",
	"Tab":       "Tab 鍵",
	"Space":     "空白鍵",
	"Backspace": "Backspace 鍵",
	"Enter":     "Enter 鍵",
	"Escape":    "Esc 鍵",
	"`":         "` 反引號鍵",
	"Insert":    "Insert 鍵",
	"Delete":    "Delete 鍵",
	"Home":      "Home 鍵",
	"End":       "End 鍵",
	"PageUp":    "Page Up 鍵",
	"PageDown":  "Page Down 鍵",
	"Left":      "← 方向鍵",
	"Up":        "↑ 方向鍵",
	"Right":     "→ 方向鍵",
	"Down":      "↓ 方向鍵",
	"Num0":      "數字鍵盤 0",
	"Num1":      "數字鍵盤 1",
	"Num2":      "數字鍵盤 2",
	"Num3":      "數字鍵盤 3",
	"Num4":      "數字鍵盤 4",
	"Num5":      "數字鍵盤 5",
	"Num6":      "數字鍵盤 6",
	"Num7":      "數字鍵盤 7",
	"Num8":      "數字鍵盤 8",
	"Num9":      "數字鍵盤 9",
	"Num*":      "數字鍵盤 *",
	"Num+":      "數字鍵盤 +",
	"Num-":      "數字鍵盤 -",
	"Num.":      "數字鍵盤 .",
	"Num/":      "數字鍵盤 /",
}

// keyDescription returns a friendly description for a key, or empty string if self-explanatory.
func keyDescription(key string) string {
	if desc, ok := keyDisplayName[key]; ok {
		return desc
	}
	// VK_0xNN 格式代表未知按鍵
	if strings.HasPrefix(key, "VK_") {
		return "未知按鍵 " + key
	}
	return ""
}

// isModifierKey returns true if the VK code is a modifier key.
func isModifierKey(vk uint32) bool {
	switch vk {
	case vkShift, vkControl, vkMenu,
		vkLShift, vkRShift, vkLControl, vkRControl, vkLMenu, vkRMenu:
		return true
	}
	return false
}

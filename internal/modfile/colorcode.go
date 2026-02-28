package modfile

import (
	"strings"
)

// d2rColorMap maps D2R color code suffixes to ANSI terminal escape sequences.
var d2rColorMap = map[byte]string{
	'0': "\033[37m",  // white
	'1': "\033[91m",  // red (bright)
	'2': "\033[92m",  // green (bright)
	'3': "\033[94m",  // blue (bright)
	'4': "\033[33m",  // gold (yellow)
	'5': "\033[90m",  // gray (dark)
	'6': "\033[30m",  // black
	'7': "\033[33m",  // tan (yellow)
	'8': "\033[38;5;208m", // orange (256-color)
	'9': "\033[93m",  // yellow (bright)
	';': "\033[95m",  // purple (bright magenta)
	'Q': "\033[38;5;136m", // dark gold (256-color)
}

const ansiReset = "\033[0m"

// d2rColorPrefix is the byte sequence that starts a D2R color code: ÿc (U+00FF + 'c').
// In UTF-8, ÿ (U+00FF) is encoded as 0xC3 0xBF.
var d2rColorPrefix = []byte{0xC3, 0xBF, 'c'}

// RenderForTerminal converts D2R color codes (ÿcX) in a string to ANSI terminal escape sequences.
func RenderForTerminal(s string) string {
	data := []byte(s)
	var buf strings.Builder
	buf.Grow(len(data) * 2)

	i := 0
	for i < len(data) {
		// Check for ÿc prefix (3 bytes: 0xC3, 0xBF, 'c')
		if i+3 < len(data) && data[i] == 0xC3 && data[i+1] == 0xBF && data[i+2] == 'c' {
			code := data[i+3]
			if ansi, ok := d2rColorMap[code]; ok {
				buf.WriteString(ansi)
				i += 4
				continue
			}
		}
		buf.WriteByte(data[i])
		i++
	}

	buf.WriteString(ansiReset)
	return buf.String()
}

// StripColorCodes removes all D2R color codes (ÿcX) from a string.
func StripColorCodes(s string) string {
	data := []byte(s)
	var buf strings.Builder
	buf.Grow(len(data))

	i := 0
	for i < len(data) {
		if i+3 < len(data) && data[i] == 0xC3 && data[i+1] == 0xBF && data[i+2] == 'c' {
			if _, ok := d2rColorMap[data[i+3]]; ok {
				i += 4
				continue
			}
		}
		buf.WriteByte(data[i])
		i++
	}

	return buf.String()
}

package modfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderForTerminal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no color codes",
			input: "plain text",
			want:  "plain text\033[0m",
		},
		{
			name:  "red color code",
			input: "\u00ffc1★★★ Ber Rune #30 ★★★",
			want:  "\033[91m★★★ Ber Rune #30 ★★★\033[0m",
		},
		{
			name:  "multiple color codes",
			input: "\u00ffc8★★\u00ffc4 Jah Rune\u00ffc5 #31",
			want:  "\033[38;5;208m★★\033[33m Jah Rune\033[90m #31\033[0m",
		},
		{
			name:  "gray color",
			input: "\u00ffc5El Rune #1",
			want:  "\033[90mEl Rune #1\033[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderForTerminal(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStripColorCodes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no color codes",
			input: "plain text",
			want:  "plain text",
		},
		{
			name:  "strip red",
			input: "\u00ffc1★★★ Ber Rune #30 ★★★",
			want:  "★★★ Ber Rune #30 ★★★",
		},
		{
			name:  "strip multiple",
			input: "\u00ffc8★★\u00ffc4 Jah Rune\u00ffc5 #31",
			want:  "★★ Jah Rune #31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripColorCodes(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

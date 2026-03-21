package main

import (
	"strconv"
	"strings"
)

type parsedMenuOption struct {
	key  string
	line string
}

type parsedMenuBlock struct {
	lines   []string
	options []parsedMenuOption
}

func normalizedOutputLines(output string) []string {
	output = strings.ReplaceAll(output, "\r\n", "\n")
	return strings.Split(output, "\n")
}

func nonEmptyOutputLines(output string) []string {
	lines := normalizedOutputLines(output)
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		filtered = append(filtered, line)
	}
	return filtered
}

func parseMenuOptionLine(line string) (parsedMenuOption, bool) {
	if !strings.HasPrefix(line, "[") {
		return parsedMenuOption{}, false
	}

	end := strings.Index(line, "]")
	if end <= 1 {
		return parsedMenuOption{}, false
	}

	return parsedMenuOption{
		key:  line[1:end],
		line: line,
	}, true
}

func buildParsedMenuBlock(lines []string) parsedMenuBlock {
	block := parsedMenuBlock{
		lines: append([]string(nil), lines...),
	}
	for _, line := range lines {
		if option, ok := parseMenuOptionLine(line); ok {
			block.options = append(block.options, option)
		}
	}
	return block
}

func parseMenuBlocks(output string) []parsedMenuBlock {
	divider := newCLIUI().style.menuDivider
	lines := normalizedOutputLines(output)
	blocks := make([]parsedMenuBlock, 0, 4)

	inBlock := false
	currentLines := make([]string, 0, 8)
	for _, line := range lines {
		if line == divider {
			if inBlock {
				blocks = append(blocks, buildParsedMenuBlock(currentLines))
				currentLines = currentLines[:0]
				inBlock = false
			} else {
				inBlock = true
			}
			continue
		}
		if !inBlock {
			continue
		}
		currentLines = append(currentLines, line)
	}

	return blocks
}

func menuBlockKeys(block parsedMenuBlock) []string {
	keys := make([]string, 0, len(block.options))
	for _, option := range block.options {
		keys = append(keys, option.key)
	}
	return keys
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func countMenuBlocksWithKeys(output string, expected []string) int {
	count := 0
	for _, block := range parseMenuBlocks(output) {
		if sameStrings(menuBlockKeys(block), expected) {
			count++
		}
	}
	return count
}

func findMenuOptionLine(output string, key string) (string, bool) {
	for _, block := range parseMenuBlocks(output) {
		for _, option := range block.options {
			if option.key == key {
				return option.line, true
			}
		}
	}
	return "", false
}

func linesWithPrefix(output string, prefix string) []string {
	lines := normalizedOutputLines(output)
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			filtered = append(filtered, line)
		}
	}
	return filtered
}

func blockLinesWithPrefix(block parsedMenuBlock, prefix string) []string {
	filtered := make([]string, 0, len(block.lines))
	for _, line := range block.lines {
		if strings.HasPrefix(line, prefix) {
			filtered = append(filtered, line)
		}
	}
	return filtered
}

func firstLineIndex(lines []string, match func(string) bool) int {
	for i, line := range lines {
		if match(line) {
			return i
		}
	}
	return -1
}

func extractIntsFromString(s string) []int {
	var (
		values  []int
		current strings.Builder
	)

	flush := func() {
		if current.Len() == 0 {
			return
		}
		value, err := strconv.Atoi(current.String())
		if err == nil {
			values = append(values, value)
		}
		current.Reset()
	}

	for _, r := range s {
		if r >= '0' && r <= '9' {
			current.WriteRune(r)
			continue
		}
		flush()
	}
	flush()

	return values
}

func between(s, start, end string) string {
	startIndex := strings.Index(s, start)
	if startIndex == -1 {
		return ""
	}
	startIndex += len(start)
	endIndex := strings.Index(s[startIndex:], end)
	if endIndex == -1 {
		return ""
	}
	return s[startIndex : startIndex+endIndex]
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

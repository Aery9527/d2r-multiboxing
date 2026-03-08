package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"d2rhl/internal/multiboxing/account"
)

func parseSelectionInput(input string, max int) ([]int, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("請至少輸入一個編號")
	}

	selected := make(map[int]bool)
	for _, part := range strings.Split(input, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("輸入中有空白項目")
		}

		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("無法辨識區間 %q", part)
			}
			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("無法辨識區間起點 %q", part)
			}
			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("無法辨識區間終點 %q", part)
			}
			if start > end {
				return nil, fmt.Errorf("區間 %q 起點不可大於終點", part)
			}
			if start < 1 || end > max {
				return nil, fmt.Errorf("區間 %q 超出可選範圍 1-%d", part, max)
			}
			for i := start; i <= end; i++ {
				selected[i-1] = true
			}
			continue
		}

		value, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("無法辨識編號 %q", part)
		}
		if value < 1 || value > max {
			return nil, fmt.Errorf("編號 %d 超出可選範圍 1-%d", value, max)
		}
		selected[value-1] = true
	}

	indexes := make([]int, 0, len(selected))
	for idx := range selected {
		indexes = append(indexes, idx)
	}
	sort.Ints(indexes)
	return indexes, nil
}

func selectedLaunchFlagMask(flagIndexes []int, options []account.LaunchFlagOption) uint32 {
	var mask uint32
	for _, idx := range flagIndexes {
		mask |= options[idx].Bit
	}
	return mask
}

func confirmChanges() bool {
	answer, ok := ui.readInputf("確認套用？([Y]/[n])：")
	if !ok {
		return false
	}
	answer = strings.ToLower(answer)
	return answer == "" || answer == "y"
}

func flagActionLabel(setMode bool) string {
	if setMode {
		return "設定"
	}
	return "取消"
}

func allLaunchFlagMask(options []account.LaunchFlagOption) uint32 {
	return selectedLaunchFlagMask(allLaunchFlagIndexes(options), options)
}

func launchFlagOptionsForMask(options []account.LaunchFlagOption, mask uint32) []account.LaunchFlagOption {
	selected := make([]account.LaunchFlagOption, 0, len(options))
	for _, option := range options {
		if mask&option.Bit == 0 {
			continue
		}
		selected = append(selected, option)
	}
	return selected
}

func allLaunchFlagIndexes(options []account.LaunchFlagOption) []int {
	indexes := make([]int, 0, len(options))
	for i := range options {
		indexes = append(indexes, i)
	}
	return indexes
}

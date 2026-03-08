package main

import (
	"fmt"
	"strconv"
	"strings"

	"d2rhl/internal/multiboxing/account"
)

func setupAccountLaunchFlags(accounts []account.Account, accountsFile string) {
	if len(accounts) == 0 {
		ui.infof("目前沒有可設定的帳號。")
		ui.blankLine()
		return
	}

	for {
		ui.blankLine()
		ui.headf("帳號啟動 flag 設定")
		printAccountList(accounts)
		ui.blankLine()
		printAccountLaunchFlagTable(accounts)
		ui.blankLine()
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option("1", "設定 flag", "")
			options.option("2", "取消 flag", "")
		})
		ui.menuBlock(func() {
			options.render()
		})
		choice, ok := ui.readInput()
		if !ok {
			return
		}
		if isMenuNav(choice) != "" {
			return
		}

		var setMode bool
		var actionLabel string
		switch choice {
		case "1":
			setMode = true
			actionLabel = "設定"
		case "2":
			setMode = false
			actionLabel = "取消"
		default:
			showInvalidInputAndPause()
			continue
		}

		for {
			ui.blankLine()
			ui.headf("%s flag：選擇操作方式", actionLabel)
			ui.promptf("這次要如何%s flag？", actionLabel)
			modeOptions := ui.subMenuOptions(func(options *cliMenuOptions) {
				options.option("1", "以 flag 為維度", "")
				options.option("2", "以帳號為維度", "")
			})
			ui.menuBlock(func() {
				modeOptions.render()
			})
			modeChoice, ok := ui.readInput()
			if !ok {
				return
			}
			if isMenuNav(modeChoice) != "" {
				return
			}

			switch modeChoice {
			case "1":
				configureFlagsByFlag(accounts, accountsFile, setMode)
				return
			case "2":
				configureFlagsByAccount(accounts, accountsFile, setMode)
				return
			default:
				showInvalidInputAndPause()
			}
		}
	}
}

func configureFlagsByFlag(accounts []account.Account, accountsFile string, setMode bool) {
	options := account.LaunchFlagOptions()
	var option account.LaunchFlagOption
	for {
		ui.blankLine()
		ui.headf("%s flag：依 flag 選帳號", flagActionLabel(setMode))
		flagOptions := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
			for i, option := range options {
				comment := option.Description
				if option.Description != "" {
					comment = fmt.Sprintf("說明：%s", option.Description)
				}
				if option.Experimental {
					if comment != "" {
						comment += "，"
					}
					comment += "效果依版本而定"
				}
				menuOptions.option(strconv.Itoa(i+1), option.Name, comment)
			}
		})
		ui.menuBlock(func() {
			flagOptions.render()
		})
		input, ok := ui.readInputf("請選擇 flag 編號：")
		if !ok {
			return
		}
		if isMenuNav(input) != "" {
			return
		}

		selected, err := strconv.Atoi(input)
		if err != nil || selected < 1 || selected > len(options) {
			showInputErrorAndPause("無效的 flag 編號。")
			continue
		}
		option = options[selected-1]
		break
	}

	actionLabel := flagActionLabel(setMode)
	for {
		ui.blankLine()
		ui.headf("%s flag：選擇帳號", actionLabel)
		accountOptions := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
			for i, acc := range accounts {
				menuOptions.option(strconv.Itoa(i+1), fmt.Sprintf("%s (%s)", acc.DisplayName, acc.Email), fmt.Sprintf("flag：%s", account.LaunchFlagsSummary(acc.LaunchFlags)))
			}
		})
		ui.menuBlock(func() {
			ui.promptf("請輸入要%s「%s」的帳號編號，可用 2,4,6 或 1-3,5-7：", actionLabel, option.Name)
			accountOptions.render()
		})
		input, ok := ui.readInputf("請輸入：")
		if !ok {
			return
		}
		if isMenuNav(input) != "" {
			return
		}

		accountIndexes, err := parseSelectionInput(input, len(accounts))
		if err != nil {
			showInputErrorAndPause(fmt.Sprintf("解析失敗：%v", err))
			continue
		}

		ui.blankLine()
		ui.infof("即將%s以下帳號的 flag「%s」：", actionLabel, option.Name)
		for _, idx := range accountIndexes {
			acc := accounts[idx]
			ui.rawlnf("  [%d] %s (%s)  目前：%s", idx+1, acc.DisplayName, acc.Email, account.LaunchFlagsSummary(acc.LaunchFlags))
		}
		if !confirmChanges() {
			ui.infof("已取消。")
			ui.blankLine()
			return
		}

		if err := applyLaunchFlagChanges(accounts, accountsFile, accountIndexes, option.Bit, setMode); err != nil {
			showInputErrorAndPause(fmt.Sprintf("儲存失敗：%v", err))
			continue
		}

		ui.successf("已完成%s。", actionLabel)
		ui.blankLine()
		return
	}
}

func configureFlagsByAccount(accounts []account.Account, accountsFile string, setMode bool) {
	options := account.LaunchFlagOptions()
	var (
		accountIndex int
		acc          account.Account
	)
	for {
		ui.blankLine()
		ui.headf("%s flag：先選帳號", flagActionLabel(setMode))
		accountOptions := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
			for i, acc := range accounts {
				menuOptions.option(strconv.Itoa(i+1), fmt.Sprintf("%s (%s)", acc.DisplayName, acc.Email), fmt.Sprintf("flag：%s", account.LaunchFlagsSummary(acc.LaunchFlags)))
			}
		})
		ui.menuBlock(func() {
			accountOptions.render()
		})
		input, ok := ui.readInputf("請選擇帳號編號：")
		if !ok {
			return
		}
		if isMenuNav(input) != "" {
			return
		}

		selected, err := strconv.Atoi(input)
		if err != nil || selected < 1 || selected > len(accounts) {
			showInputErrorAndPause("無效的帳號編號。")
			continue
		}

		accountIndex = selected - 1
		acc = accounts[accountIndex]
		break
	}

	actionLabel := flagActionLabel(setMode)
	for {
		ui.blankLine()
		ui.headf("%s flag：選擇旗標", actionLabel)
		flagOptions := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
			for i, option := range options {
				comment := option.Description
				if option.Description != "" {
					comment = fmt.Sprintf("說明：%s", option.Description)
				}
				if option.Experimental {
					if comment != "" {
						comment += "，"
					}
					comment += "效果依版本而定"
				}
				menuOptions.option(strconv.Itoa(i+1), option.Name, comment)
			}
		})
		ui.menuBlock(func() {
			ui.promptf("請輸入要對帳號「%s」%s的 flag 編號，可用 1,3 或 2-4：", acc.DisplayName, actionLabel)
			flagOptions.render()
		})
		input, ok := ui.readInputf("請輸入：")
		if !ok {
			return
		}
		if isMenuNav(input) != "" {
			return
		}

		flagIndexes, err := parseSelectionInput(input, len(options))
		if err != nil {
			showInputErrorAndPause(fmt.Sprintf("解析失敗：%v", err))
			continue
		}

		mask := selectedLaunchFlagMask(flagIndexes, options)
		ui.blankLine()
		ui.infof("即將對帳號「%s」%s以下 flag：", acc.DisplayName, actionLabel)
		for _, idx := range flagIndexes {
			option := options[idx]
			ui.rawlnf("  [%d] %s（%s）", idx+1, option.Name, option.Description)
		}
		if !confirmChanges() {
			ui.infof("已取消。")
			ui.blankLine()
			return
		}

		if err := applyLaunchFlagChanges(accounts, accountsFile, []int{accountIndex}, mask, setMode); err != nil {
			showInputErrorAndPause(fmt.Sprintf("儲存失敗：%v", err))
			continue
		}

		ui.successf("已完成%s。", actionLabel)
		ui.blankLine()
		return
	}
}

func applyLaunchFlagChanges(accounts []account.Account, accountsFile string, accountIndexes []int, mask uint32, setMode bool) error {
	if setMode && hasConflictingLaunchFlags(mask) {
		return fmt.Errorf("關閉聲音與背景保留聲音不可同時設定，請分開操作")
	}

	previous := make(map[int]uint32, len(accountIndexes))
	for _, idx := range accountIndexes {
		previous[idx] = accounts[idx].LaunchFlags
		if setMode {
			accounts[idx].LaunchFlags |= mask
			accounts[idx].LaunchFlags = normalizeLaunchFlags(accounts[idx].LaunchFlags, mask)
			continue
		}
		accounts[idx].LaunchFlags &^= mask
	}

	if err := account.SaveAccounts(accountsFile, accounts); err != nil {
		for idx, flags := range previous {
			accounts[idx].LaunchFlags = flags
		}
		return err
	}
	return nil
}

func printAccountList(accounts []account.Account) {
	ui.infof("帳號列表：")
	for i, acc := range accounts {
		status := "未啟動"
		if isAccountRunning(acc.DisplayName) {
			status = "已啟動"
		}
		ui.rawlnf("[%d] <%s> %-15s (%s) ", i+1, status, acc.DisplayName, acc.Email)
	}
}

func printAccountLaunchFlagTable(accounts []account.Account) {
	ui.infof("flag 對照表：")
	for _, line := range buildAccountLaunchFlagTableLines(accounts) {
		ui.rawln(line)
	}
}

func buildAccountLaunchFlagTableLines(accounts []account.Account) []string {
	options := account.LaunchFlagOptions()
	headerTop := make([]string, 0, len(options)+1)
	headerBottom := make([]string, 0, len(options)+1)
	headerTop = append(headerTop, "帳號編號")
	headerBottom = append(headerBottom, "")
	for _, option := range options {
		title, flag := launchFlagTableHeaderLines(option)
		headerTop = append(headerTop, title)
		headerBottom = append(headerBottom, flag)
	}

	widths := launchFlagTableColumnWidths(accounts, headerTop, headerBottom, options)

	lines := make([]string, 0, len(accounts)+5)
	separator := buildLaunchFlagTableSeparator(widths)
	lines = append(lines, separator)
	lines = append(lines, buildLaunchFlagTableRow(headerTop, widths))
	lines = append(lines, buildLaunchFlagTableRow(headerBottom, widths))
	lines = append(lines, separator)
	for i, acc := range accounts {
		cells := make([]string, 0, len(options)+1)
		cells = append(cells, strconv.Itoa(i+1))
		for _, option := range options {
			cell := ""
			if acc.LaunchFlags&option.Bit != 0 {
				cell = "v"
			}
			cells = append(cells, cell)
		}
		lines = append(lines, buildLaunchFlagTableRow(cells, widths))
	}
	lines = append(lines, separator)
	return lines
}

func launchFlagTableHeaderLines(option account.LaunchFlagOption) (string, string) {
	if option.Description == "" {
		return option.Name, ""
	}
	return option.Name, option.Description
}

func launchFlagTableColumnWidths(accounts []account.Account, headerTop []string, headerBottom []string, options []account.LaunchFlagOption) []int {
	widths := make([]int, len(headerTop))
	for i := range headerTop {
		widths[i] = maxDisplayWidth(headerTop[i], headerBottom[i])
	}
	for i := range accounts {
		indexWidth := displayWidth(strconv.Itoa(i + 1))
		if indexWidth > widths[0] {
			widths[0] = indexWidth
		}
		for j, option := range options {
			cellWidth := 0
			if accounts[i].LaunchFlags&option.Bit != 0 {
				cellWidth = displayWidth("v")
			}
			if cellWidth > widths[j+1] {
				widths[j+1] = cellWidth
			}
		}
	}
	return widths
}

func maxDisplayWidth(values ...string) int {
	maxWidth := 0
	for _, value := range values {
		if width := displayWidth(value); width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth
}

func buildLaunchFlagTableSeparator(widths []int) string {
	var builder strings.Builder
	builder.WriteString("+")
	for _, width := range widths {
		builder.WriteString(strings.Repeat("-", width+2))
		builder.WriteString("+")
	}
	return builder.String()
}

func buildLaunchFlagTableRow(cells []string, widths []int) string {
	var builder strings.Builder
	builder.WriteString("|")
	for i, cell := range cells {
		builder.WriteString(" ")
		builder.WriteString(centerLaunchFlagTableCell(cell, widths[i]))
		builder.WriteString(" ")
		builder.WriteString("|")
	}
	return builder.String()
}

func centerLaunchFlagTableCell(value string, width int) string {
	padding := width - displayWidth(value)
	if padding <= 0 {
		return value
	}
	left := padding / 2
	right := padding - left
	return strings.Repeat(" ", left) + value + strings.Repeat(" ", right)
}

func printLaunchFlagOptions(options []account.LaunchFlagOption) {
	for i, option := range options {
		line := fmt.Sprintf("  [%d] %s", i+1, option.Name)
		if option.Description != "" {
			line += fmt.Sprintf("（%s）", option.Description)
		}
		if option.Experimental {
			line += "，效果依版本而定"
		}
		ui.rawln(line)
	}
}

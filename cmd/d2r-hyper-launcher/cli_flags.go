package main

import (
	"fmt"
	"strconv"
	"strings"

	"d2rhl/internal/multiboxing/account"
)

func setupAccountLaunchFlags(accounts []account.Account, accountsFile string) {
	if len(accounts) == 0 {
		ui.infof("%s", lang.Flags.NoAccounts)
		ui.blankLine()
		return
	}

menuLoop:
	for {
		ui.headf("%s", lang.Flags.Title)
		printAccountList(accounts)
		ui.blankLine()
		printAccountLaunchFlagTable(accounts)
		ui.blankLine()
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option("1", lang.Flags.OptSetFlag, "")
			options.option("2", lang.Flags.OptClearFlag, "")
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
			actionLabel = lang.Flags.ActionSet
		case "2":
			setMode = false
			actionLabel = lang.Flags.ActionClear
		default:
			showInvalidInputAndPause()
			continue
		}

		for {
			ui.headf(lang.Flags.ModeTitle, actionLabel)
			ui.infof(lang.Flags.ModeQuestion, actionLabel)
			modeOptions := ui.subMenuOptions(func(options *cliMenuOptions) {
				options.option("1", fmt.Sprintf(lang.Flags.OptFlagToAccounts, actionLabel), "")
				options.option("2", fmt.Sprintf(lang.Flags.OptAccountToFlags, actionLabel), "")
				options.option("3", fmt.Sprintf(lang.Flags.OptAllFlagsAll, actionLabel), "")
			})
			ui.menuBlock(func() {
				modeOptions.render()
			})
			modeChoice, ok := ui.readInput()
			if !ok {
				return
			}
			switch isMenuNav(modeChoice) {
			case "back":
				continue menuLoop
			case "home":
				return
			}

			switch modeChoice {
			case "1":
				switch configureFlagsByFlag(accounts, accountsFile, setMode) {
				case "back":
					continue
				case "home":
					return
				default:
					continue menuLoop
				}
			case "2":
				switch configureFlagsByAccount(accounts, accountsFile, setMode) {
				case "back":
					continue
				case "home":
					return
				default:
					continue menuLoop
				}
			case "3":
				switch configureAllFlagsForAllAccounts(accounts, accountsFile, setMode) {
				case "back":
					continue
				case "home":
					return
				default:
					continue menuLoop
				}
			default:
				showInvalidInputAndPause()
			}
		}
	}
}

func configureAllFlagsForAllAccounts(accounts []account.Account, accountsFile string, setMode bool) string {
	options := account.LaunchFlagOptions()
	actionLabel := flagActionLabel(setMode)
	accountIndexes := make([]int, 0, len(accounts))
	for i := range accounts {
		accountIndexes = append(accountIndexes, i)
	}

	mask := allLaunchFlagMask(options)
	affectedOptions := launchFlagOptionsForMask(options, mask)

	ui.headf(lang.Flags.FlagAllTitle, actionLabel)
	ui.infof(lang.Flags.FlagAllAbout, actionLabel)
	for _, option := range affectedOptions {
		ui.rawlnf("  - %s（%s）", option.Name, option.Description)
	}
	ui.infof(lang.Flags.FlagAllCount, len(accounts))
	if !confirmChanges() {
		ui.infof("%s", lang.Common.Cancelled)
		ui.blankLine()
		return ""
	}

	if err := applyLaunchFlagChanges(accounts, accountsFile, accountIndexes, mask, setMode); err != nil {
		showInputErrorAndPause(fmt.Sprintf(lang.Flags.SaveFailed, err))
		return ""
	}

	ui.successf(lang.Flags.Done, actionLabel)
	ui.blankLine()
	return ""
}

func configureFlagsByFlag(accounts []account.Account, accountsFile string, setMode bool) string {
	options := account.LaunchFlagOptions()
	var option account.LaunchFlagOption

selectFlag:
	for {
		ui.headf(lang.Flags.FlagByFlagTitle, flagActionLabel(setMode))
		flagOptions := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
			for i, option := range options {
				comment := ""
				if option.Description != "" {
					comment = fmt.Sprintf(lang.Flags.FlagDescPrefix, option.Description)
				}
				if option.Experimental {
					if comment != "" {
						comment += "，"
					}
					comment += lang.Flags.FlagExperimental
				}
				menuOptions.option(strconv.Itoa(i+1), option.Name, comment)
			}
		})
		ui.menuBlock(func() {
			flagOptions.render()
		})
		input, ok := ui.readInputf("%s", lang.Flags.FlagByFlagSelectPrompt)
		if !ok {
			return ""
		}
		switch isMenuNav(input) {
		case "back":
			return "back"
		case "home":
			return "home"
		}

		selected, err := strconv.Atoi(input)
		if err != nil || selected < 1 || selected > len(options) {
			showInputErrorAndPause(lang.Flags.InvalidFlagID)
			continue
		}
		option = options[selected-1]
		break
	}

	actionLabel := flagActionLabel(setMode)
	for {
		ui.headf(lang.Flags.FlagByFlagAccountTitle, actionLabel)
		accountOptions := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
			for i, acc := range accounts {
				menuOptions.option(strconv.Itoa(i+1), fmt.Sprintf("%s (%s)", acc.DisplayName, acc.Email), fmt.Sprintf(lang.Flags.FlagComment, account.LaunchFlagsSummary(acc.LaunchFlags)))
			}
		})
		ui.menuBlock(func() {
			ui.promptf(lang.Flags.FlagByFlagAccountPrompt, actionLabel, option.Name)
			accountOptions.render()
		})
		input, ok := ui.readInputf("%s", lang.Flags.FlagInputPrompt)
		if !ok {
			return ""
		}
		switch isMenuNav(input) {
		case "back":
			goto selectFlag
		case "home":
			return "home"
		}

		accountIndexes, err := parseSelectionInput(input, len(accounts))
		if err != nil {
			showInputErrorAndPause(fmt.Sprintf(lang.Common.ParseFailed, err))
			continue
		}

		ui.blankLine()
		ui.infof(lang.Flags.FlagByFlagAbout, actionLabel, option.Name)
		for _, idx := range accountIndexes {
			acc := accounts[idx]
			ui.rawlnf(lang.Flags.FlagAccountItemFmt, idx+1, acc.DisplayName, acc.Email, account.LaunchFlagsSummary(acc.LaunchFlags))
		}
		if !confirmChanges() {
			ui.infof("%s", lang.Common.Cancelled)
			ui.blankLine()
			return ""
		}

		if err := applyLaunchFlagChanges(accounts, accountsFile, accountIndexes, option.Bit, setMode); err != nil {
			showInputErrorAndPause(fmt.Sprintf(lang.Flags.SaveFailed, err))
			continue
		}

		ui.successf(lang.Flags.Done, actionLabel)
		ui.blankLine()
		return ""
	}
}

func configureFlagsByAccount(accounts []account.Account, accountsFile string, setMode bool) string {
	options := account.LaunchFlagOptions()
	var (
		accountIndex int
		acc          account.Account
	)

selectAccount:
	for {
		ui.headf(lang.Flags.FlagByAccountTitle, flagActionLabel(setMode))
		accountOptions := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
			for i, acc := range accounts {
				menuOptions.option(strconv.Itoa(i+1), fmt.Sprintf("%s (%s)", acc.DisplayName, acc.Email), fmt.Sprintf(lang.Flags.FlagComment, account.LaunchFlagsSummary(acc.LaunchFlags)))
			}
		})
		ui.menuBlock(func() {
			accountOptions.render()
		})
		input, ok := ui.readInputf("%s", lang.Flags.FlagByAccountSelectPrompt)
		if !ok {
			return ""
		}
		switch isMenuNav(input) {
		case "back":
			return "back"
		case "home":
			return "home"
		}

		selected, err := strconv.Atoi(input)
		if err != nil || selected < 1 || selected > len(accounts) {
			showInputErrorAndPause(lang.Flags.InvalidAccountID)
			continue
		}

		accountIndex = selected - 1
		acc = accounts[accountIndex]
		break
	}

	actionLabel := flagActionLabel(setMode)
	for {
		ui.headf(lang.Flags.FlagByAccountFlagTitle, actionLabel)
		flagOptions := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
			for i, option := range options {
				comment := ""
				if option.Description != "" {
					comment = fmt.Sprintf(lang.Flags.FlagDescPrefix, option.Description)
				}
				if option.Experimental {
					if comment != "" {
						comment += "，"
					}
					comment += lang.Flags.FlagExperimental
				}
				menuOptions.option(strconv.Itoa(i+1), option.Name, comment)
			}
		})
		ui.menuBlock(func() {
			ui.promptf(lang.Flags.FlagByAccountFlagPrompt, acc.DisplayName, actionLabel)
			flagOptions.render()
		})
		input, ok := ui.readInputf("%s", lang.Flags.FlagInputPrompt)
		if !ok {
			return ""
		}
		switch isMenuNav(input) {
		case "back":
			goto selectAccount
		case "home":
			return "home"
		}

		flagIndexes, err := parseSelectionInput(input, len(options))
		if err != nil {
			showInputErrorAndPause(fmt.Sprintf(lang.Common.ParseFailed, err))
			continue
		}

		mask := selectedLaunchFlagMask(flagIndexes, options)
		ui.blankLine()
		ui.infof(lang.Flags.FlagByAccountAbout, acc.DisplayName, actionLabel)
		for _, idx := range flagIndexes {
			option := options[idx]
			ui.rawlnf("  [%d] %s（%s）", idx+1, option.Name, option.Description)
		}
		if !confirmChanges() {
			ui.infof("%s", lang.Common.Cancelled)
			ui.blankLine()
			return ""
		}

		if err := applyLaunchFlagChanges(accounts, accountsFile, []int{accountIndex}, mask, setMode); err != nil {
			showInputErrorAndPause(fmt.Sprintf(lang.Flags.SaveFailed, err))
			continue
		}

		ui.successf(lang.Flags.Done, actionLabel)
		ui.blankLine()
		return ""
	}
}

func applyLaunchFlagChanges(accounts []account.Account, accountsFile string, accountIndexes []int, mask uint32, setMode bool) error {
	previous := make(map[int]uint32, len(accountIndexes))
	for _, idx := range accountIndexes {
		previous[idx] = accounts[idx].LaunchFlags
		if setMode {
			accounts[idx].LaunchFlags |= mask
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
	ui.infof("%s", lang.MainMenu.AccountListHeader)
	for i, acc := range accounts {
		status := lang.Launch.StatusStopped
		if isAccountRunning(acc.DisplayName) {
			status = lang.Launch.StatusRunning
		}
		ui.rawlnf("[%d] <%s> %-15s (%s) ", i+1, status, acc.DisplayName, acc.Email)
	}
}

func printAccountLaunchFlagTable(accounts []account.Account) {
	ui.infof("%s", lang.Flags.FlagTableHeader)
	for _, line := range buildAccountLaunchFlagTableLines(accounts) {
		ui.rawln(line)
	}
}

func buildAccountLaunchFlagTableLines(accounts []account.Account) []string {
	options := account.LaunchFlagOptions()
	headerTop := make([]string, 0, len(options)+1)
	headerBottom := make([]string, 0, len(options)+1)
	headerTop = append(headerTop, lang.Flags.FlagTableAccountHeader)
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
			line += "，術士版本似乎已失效"
		}
		ui.rawln(line)
	}
}

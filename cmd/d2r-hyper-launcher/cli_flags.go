package main

import (
	"fmt"
	"strconv"

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
		printAccountLaunchFlagSummary(accounts)
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

func printAccountLaunchFlagSummary(accounts []account.Account) {
	for i, acc := range accounts {
		ui.rawlnf("  [%d] %s (%s)  flag：%s", i+1, acc.DisplayName, acc.Email, account.LaunchFlagsSummary(acc.LaunchFlags))
	}
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

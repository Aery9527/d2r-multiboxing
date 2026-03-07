package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"d2rhl/internal/multiboxing/account"
)

func setupAccountLaunchFlags(accounts []account.Account, accountsFile string, scanner *bufio.Scanner) {
	if len(accounts) == 0 {
		fmt.Println("  目前沒有可設定的帳號。")
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Println("  === 帳號啟動 flag 設定 ===")
	printAccountLaunchFlagSummary(accounts)
	fmt.Println()
	fmt.Println("  [1] 設定 flag")
	fmt.Println("  [2] 取消 flag")
	printSubMenuNav()
	fmt.Print("  > 請選擇：")

	if !scanner.Scan() {
		return
	}
	choice := strings.TrimSpace(scanner.Text())
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
		return
	}

	fmt.Println()
	fmt.Printf("  這次要如何%s flag？\n", actionLabel)
	fmt.Println("  [1] 以 flag 為維度")
	fmt.Println("  [2] 以帳號為維度")
	printSubMenuNav()
	fmt.Print("  > 請選擇：")

	if !scanner.Scan() {
		return
	}
	modeChoice := strings.TrimSpace(scanner.Text())
	if isMenuNav(modeChoice) != "" {
		return
	}

	switch modeChoice {
	case "1":
		configureFlagsByFlag(accounts, accountsFile, scanner, setMode)
	case "2":
		configureFlagsByAccount(accounts, accountsFile, scanner, setMode)
	default:
		showInvalidInputAndPause()
	}
}

func configureFlagsByFlag(accounts []account.Account, accountsFile string, scanner *bufio.Scanner, setMode bool) {
	options := account.LaunchFlagOptions()
	fmt.Println()
	fmt.Println("  可用 flag：")
	printLaunchFlagOptions(options)
	printSubMenuNav()
	fmt.Print("  > 請選擇 flag 編號：")

	if !scanner.Scan() {
		return
	}
	input := strings.TrimSpace(scanner.Text())
	if isMenuNav(input) != "" {
		return
	}

	selected, err := strconv.Atoi(input)
	if err != nil || selected < 1 || selected > len(options) {
		showInputErrorAndPause("無效的 flag 編號。")
		return
	}

	option := options[selected-1]
	actionLabel := flagActionLabel(setMode)
	fmt.Println()
	fmt.Printf("  請輸入要%s「%s」的帳號編號，可用 2,4,6 或 1-3,5-7：\n", actionLabel, option.Name)
	printAccountLaunchFlagSummary(accounts)
	printSubMenuNav()
	fmt.Print("  > 請輸入：")

	if !scanner.Scan() {
		return
	}
	input = strings.TrimSpace(scanner.Text())
	if isMenuNav(input) != "" {
		return
	}

	accountIndexes, err := parseSelectionInput(input, len(accounts))
	if err != nil {
		showInputErrorAndPause(fmt.Sprintf("解析失敗：%v", err))
		return
	}

	fmt.Println()
	fmt.Printf("  即將%s以下帳號的 flag「%s」：\n", actionLabel, option.Name)
	for _, idx := range accountIndexes {
		acc := accounts[idx]
		fmt.Printf("  [%d] %s (%s)  目前：%s\n", idx+1, acc.DisplayName, acc.Email, account.LaunchFlagsSummary(acc.LaunchFlags))
	}
	if !confirmChanges(scanner) {
		fmt.Println("  已取消。")
		fmt.Println()
		return
	}

	if err := applyLaunchFlagChanges(accounts, accountsFile, accountIndexes, option.Bit, setMode); err != nil {
		showInputErrorAndPause(fmt.Sprintf("儲存失敗：%v", err))
		return
	}

	fmt.Printf("  ✔ 已完成%s。\n", actionLabel)
	fmt.Println()
}

func configureFlagsByAccount(accounts []account.Account, accountsFile string, scanner *bufio.Scanner, setMode bool) {
	options := account.LaunchFlagOptions()
	fmt.Println()
	fmt.Println("  帳號列表：")
	printAccountLaunchFlagSummary(accounts)
	printSubMenuNav()
	fmt.Print("  > 請選擇帳號編號：")

	if !scanner.Scan() {
		return
	}
	input := strings.TrimSpace(scanner.Text())
	if isMenuNav(input) != "" {
		return
	}

	selected, err := strconv.Atoi(input)
	if err != nil || selected < 1 || selected > len(accounts) {
		showInputErrorAndPause("無效的帳號編號。")
		return
	}

	accountIndex := selected - 1
	acc := accounts[accountIndex]
	actionLabel := flagActionLabel(setMode)
	fmt.Println()
	fmt.Printf("  請輸入要對帳號「%s」%s的 flag 編號，可用 1,3 或 2-4：\n", acc.DisplayName, actionLabel)
	printLaunchFlagOptions(options)
	printSubMenuNav()
	fmt.Print("  > 請輸入：")

	if !scanner.Scan() {
		return
	}
	input = strings.TrimSpace(scanner.Text())
	if isMenuNav(input) != "" {
		return
	}

	flagIndexes, err := parseSelectionInput(input, len(options))
	if err != nil {
		showInputErrorAndPause(fmt.Sprintf("解析失敗：%v", err))
		return
	}

	mask := selectedLaunchFlagMask(flagIndexes, options)
	fmt.Println()
	fmt.Printf("  即將對帳號「%s」%s以下 flag：\n", acc.DisplayName, actionLabel)
	for _, idx := range flagIndexes {
		option := options[idx]
		fmt.Printf("  [%d] %s（%s）\n", idx+1, option.Name, option.Description)
	}
	if !confirmChanges(scanner) {
		fmt.Println("  已取消。")
		fmt.Println()
		return
	}

	if err := applyLaunchFlagChanges(accounts, accountsFile, []int{accountIndex}, mask, setMode); err != nil {
		showInputErrorAndPause(fmt.Sprintf("儲存失敗：%v", err))
		return
	}

	fmt.Printf("  ✔ 已完成%s。\n", actionLabel)
	fmt.Println()
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
		fmt.Printf("  [%d] %s (%s)  flag：%s\n", i+1, acc.DisplayName, acc.Email, account.LaunchFlagsSummary(acc.LaunchFlags))
	}
}

func printLaunchFlagOptions(options []account.LaunchFlagOption) {
	for i, option := range options {
		fmt.Printf("  [%d] %s", i+1, option.Name)
		if option.Description != "" {
			fmt.Printf("（%s）", option.Description)
		}
		if option.Experimental {
			fmt.Print("，效果依版本而定")
		}
		fmt.Println()
	}
}

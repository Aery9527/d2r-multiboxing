package main

import (
	"bufio"
	"fmt"
	"strings"

	"d2rhl/internal/common/config"
)

func setupLaunchDelay(cfg *config.Config, scanner *bufio.Scanner) {
	fmt.Println()
	fmt.Println("  === 啟動間隔設定 ===")
	fmt.Printf("  目前設定：%s\n", cfg.LaunchDelay.DisplayString())
	fmt.Println("  說明：這會影響主選單 a「啟動所有帳號」時，每個帳號之間的等待秒數。")
	fmt.Printf("  固定下限：%d 秒\n", config.MinLaunchDelaySeconds)
	fmt.Println("  可輸入固定秒數，例如：30")
	fmt.Println("  也可輸入隨機範圍，例如：30-60")
	printSubMenuNav()
	fmt.Print("  > 請輸入新的秒數或範圍：")

	if !scanner.Scan() {
		return
	}

	input := strings.TrimSpace(scanner.Text())
	if isMenuNav(input) != "" {
		return
	}

	delay, err := parseLaunchDelayInput(input)
	if err != nil {
		showInputErrorAndPause(err.Error())
		return
	}

	cfg.LaunchDelay = delay
	if err := config.Save(cfg); err != nil {
		showInputErrorAndPause(fmt.Sprintf("設定儲存失敗：%v", err))
		return
	}

	fmt.Printf("  ✔ 已更新啟動間隔：%s\n", delay.DisplayString())
	fmt.Println()
}

func parseLaunchDelayInput(input string) (config.LaunchDelayRange, error) {
	return config.ParseLaunchDelayRange(strings.TrimSpace(input))
}

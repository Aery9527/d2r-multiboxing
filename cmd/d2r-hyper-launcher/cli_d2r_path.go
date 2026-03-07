package main

import (
	"bufio"
	"fmt"
	"strings"

	"d2rhl/internal/common/config"
)

func ensureLaunchReadyD2RPath(cfg *config.Config, scanner *bufio.Scanner) bool {
	return ensureLaunchReadyD2RPathWithSetup(cfg, scanner, setupD2RPath)
}

func ensureLaunchReadyD2RPathWithSetup(cfg *config.Config, scanner *bufio.Scanner, setup func(*config.Config) bool) bool {
	for {
		err := config.ValidateD2RPath(cfg.D2RPath)
		if err == nil {
			return true
		}

		fmt.Println()
		fmt.Printf("  ⚠ 找不到可啟動的 D2R.exe：%s\n", cfg.D2RPath)
		fmt.Printf("  原因：%v\n", err)
		fmt.Println("  請先設定正確的 D2R.exe 路徑，完成後再繼續啟動。")
		fmt.Println("  p       立即設定 D2R.exe 路徑")
		printSubMenuNav()
		fmt.Print("  > 請選擇：")
		if !scanner.Scan() {
			return false
		}

		input := strings.TrimSpace(scanner.Text())
		if nav := isMenuNav(input); nav != "" {
			return false
		}
		if strings.EqualFold(input, "p") {
			if !setup(cfg) {
				return false
			}
			continue
		}

		showInputErrorAndPause("無效輸入，請輸入 p / b / h / q。")
	}
}

func setupD2RPath(cfg *config.Config) bool {
	fmt.Println()
	fmt.Println("  === 設定 D2R 路徑 ===")
	fmt.Println("  即將開啟 Windows 檔案選擇視窗，請選擇 D2R.exe。")

	selectedPath, err := PickD2RPath(cfg.D2RPath)
	if err != nil {
		fmt.Printf("  ⚠ D2R 路徑設定失敗：%v\n", err)
		fmt.Println()
		return false
	}
	if selectedPath == "" {
		fmt.Println("  已取消。")
		fmt.Println()
		return false
	}

	cfg.D2RPath = selectedPath
	if err := config.Save(cfg); err != nil {
		fmt.Printf("  ⚠ 設定儲存失敗：%v\n", err)
		fmt.Println()
		return false
	}

	fmt.Printf("  ✔ 已更新 D2R 路徑：%s\n", cfg.D2RPath)
	fmt.Println()
	return true
}

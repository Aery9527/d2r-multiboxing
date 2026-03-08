package main

import (
	"os/exec"

	"golang.org/x/sys/windows"
)

func handleCreatedAccountsFile(cfgDir, accountsFile string) {
	ui.successf("已自動建立帳號設定檔 accounts.csv，建立位置：%s", accountsFile)
	ui.infof("已先幫你放入兩筆範例資料，請把它們改成你自己的 Battle.net 帳號。")
	ui.infof("CSV 格式：Email,Password,DisplayName,LaunchFlags")
	ui.infof("範例：your-account1@example.com,your-password-here,主帳號-法師(倉庫/武器/飾品),")
	ui.infof("LaunchFlags 可先留空；之後可回到工具主選單用 [f] 再設定各帳號的啟動旗標。")
	ui.blankLine()
	ui.promptf("按任意鍵後，程式會結束並自動開啟資料目錄，方便你直接修改剛建立好的 accounts.csv。")

	if err := ui.anyKeyContinue(); err != nil {
		ui.warningf("等待按鍵失敗：%v", err)
		return
	}

	if err := openFolder(cfgDir); err != nil {
		ui.warningf("無法自動開啟資料目錄：%v", err)
	}
}

func waitForAnyKey() error {
	inputHandle, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE)
	if err != nil {
		return err
	}

	var originalMode uint32
	if err := windows.GetConsoleMode(inputHandle, &originalMode); err != nil {
		return err
	}
	defer func() {
		_ = windows.SetConsoleMode(inputHandle, originalMode)
	}()

	// Disable line buffering so a single key press is enough, while keeping
	// processed input enabled so Ctrl+C still behaves like an interrupt.
	rawMode := originalMode &^ (windows.ENABLE_LINE_INPUT | windows.ENABLE_ECHO_INPUT)
	if err := windows.SetConsoleMode(inputHandle, rawMode); err != nil {
		return err
	}

	var (
		buffer [1]uint16
		read   uint32
	)
	return windows.ReadConsole(inputHandle, &buffer[0], 1, &read, nil)
}

func consoleSupportsSingleKeyContinue() bool {
	inputHandle, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE)
	if err != nil {
		return false
	}

	var mode uint32
	return windows.GetConsoleMode(inputHandle, &mode) == nil
}

func openFolder(path string) error {
	cmd := exec.Command("explorer.exe", path)
	return cmd.Start()
}

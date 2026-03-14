package main

import (
	"os/exec"

	"golang.org/x/sys/windows"
)

func handleCreatedAccountsFile(cfgDir, accountsFile string) {
	ui.successf(lang.Accounts.CreatedOK, accountsFile)
	ui.infof("%s", lang.Accounts.CreatedInfo1)
	ui.infof("%s", lang.Accounts.CreatedInfo2)
	ui.infof("%s", lang.Accounts.CreatedInfo3)
	ui.infof("%s", lang.Accounts.CreatedInfo4)
	ui.blankLine()
	ui.promptf("%s", lang.Accounts.CreatedPressAny)

	if err := ui.anyKeyContinue(); err != nil {
		ui.warningf(lang.Common.WaitKeyFailed, err)
		return
	}

	if err := openFolder(cfgDir); err != nil {
		ui.warningf(lang.Accounts.OpenFolderFailed, err)
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

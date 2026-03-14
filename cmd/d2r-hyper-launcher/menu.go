package main

import (
	"fmt"
	"os"
	"strings"

	"d2rhl/internal/common/config"
	"d2rhl/internal/multiboxing/account"
	"d2rhl/internal/switcher"
)

const (
	menuBack = "b"
	menuHome = "h"
	menuQuit = "q"
)

func printStartupAnnouncement(cfgDir string) {
	ui.infof(lang.Startup.VersionLabel, displayReleaseSummary(version, releaseTime))
	ui.infof(lang.Startup.DataDirLabel, cfgDir)
	ui.warningLines(
		lang.Startup.WarnStatusDetect1,
		lang.Startup.WarnStatusDetect2,
		lang.Startup.WarnStatusDetect3,
	)
	ui.warningLines(
		lang.Startup.WarnNote,
		lang.Startup.WarnNoteWindowed,
		lang.Startup.WarnNoteGamepad,
		lang.Startup.WarnNoteSwitcher,
		lang.Startup.WarnNoteDelay,
		lang.Startup.WarnNoteConfig,
		lang.Startup.WarnNoteBattleNet,
		lang.Startup.WarnNoteNoModify,
		lang.Startup.WarnNoteCommunity,
	)
}

func printStartupHeader() {
	ui.headf("d2r-hyper-launcher (%s)", displayVersion(version))
}

func pauseAfterStartupAnnouncement() {
	if err := ui.anyKeyContinue(); err != nil {
		ui.warningf(lang.Common.WaitKeyFailed, err)
	}
}

func printMenu(accounts []account.Account, cfg *config.Config) {
	ui.headf("%s", lang.MainMenu.Title)
	printAccountList(accounts)

	ui.blankLine()
	options := ui.mainMenuOptions(func(options *cliMenuOptions) {
		options.option(lang.MainMenu.OptByNumberKey, lang.MainMenu.OptByNumber, "")
		options.option("0", lang.MainMenu.OptOffline, lang.MainMenu.OptOfflineComment)
		options.option("a", lang.MainMenu.OptLaunchAll, lang.MainMenu.OptLaunchAllComment)
		options.option("d", lang.MainMenu.OptDelay, displayDelay(cfg.LaunchDelay))
		options.option("f", lang.MainMenu.OptFlags, lang.MainMenu.OptFlagsComment)
		options.option("p", lang.MainMenu.OptD2RPath, cfg.D2RPath)
		options.option("s", lang.MainMenu.OptSwitcher, switcherMenuOptionStatus(cfg))
		options.option("r", lang.MainMenu.OptRefresh, "")
		options.option("l", lang.MainMenu.OptLanguage, "")
	})
	ui.menuBlock(func() {
		options.render()
	})
}

func switcherMenuOptionStatus(cfg *config.Config) string {
	display, ok := switcherSavedDisplay(cfg)
	if !ok {
		return lang.MainMenu.SwitcherNotSet
	}
	if !cfg.Switcher.Enabled {
		return fmt.Sprintf(lang.MainMenu.SwitcherDisabled, display)
	}
	return fmt.Sprintf(lang.MainMenu.SwitcherEnabled, display)
}

func switcherSavedDisplay(cfg *config.Config) (string, bool) {
	if cfg == nil || cfg.Switcher == nil || cfg.Switcher.Key == "" {
		return "", false
	}
	return switcher.FormatSwitcherDisplay(cfg.Switcher.Modifiers, cfg.Switcher.Key, cfg.Switcher.GamepadIndex), true
}

func isMenuNav(input string) string {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case menuBack:
		return "back"
	case menuHome:
		return "home"
	case menuQuit:
		ui.infof("%s", lang.Common.Goodbye)
		os.Exit(0)
		return ""
	default:
		return ""
	}
}

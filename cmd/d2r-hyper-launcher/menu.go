package main

import (
	"errors"
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
	printAccountList(accounts, runningStatusLabel)

	ui.blankLine()
	options := ui.mainMenuOptions(func(options *cliMenuOptions) {
		options.option(lang.MainMenu.OptByNumberKey, lang.MainMenu.OptByNumber, "")
		options.option("0", lang.MainMenu.OptOffline, lang.MainMenu.OptOfflineComment)
		options.option("a", lang.MainMenu.OptLaunchAll, lang.MainMenu.OptLaunchAllComment)
		options.option("d", lang.MainMenu.OptDelay, displayDelay(cfg.LaunchDelay))
		options.option("f", lang.MainMenu.OptFlags, lang.MainMenu.OptFlagsComment)
		options.option("g", lang.MainMenu.OptGraphicsProfiles, lang.MainMenu.OptGraphicsProfilesComment)
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

// ErrNavHome is returned by runMenu when the player presses h.
// Nested submenu callers must propagate this to ensure h reaches the main menu.
var ErrNavHome = errors.New("nav:home")

// errNavDone signals that a runMenu handler has completed its action.
// runMenu exits the current loop and returns nil to the caller (equivalent to pressing b).
var errNavDone = errors.New("nav:done")

// runMenu is the canonical submenu input loop. It handles b/h/q navigation
// centrally. All submenu loops MUST use runMenu (or runMenuRead) so that h
// always propagates correctly to the main menu regardless of nesting depth.
//
// display is called before each input read (pass nil if not needed).
// handle receives non-nav input. Return values:
//   - nil        → continue loop
//   - errNavDone → exit loop, returning nil to caller
//   - ErrNavHome → exit loop and propagate ErrNavHome to caller
func runMenu(display func(), handle func(input string) error) error {
	return runMenuRead(display, ui.readInput, handle)
}

// runMenuRead is like runMenu but accepts a custom read function.
// Use when a non-standard input prompt is needed (e.g. ui.readInputf).
func runMenuRead(display func(), readFn func() (string, bool), handle func(input string) error) error {
	for {
		if display != nil {
			display()
		}
		input, ok := readFn()
		if !ok {
			return nil
		}
		switch isMenuNav(input) {
		case "back":
			return nil
		case "home":
			return ErrNavHome
		}
		err := handle(input)
		if errors.Is(err, ErrNavHome) {
			return ErrNavHome
		}
		if errors.Is(err, errNavDone) {
			return nil
		}
	}
}

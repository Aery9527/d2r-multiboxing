package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"d2rhl/internal/common/config"
	"d2rhl/internal/multiboxing/account"
	"d2rhl/internal/multiboxing/mods"
)

func setupAccountDefaultMods(accounts []account.Account, accountsFile string, cfg *config.Config) {
	if len(accounts) == 0 {
		ui.infof("%s", lang.DefaultMods.NoAccounts)
		ui.blankLine()
		return
	}
	if !ensureLaunchReadyD2RPath(cfg) {
		return
	}

	installedMods, ok := discoverInstalledMods(cfg.D2RPath)
	if !ok {
		return
	}

	_ = runMenu(func() {
		ui.headf("%s", lang.DefaultMods.Title)
		ui.infof("%s", lang.DefaultMods.Intro1)
		ui.infof("%s", lang.DefaultMods.Intro2)
		ui.infof("%s", lang.DefaultMods.Intro3)
		ui.blankLine()
		printAccountList(accounts, func(acc account.Account) string {
			return defaultModStatusLabel(acc, installedMods)
		})
		ui.blankLine()
		printDefaultModList(installedMods)
		ui.blankLine()
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option("1", lang.DefaultMods.OptAssign, "")
			options.option("2", lang.DefaultMods.OptClear, "")
		})
		ui.menuBlock(func() {
			options.render()
		})
	}, func(choice string) error {
		switch choice {
		case "1":
			return assignDefaultMods(accounts, accountsFile, installedMods)
		case "2":
			return clearDefaultMods(accounts, accountsFile, installedMods)
		default:
			showInvalidInputAndPause()
			return nil
		}
	})
}

func assignDefaultMods(accounts []account.Account, accountsFile string, installedMods []string) error {
	return runMenu(func() {
		ui.headf("%s", lang.DefaultMods.AssignModeTitle)
		ui.infof("%s", lang.DefaultMods.AssignModeQuestion)
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option("1", lang.DefaultMods.OptModToAccounts, "")
			options.option("2", lang.DefaultMods.OptAccountToMod, "")
		})
		ui.menuBlock(func() {
			options.render()
		})
	}, func(choice string) error {
		switch choice {
		case "1":
			return assignDefaultModsByMod(accounts, accountsFile, installedMods)
		case "2":
			return assignDefaultModsByAccount(accounts, accountsFile, installedMods)
		default:
			showInvalidInputAndPause()
			return nil
		}
	})
}

func assignDefaultModsByMod(accounts []account.Account, accountsFile string, installedMods []string) error {
	var completed bool

	err := runMenuRead(
		func() {
			ui.headf("%s", lang.DefaultMods.AssignByModTitle)
			ui.menuBlock(func() {
				renderLaunchModOptions(installedMods)
			})
		},
		func() (string, bool) {
			return ui.readInputf("%s", lang.DefaultMods.AssignByModSelectPrompt)
		},
		func(input string) error {
			selectedMod, ok := parseLaunchModInput(input, installedMods)
			if !ok {
				showInvalidInputAndPause()
				return nil
			}

			modLabel := defaultModOptionLabel(selectedMod)
			var done bool
			innerErr := runMenuRead(
				func() {
					ui.headf("%s", lang.DefaultMods.AssignByModAccountTitle)
					options := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
						for i, acc := range accounts {
							menuOptions.option(strconv.Itoa(i+1), fmt.Sprintf("%s (%s)", acc.DisplayName, acc.Email), fmt.Sprintf(lang.DefaultMods.AccountComment, defaultModStatusLabel(acc, installedMods)))
						}
					})
					ui.menuBlock(func() {
						options.render()
					})
				},
				func() (string, bool) {
					return ui.readInputf(lang.DefaultMods.AssignByModAccountPrompt, modLabel)
				},
				func(input string) error {
					accountIndexes, err := parseSelectionInput(input, len(accounts))
					if err != nil {
						showInputErrorAndPause(fmt.Sprintf(lang.Common.ParseFailed, err))
						return nil
					}

					ui.blankLine()
					ui.infof(lang.DefaultMods.AssignByModAbout, modLabel)
					for _, idx := range accountIndexes {
						acc := accounts[idx]
						ui.rawlnf(lang.DefaultMods.AccountItemFmt, idx+1, acc.DisplayName, acc.Email, defaultModStatusLabel(acc, installedMods))
					}
					if !confirmChanges() {
						ui.infof("%s", lang.Common.Cancelled)
						ui.blankLine()
						return nil
					}

					if err := applyDefaultModAssignments(accounts, accountsFile, accountIndexes, selectedMod); err != nil {
						showInputErrorAndPause(fmt.Sprintf(lang.Common.SaveFailed, err))
						return nil
					}

					ui.successf(lang.DefaultMods.AssignDone, modLabel)
					ui.blankLine()
					done = true
					return errNavDone
				},
			)
			if errors.Is(innerErr, ErrNavHome) {
				return ErrNavHome
			}
			if done {
				completed = true
				return errNavDone
			}
			return nil
		},
	)
	if errors.Is(err, ErrNavHome) {
		return ErrNavHome
	}
	if completed {
		return errNavDone
	}
	return nil
}

func assignDefaultModsByAccount(accounts []account.Account, accountsFile string, installedMods []string) error {
	var completed bool

	err := runMenuRead(
		func() {
			ui.headf("%s", lang.DefaultMods.AssignByAccountTitle)
			options := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
				for i, acc := range accounts {
					menuOptions.option(strconv.Itoa(i+1), fmt.Sprintf("%s (%s)", acc.DisplayName, acc.Email), fmt.Sprintf(lang.DefaultMods.AccountComment, defaultModStatusLabel(acc, installedMods)))
				}
			})
			ui.menuBlock(func() {
				options.render()
			})
		},
		func() (string, bool) {
			return ui.readInputf("%s", lang.DefaultMods.AssignByAccountSelectPrompt)
		},
		func(input string) error {
			selected, err := strconv.Atoi(strings.TrimSpace(input))
			if err != nil || selected < 1 || selected > len(accounts) {
				showInvalidInputAndPause()
				return nil
			}

			accountIndex := selected - 1
			acc := accounts[accountIndex]
			var done bool
			innerErr := runMenuRead(
				func() {
					ui.headf("%s", lang.DefaultMods.AssignByAccountModTitle)
					ui.infof(lang.DefaultMods.AssignByAccountModPrompt, acc.DisplayName)
					ui.menuBlock(func() {
						renderLaunchModOptions(installedMods)
					})
				},
				func() (string, bool) {
					return ui.readInput()
				},
				func(input string) error {
					selectedMod, ok := parseLaunchModInput(input, installedMods)
					if !ok {
						showInvalidInputAndPause()
						return nil
					}

					modLabel := defaultModOptionLabel(selectedMod)
					ui.blankLine()
					ui.infof(lang.DefaultMods.AssignByAccountAbout, acc.DisplayName, modLabel)
					ui.rawlnf(lang.DefaultMods.AccountItemFmt, selected, acc.DisplayName, acc.Email, defaultModStatusLabel(acc, installedMods))
					if !confirmChanges() {
						ui.infof("%s", lang.Common.Cancelled)
						ui.blankLine()
						return nil
					}

					if err := applyDefaultModAssignments(accounts, accountsFile, []int{accountIndex}, selectedMod); err != nil {
						showInputErrorAndPause(fmt.Sprintf(lang.Common.SaveFailed, err))
						return nil
					}

					ui.successf(lang.DefaultMods.AssignDone, modLabel)
					ui.blankLine()
					done = true
					return errNavDone
				},
			)
			if errors.Is(innerErr, ErrNavHome) {
				return ErrNavHome
			}
			if done {
				completed = true
				return errNavDone
			}
			return nil
		},
	)
	if errors.Is(err, ErrNavHome) {
		return ErrNavHome
	}
	if completed {
		return errNavDone
	}
	return nil
}

func clearDefaultMods(accounts []account.Account, accountsFile string, installedMods []string) error {
	assignedIndexes := assignedDefaultModAccountIndexes(accounts)
	if len(assignedIndexes) == 0 {
		showInfoAndPause(lang.DefaultMods.ClearNoAssignments)
		return nil
	}

	err := runMenuRead(
		func() {
			ui.headf("%s", lang.DefaultMods.ClearTitle)
			options := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
				for i, accountIndex := range assignedIndexes {
					acc := accounts[accountIndex]
					menuOptions.option(strconv.Itoa(i+1), fmt.Sprintf("%s (%s)", acc.DisplayName, acc.Email), fmt.Sprintf(lang.DefaultMods.AccountComment, defaultModStatusLabel(acc, installedMods)))
				}
			})
			ui.menuBlock(func() {
				options.render()
			})
		},
		func() (string, bool) {
			return ui.readInputf("%s", lang.DefaultMods.ClearPrompt)
		},
		func(input string) error {
			selectionIndexes, err := parseSelectionInput(input, len(assignedIndexes))
			if err != nil {
				showInputErrorAndPause(fmt.Sprintf(lang.Common.ParseFailed, err))
				return nil
			}

			actualIndexes := make([]int, 0, len(selectionIndexes))
			ui.blankLine()
			ui.infof("%s", lang.DefaultMods.ClearAbout)
			for _, idx := range selectionIndexes {
				accountIndex := assignedIndexes[idx]
				actualIndexes = append(actualIndexes, accountIndex)
				acc := accounts[accountIndex]
				ui.rawlnf(lang.DefaultMods.AccountItemFmt, accountIndex+1, acc.DisplayName, acc.Email, defaultModStatusLabel(acc, installedMods))
			}

			if !confirmChanges() {
				ui.infof("%s", lang.Common.Cancelled)
				ui.blankLine()
				return nil
			}

			if err := clearDefaultModAssignments(accounts, accountsFile, actualIndexes); err != nil {
				showInputErrorAndPause(fmt.Sprintf(lang.Common.SaveFailed, err))
				return nil
			}

			ui.successf("%s", lang.DefaultMods.ClearDone)
			ui.blankLine()
			return errNavDone
		},
	)
	if errors.Is(err, ErrNavHome) {
		return ErrNavHome
	}
	return nil
}

func applyDefaultModAssignments(accounts []account.Account, accountsFile string, accountIndexes []int, defaultMod string) error {
	previous := make(map[int]string, len(accountIndexes))
	normalizedDefaultMod := mods.NormalizeSavedDefaultMod(defaultMod)
	for _, idx := range accountIndexes {
		previous[idx] = accounts[idx].DefaultMod
		accounts[idx].DefaultMod = normalizedDefaultMod
	}

	if err := account.SaveAccounts(accountsFile, accounts); err != nil {
		for idx, previousMod := range previous {
			accounts[idx].DefaultMod = previousMod
		}
		return err
	}
	return nil
}

func clearDefaultModAssignments(accounts []account.Account, accountsFile string, accountIndexes []int) error {
	return applyDefaultModAssignments(accounts, accountsFile, accountIndexes, "")
}

func defaultModStatusLabel(acc account.Account, installedMods []string) string {
	savedDefault := mods.NormalizeSavedDefaultMod(acc.DefaultMod)
	switch {
	case savedDefault == "":
		return lang.DefaultMods.StatusUnassigned
	case savedDefault == mods.DefaultModVanilla:
		return lang.DefaultMods.StatusVanilla
	}

	resolved := mods.ResolveSavedDefaultMod(savedDefault, installedMods)
	if resolved != "" {
		return resolved
	}
	return fmt.Sprintf(lang.DefaultMods.StatusMissing, savedDefault)
}

func defaultModOptionLabel(defaultMod string) string {
	if mods.IsDefaultModVanilla(defaultMod) {
		return lang.DefaultMods.StatusVanilla
	}
	return strings.TrimSpace(defaultMod)
}

func printDefaultModList(installedMods []string) {
	ui.infof("%s", lang.DefaultMods.ModListHeader)
	ui.rawlnf("  [0] %s", lang.DefaultMods.StatusVanilla)
	if len(installedMods) == 0 {
		ui.rawlnf("  - %s", lang.DefaultMods.NoInstalledMods)
		return
	}
	for i, modName := range installedMods {
		ui.rawlnf("  [%d] %s", i+1, modName)
	}
}

func assignedDefaultModAccountIndexes(accounts []account.Account) []int {
	indexes := make([]int, 0, len(accounts))
	for i, acc := range accounts {
		if mods.NormalizeSavedDefaultMod(acc.DefaultMod) == "" {
			continue
		}
		indexes = append(indexes, i)
	}
	return indexes
}

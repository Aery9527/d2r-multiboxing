package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"d2rhl/internal/multiboxing/account"
	"d2rhl/internal/multiboxing/graphicsprofile"
)

var newGraphicsProfileStore = graphicsprofile.NewDefaultStore

func setupAccountGraphicsProfiles(accounts []account.Account, accountsFile string) {
	if len(accounts) == 0 {
		ui.infof("%s", lang.GraphicsProfiles.NoAccounts)
		ui.blankLine()
		return
	}

	_ = runMenu(func() {
		ui.headf("%s", lang.GraphicsProfiles.Title)
		ui.infof("%s", lang.GraphicsProfiles.Intro1)
		ui.infof("%s", lang.GraphicsProfiles.Intro2)
		ui.infof("%s", lang.GraphicsProfiles.Intro3)
		ui.infof("%s", lang.GraphicsProfiles.Intro4)
		ui.blankLine()
		printAccountList(accounts, graphicsProfileStatusLabel)
		ui.blankLine()
		profiles, err := listGraphicsProfiles()
		if err != nil {
			ui.warningf(lang.GraphicsProfiles.StoreOpenFailed, err)
		} else {
			printGraphicsProfileList(profiles)
		}
		ui.blankLine()
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option("1", lang.GraphicsProfiles.OptSaveCurrent, "")
			options.option("2", lang.GraphicsProfiles.OptAssign, "")
			options.option("3", lang.GraphicsProfiles.OptClear, "")
			options.option("4", lang.GraphicsProfiles.OptDeleteSaved, "")
		})
		ui.menuBlock(func() {
			options.render()
		})
	}, func(choice string) error {
		switch choice {
		case "1":
			return saveCurrentGraphicsProfile()
		case "2":
			return assignGraphicsProfiles(accounts, accountsFile)
		case "3":
			return clearGraphicsProfiles(accounts, accountsFile)
		case "4":
			return deleteSavedGraphicsProfiles(accounts)
		default:
			showInvalidInputAndPause()
			return nil
		}
	})
}

func saveCurrentGraphicsProfile() error {
	store, err := newGraphicsProfileStore()
	if err != nil {
		showInputErrorAndPause(fmt.Sprintf(lang.GraphicsProfiles.StoreOpenFailed, err))
		return nil
	}

	var profiles []string
	err = runMenuRead(
		func() {
			profiles, err = store.List()
			if err != nil {
				ui.warningf(lang.GraphicsProfiles.StoreOpenFailed, err)
				profiles = nil
			}
			ui.headf("%s", lang.GraphicsProfiles.SaveTitle)
			ui.infoLines(
				lang.GraphicsProfiles.SaveIntro1,
				lang.GraphicsProfiles.SaveIntro2,
			)
			ui.infof(lang.GraphicsProfiles.CurrentSettingsLabel, store.SettingsPath())
			ui.blankLine()
			ui.infof("%s", lang.GraphicsProfiles.ProfileListHeader)
			options := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
				for i, profileName := range profiles {
					menuOptions.option(strconv.Itoa(i+1), profileName, lang.GraphicsProfiles.SaveOptionComment)
				}
			})
			ui.menuBlock(func() {
				if len(profiles) == 0 {
					ui.infof("%s", lang.GraphicsProfiles.NoProfiles)
					ui.blankLine()
				}
				options.render()
			})
		},
		func() (string, bool) {
			return ui.readInputf("%s", lang.GraphicsProfiles.SavePrompt)
		},
		func(input string) error {
			rawInput := strings.TrimSpace(input)
			profileName := rawInput
			overwrite := false
			if selected, parseErr := strconv.Atoi(rawInput); parseErr == nil {
				if selected < 1 || selected > len(profiles) {
					showInputErrorAndPause(fmt.Sprintf(lang.GraphicsProfiles.SaveInvalidProfileID, selected))
					return nil
				}
				profileName = profiles[selected-1]
				overwrite = true
			} else {
				if err := graphicsprofile.ValidateProfileName(profileName); err != nil {
					showInputErrorAndPause(fmt.Sprintf(lang.GraphicsProfiles.SaveInvalidName, err))
					return nil
				}

				exists, err := store.Exists(profileName)
				if err != nil {
					showInputErrorAndPause(fmt.Sprintf(lang.GraphicsProfiles.SaveFailed, err))
					return nil
				}
				if exists {
					showInputErrorAndPause(fmt.Sprintf(lang.GraphicsProfiles.SaveExistingUseNumber, profileName))
					return nil
				}
			}

			if err := store.SaveCurrentAs(profileName, overwrite); err != nil {
				showInputErrorAndPause(fmt.Sprintf(lang.GraphicsProfiles.SaveFailed, err))
				return nil
			}

			ui.successf(lang.GraphicsProfiles.SaveDone, profileName)
			ui.blankLine()
			return errNavDone
		},
	)
	if errors.Is(err, ErrNavHome) {
		return ErrNavHome
	}
	return nil
}

func assignGraphicsProfiles(accounts []account.Account, accountsFile string) error {
	profiles, err := listGraphicsProfiles()
	if err != nil {
		showInputErrorAndPause(fmt.Sprintf(lang.GraphicsProfiles.StoreOpenFailed, err))
		return nil
	}
	if len(profiles) == 0 {
		showInfoAndPause(lang.GraphicsProfiles.AssignNoProfiles)
		return nil
	}

	return runMenu(func() {
		ui.headf("%s", lang.GraphicsProfiles.AssignModeTitle)
		ui.infof("%s", lang.GraphicsProfiles.AssignModeQuestion)
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option("1", lang.GraphicsProfiles.OptProfileToAccounts, "")
			options.option("2", lang.GraphicsProfiles.OptAccountToProfile, "")
		})
		ui.menuBlock(func() {
			options.render()
		})
	}, func(choice string) error {
		switch choice {
		case "1":
			return assignGraphicsProfilesByProfile(accounts, accountsFile, profiles)
		case "2":
			return assignGraphicsProfilesByAccount(accounts, accountsFile, profiles)
		default:
			showInvalidInputAndPause()
			return nil
		}
	})
}

func assignGraphicsProfilesByProfile(accounts []account.Account, accountsFile string, profiles []string) error {
	var completed bool

	err := runMenuRead(
		func() {
			ui.headf("%s", lang.GraphicsProfiles.AssignByProfileTitle)
			ui.menuBlock(func() {
				renderGraphicsProfileOptions(profiles)
			})
		},
		func() (string, bool) {
			return ui.readInputf("%s", lang.GraphicsProfiles.AssignByProfileSelectPrompt)
		},
		func(input string) error {
			selected, err := strconv.Atoi(strings.TrimSpace(input))
			if err != nil || selected < 1 || selected > len(profiles) {
				showInvalidInputAndPause()
				return nil
			}

			profileName := profiles[selected-1]
			var done bool
			innerErr := runMenuRead(
				func() {
					ui.headf("%s", lang.GraphicsProfiles.AssignByProfileAccountTitle)
					options := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
						for i, acc := range accounts {
							menuOptions.option(strconv.Itoa(i+1), fmt.Sprintf("%s (%s)", acc.DisplayName, acc.Email), fmt.Sprintf(lang.GraphicsProfiles.AccountComment, graphicsProfileStatusLabel(acc)))
						}
					})
					ui.menuBlock(func() {
						options.render()
					})
				},
				func() (string, bool) {
					return ui.readInputf(lang.GraphicsProfiles.AssignByProfileAccountPrompt, profileName)
				},
				func(input string) error {
					accountIndexes, err := parseSelectionInput(input, len(accounts))
					if err != nil {
						showInputErrorAndPause(fmt.Sprintf(lang.Common.ParseFailed, err))
						return nil
					}

					ui.blankLine()
					ui.infof(lang.GraphicsProfiles.AssignByProfileAbout, profileName)
					for _, idx := range accountIndexes {
						acc := accounts[idx]
						ui.rawlnf(lang.GraphicsProfiles.AccountItemFmt, idx+1, acc.DisplayName, acc.Email, graphicsProfileStatusLabel(acc))
					}
					if !confirmChanges() {
						ui.infof("%s", lang.Common.Cancelled)
						ui.blankLine()
						return nil
					}

					if err := applyGraphicsProfileAssignments(accounts, accountsFile, accountIndexes, profileName); err != nil {
						showInputErrorAndPause(fmt.Sprintf(lang.GraphicsProfiles.SaveFailed, err))
						return nil
					}

					ui.successf(lang.GraphicsProfiles.AssignDone, profileName)
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

func assignGraphicsProfilesByAccount(accounts []account.Account, accountsFile string, profiles []string) error {
	var completed bool

	err := runMenuRead(
		func() {
			ui.headf("%s", lang.GraphicsProfiles.AssignByAccountTitle)
			options := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
				for i, acc := range accounts {
					menuOptions.option(strconv.Itoa(i+1), fmt.Sprintf("%s (%s)", acc.DisplayName, acc.Email), fmt.Sprintf(lang.GraphicsProfiles.AccountComment, graphicsProfileStatusLabel(acc)))
				}
			})
			ui.menuBlock(func() {
				options.render()
			})
		},
		func() (string, bool) {
			return ui.readInputf("%s", lang.GraphicsProfiles.AssignByAccountSelectPrompt)
		},
		func(input string) error {
			selected, err := strconv.Atoi(strings.TrimSpace(input))
			if err != nil || selected < 1 || selected > len(accounts) {
				showInvalidInputAndPause()
				return nil
			}

			acc := accounts[selected-1]
			var done bool
			innerErr := runMenuRead(
				func() {
					ui.headf("%s", lang.GraphicsProfiles.AssignByAccountProfileTitle)
					ui.infof(lang.GraphicsProfiles.AssignByAccountProfilePrompt, acc.DisplayName)
					options := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
						for i, profileName := range profiles {
							menuOptions.option(strconv.Itoa(i+1), profileName, "")
						}
					})
					ui.menuBlock(func() {
						options.render()
					})
				},
				func() (string, bool) {
					return ui.readInput()
				},
				func(input string) error {
					profileSelection, err := strconv.Atoi(strings.TrimSpace(input))
					if err != nil || profileSelection < 1 || profileSelection > len(profiles) {
						showInvalidInputAndPause()
						return nil
					}

					profileName := profiles[profileSelection-1]
					ui.blankLine()
					ui.infof(lang.GraphicsProfiles.AssignByAccountAbout, acc.DisplayName, profileName)
					ui.rawlnf(lang.GraphicsProfiles.AccountItemFmt, selected, acc.DisplayName, acc.Email, graphicsProfileStatusLabel(acc))
					if !confirmChanges() {
						ui.infof("%s", lang.Common.Cancelled)
						ui.blankLine()
						return nil
					}

					if err := applyGraphicsProfileAssignments(accounts, accountsFile, []int{selected - 1}, profileName); err != nil {
						showInputErrorAndPause(fmt.Sprintf(lang.GraphicsProfiles.SaveFailed, err))
						return nil
					}

					ui.successf(lang.GraphicsProfiles.AssignDone, profileName)
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

func clearGraphicsProfiles(accounts []account.Account, accountsFile string) error {
	assignedIndexes := assignedGraphicsProfileAccountIndexes(accounts)
	if len(assignedIndexes) == 0 {
		showInfoAndPause(lang.GraphicsProfiles.ClearNoAssignments)
		return nil
	}

	err := runMenuRead(
		func() {
			ui.headf("%s", lang.GraphicsProfiles.ClearTitle)
			options := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
				for i, accountIndex := range assignedIndexes {
					acc := accounts[accountIndex]
					menuOptions.option(strconv.Itoa(i+1), fmt.Sprintf("%s (%s)", acc.DisplayName, acc.Email), fmt.Sprintf(lang.GraphicsProfiles.AccountComment, graphicsProfileStatusLabel(acc)))
				}
			})
			ui.menuBlock(func() {
				options.render()
			})
		},
		func() (string, bool) {
			return ui.readInputf("%s", lang.GraphicsProfiles.ClearPrompt)
		},
		func(input string) error {
			selectionIndexes, err := parseSelectionInput(input, len(assignedIndexes))
			if err != nil {
				showInputErrorAndPause(fmt.Sprintf(lang.Common.ParseFailed, err))
				return nil
			}

			actualIndexes := make([]int, 0, len(selectionIndexes))
			ui.blankLine()
			ui.infof("%s", lang.GraphicsProfiles.ClearAbout)
			for _, idx := range selectionIndexes {
				accountIndex := assignedIndexes[idx]
				actualIndexes = append(actualIndexes, accountIndex)
				acc := accounts[accountIndex]
				ui.rawlnf(lang.GraphicsProfiles.AccountItemFmt, accountIndex+1, acc.DisplayName, acc.Email, graphicsProfileStatusLabel(acc))
			}

			if !confirmChanges() {
				ui.infof("%s", lang.Common.Cancelled)
				ui.blankLine()
				return nil
			}

			if err := clearGraphicsProfileAssignments(accounts, accountsFile, actualIndexes); err != nil {
				showInputErrorAndPause(fmt.Sprintf(lang.GraphicsProfiles.SaveFailed, err))
				return nil
			}

			ui.successf("%s", lang.GraphicsProfiles.ClearDone)
			ui.blankLine()
			return errNavDone
		},
	)
	if errors.Is(err, ErrNavHome) {
		return ErrNavHome
	}
	return nil
}

func deleteSavedGraphicsProfiles(accounts []account.Account) error {
	store, err := newGraphicsProfileStore()
	if err != nil {
		showInputErrorAndPause(fmt.Sprintf(lang.GraphicsProfiles.StoreOpenFailed, err))
		return nil
	}

	profiles, err := store.List()
	if err != nil {
		showInputErrorAndPause(fmt.Sprintf(lang.GraphicsProfiles.StoreOpenFailed, err))
		return nil
	}
	if len(profiles) == 0 {
		showInfoAndPause(lang.GraphicsProfiles.DeleteNoProfiles)
		return nil
	}

	err = runMenuRead(
		func() {
			ui.headf("%s", lang.GraphicsProfiles.DeleteTitle)
			options := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
				for i, profileName := range profiles {
					menuOptions.option(strconv.Itoa(i+1), profileName, deleteGraphicsProfileComment(accounts, profileName))
				}
			})
			ui.menuBlock(func() {
				options.render()
			})
		},
		func() (string, bool) {
			return ui.readInputf("%s", lang.GraphicsProfiles.DeletePrompt)
		},
		func(input string) error {
			selectionIndexes, err := parseSelectionInput(input, len(profiles))
			if err != nil {
				showInputErrorAndPause(fmt.Sprintf(lang.Common.ParseFailed, err))
				return nil
			}

			selectedProfiles := make([]string, 0, len(selectionIndexes))
			for _, idx := range selectionIndexes {
				profileName := profiles[idx]
				if warningMessage, inUse := deleteGraphicsProfileInUseMessage(accounts, profileName); inUse {
					showWarningAndPause(warningMessage)
					return nil
				}
				selectedProfiles = append(selectedProfiles, profileName)
			}

			ui.blankLine()
			ui.infof("%s", lang.GraphicsProfiles.DeleteAbout)
			for _, profileName := range selectedProfiles {
				ui.rawlnf("  - %s", profileName)
			}
			if !confirmChanges() {
				ui.infof("%s", lang.Common.Cancelled)
				ui.blankLine()
				return nil
			}

			if err := deleteGraphicsProfileSelections(store, selectedProfiles); err != nil {
				showInputErrorAndPause(fmt.Sprintf(lang.GraphicsProfiles.DeleteFailed, err))
				return nil
			}

			ui.successf("%s", lang.GraphicsProfiles.DeleteDone)
			ui.blankLine()
			return errNavDone
		},
	)
	if errors.Is(err, ErrNavHome) {
		return ErrNavHome
	}
	return nil
}

func applyGraphicsProfileAssignments(accounts []account.Account, accountsFile string, accountIndexes []int, profileName string) error {
	previous := make(map[int]string, len(accountIndexes))
	for _, idx := range accountIndexes {
		previous[idx] = accounts[idx].GraphicsProfile
		accounts[idx].GraphicsProfile = strings.TrimSpace(profileName)
	}

	if err := account.SaveAccounts(accountsFile, accounts); err != nil {
		for idx, previousProfile := range previous {
			accounts[idx].GraphicsProfile = previousProfile
		}
		return err
	}
	return nil
}

func clearGraphicsProfileAssignments(accounts []account.Account, accountsFile string, accountIndexes []int) error {
	return applyGraphicsProfileAssignments(accounts, accountsFile, accountIndexes, "")
}

func deleteGraphicsProfileSelections(store *graphicsprofile.Store, profileNames []string) error {
	for _, profileName := range profileNames {
		if err := store.Delete(profileName); err != nil {
			return err
		}
	}
	return nil
}

func graphicsProfileStatusLabel(acc account.Account) string {
	if strings.TrimSpace(acc.GraphicsProfile) == "" {
		return lang.GraphicsProfiles.StatusUnassigned
	}
	return acc.GraphicsProfile
}

func deleteGraphicsProfileComment(accounts []account.Account, profileName string) string {
	assigned := graphicsProfileAssignedAccountLabels(accounts, profileName)
	if len(assigned) == 0 {
		return lang.GraphicsProfiles.DeleteUnusedComment
	}
	return fmt.Sprintf(lang.GraphicsProfiles.DeleteUsedComment, len(assigned))
}

func deleteGraphicsProfileInUseMessage(accounts []account.Account, profileName string) (string, bool) {
	assigned := graphicsProfileAssignedAccountLabels(accounts, profileName)
	if len(assigned) == 0 {
		return "", false
	}
	return fmt.Sprintf(lang.GraphicsProfiles.DeleteInUse, profileName, strings.Join(assigned, ", ")), true
}

func graphicsProfileAssignedAccountLabels(accounts []account.Account, profileName string) []string {
	labels := make([]string, 0, len(accounts))
	normalizedProfile := strings.TrimSpace(profileName)
	for _, acc := range accounts {
		if !strings.EqualFold(strings.TrimSpace(acc.GraphicsProfile), normalizedProfile) {
			continue
		}
		labels = append(labels, graphicsProfileAccountLabel(acc))
	}
	return labels
}

func graphicsProfileAccountLabel(acc account.Account) string {
	displayName := strings.TrimSpace(acc.DisplayName)
	email := strings.TrimSpace(acc.Email)
	switch {
	case displayName == "":
		return email
	case email == "":
		return displayName
	default:
		return fmt.Sprintf("%s (%s)", displayName, email)
	}
}

func graphicsProfileAccountIndex(accounts []account.Account, target *account.Account) int {
	if target == nil {
		return -1
	}
	for i := range accounts {
		if &accounts[i] == target {
			return i
		}
	}

	targetDisplayName := strings.TrimSpace(target.DisplayName)
	targetEmail := strings.TrimSpace(target.Email)
	for i := range accounts {
		if !strings.EqualFold(strings.TrimSpace(accounts[i].DisplayName), targetDisplayName) {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(accounts[i].Email), targetEmail) {
			continue
		}
		return i
	}
	return -1
}

func listGraphicsProfiles() ([]string, error) {
	store, err := newGraphicsProfileStore()
	if err != nil {
		return nil, err
	}
	return store.List()
}

func printGraphicsProfileList(profiles []string) {
	ui.infof("%s", lang.GraphicsProfiles.ProfileListHeader)
	if len(profiles) == 0 {
		ui.rawlnf("  - %s", lang.GraphicsProfiles.NoProfiles)
		return
	}
	for i, profileName := range profiles {
		ui.rawlnf("  [%d] %s", i+1, profileName)
	}
}

func renderGraphicsProfileOptions(profiles []string) {
	options := ui.subMenuOptions(func(menuOptions *cliMenuOptions) {
		for i, profileName := range profiles {
			menuOptions.option(strconv.Itoa(i+1), profileName, "")
		}
	})
	options.render()
}

func assignedGraphicsProfileAccountIndexes(accounts []account.Account) []int {
	indexes := make([]int, 0, len(accounts))
	for i, acc := range accounts {
		if strings.TrimSpace(acc.GraphicsProfile) == "" {
			continue
		}
		indexes = append(indexes, i)
	}
	return indexes
}

func applyGraphicsProfileForLaunch(acc account.Account, store *graphicsprofile.Store) (*graphicsprofile.Store, error) {
	return applyNamedGraphicsProfileForLaunch(acc.GraphicsProfile, store)
}

func applyNamedGraphicsProfileForLaunch(profileName string, store *graphicsprofile.Store) (*graphicsprofile.Store, error) {
	normalizedProfile := strings.TrimSpace(profileName)
	if normalizedProfile == "" {
		return store, nil
	}
	if store == nil {
		var err error
		store, err = newGraphicsProfileStore()
		if err != nil {
			return nil, err
		}
	}

	ui.infof(lang.GraphicsProfiles.ApplyingDuringLaunch, normalizedProfile)
	if err := store.Apply(normalizedProfile); err != nil {
		return store, err
	}
	return store, nil
}

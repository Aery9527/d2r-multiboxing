package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"d2rhl/internal/common/config"
	"d2rhl/internal/common/d2r"
	"d2rhl/internal/common/process"
	"d2rhl/internal/multiboxing/account"
	"d2rhl/internal/multiboxing/graphicsprofile"
	"d2rhl/internal/multiboxing/launcher"
	"d2rhl/internal/multiboxing/mods"
)

var launchDelayRandIntN = rand.Intn
var launchDelaySleep = time.Sleep
var launchSuccessPauseSleep = time.Sleep

const launchSuccessPauseDuration = 3 * time.Second

type launchRegionChoice struct {
	ManualRegion *d2r.Region
	UseDefaults  bool
}

type launchModChoice struct {
	ManualMod   string
	HasManual   bool
	UseDefaults bool
}

func launchAccount(acc *account.Account, accounts []account.Account, accountsFile string, cfg *config.Config) {
	if !ensureLaunchReadyD2RPath(cfg) {
		return
	}
	if isAccountRunning(acc.DisplayName) {
		showWarningAndPause(fmt.Sprintf(lang.Launch.AlreadyRunning, acc.DisplayName))
		return
	}

	regionChoice, ok := promptLaunchRegion(lang.Launch.RegionSingleTitle, []*account.Account{acc})
	if !ok {
		return
	}

	installedMods, ok := discoverInstalledMods(cfg.D2RPath)
	if !ok {
		return
	}

	modChoice, ok := promptLaunchMod(lang.Launch.ModSingleTitle, accounts, accountsFile, []*account.Account{acc}, installedMods)
	if !ok {
		return
	}

	_, err := prepareGraphicsProfileForLaunch(accounts, accountsFile, acc, nil)
	if err != nil {
		showInputErrorAndPause(fmt.Sprintf(lang.GraphicsProfiles.ApplyFailed, acc.GraphicsProfile, err))
		return
	}

	password, err := account.GetDecryptedPassword(acc)
	if err != nil {
		ui.errorf(lang.Launch.DecryptFailed, err)
		return
	}

	region := resolveLaunchRegionChoice(regionChoice, *acc)
	if region == nil {
		showInputErrorAndPause(fmt.Sprintf(lang.Launch.RegionMissing, launchTargetAccountLabel(acc)))
		return
	}

	modArgs, modOK := resolveLaunchModChoice(modChoice, *acc, installedMods)
	if !modOK {
		showInputErrorAndPause(fmt.Sprintf(lang.Launch.ModMissing, launchTargetAccountLabel(acc)))
		return
	}

	ui.infof(lang.Launch.Starting, acc.DisplayName, region.Name)
	pid, err := launcher.LaunchD2R(cfg.D2RPath, acc.Email, password, region.Address, accountLaunchArgs(*acc, modArgs)...)
	if err != nil {
		ui.errorf(lang.Launch.LaunchFailed, err)
		return
	}
	ui.successf(lang.Launch.LaunchOK, pid)

	pauseAfterSuccessfulLaunch()
	closed, err := launcher.CloseHandlesByName(pid, d2r.SingleInstanceEventName)
	if err != nil {
		ui.warningf(lang.Launch.CloseHandleFailed, err)
	} else if closed > 0 {
		ui.successf(lang.Launch.HandlesClosed, closed)
	}

	renameLaunchedWindow(pid, acc.DisplayName)
	ui.blankLine()
}

func launchAll(accounts []account.Account, accountsFile string, cfg *config.Config) {
	if !ensureLaunchReadyD2RPath(cfg) {
		return
	}

	runningTitles := runningAccountWindowTitles()
	pendingAccounts := pendingBatchAccounts(accounts, runningTitles)
	ui.infof("%s", lang.Launch.BatchScanHeader)
	for _, line := range batchAccountStatusLines(accounts, runningTitles) {
		ui.rawln(line)
	}
	if len(pendingAccounts) == 0 {
		showInfoAndPause(lang.Launch.AllRunning)
		return
	}
	ui.infof(lang.Launch.BatchOnlyPending, len(pendingAccounts))

	regionChoice, ok := promptLaunchRegion(lang.Launch.RegionBatchTitle, pendingAccounts)
	if !ok {
		return
	}

	installedMods, ok := discoverInstalledMods(cfg.D2RPath)
	if !ok {
		return
	}

	modChoice, ok := promptLaunchMod(lang.Launch.ModBatchTitle, accounts, accountsFile, pendingAccounts, installedMods)
	if !ok {
		return
	}

	var graphicsStore *graphicsprofile.Store
	for i, acc := range pendingAccounts {
		var applyErr error
		graphicsStore, applyErr = prepareGraphicsProfileForLaunch(accounts, accountsFile, acc, graphicsStore)
		if applyErr != nil {
			ui.warningf(lang.GraphicsProfiles.BatchApplyFailed, acc.DisplayName, acc.GraphicsProfile, applyErr)
			continue
		}

		password, err := account.GetDecryptedPassword(acc)
		if err != nil {
			ui.warningf(lang.Launch.BatchDecryptFailed, acc.DisplayName, err)
			continue
		}

		region := resolveLaunchRegionChoice(regionChoice, *acc)
		if region == nil {
			ui.warningf(lang.Launch.RegionMissing, launchTargetAccountLabel(acc))
			continue
		}

		modArgs, modOK := resolveLaunchModChoice(modChoice, *acc, installedMods)
		if !modOK {
			ui.warningf(lang.Launch.ModMissing, launchTargetAccountLabel(acc))
			continue
		}

		ui.infof(lang.Launch.Starting, acc.DisplayName, region.Name)
		pid, err := launcher.LaunchD2R(cfg.D2RPath, acc.Email, password, region.Address, accountLaunchArgs(*acc, modArgs)...)
		if err != nil {
			ui.warningf(lang.Launch.BatchLaunchFailed, acc.DisplayName, err)
			continue
		}
		ui.successf(lang.Launch.BatchLaunchOK, acc.DisplayName, pid)

		pauseAfterSuccessfulLaunch()
		closed, err := launcher.CloseHandlesByName(pid, d2r.SingleInstanceEventName)
		if err != nil {
			ui.warningf(lang.Launch.BatchHandleCloseFailed, acc.DisplayName, err)
		} else if closed > 0 {
			ui.successf(lang.Launch.BatchHandlesClosed, acc.DisplayName, closed)
		}

		renameLaunchedWindow(pid, acc.DisplayName)

		if i+1 < len(pendingAccounts) {
			delaySeconds := cfg.LaunchDelay.NextSeconds(launchDelayRandIntN)
			waitForNextBatchLaunch(delaySeconds, pendingAccounts[i+1].DisplayName)
		}
	}
	ui.blankLine()
}

func prepareGraphicsProfileForLaunch(accounts []account.Account, accountsFile string, acc *account.Account, store *graphicsprofile.Store) (*graphicsprofile.Store, error) {
	profileName := strings.TrimSpace(acc.GraphicsProfile)
	store, err := applyGraphicsProfileForLaunch(*acc, store)
	if err == nil {
		return store, nil
	}
	if !errors.Is(err, graphicsprofile.ErrProfileNotFound) {
		return store, err
	}

	if clearErr := clearMissingGraphicsProfileAssignment(accounts, accountsFile, acc); clearErr != nil {
		return store, fmt.Errorf(lang.GraphicsProfiles.MissingProfileClearFailed, clearErr)
	}

	ui.warningf(lang.GraphicsProfiles.MissingProfileCleared, graphicsProfileAccountLabel(*acc), profileName)
	return store, nil
}

func clearMissingGraphicsProfileAssignment(accounts []account.Account, accountsFile string, acc *account.Account) error {
	accountIndex := graphicsProfileAccountIndex(accounts, acc)
	if accountIndex < 0 {
		return errors.New("target account was not found in current account list")
	}
	return clearGraphicsProfileAssignments(accounts, accountsFile, []int{accountIndex})
}

func launchOffline(cfg *config.Config) {
	if !ensureLaunchReadyD2RPath(cfg) {
		return
	}

	ui.headf("%s", lang.Launch.OfflineTitle)

	installedMods, ok := discoverInstalledMods(cfg.D2RPath)
	if !ok {
		return
	}

	modArgs, ok := selectOfflineLaunchMod(installedMods)
	if !ok {
		return
	}

	ui.infof("%s", lang.Launch.OfflineLaunching)
	pid, err := launcher.LaunchD2ROffline(cfg.D2RPath, modArgs...)
	if err != nil {
		ui.errorf(lang.Launch.OfflineLaunchFailed, err)
		return
	}
	ui.successf(lang.Launch.OfflineLaunchOK, pid)
	pauseAfterSuccessfulLaunch()
	ui.blankLine()
}

func promptLaunchRegion(title string, accounts []*account.Account) (launchRegionChoice, bool) {
	var result launchRegionChoice
	_ = runMenu(func() {
		ui.headf("%s", title)
		ui.infof("%s", lang.Launch.RegionTargetLabel)
		for _, line := range launchTargetAccountLines(accounts) {
			ui.rawln(line)
		}
		ui.infof("%s", lang.Launch.RegionUseDefaults)
		ui.infof("%s", lang.Launch.RegionOverride)
		options := ui.subMenuOptions(func(options *cliMenuOptions) {
			options.option("1", "NA", "")
			options.option("2", "EU", "")
			options.option("3", "Asia", "")
		})
		ui.menuBlock(func() {
			options.render()
		})
	}, func(input string) error {
		if strings.TrimSpace(input) == "" {
			missing := missingDefaultRegionAccountLabels(accounts)
			if len(missing) > 0 {
				showInputErrorAndPause(fmt.Sprintf(lang.Launch.RegionMissing, strings.Join(missing, ", ")))
				return nil
			}
			result.UseDefaults = true
			return errNavDone
		}

		region := parseRegionInput(input)
		if region == nil {
			showInputErrorAndPause(lang.Launch.RegionInvalid)
			return nil
		}
		result.ManualRegion = region
		return errNavDone
	})
	return result, result.UseDefaults || result.ManualRegion != nil
}

func promptLaunchMod(title string, accounts []account.Account, accountsFile string, targets []*account.Account, installedMods []string) (launchModChoice, bool) {
	var result launchModChoice
	_ = runMenu(func() {
		ui.headf("%s", title)
		ui.infof("%s", lang.Launch.RegionTargetLabel)
		for _, line := range launchTargetAccountLines(targets) {
			ui.rawln(line)
		}
		ui.infof("%s", lang.Launch.ModUseDefaults)
		ui.infof("%s", lang.Launch.ModOverride)
		if len(installedMods) == 0 {
			ui.infof("%s", lang.Launch.ModNoMods)
		}
		ui.menuBlock(func() {
			renderLaunchModOptions(installedMods)
		})
	}, func(input string) error {
		if strings.TrimSpace(input) == "" {
			if err := reconcileDefaultModAssignmentsForLaunch(accounts, accountsFile, targets, installedMods); err != nil {
				showInputErrorAndPause(fmt.Sprintf(lang.Common.SaveFailed, err))
				return nil
			}
			missing := missingDefaultModAccountLabels(targets, installedMods)
			if len(missing) > 0 {
				showInputErrorAndPause(fmt.Sprintf(lang.Launch.ModMissing, strings.Join(missing, ", ")))
				return nil
			}
			result.UseDefaults = true
			return errNavDone
		}

		selectedMod, ok := parseLaunchModInput(input, installedMods)
		if !ok {
			showInvalidInputAndPause()
			return nil
		}
		if selectedMod == mods.DefaultModVanilla {
			ui.infof("%s", lang.Launch.ModNoneChosen)
		} else {
			ui.infof(lang.Launch.ModUsing, selectedMod)
		}
		result.ManualMod = selectedMod
		result.HasManual = true
		return errNavDone
	})
	return result, result.UseDefaults || result.HasManual
}

func launchTargetAccountLines(accounts []*account.Account) []string {
	lines := make([]string, 0, len(accounts))
	for _, acc := range accounts {
		if acc == nil {
			continue
		}
		lines = append(lines, fmt.Sprintf("  %s", launchTargetAccountLabel(acc)))
	}
	return lines
}

func missingDefaultRegionAccountLabels(accounts []*account.Account) []string {
	labels := make([]string, 0, len(accounts))
	for _, acc := range accounts {
		if acc == nil {
			continue
		}
		if d2r.NormalizeRegionName(acc.DefaultRegion) != "" {
			continue
		}
		labels = append(labels, launchTargetAccountLabel(acc))
	}
	return labels
}

func resolveLaunchRegionChoice(choice launchRegionChoice, acc account.Account) *d2r.Region {
	if choice.ManualRegion != nil {
		return choice.ManualRegion
	}
	if !choice.UseDefaults {
		return nil
	}
	return d2r.FindRegion(acc.DefaultRegion)
}

func reconcileDefaultModAssignmentsForLaunch(accounts []account.Account, accountsFile string, targets []*account.Account, installedMods []string) error {
	previous := make(map[int]string, len(targets))
	changed := false

	for _, target := range targets {
		accountIndex := graphicsProfileAccountIndex(accounts, target)
		if accountIndex < 0 {
			return errors.New("target account was not found in current account list")
		}

		current := accounts[accountIndex].DefaultMod
		resolved := mods.ResolveSavedDefaultMod(current, installedMods)
		normalized := mods.NormalizeSavedDefaultMod(current)

		desired := resolved
		if normalized == "" || resolved == "" {
			desired = resolved
			if normalized != "" && resolved == "" {
				desired = ""
			}
		}

		if current == desired {
			continue
		}
		if _, exists := previous[accountIndex]; !exists {
			previous[accountIndex] = current
		}
		accounts[accountIndex].DefaultMod = desired
		changed = true
	}

	if !changed {
		return nil
	}
	if err := account.SaveAccounts(accountsFile, accounts); err != nil {
		for idx, previousMod := range previous {
			accounts[idx].DefaultMod = previousMod
		}
		return err
	}
	return nil
}

func missingDefaultModAccountLabels(accounts []*account.Account, installedMods []string) []string {
	labels := make([]string, 0, len(accounts))
	for _, acc := range accounts {
		if acc == nil {
			continue
		}
		if mods.ResolveSavedDefaultMod(acc.DefaultMod, installedMods) != "" {
			continue
		}
		labels = append(labels, launchTargetAccountLabel(acc))
	}
	return labels
}

func resolveLaunchModChoice(choice launchModChoice, acc account.Account, installedMods []string) ([]string, bool) {
	selectedMod := ""
	switch {
	case choice.HasManual:
		selectedMod = choice.ManualMod
	case choice.UseDefaults:
		selectedMod = mods.ResolveSavedDefaultMod(acc.DefaultMod, installedMods)
	default:
		return nil, false
	}

	switch selectedMod {
	case "":
		return nil, false
	case mods.DefaultModVanilla:
		return nil, true
	default:
		return mods.BuildLaunchArgs(selectedMod), true
	}
}

func launchTargetAccountLabel(acc *account.Account) string {
	if acc == nil {
		return ""
	}
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

func pauseAfterSuccessfulLaunch() {
	launchSuccessPauseSleep(launchSuccessPauseDuration)
}

func accountLaunchArgs(acc account.Account, modArgs []string) []string {
	args := make([]string, 0, len(modArgs)+4)
	args = append(args, modArgs...)
	args = append(args, account.LaunchArgs(acc.LaunchFlags)...)
	return args
}

func renameLaunchedWindow(pid uint32, displayName string) {
	ui.infof(lang.Launch.WindowRenaming, displayName)
	err := process.RenameWindow(pid, d2r.WindowTitle(displayName), 15, 2*time.Second)
	if err != nil {
		ui.warningf(lang.Launch.WindowRenameFailed, displayName, err)
		return
	}

	ui.successf(lang.Launch.WindowRenamed, d2r.WindowTitle(displayName))
}

func runningAccountWindowTitles() map[string]bool {
	titles := process.FindWindowTitlesByPrefix(d2r.WindowTitlePrefix)
	running := make(map[string]bool, len(titles))
	for _, title := range titles {
		running[title] = true
	}
	return running
}

func isAccountRunning(displayName string) bool {
	return process.FindWindowByTitle(d2r.WindowTitle(displayName))
}

func pendingBatchAccounts(accounts []account.Account, runningTitles map[string]bool) []*account.Account {
	pending := make([]*account.Account, 0, len(accounts))
	for i := range accounts {
		if runningTitles[d2r.WindowTitle(accounts[i].DisplayName)] {
			continue
		}
		pending = append(pending, &accounts[i])
	}
	return pending
}

func runningBatchAccounts(accounts []account.Account, runningTitles map[string]bool) []*account.Account {
	running := make([]*account.Account, 0, len(accounts))
	for i := range accounts {
		if !runningTitles[d2r.WindowTitle(accounts[i].DisplayName)] {
			continue
		}
		running = append(running, &accounts[i])
	}
	return running
}

func batchAccountStatusLines(accounts []account.Account, runningTitles map[string]bool) []string {
	lines := make([]string, 0, len(accounts))
	for i := range accounts {
		status := lang.Launch.StatusStopped
		if runningTitles[d2r.WindowTitle(accounts[i].DisplayName)] {
			status = lang.Launch.StatusRunning
		}
		lines = append(lines, fmt.Sprintf("  <%s> %s (%s)", status, accounts[i].DisplayName, accounts[i].Email))
	}
	return lines
}

func formatLaunchDelayMessage(delaySeconds int, nextDisplayName string) string {
	return fmt.Sprintf("  "+lang.Launch.BatchDelayMsg, delaySeconds, nextDisplayName)
}

func formatLaunchDelayRemainingMessage(remainingSeconds int, nextDisplayName string) string {
	return fmt.Sprintf("  "+lang.Launch.BatchDelayRemaining, remainingSeconds, nextDisplayName)
}

func waitForNextBatchLaunch(delaySeconds int, nextDisplayName string) {
	if delaySeconds <= 0 {
		return
	}

	ui.infof("%s", strings.TrimSpace(formatLaunchDelayMessage(delaySeconds, nextDisplayName)))
	remainingSeconds := delaySeconds
	for remainingSeconds > 0 {
		stepSeconds := 5
		if remainingSeconds < stepSeconds {
			stepSeconds = remainingSeconds
		}

		launchDelaySleep(time.Duration(stepSeconds) * time.Second)
		remainingSeconds -= stepSeconds
		if remainingSeconds > 0 {
			ui.infof("%s", strings.TrimSpace(formatLaunchDelayRemainingMessage(remainingSeconds, nextDisplayName)))
		}
	}
}

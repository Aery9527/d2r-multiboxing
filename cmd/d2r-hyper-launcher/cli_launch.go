package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
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

type launchGraphicsChoice struct {
	ManualProfile string
	HasManual     bool
	UseDefaults   bool
}

type launchGraphicsResolution struct {
	ProfileName string
	Apply       bool
}

func launchAccount(acc *account.Account, accounts []account.Account, accountsFile string, cfg *config.Config) {
	if !ensureLaunchReadyD2RPath(cfg) {
		return
	}
	if isAccountRunning(acc.DisplayName) {
		showWarningAndPause(fmt.Sprintf(lang.Launch.AlreadyRunning, acc.DisplayName))
		return
	}

	installedMods, ok := discoverInstalledMods(cfg.D2RPath)
	if !ok {
		return
	}

	availableGraphicsProfiles, ok := loadLaunchGraphicsProfiles()
	if !ok {
		return
	}

	regionChoice, ok := promptLaunchRegion(lang.Launch.RegionSingleTitle, []*account.Account{acc}, installedMods, availableGraphicsProfiles)
	if !ok {
		return
	}

	modChoice, ok := promptLaunchMod(lang.Launch.ModSingleTitle, accounts, accountsFile, []*account.Account{acc}, installedMods, availableGraphicsProfiles)
	if !ok {
		return
	}

	graphicsChoice, ok := promptLaunchGraphics(
		lang.Launch.GraphicsSingleTitle,
		accounts,
		accountsFile,
		[]*account.Account{acc},
		installedMods,
		availableGraphicsProfiles,
	)
	if !ok {
		return
	}

	graphicsResolution, graphicsOK := resolveLaunchGraphicsChoice(graphicsChoice, *acc, availableGraphicsProfiles)
	if !graphicsOK {
		showInputErrorAndPause(formatLaunchMissingAccountsMessage(lang.Launch.GraphicsMissing, []string{launchTargetAccountLabel(acc)}))
		return
	}

	password, err := account.GetDecryptedPassword(acc)
	if err != nil {
		ui.errorf(lang.Launch.DecryptFailed, err)
		return
	}

	region := resolveLaunchRegionChoice(regionChoice, *acc)
	if region == nil {
		showInputErrorAndPause(formatLaunchMissingAccountsMessage(lang.Launch.RegionMissing, []string{launchTargetAccountLabel(acc)}))
		return
	}

	modArgs, modOK := resolveLaunchModChoice(modChoice, *acc, installedMods)
	if !modOK {
		showInputErrorAndPause(formatLaunchMissingAccountsMessage(lang.Launch.ModMissing, []string{launchTargetAccountLabel(acc)}))
		return
	}

	_, err = applyResolvedLaunchGraphics(graphicsResolution, nil)
	if err != nil {
		showInputErrorAndPause(fmt.Sprintf(lang.GraphicsProfiles.ApplyFailed, graphicsResolution.ProfileName, err))
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

	installedMods, ok := discoverInstalledMods(cfg.D2RPath)
	if !ok {
		return
	}

	availableGraphicsProfiles, ok := loadLaunchGraphicsProfiles()
	if !ok {
		return
	}

	regionChoice, ok := promptLaunchRegion(lang.Launch.RegionBatchTitle, pendingAccounts, installedMods, availableGraphicsProfiles)
	if !ok {
		return
	}

	modChoice, ok := promptLaunchMod(lang.Launch.ModBatchTitle, accounts, accountsFile, pendingAccounts, installedMods, availableGraphicsProfiles)
	if !ok {
		return
	}

	graphicsChoice, ok := promptLaunchGraphics(
		lang.Launch.GraphicsBatchTitle,
		accounts,
		accountsFile,
		pendingAccounts,
		installedMods,
		availableGraphicsProfiles,
	)
	if !ok {
		return
	}

	var graphicsStore *graphicsprofile.Store
	for i, acc := range pendingAccounts {
		password, err := account.GetDecryptedPassword(acc)
		if err != nil {
			ui.warningf(lang.Launch.BatchDecryptFailed, acc.DisplayName, err)
			continue
		}

		region := resolveLaunchRegionChoice(regionChoice, *acc)
		if region == nil {
			ui.warningf("%s", formatLaunchMissingAccountsMessage(lang.Launch.RegionMissing, []string{launchTargetAccountLabel(acc)}))
			continue
		}

		modArgs, modOK := resolveLaunchModChoice(modChoice, *acc, installedMods)
		if !modOK {
			ui.warningf("%s", formatLaunchMissingAccountsMessage(lang.Launch.ModMissing, []string{launchTargetAccountLabel(acc)}))
			continue
		}

		graphicsResolution, graphicsOK := resolveLaunchGraphicsChoice(graphicsChoice, *acc, availableGraphicsProfiles)
		if !graphicsOK {
			ui.warningf("%s", formatLaunchMissingAccountsMessage(lang.Launch.GraphicsMissing, []string{launchTargetAccountLabel(acc)}))
			continue
		}

		var applyErr error
		graphicsStore, applyErr = applyResolvedLaunchGraphics(graphicsResolution, graphicsStore)
		if applyErr != nil {
			ui.warningf(lang.GraphicsProfiles.BatchApplyFailed, acc.DisplayName, graphicsResolution.ProfileName, applyErr)
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

func loadLaunchGraphicsProfiles() ([]string, bool) {
	profiles, err := listGraphicsProfiles()
	if err != nil {
		showInputErrorAndPause(fmt.Sprintf(lang.GraphicsProfiles.StoreOpenFailed, err))
		return nil, false
	}
	if profiles == nil {
		profiles = []string{}
	}
	return profiles, true
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

func promptLaunchRegion(title string, accounts []*account.Account, installedMods []string, availableGraphicsProfiles []string) (launchRegionChoice, bool) {
	var result launchRegionChoice
	_ = runMenu(func() {
		ui.headf("%s", title)
		ui.infof("%s", lang.Launch.RegionUseDefaults)
		ui.infof("%s", lang.Launch.RegionOverride)
		ui.infof("%s", lang.Launch.RegionTargetLabel)
		for _, line := range launchTargetAccountLines(accounts, installedMods, availableGraphicsProfiles) {
			ui.rawln(line)
		}
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
				showInputErrorAndPause(formatLaunchMissingAccountsMessage(lang.Launch.RegionMissing, missing))
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

func promptLaunchMod(title string, accounts []account.Account, accountsFile string, targets []*account.Account, installedMods []string, availableGraphicsProfiles []string) (launchModChoice, bool) {
	var result launchModChoice
	_ = runMenu(func() {
		ui.headf("%s", title)
		ui.infof("%s", lang.Launch.RegionTargetLabel)
		for _, line := range launchTargetAccountLines(targets, installedMods, availableGraphicsProfiles) {
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
				showInputErrorAndPause(formatLaunchMissingAccountsMessage(lang.Launch.ModMissing, missing))
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

func promptLaunchGraphics(title string, accounts []account.Account, accountsFile string, targets []*account.Account, installedMods []string, availableGraphicsProfiles []string) (launchGraphicsChoice, bool) {
	var result launchGraphicsChoice
	_ = runMenu(func() {
		ui.headf("%s", title)
		ui.infof("%s", lang.Launch.RegionTargetLabel)
		for _, line := range launchTargetAccountLines(targets, installedMods, availableGraphicsProfiles) {
			ui.rawln(line)
		}
		ui.infof("%s", lang.Launch.GraphicsUseDefaults)
		ui.infof("%s", lang.Launch.GraphicsOverride)
		if len(availableGraphicsProfiles) == 0 {
			ui.infof("%s", lang.Launch.GraphicsNoProfiles)
		}
		ui.menuBlock(func() {
			renderLaunchGraphicsOptions(availableGraphicsProfiles)
		})
	}, func(input string) error {
		if strings.TrimSpace(input) == "" {
			if err := reconcileDefaultGraphicsProfileAssignmentsForLaunch(accounts, accountsFile, targets, availableGraphicsProfiles); err != nil {
				showInputErrorAndPause(fmt.Sprintf(lang.Common.SaveFailed, err))
				return nil
			}
			missing := missingDefaultGraphicsProfileAccountLabels(targets, availableGraphicsProfiles)
			if len(missing) > 0 {
				showInputErrorAndPause(formatLaunchMissingAccountsMessage(lang.Launch.GraphicsMissing, missing))
				return nil
			}
			result.UseDefaults = true
			return errNavDone
		}

		selectedProfile, ok := parseLaunchGraphicsInput(input, availableGraphicsProfiles)
		if !ok {
			showInvalidInputAndPause()
			return nil
		}
		if strings.TrimSpace(selectedProfile) == "" {
			ui.infof("%s", lang.Launch.GraphicsNoneChosen)
		} else {
			ui.infof(lang.Launch.GraphicsUsing, selectedProfile)
		}
		result.ManualProfile = selectedProfile
		result.HasManual = true
		return errNavDone
	})
	return result, result.UseDefaults || result.HasManual
}

func launchTargetAccountLines(accounts []*account.Account, installedMods []string, availableGraphicsProfiles []string) []string {
	lines := make([]string, 0, len(accounts))
	for _, acc := range accounts {
		if acc == nil {
			continue
		}
		lines = append(lines, fmt.Sprintf(
			lang.Launch.TargetDefaultSummary,
			launchTargetAccountLabel(acc),
			defaultRegionStatusLabel(*acc),
			defaultModStatusLabel(*acc, installedMods),
			defaultGraphicsProfileStatusLabel(*acc, availableGraphicsProfiles),
		))
	}
	return lines
}

func formatLaunchMissingAccountsMessage(template string, labels []string) string {
	return fmt.Sprintf(template, formatLaunchMissingAccountLines(labels))
}

func formatLaunchMissingAccountLines(labels []string) string {
	lines := make([]string, 0, len(labels))
	for _, label := range labels {
		trimmed := strings.TrimSpace(label)
		if trimmed == "" {
			continue
		}
		lines = append(lines, "  - "+trimmed)
	}
	return strings.Join(lines, "\n")
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

func resolveSavedGraphicsProfile(profileName string, availableGraphicsProfiles []string) string {
	normalizedProfile := strings.TrimSpace(profileName)
	if normalizedProfile == "" {
		return ""
	}
	for _, availableProfile := range availableGraphicsProfiles {
		trimmedAvailableProfile := strings.TrimSpace(availableProfile)
		if !strings.EqualFold(trimmedAvailableProfile, normalizedProfile) {
			continue
		}
		return trimmedAvailableProfile
	}
	return ""
}

func defaultGraphicsProfileStatusLabel(acc account.Account, availableGraphicsProfiles []string) string {
	profileName := strings.TrimSpace(acc.GraphicsProfile)
	if profileName == "" {
		return lang.GraphicsProfiles.StatusUnassigned
	}
	resolvedProfile := resolveSavedGraphicsProfile(profileName, availableGraphicsProfiles)
	if resolvedProfile == "" {
		return fmt.Sprintf(lang.GraphicsProfiles.StatusMissing, profileName)
	}
	return resolvedProfile
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

func reconcileDefaultGraphicsProfileAssignmentsForLaunch(accounts []account.Account, accountsFile string, targets []*account.Account, availableGraphicsProfiles []string) error {
	previous := make(map[int]string, len(targets))
	cleared := make([]struct {
		label       string
		profileName string
	}, 0)
	changed := false

	for _, target := range targets {
		accountIndex := graphicsProfileAccountIndex(accounts, target)
		if accountIndex < 0 {
			return errors.New("target account was not found in current account list")
		}

		current := accounts[accountIndex].GraphicsProfile
		normalizedProfile := strings.TrimSpace(current)
		desiredProfile := ""
		if normalizedProfile != "" {
			desiredProfile = resolveSavedGraphicsProfile(normalizedProfile, availableGraphicsProfiles)
		}

		if current == desiredProfile {
			continue
		}
		if _, exists := previous[accountIndex]; !exists {
			previous[accountIndex] = current
		}
		accounts[accountIndex].GraphicsProfile = desiredProfile
		changed = true
		if normalizedProfile != "" && desiredProfile == "" {
			cleared = append(cleared, struct {
				label       string
				profileName string
			}{
				label:       graphicsProfileAccountLabel(accounts[accountIndex]),
				profileName: normalizedProfile,
			})
		}
	}
	if !changed {
		return nil
	}
	if err := account.SaveAccounts(accountsFile, accounts); err != nil {
		for idx, previousProfile := range previous {
			accounts[idx].GraphicsProfile = previousProfile
		}
		return err
	}
	for _, item := range cleared {
		ui.warningf(lang.GraphicsProfiles.MissingProfileCleared, item.label, item.profileName)
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

func missingDefaultGraphicsProfileAccountLabels(accounts []*account.Account, availableGraphicsProfiles []string) []string {
	labels := make([]string, 0, len(accounts))
	for _, acc := range accounts {
		if acc == nil {
			continue
		}
		if resolveSavedGraphicsProfile(acc.GraphicsProfile, availableGraphicsProfiles) != "" {
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

func resolveLaunchGraphicsChoice(choice launchGraphicsChoice, acc account.Account, availableGraphicsProfiles []string) (launchGraphicsResolution, bool) {
	switch {
	case choice.HasManual:
		selectedProfile := strings.TrimSpace(choice.ManualProfile)
		return launchGraphicsResolution{ProfileName: selectedProfile, Apply: selectedProfile != ""}, true
	case choice.UseDefaults:
		selectedProfile := resolveSavedGraphicsProfile(acc.GraphicsProfile, availableGraphicsProfiles)
		if selectedProfile == "" {
			return launchGraphicsResolution{}, false
		}
		return launchGraphicsResolution{ProfileName: selectedProfile, Apply: true}, true
	default:
		return launchGraphicsResolution{}, false
	}
}

func applyResolvedLaunchGraphics(resolution launchGraphicsResolution, store *graphicsprofile.Store) (*graphicsprofile.Store, error) {
	if !resolution.Apply {
		return store, nil
	}
	return applyNamedGraphicsProfileForLaunch(resolution.ProfileName, store)
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

func renderLaunchGraphicsOptions(availableGraphicsProfiles []string) {
	options := ui.subMenuOptions(func(options *cliMenuOptions) {
		options.option("0", lang.Launch.GraphicsOptNone, "")
		for i, profileName := range availableGraphicsProfiles {
			options.option(strconv.Itoa(i+1), profileName, "")
		}
	})
	options.render()
}

func parseLaunchGraphicsInput(input string, availableGraphicsProfiles []string) (string, bool) {
	selected, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		return "", false
	}
	if selected == 0 {
		return "", true
	}
	if selected < 1 || selected > len(availableGraphicsProfiles) {
		return "", false
	}
	return availableGraphicsProfiles[selected-1], true
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

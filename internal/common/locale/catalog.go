// Package locale provides multi-language catalog support.
// Catalogs are nested structs of format strings, keyed by Locale constant.
// All user-visible strings in the CLI layer are sourced from the active Catalog.
package locale

import "strings"

// Locale identifies a supported display language.
type Locale string

const (
	LocaleZhTW Locale = "zh-TW"
	LocaleEn   Locale = "en"
)

// KnownLocales returns all supported locales in display order.
func KnownLocales() []Locale {
	return []Locale{LocaleZhTW, LocaleEn}
}

// ParseLocale validates and normalises a locale string.
// Returns (locale, true) on match; (LocaleZhTW, false) on unknown input.
func ParseLocale(s string) (Locale, bool) {
	for _, l := range KnownLocales() {
		if strings.EqualFold(s, string(l)) {
			return l, true
		}
	}
	return LocaleZhTW, false
}

// Get returns the Catalog for the given locale.
// Falls back to zh-TW for any unrecognised locale.
func Get(l Locale) *Catalog {
	switch l {
	case LocaleEn:
		c := catalogEn
		return &c
	default:
		c := catalogZhTW
		return &c
	}
}

// ---------------------------------------------------------------------------
// Top-level Catalog
// ---------------------------------------------------------------------------

// Catalog is the root container for all user-visible strings.
// Sub-catalogs group strings by functional area.
type Catalog struct {
	Locale           Locale
	Common           CommonCatalog
	Startup          StartupCatalog
	MainMenu         MainMenuCatalog
	Launch           LaunchCatalog
	Delay            DelayCatalog
	Switcher         SwitcherCatalog
	Flags            FlagsCatalog
	GraphicsProfiles GraphicsProfilesCatalog
	DefaultMods      DefaultModsCatalog
	RegionDefaults   RegionDefaultsCatalog
	D2RPath          D2RPathCatalog
	Accounts         AccountsCatalog
	Language         LanguageCatalog
}

// ---------------------------------------------------------------------------
// Sub-catalogs
// ---------------------------------------------------------------------------

// CommonCatalog holds navigation labels and generic prompts shared across all flows.
type CommonCatalog struct {
	SelectPrompt   string // default readline prompt: "請選擇："
	AnyKeyPrompt   string // single-key continue
	EnterKeyPrompt string // Enter-only continue
	NavBack        string // [b] submenu nav label
	NavHome        string // [h] submenu nav label
	NavQuit        string // [q] submenu nav label
	QuitLabel      string // [q] main menu quit label
	Cancelled      string // user aborted an action
	InvalidInput   string // unrecognised menu input
	WaitKeyFailed  string // fmt("%v") – wait-for-key error
	Goodbye        string // farewell on exit
	SaveFailed     string // fmt("%v") – config save error
	ParseFailed    string // fmt("%v") – input parse error (used in flags)
}

// StartupCatalog holds strings shown during application startup.
type StartupCatalog struct {
	// Version / directory line in announcement
	VersionLabel string // fmt("%s") – "目前版本：%s"
	DataDirLabel string // fmt("%s") – "資料目錄：%s"

	// displayReleaseTime helpers
	UnreleasedTag string // shown when releaseTime is empty
	ReleaseSuffix string // fmt("%s") – wraps the releaseTime value

	// Startup error wrappers (shown before any menu is drawn)
	ConfigLoadFailed    string // fmt("%v")
	AccountsPathFailed  string // fmt("%v")
	AccountsFileFailed  string // fmt("%v")
	AccountsLoadFailed  string // fmt("%v")
	EncryptFailed       string // fmt("%v")
	SwitcherStartFail   string // fmt("%v")
	AccountsLoadRefresh string // fmt("%v") – on 'r' refresh error

	// Success notice
	PasswordEncrypted string

	// Announcement warning lines (each is a standalone line)
	WarnStatusDetect1 string
	WarnStatusDetect2 string
	WarnStatusDetect3 string
	WarnNote          string
	WarnNoteWindowed  string
	WarnNoteGamepad   string
	WarnNoteSwitcher  string
	WarnNoteDelay     string
	WarnNoteConfig    string
	WarnNoteBattleNet string
	WarnNoteNoModify  string
	WarnNoteCommunity string
}

// MainMenuCatalog holds the main menu title, option labels, and account-list header.
type MainMenuCatalog struct {
	Title             string
	AccountListHeader string // "帳號列表：" (also used in flags sub-menu)

	// Menu option keys and labels
	OptByNumberKey             string // display key "數字" / "1-N"
	OptByNumber                string // label "啟動指定帳號"
	OptOffline                 string
	OptOfflineComment          string
	OptLaunchAll               string
	OptLaunchAllComment        string
	OptDelay                   string // comment is dynamic (displayDelay)
	OptFlags                   string
	OptFlagsComment            string
	OptGraphicsProfiles        string
	OptGraphicsProfilesComment string
	OptDefaultMods             string
	OptDefaultModsComment      string
	OptDefaultRegions          string
	OptDefaultRegionsComment   string
	OptD2RPath                 string // comment is dynamic (cfg.D2RPath)
	OptSwitcher                string // comment is dynamic (switcherMenuOptionStatus)
	OptRefresh                 string
	OptLanguage                string // new – language settings entry

	// Dynamic switcher status strings used as the [s] menu comment
	SwitcherNotSet   string
	SwitcherDisabled string // fmt("%s") – "未啟用設定：%s"
	SwitcherEnabled  string // fmt("%s") – "已啟用設定：%s"
}

// LaunchCatalog holds strings for account launch flows.
type LaunchCatalog struct {
	// Single-account launch
	AlreadyRunning     string // fmt("%s") – account display name
	Starting           string // fmt("%s", "%s") – display name, region name
	LaunchOK           string // fmt("%d") – PID
	LaunchFailed       string // fmt("%v")
	DecryptFailed      string // fmt("%v")
	CloseHandleFailed  string // fmt("%v")
	HandlesClosed      string // fmt("%d") – event handle count
	WindowRenaming     string // fmt("%s") – display name
	WindowRenamed      string // fmt("%s") – window title (quoted)
	WindowRenameFailed string // fmt("%s", "%v") – display name, error

	// Batch launch
	BatchScanHeader        string
	BatchOnlyPending       string // fmt("%d") – pending count
	AllRunning             string
	BatchDecryptFailed     string // fmt("%s", "%v") – display name, error
	BatchLaunchFailed      string // fmt("%s", "%v") – display name, error
	BatchLaunchOK          string // fmt("%s", "%d") – display name, PID
	BatchHandleCloseFailed string // fmt("%s", "%v") – display name, error
	BatchHandlesClosed     string // fmt("%s", "%d") – display name, count
	BatchDelayMsg          string // fmt("%d", "%s") – seconds, next display name
	BatchDelayRemaining    string // fmt("%d", "%s") – remaining seconds, next display name

	// Offline launch
	OfflineTitle        string
	OfflineLaunching    string
	OfflineLaunchOK     string // fmt("%d") – PID
	OfflineLaunchFailed string // fmt("%v")

	// Region selector
	RegionSingleTitle string // heading for single-account region pick
	RegionBatchTitle  string // heading for batch-launch region pick
	RegionTargetLabel string // "準備啟動的帳號："
	RegionInvalid     string
	RegionUseDefaults string
	RegionOverride    string
	RegionMissing     string // fmt("%s")

	// Mod selector
	ModSingleTitle  string
	ModBatchTitle   string
	ModOfflineTitle string
	ModLoadFailed   string // fmt("%v") – mods discovery error
	ModNoMods       string
	ModOptNone      string // option label "不使用 mod"
	ModUsing        string // fmt("%s") – selected mod name
	ModNoneChosen   string // confirmed no-mod info line
	ModUseDefaults  string
	ModOverride     string
	ModMissing      string // fmt("%s")

	// Account status labels
	StatusRunning string // "已啟動"
	StatusStopped string // "未啟動"
}

// DelayCatalog holds strings for launch-delay settings.
type DelayCatalog struct {
	Title          string
	CurrentSetting string // fmt("%s")
	Description    string
	MinLabel       string // fmt("%d") – minimum seconds info
	HintFixed      string
	HintRange      string
	InputPrompt    string
	Updated        string // fmt("%s")

	// CLI-layer display format strings (replace domain DisplayString in UI context).
	// These are used by the displayDelay() helper in the CLI layer.
	DisplayFixed  string // fmt("%d") → "10 秒" / "10 seconds"
	DisplayRandom string // fmt("%d", "%d") → "10-30 秒（隨機）" / "10-30 seconds (random)"
}

// SwitcherCatalog holds strings for window-switcher configuration.
type SwitcherCatalog struct {
	Title string

	// Status display
	StatusEnabled  string
	StatusDisabled string
	StatusNotSet   string
	StatusLabel    string // fmt("%s") – "目前狀態：%s"
	SettingLabel   string // fmt("%s") – "目前設定：%s"
	SavedLabel     string // fmt("%s") – "已保存設定：%s"

	// Menu options
	OptSetKey  string
	OptEnable  string
	OptDisable string

	// Key detection flow
	DetectInstruction   string
	DetectSupport       string
	DetectGamepad       string
	DetectEscCancel     string
	DetectFailed        string // fmt("%v")
	DetectCancelled     string
	DetectedKey         string // fmt("%s") – display of detected combo
	DetectConfirmPrompt string

	// Outcomes
	KeySet        string // fmt("%s") – saved key display
	StartFailed   string // fmt("%v") – switcher start error
	RestartFailed string // fmt("%v")
	ToggleNotSet  string // no key configured yet

	// Enable/disable outcomes
	Enabled                     string // fmt("%s") – key display
	Disabled                    string
	DisableSaveAndRestoreFailed string // fmt("%v", "%v") – save err, restart err

	// Account filter
	OptSetAccounts                string // [2] menu option label
	AccountFilterTitle            string // sub-menu header
	AccountFilterDescIncluded     string // description line: what "included" means
	AccountFilterDescExcluded     string // description line: what "excluded" means
	AccountIncluded               string // per-account status label – included
	AccountExcluded               string // per-account status label – excluded
	AccountFilterSaved            string // success after save
	AccountFilterNoAccounts       string // warning when accounts list is empty
	AccountFilterOptToggle        string // [1~N] toggle single account
	AccountFilterOptAll           string // [a] include-all option label
	AccountFilterOptNone          string // [n] exclude-all option label
	AccountFilterWarnOneIncluded  string // warning: only 1 account in cycle
	AccountFilterWarnNoneIncluded string // warning: all accounts excluded
}

// FlagsCatalog holds strings for account launch-flag configuration.
type FlagsCatalog struct {
	Title      string
	NoAccounts string

	// Main sub-menu options
	OptSetFlag   string
	OptClearFlag string

	// Action verbs injected into format strings below
	ActionSet   string // "設定" / "Set"
	ActionClear string // "取消" / "Clear"

	// Mode selection
	ModeTitle         string // fmt("%s") – action verb
	ModeQuestion      string // fmt("%s")
	OptFlagToAccounts string // fmt("%s") – label
	OptAccountToFlags string // fmt("%s") – label
	OptAllFlagsAll    string // fmt("%s") – label

	// "By flag" flow headings and prompts
	FlagByFlagTitle         string // fmt("%s")
	FlagByFlagSelectPrompt  string
	FlagByFlagAccountTitle  string // fmt("%s")
	FlagByFlagAccountPrompt string // fmt("%s", "%s") – action, flag name

	// "By account" flow headings and prompts
	FlagByAccountTitle        string // fmt("%s")
	FlagByAccountSelectPrompt string
	FlagByAccountFlagTitle    string // fmt("%s")
	FlagByAccountFlagPrompt   string // fmt("%s", "%s") – account name, action

	// "All accounts all flags" flow
	FlagAllTitle string // fmt("%s") – heading
	FlagAllAbout string // fmt("%s") – action verb
	FlagAllCount string // fmt("%d")

	// Scope summaries before confirmation
	FlagByFlagAbout    string // fmt("%s", "%s") – action, flag name
	FlagByAccountAbout string // fmt("%s", "%s") – account name, action

	// Shared input prompt in flag/account selection loops
	FlagInputPrompt string

	// Confirmation in confirmChanges()
	ConfirmPrompt string

	// Error messages
	InvalidFlagID    string
	InvalidAccountID string
	SaveFailed       string // fmt("%v")

	// Success
	Done string // fmt("%s") – action verb

	// Flag table
	FlagTableHeader string

	// Per-option comment rendering helpers
	FlagDescPrefix   string // fmt("%s") – "說明：%s" / "Info: %s"
	FlagExperimental string // appended when option.Experimental is true

	// Account item format in flag selection menus
	// Args: index (int), display name, email, flag summary
	FlagAccountItemFmt string // fmt("%d", "%s", "%s", "%s")

	// Account item comment in flag selection menus: fmt("%s") – flag summary
	FlagComment string // "flag：%s" / "flags: %s"

	// Header of the first column in the flag table
	FlagTableAccountHeader string // "帳號編號" / "#"
}

// GraphicsProfilesCatalog holds strings for per-account graphics profile flows.
type GraphicsProfilesCatalog struct {
	Title             string
	NoAccounts        string
	ProfileListHeader string
	NoProfiles        string
	StatusUnassigned  string
	Intro1            string
	Intro2            string
	Intro3            string
	Intro4            string

	// Main sub-menu options
	OptSaveCurrent string
	OptAssign      string
	OptClear       string
	OptDeleteSaved string

	// Save-current flow
	SaveTitle             string
	SaveIntro1            string
	SaveIntro2            string
	CurrentSettingsLabel  string // fmt("%s")
	SaveOptionComment     string
	SavePrompt            string
	SaveInvalidProfileID  string // fmt("%d")
	SaveExistingUseNumber string // fmt("%s")
	SaveInvalidName       string // fmt("%v")
	SaveDone              string // fmt("%s")
	SaveFailed            string // fmt("%v")
	StoreOpenFailed       string // fmt("%v")

	// Assignment flow
	AssignTitle                  string
	AssignNoProfiles             string
	AssignModeTitle              string
	AssignModeQuestion           string
	OptProfileToAccounts         string
	OptAccountToProfile          string
	AssignByProfileTitle         string
	AssignByProfileSelectPrompt  string
	AssignByProfileAccountTitle  string // fmt("%s")
	AssignByProfileAccountPrompt string // fmt("%s")
	AssignByProfileAbout         string // fmt("%s")
	AssignByAccountTitle         string
	AssignByAccountSelectPrompt  string
	AssignByAccountProfileTitle  string // fmt("%s")
	AssignByAccountProfilePrompt string // fmt("%s")
	AssignByAccountAbout         string // fmt("%s", "%s")
	AssignDone                   string // fmt("%s")

	// Clear flow
	ClearTitle         string
	ClearNoAssignments string
	ClearPrompt        string
	ClearAbout         string
	ClearDone          string

	// Delete saved-profile flow
	DeleteTitle         string
	DeleteNoProfiles    string
	DeletePrompt        string
	DeleteAbout         string
	DeleteDone          string
	DeleteFailed        string // fmt("%v")
	DeleteUnusedComment string
	DeleteUsedComment   string // fmt("%d")
	DeleteInUse         string // fmt("%s", "%s")

	// Shared account rendering
	AccountComment string // fmt("%s")
	AccountItemFmt string // fmt("%d", "%s", "%s", "%s")

	// Launch-time apply
	ApplyingDuringLaunch      string // fmt("%s")
	ApplyFailed               string // fmt("%s", "%v")
	BatchApplyFailed          string // fmt("%s", "%s", "%v")
	MissingProfileCleared     string // fmt("%s", "%s")
	MissingProfileClearFailed string // fmt("%v")
}

// DefaultModsCatalog holds strings for per-account default mod flows.
type DefaultModsCatalog struct {
	Title            string
	NoAccounts       string
	StatusUnassigned string
	StatusVanilla    string
	StatusMissing    string // fmt("%s")
	Intro1           string
	Intro2           string
	Intro3           string
	ModListHeader    string
	NoInstalledMods  string

	// Main sub-menu options
	OptAssign string
	OptClear  string

	// Assignment flow
	AssignModeTitle             string
	AssignModeQuestion          string
	OptModToAccounts            string
	OptAccountToMod             string
	AssignByModTitle            string
	AssignByModSelectPrompt     string
	AssignByModAccountTitle     string // fmt("%s")
	AssignByModAccountPrompt    string // fmt("%s")
	AssignByModAbout            string // fmt("%s")
	AssignByAccountTitle        string
	AssignByAccountSelectPrompt string
	AssignByAccountModTitle     string // fmt("%s")
	AssignByAccountModPrompt    string // fmt("%s")
	AssignByAccountAbout        string // fmt("%s", "%s")
	AssignDone                  string // fmt("%s")

	// Clear flow
	ClearTitle         string
	ClearNoAssignments string
	ClearPrompt        string
	ClearAbout         string
	ClearDone          string

	// Shared account rendering
	AccountComment string // fmt("%s")
	AccountItemFmt string // fmt("%d", "%s", "%s", "%s")
}

// RegionDefaultsCatalog holds strings for per-account default region flows.
type RegionDefaultsCatalog struct {
	Title            string
	NoAccounts       string
	StatusUnassigned string
	Intro1           string
	Intro2           string
	Intro3           string

	// Main sub-menu options
	OptAssign string
	OptClear  string

	// Assignment flow
	AssignModeTitle             string
	AssignModeQuestion          string
	OptRegionToAccounts         string
	OptAccountToRegion          string
	AssignByRegionTitle         string
	AssignByRegionSelectPrompt  string
	AssignByRegionAccountTitle  string // fmt("%s")
	AssignByRegionAccountPrompt string // fmt("%s")
	AssignByRegionAbout         string // fmt("%s")
	AssignByAccountTitle        string
	AssignByAccountSelectPrompt string
	AssignByAccountRegionTitle  string // fmt("%s")
	AssignByAccountRegionPrompt string // fmt("%s")
	AssignByAccountAbout        string // fmt("%s", "%s")
	AssignDone                  string // fmt("%s")

	// Clear flow
	ClearTitle         string
	ClearNoAssignments string
	ClearPrompt        string
	ClearAbout         string
	ClearDone          string

	// Shared account rendering
	AccountComment string // fmt("%s")
	AccountItemFmt string // fmt("%d", "%s", "%s", "%s")
}

// D2RPathCatalog holds strings for D2R.exe path configuration.
type D2RPathCatalog struct {
	// Pre-launch path check (ensureLaunchReadyD2RPath)
	PreCheckTitle        string
	PathNotFound         string // fmt("%s") – current path
	PathError            string // fmt("%v")
	PromptFix            string
	OptSetPath           string
	PreCheckInvalidInput string

	// Path setup (setupD2RPath)
	SetTitle  string
	SetPrompt string
	SetFailed string // fmt("%v")
	SetOK     string // fmt("%s")

	// PowerShell file picker dialog title
	PickerDialogTitle string
}

// AccountsCatalog holds strings for the first-run accounts-file creation flow.
type AccountsCatalog struct {
	CreatedOK        string // fmt("%s") – file path
	CreatedInfo1     string
	CreatedInfo2     string
	CreatedInfo3     string
	CreatedInfo4     string
	CreatedPressAny  string
	OpenFolderFailed string // fmt("%v")
}

// LanguageCatalog holds strings for the language-selection flow.
type LanguageCatalog struct {
	MenuLabel    string // main menu option label
	PickerTitle  string
	AutoDetected string // fmt("%s") – detected locale display name
	ConfirmUse   string
	PromptSelect string
	Saved        string // fmt("%s") – new locale display name

	NameZhTW string
	NameEn   string
}

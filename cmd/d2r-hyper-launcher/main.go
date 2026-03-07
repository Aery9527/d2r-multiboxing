package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"d2rhl/internal/account"
	"d2rhl/internal/config"
	"d2rhl/internal/d2r"
	"d2rhl/internal/handle"
	"d2rhl/internal/mods"
	"d2rhl/internal/process"
	"d2rhl/internal/switcher"

	"golang.org/x/sys/windows"
)

// version is set at build time via -ldflags "-X main.version=x.y.z".
var version = "dev"

func displayVersion(version string) string {
	if strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

// 子選單統一導航指令（所有子選單必須支援這三個選項）
const (
	menuBack = "b" // 回上一層
	menuHome = "h" // 回主選單
	menuQuit = "q" // 離開程式
)

func main() {
	// 設定 Windows console 輸出為 UTF-8，避免中文亂碼
	_ = windows.SetConsoleCP(65001)
	_ = windows.SetConsoleOutputCP(65001)

	fmt.Println("============================================")
	fmt.Printf("  d2r-hyper-launcher  %s\n", displayVersion(version))
	fmt.Println("============================================")
	fmt.Println()

	// 載入設定檔
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("  設定檔載入失敗：%v\n", err)
		return
	}
	cfgDir, _ := config.Dir()
	fmt.Printf("  資料目錄：%s\n", cfgDir)
	fmt.Printf("  D2R 路徑：%s\n", cfg.D2RPath)
	fmt.Printf("  啟動間隔：%d 秒\n", cfg.LaunchDelay)

	// 啟動視窗切換功能
	if cfg.Switcher != nil && cfg.Switcher.Enabled {
		if err := switcher.Start(cfg.Switcher); err != nil {
			fmt.Printf("  ⚠ 視窗切換啟動失敗：%v\n", err)
		} else {
			fmt.Printf("  ✔ 視窗切換已啟用：%s\n", switcher.FormatSwitcherDisplay(cfg.Switcher.Modifiers, cfg.Switcher.Key, cfg.Switcher.GamepadIndex))
		}
	}
	fmt.Println()

	accountsFile, err := config.AccountsPath()
	if err != nil {
		fmt.Printf("  無法取得帳號檔案路徑：%v\n", err)
		return
	}

	createdAccountsFile, err := account.EnsureAccountsFile(accountsFile)
	if err != nil {
		fmt.Printf("  建立帳號檔案失敗：%v\n", err)
		return
	}
	if createdAccountsFile {
		handleCreatedAccountsFile(cfgDir, accountsFile)
		return
	}

	// 讀取帳號
	accounts, err := account.LoadAccounts(accountsFile)
	if err != nil {
		fmt.Printf("  讀取帳號失敗：%v\n", err)
		return
	}

	// 首次執行時加密明文密碼
	changed, err := account.EncryptPlaintextPasswords(accountsFile, accounts)
	if err != nil {
		fmt.Printf("  密碼加密失敗：%v\n", err)
		return
	}
	if changed {
		fmt.Println("  ✔ 已加密明文密碼並回寫至 CSV")
	}

	// 啟動背景 Handle 監控
	go handleMonitor()

	// CLI 主迴圈
	scanner := bufio.NewScanner(os.Stdin)
	for {
		printMenu(accounts)
		fmt.Print("  > 請選擇：")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())

		switch strings.ToLower(input) {
		case "q":
			switcher.Stop()
			fmt.Println("  再見！")
			return
		case "r":
			// 重新整理，重新讀取帳號
			accounts, err = account.LoadAccounts(accountsFile)
			if err != nil {
				fmt.Printf("  讀取帳號失敗：%v\n", err)
			}
			continue
		case "0":
			launchOffline(cfg, scanner)
			continue
		case "a":
			launchAll(accounts, cfg, scanner)
			continue
		case "p":
			setupD2RPath(cfg)
			continue
		case "s":
			setupSwitcher(cfg, scanner)
			continue
		case "f":
			setupAccountLaunchFlags(accounts, accountsFile, scanner)
			continue
		default:
			id, err := strconv.Atoi(input)
			if err != nil || id < 1 || id > len(accounts) {
				fmt.Println("  無效輸入，請重試。")
				continue
			}
			acc := &accounts[id-1]
			launchAccount(acc, cfg, scanner)
		}
	}
}
func printMenu(accounts []account.Account) {
	fmt.Println("  帳號列表：")
	for i, acc := range accounts {
		status := "未啟動"
		if process.FindWindowByTitle(d2r.WindowTitle(acc.DisplayName)) {
			status = "已啟動"
		}
		fmt.Printf("  [%d] %-15s (%s)  [%s]\n",
			i+1, acc.DisplayName, acc.Email, status)
	}

	fmt.Println()
	fmt.Println("  *是否已啟動的判斷基準是用 account.csv 裡的 DisplayName 來對應視窗名稱。")
	fmt.Println("   如果 D2R 還開著就先關掉工具再去改 DisplayName，之後這裡的啟動狀態偵測可能會不正確。")
	fmt.Println()
	fmt.Println("--------------------------------------------")
	fmt.Println("  <數字>  啟動指定帳號")
	fmt.Println("  0       離線遊玩（可選 mod，不需帳密）")
	fmt.Println("  a       啟動所有帳號（可選 mod，只啟動未啟動的）")
	fmt.Println("  f       設定帳號啟動 flag")
	fmt.Println("  p       選擇 D2R.exe 路徑")
	fmt.Println("  s       視窗切換設定")
	fmt.Println("  r       重新整理狀態")
	fmt.Println("  q       退出")
	fmt.Println("--------------------------------------------")
}

func launchAccount(acc *account.Account, cfg *config.Config, scanner *bufio.Scanner) {
	if !ensureLaunchReadyD2RPath(cfg, scanner) {
		return
	}
	if isAccountRunning(acc.DisplayName) {
		fmt.Printf("  ⏭ %s 已在執行中，請先切回既有視窗或改用其他帳號。\n", acc.DisplayName)
		fmt.Println()
		return
	}

	// 選擇區域
	fmt.Println()
	fmt.Println("  選擇區域 (1=NA, 2=EU, 3=Asia)")
	printSubMenuNav()
	fmt.Print("  > 請選擇：")
	if !scanner.Scan() {
		return
	}
	input := strings.TrimSpace(scanner.Text())
	if nav := isMenuNav(input); nav != "" {
		return
	}
	region := parseRegionInput(input)
	if region == nil {
		fmt.Println("  無效的區域選擇。")
		return
	}

	modArgs, ok := selectLaunchMod(cfg.D2RPath, scanner)
	if !ok {
		return
	}
	launchArgs := accountLaunchArgs(*acc, modArgs)

	// 解密密碼
	password, err := account.GetDecryptedPassword(acc)
	if err != nil {
		fmt.Printf("  密碼解密失敗：%v\n", err)
		return
	}

	fmt.Printf("  正在啟動 %s (%s)...\n", acc.DisplayName, region.Name)

	// 啟動 D2R
	pid, err := process.LaunchD2R(cfg.D2RPath, acc.Email, password, region.Address, launchArgs...)
	if err != nil {
		fmt.Printf("  啟動失敗：%v\n", err)
		return
	}
	fmt.Printf("  ✔ D2R 已啟動 (PID: %d)\n", pid)

	// 等待進程初始化後關閉 Handle
	time.Sleep(2 * time.Second)
	closed, err := handle.CloseHandlesByName(pid, d2r.SingleInstanceEventName)
	if err != nil {
		fmt.Printf("  ⚠ 關閉 Handle 失敗：%v\n", err)
	} else if closed > 0 {
		fmt.Printf("  ✔ 已關閉 %d 個 Event Handle\n", closed)
	}

	renameLaunchedWindow(pid, acc.DisplayName)

	fmt.Println()
}

func launchAll(accounts []account.Account, cfg *config.Config, scanner *bufio.Scanner) {
	if !ensureLaunchReadyD2RPath(cfg, scanner) {
		return
	}

	runningTitles := runningAccountWindowTitles()
	pendingAccounts := pendingBatchAccounts(accounts, runningTitles)
	fmt.Println("  已預先掃描目前執行中的 D2R 視窗：")
	for _, line := range batchAccountStatusLines(accounts, runningTitles) {
		fmt.Println(line)
	}
	if len(pendingAccounts) == 0 {
		fmt.Println("  所有帳號都已在執行中。")
		fmt.Println()
		return
	}
	fmt.Printf("  本次只會啟動上面標示為 [未啟動] 的帳號，共 %d 個。\n", len(pendingAccounts))

	fmt.Println()
	fmt.Println("  選擇區域 (1=NA, 2=EU, 3=Asia)")
	printSubMenuNav()
	fmt.Print("  > 請選擇：")
	if !scanner.Scan() {
		return
	}
	input := strings.TrimSpace(scanner.Text())
	if nav := isMenuNav(input); nav != "" {
		return
	}
	region := parseRegionInput(input)
	if region == nil {
		fmt.Println("  無效的區域選擇。")
		return
	}

	modArgs, ok := selectLaunchMod(cfg.D2RPath, scanner)
	if !ok {
		return
	}

	for i, acc := range pendingAccounts {
		launchArgs := accountLaunchArgs(*acc, modArgs)

		password, err := account.GetDecryptedPassword(acc)
		if err != nil {
			fmt.Printf("  ⚠ 帳號 %s 密碼解密失敗：%v\n", acc.DisplayName, err)
			continue
		}

		fmt.Printf("  正在啟動 %s (%s)...\n", acc.DisplayName, region.Name)
		pid, err := process.LaunchD2R(cfg.D2RPath, acc.Email, password, region.Address, launchArgs...)
		if err != nil {
			fmt.Printf("  ⚠ 帳號 %s 啟動失敗：%v\n", acc.DisplayName, err)
			continue
		}
		fmt.Printf("  ✔ %s 已啟動 (PID: %d)\n", acc.DisplayName, pid)

		// 等待並關閉 Handle
		time.Sleep(3 * time.Second)
		closed, err := handle.CloseHandlesByName(pid, d2r.SingleInstanceEventName)
		if err != nil {
			fmt.Printf("  ⚠ %s Handle 關閉失敗：%v\n", acc.DisplayName, err)
		} else if closed > 0 {
			fmt.Printf("  ✔ %s 已關閉 %d 個 Handle\n", acc.DisplayName, closed)
		}

		renameLaunchedWindow(pid, acc.DisplayName)

		if cfg.LaunchDelay > 0 && i+1 < len(pendingAccounts) {
			fmt.Println(formatLaunchDelayMessage(cfg.LaunchDelay, pendingAccounts[i+1].DisplayName))
			time.Sleep(time.Duration(cfg.LaunchDelay) * time.Second)
		}
	}
	fmt.Println()
}

// handleMonitor 背景持續監控 D2R 進程並自動關閉 Handle
func handleMonitor() {
	processedPIDs := make(map[uint32]bool)

	for {
		time.Sleep(2 * time.Second)

		d2rProcesses, err := process.FindProcessesByName(d2r.ProcessName)
		if err != nil {
			continue
		}

		// 清理已結束的 PID
		activePIDs := make(map[uint32]bool)
		for _, p := range d2rProcesses {
			activePIDs[p.PID] = true
		}
		for pid := range processedPIDs {
			if !activePIDs[pid] {
				delete(processedPIDs, pid)
			}
		}

		for _, p := range d2rProcesses {
			if processedPIDs[p.PID] {
				continue
			}
			processedPIDs[p.PID] = true

			_, _ = handle.CloseHandlesByName(p.PID, d2r.SingleInstanceEventName)
		}
	}
}

func parseRegionInput(input string) *d2r.Region {
	input = strings.TrimSpace(strings.ToUpper(input))
	switch input {
	case "1", "NA":
		return d2r.FindRegion("NA")
	case "2", "EU":
		return d2r.FindRegion("EU")
	case "3", "ASIA":
		return d2r.FindRegion("Asia")
	default:
		return d2r.FindRegion(input)
	}
}

func handleCreatedAccountsFile(cfgDir, accountsFile string) {
	fmt.Println("  ✔ 已自動建立帳號設定檔 accounts.csv。")
	fmt.Printf("  建立位置：%s\n", accountsFile)
	fmt.Println("  工具已先幫你放入兩筆範例資料，請把它們改成你自己的 Battle.net 帳號。")
	fmt.Println("  CSV 格式：Email,Password,DisplayName,LaunchFlags")
	fmt.Println("  範例：your-account1@example.com,your-password-here,主帳號-法師(倉庫/武器/飾品),")
	fmt.Println("  LaunchFlags 可先留空；之後可回到工具主選單用 f 再設定各帳號的啟動旗標。")
	fmt.Println()
	fmt.Println("  按任意鍵後，程式會結束並自動開啟資料目錄，方便你直接修改剛建立好的 accounts.csv。")

	if err := waitForAnyKey(); err != nil {
		fmt.Printf("  ⚠ 等待按鍵失敗：%v\n", err)
		return
	}

	if err := openFolder(cfgDir); err != nil {
		fmt.Printf("  ⚠ 無法自動開啟資料目錄：%v\n", err)
	}
}

func waitForAnyKey() error {
	fmt.Print("  > 請按任意鍵繼續...")

	cmd := exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-Command",
		"$Host.UI.RawUI.ReadKey('NoEcho,IncludeKeyDown') | Out-Null",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	fmt.Println()
	return err
}

func openFolder(path string) error {
	cmd := exec.Command("explorer.exe", path)
	return cmd.Start()
}

// printSubMenuNav prints the standard sub-menu navigation options.
func printSubMenuNav() {
	fmt.Printf("  %s       回上一層\n", menuBack)
	fmt.Printf("  %s       回主選單\n", menuHome)
	fmt.Printf("  %s       離開程式\n", menuQuit)
}

// isMenuNav returns "back" if the input is menuBack, "home" if menuHome, or "" otherwise.
// If the input is menuQuit, the program exits immediately.
func isMenuNav(input string) string {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case menuBack:
		return "back"
	case menuHome:
		return "home"
	case menuQuit:
		fmt.Println("  再見！")
		os.Exit(0)
		return "" // unreachable
	default:
		return ""
	}
}

func launchOffline(cfg *config.Config, scanner *bufio.Scanner) {
	if !ensureLaunchReadyD2RPath(cfg, scanner) {
		return
	}

	fmt.Println()
	fmt.Println("  === 離線遊玩模式 ===")

	modArgs, ok := selectLaunchMod(cfg.D2RPath, scanner)
	if !ok {
		return
	}

	fmt.Println("  正在啟動 D2R（離線模式）...")
	pid, err := process.LaunchD2ROffline(cfg.D2RPath, modArgs...)
	if err != nil {
		fmt.Printf("  啟動失敗：%v\n", err)
		return
	}
	fmt.Printf("  ✔ D2R 已啟動 (PID: %d)\n", pid)
	fmt.Println()
}

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

		fmt.Println("  無效輸入，請輸入 p / b / h / q。")
	}
}

func setupD2RPath(cfg *config.Config) bool {
	fmt.Println()
	fmt.Println("  === 設定 D2R 路徑 ===")
	fmt.Println("  即將開啟 Windows 檔案選擇視窗，請選擇 D2R.exe。")

	selectedPath, err := config.PickD2RPath(cfg.D2RPath)
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

func renameLaunchedWindow(pid uint32, displayName string) {
	fmt.Printf("  正在準備重命名視窗：%s\n", displayName)
	err := process.RenameWindow(pid, d2r.WindowTitle(displayName), 15, 2*time.Second)
	if err != nil {
		fmt.Printf("  ⚠ 視窗重命名失敗 (%s)：%v\n", displayName, err)
		return
	}

	fmt.Printf("  ✔ 視窗已重命名為 \"%s\"\n", d2r.WindowTitle(displayName))
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
		status := "未啟動"
		if runningTitles[d2r.WindowTitle(accounts[i].DisplayName)] {
			status = "已啟動"
		}
		lines = append(lines, fmt.Sprintf("  [%s] %s (%s)", status, accounts[i].DisplayName, accounts[i].Email))
	}

	return lines
}

func formatLaunchDelayMessage(delaySeconds int, nextDisplayName string) string {
	return fmt.Sprintf("  等待 %d 秒後啟動下一個帳號：%s", delaySeconds, nextDisplayName)
}

func selectLaunchMod(d2rPath string, scanner *bufio.Scanner) ([]string, bool) {
	installedMods, err := mods.DiscoverInstalled(d2rPath)
	if err != nil {
		fmt.Printf("  讀取 mods 失敗：%v\n", err)
		return nil, false
	}

	if len(installedMods) == 0 {
		fmt.Println("  找不到已安裝 mod，將以原版啟動。")
		return nil, true
	}

	fmt.Println()
	fmt.Println("  選擇 mod")
	for {
		fmt.Println("  [0] 不使用 mod")
		for i, modName := range installedMods {
			fmt.Printf("  [%d] %s\n", i+1, modName)
		}
		printSubMenuNav()
		fmt.Print("  > 請選擇：")

		if !scanner.Scan() {
			return nil, false
		}
		input := strings.TrimSpace(scanner.Text())
		if nav := isMenuNav(input); nav != "" {
			return nil, false
		}

		selected, err := strconv.Atoi(input)
		if err != nil || selected < 0 || selected > len(installedMods) {
			fmt.Println("  無效輸入，請重試。")
			continue
		}

		if selected == 0 {
			fmt.Println("  本次啟動不使用 mod。")
			return nil, true
		}

		modName := installedMods[selected-1]
		fmt.Printf("  本次使用 mod：%s\n", modName)
		return mods.BuildLaunchArgs(modName), true
	}
}

func accountLaunchArgs(acc account.Account, modArgs []string) []string {
	args := make([]string, 0, len(modArgs)+4)
	args = append(args, modArgs...)
	args = append(args, account.LaunchArgs(acc.LaunchFlags)...)
	return args
}

func setupAccountLaunchFlags(accounts []account.Account, accountsFile string, scanner *bufio.Scanner) {
	if len(accounts) == 0 {
		fmt.Println("  目前沒有可設定的帳號。")
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Println("  === 帳號啟動 flag 設定 ===")
	printAccountLaunchFlagSummary(accounts)
	fmt.Println()
	fmt.Println("  [1] 設定 flag")
	fmt.Println("  [2] 取消 flag")
	printSubMenuNav()
	fmt.Print("  > 請選擇：")

	if !scanner.Scan() {
		return
	}
	choice := strings.TrimSpace(scanner.Text())
	if isMenuNav(choice) != "" {
		return
	}

	var setMode bool
	var actionLabel string
	switch choice {
	case "1":
		setMode = true
		actionLabel = "設定"
	case "2":
		setMode = false
		actionLabel = "取消"
	default:
		fmt.Println("  無效輸入，請重試。")
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Printf("  這次要如何%s flag？\n", actionLabel)
	fmt.Println("  [1] 以 flag 為維度")
	fmt.Println("  [2] 以帳號為維度")
	printSubMenuNav()
	fmt.Print("  > 請選擇：")

	if !scanner.Scan() {
		return
	}
	modeChoice := strings.TrimSpace(scanner.Text())
	if isMenuNav(modeChoice) != "" {
		return
	}

	switch modeChoice {
	case "1":
		configureFlagsByFlag(accounts, accountsFile, scanner, setMode)
	case "2":
		configureFlagsByAccount(accounts, accountsFile, scanner, setMode)
	default:
		fmt.Println("  無效輸入，請重試。")
		fmt.Println()
	}
}

func configureFlagsByFlag(accounts []account.Account, accountsFile string, scanner *bufio.Scanner, setMode bool) {
	options := account.LaunchFlagOptions()
	fmt.Println()
	fmt.Println("  可用 flag：")
	printLaunchFlagOptions(options)
	printSubMenuNav()
	fmt.Print("  > 請選擇 flag 編號：")

	if !scanner.Scan() {
		return
	}
	input := strings.TrimSpace(scanner.Text())
	if isMenuNav(input) != "" {
		return
	}

	selected, err := strconv.Atoi(input)
	if err != nil || selected < 1 || selected > len(options) {
		fmt.Println("  無效的 flag 編號。")
		fmt.Println()
		return
	}

	option := options[selected-1]
	actionLabel := flagActionLabel(setMode)
	fmt.Println()
	fmt.Printf("  請輸入要%s「%s」的帳號編號，可用 2,4,6 或 1-3,5-7：\n", actionLabel, option.Name)
	printAccountLaunchFlagSummary(accounts)
	printSubMenuNav()
	fmt.Print("  > 請輸入：")

	if !scanner.Scan() {
		return
	}
	input = strings.TrimSpace(scanner.Text())
	if isMenuNav(input) != "" {
		return
	}

	accountIndexes, err := parseSelectionInput(input, len(accounts))
	if err != nil {
		fmt.Printf("  解析失敗：%v\n", err)
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Printf("  即將%s以下帳號的 flag「%s」：\n", actionLabel, option.Name)
	for _, idx := range accountIndexes {
		acc := accounts[idx]
		fmt.Printf("  [%d] %s (%s)  目前：%s\n", idx+1, acc.DisplayName, acc.Email, account.LaunchFlagsSummary(acc.LaunchFlags))
	}
	if !confirmChanges(scanner) {
		fmt.Println("  已取消。")
		fmt.Println()
		return
	}

	if err := applyLaunchFlagChanges(accounts, accountsFile, accountIndexes, option.Bit, setMode); err != nil {
		fmt.Printf("  儲存失敗：%v\n", err)
		fmt.Println()
		return
	}

	fmt.Printf("  ✔ 已完成%s。\n", actionLabel)
	fmt.Println()
}

func configureFlagsByAccount(accounts []account.Account, accountsFile string, scanner *bufio.Scanner, setMode bool) {
	options := account.LaunchFlagOptions()
	fmt.Println()
	fmt.Println("  帳號列表：")
	printAccountLaunchFlagSummary(accounts)
	printSubMenuNav()
	fmt.Print("  > 請選擇帳號編號：")

	if !scanner.Scan() {
		return
	}
	input := strings.TrimSpace(scanner.Text())
	if isMenuNav(input) != "" {
		return
	}

	selected, err := strconv.Atoi(input)
	if err != nil || selected < 1 || selected > len(accounts) {
		fmt.Println("  無效的帳號編號。")
		fmt.Println()
		return
	}

	accountIndex := selected - 1
	acc := accounts[accountIndex]
	actionLabel := flagActionLabel(setMode)
	fmt.Println()
	fmt.Printf("  請輸入要對帳號「%s」%s的 flag 編號，可用 1,3 或 2-4：\n", acc.DisplayName, actionLabel)
	printLaunchFlagOptions(options)
	printSubMenuNav()
	fmt.Print("  > 請輸入：")

	if !scanner.Scan() {
		return
	}
	input = strings.TrimSpace(scanner.Text())
	if isMenuNav(input) != "" {
		return
	}

	flagIndexes, err := parseSelectionInput(input, len(options))
	if err != nil {
		fmt.Printf("  解析失敗：%v\n", err)
		fmt.Println()
		return
	}

	mask := selectedLaunchFlagMask(flagIndexes, options)
	fmt.Println()
	fmt.Printf("  即將對帳號「%s」%s以下 flag：\n", acc.DisplayName, actionLabel)
	for _, idx := range flagIndexes {
		option := options[idx]
		fmt.Printf("  [%d] %s（%s）\n", idx+1, option.Name, option.Description)
	}
	if !confirmChanges(scanner) {
		fmt.Println("  已取消。")
		fmt.Println()
		return
	}

	if err := applyLaunchFlagChanges(accounts, accountsFile, []int{accountIndex}, mask, setMode); err != nil {
		fmt.Printf("  儲存失敗：%v\n", err)
		fmt.Println()
		return
	}

	fmt.Printf("  ✔ 已完成%s。\n", actionLabel)
	fmt.Println()
}

func applyLaunchFlagChanges(accounts []account.Account, accountsFile string, accountIndexes []int, mask uint32, setMode bool) error {
	if setMode && hasConflictingLaunchFlags(mask) {
		return fmt.Errorf("關閉聲音與背景保留聲音不可同時設定，請分開操作")
	}

	previous := make(map[int]uint32, len(accountIndexes))
	for _, idx := range accountIndexes {
		previous[idx] = accounts[idx].LaunchFlags
		if setMode {
			accounts[idx].LaunchFlags |= mask
			accounts[idx].LaunchFlags = normalizeLaunchFlags(accounts[idx].LaunchFlags, mask)
			continue
		}
		accounts[idx].LaunchFlags &^= mask
	}

	if err := account.SaveAccounts(accountsFile, accounts); err != nil {
		for idx, flags := range previous {
			accounts[idx].LaunchFlags = flags
		}
		return err
	}
	return nil
}

func printAccountLaunchFlagSummary(accounts []account.Account) {
	for i, acc := range accounts {
		fmt.Printf("  [%d] %s (%s)  flag：%s\n", i+1, acc.DisplayName, acc.Email, account.LaunchFlagsSummary(acc.LaunchFlags))
	}
}

func printLaunchFlagOptions(options []account.LaunchFlagOption) {
	for i, option := range options {
		fmt.Printf("  [%d] %s", i+1, option.Name)
		if option.Description != "" {
			fmt.Printf("（%s）", option.Description)
		}
		if option.Experimental {
			fmt.Print("，效果依版本而定")
		}
		fmt.Println()
	}
}

func parseSelectionInput(input string, max int) ([]int, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("請至少輸入一個編號")
	}

	selected := make(map[int]bool)
	for _, part := range strings.Split(input, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("輸入中有空白項目")
		}

		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("無法辨識區間 %q", part)
			}
			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("無法辨識區間起點 %q", part)
			}
			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("無法辨識區間終點 %q", part)
			}
			if start > end {
				return nil, fmt.Errorf("區間 %q 起點不可大於終點", part)
			}
			if start < 1 || end > max {
				return nil, fmt.Errorf("區間 %q 超出可選範圍 1-%d", part, max)
			}
			for i := start; i <= end; i++ {
				selected[i-1] = true
			}
			continue
		}

		value, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("無法辨識編號 %q", part)
		}
		if value < 1 || value > max {
			return nil, fmt.Errorf("編號 %d 超出可選範圍 1-%d", value, max)
		}
		selected[value-1] = true
	}

	indexes := make([]int, 0, len(selected))
	for idx := range selected {
		indexes = append(indexes, idx)
	}
	sort.Ints(indexes)
	return indexes, nil
}

func selectedLaunchFlagMask(flagIndexes []int, options []account.LaunchFlagOption) uint32 {
	var mask uint32
	for _, idx := range flagIndexes {
		mask |= options[idx].Bit
	}
	return mask
}

func hasConflictingLaunchFlags(mask uint32) bool {
	return mask&account.LaunchFlagNoSound != 0 && mask&account.LaunchFlagSoundInBackground != 0
}

func normalizeLaunchFlags(flags uint32, changedMask uint32) uint32 {
	if changedMask&account.LaunchFlagNoSound != 0 {
		flags &^= account.LaunchFlagSoundInBackground
	}
	if changedMask&account.LaunchFlagSoundInBackground != 0 {
		flags &^= account.LaunchFlagNoSound
	}
	return flags
}

func confirmChanges(scanner *bufio.Scanner) bool {
	fmt.Print("  > 確認套用？(Y/n)：")
	if !scanner.Scan() {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "" || answer == "y"
}

func flagActionLabel(setMode bool) string {
	if setMode {
		return "設定"
	}
	return "取消"
}

func setupSwitcher(cfg *config.Config, scanner *bufio.Scanner) {
	fmt.Println()
	fmt.Println("  === 視窗切換設定 ===")

	if cfg.Switcher != nil && cfg.Switcher.Enabled {
		fmt.Printf("  目前設定：%s\n", switcher.FormatSwitcherDisplay(cfg.Switcher.Modifiers, cfg.Switcher.Key, cfg.Switcher.GamepadIndex))
	} else {
		fmt.Println("  目前狀態：未啟用")
	}

	fmt.Println()
	fmt.Println("  [1] 設定切換按鍵")
	fmt.Println("  [0] 關閉切換功能")
	printSubMenuNav()
	fmt.Print("  > 請選擇：")

	if !scanner.Scan() {
		return
	}
	choice := strings.TrimSpace(scanner.Text())

	if isMenuNav(choice) != "" {
		return
	}

	switch choice {
	case "1":
		// 先停止現有的 switcher 以避免衝突
		wasRunning := switcher.IsRunning()
		switcher.Stop()

		fmt.Println()
		fmt.Println("  請按下想用來切換視窗的按鍵組合...")
		fmt.Println("  （支援：鍵盤任意鍵 + Ctrl/Alt/Shift、滑鼠側鍵、搖桿按鈕）")
		fmt.Println("  （搖桿組合鍵：先按住修飾按鈕，再按觸發按鈕，放開後完成偵測）")
		fmt.Println("  （按 Esc 取消）")
		fmt.Println()

		modifiers, key, gamepadIndex, err := switcher.DetectKeyPress()
		if err != nil {
			fmt.Printf("  ⚠ 偵測失敗：%v\n", err)
			restartSwitcherIfNeeded(cfg, wasRunning)
			return
		}
		if key == "" {
			fmt.Println("  已取消。")
			restartSwitcherIfNeeded(cfg, wasRunning)
			return
		}

		display := switcher.FormatSwitcherDisplay(modifiers, key, gamepadIndex)
		fmt.Printf("  偵測到：%s\n", display)
		fmt.Print("  確認使用此組合？(Y/n)：")

		if !scanner.Scan() {
			return
		}
		answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if answer != "" && answer != "y" {
			fmt.Println("  已取消。")
			restartSwitcherIfNeeded(cfg, wasRunning)
			return
		}

		cfg.Switcher = &config.SwitcherConfig{
			Enabled:      true,
			Modifiers:    modifiers,
			Key:          key,
			GamepadIndex: gamepadIndex,
		}
		if err := config.Save(cfg); err != nil {
			fmt.Printf("  ⚠ 設定儲存失敗：%v\n", err)
			return
		}

		if err := switcher.Start(cfg.Switcher); err != nil {
			fmt.Printf("  ⚠ 切換功能啟動失敗：%v\n", err)
			return
		}

		fmt.Printf("  ✔ 已儲存切換設定：%s\n", display)

	case "0":
		switcher.Stop()
		if cfg.Switcher != nil {
			cfg.Switcher.Enabled = false
		}
		if err := config.Save(cfg); err != nil {
			fmt.Printf("  ⚠ 設定儲存失敗：%v\n", err)
			return
		}
		fmt.Println("  ✔ 已關閉切換功能")
	}

	fmt.Println()
}

func restartSwitcherIfNeeded(cfg *config.Config, wasRunning bool) {
	if wasRunning && cfg.Switcher != nil && cfg.Switcher.Enabled {
		if err := switcher.Start(cfg.Switcher); err != nil {
			fmt.Printf("  ⚠ 重新啟動切換功能失敗：%v\n", err)
		}
	}
}

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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
	fmt.Printf("  d2r-hyper-launcher  v%s\n", version)
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
			launchOffline(cfg.D2RPath, scanner)
			continue
		case "a":
			launchAll(accounts, cfg.D2RPath, cfg.LaunchDelay, scanner)
			continue
		case "p":
			setupD2RPath(cfg)
			continue
		case "s":
			setupSwitcher(cfg, scanner)
			continue
		default:
			id, err := strconv.Atoi(input)
			if err != nil || id < 1 || id > len(accounts) {
				fmt.Println("  無效輸入，請重試。")
				continue
			}
			acc := &accounts[id-1]
			launchAccount(acc, cfg.D2RPath, scanner)
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
	fmt.Println("--------------------------------------------")
	fmt.Println("  <數字>  啟動指定帳號")
	fmt.Println("  0       離線遊玩（可選 mod，不需帳密）")
	fmt.Println("  a       啟動所有帳號（可選 mod，只啟動未啟動的）")
	fmt.Println("  p       選擇 D2R.exe 路徑")
	fmt.Println("  s       視窗切換設定")
	fmt.Println("  r       重新整理狀態")
	fmt.Println("  q       退出")
	fmt.Println("--------------------------------------------")
}

func launchAccount(acc *account.Account, d2rPath string, scanner *bufio.Scanner) {
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

	// 解密密碼
	password, err := account.GetDecryptedPassword(acc)
	if err != nil {
		fmt.Printf("  密碼解密失敗：%v\n", err)
		return
	}

	fmt.Printf("  正在啟動 %s (%s)...\n", acc.DisplayName, region.Name)

	// 啟動 D2R
	pid, err := process.LaunchD2R(d2rPath, acc.Email, password, region.Address)
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

func launchAll(accounts []account.Account, d2rPath string, launchDelay int, scanner *bufio.Scanner) {
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

	modArgs, ok := selectLaunchMod(d2rPath, scanner)
	if !ok {
		return
	}

	runningTitles := runningAccountWindowTitles()
	pendingAccounts := pendingBatchAccounts(accounts, runningTitles)
	if len(pendingAccounts) == 0 {
		fmt.Println("  所有帳號都已在執行中。")
		fmt.Println()
		return
	}

	for i := range accounts {
		if runningTitles[d2r.WindowTitle(accounts[i].DisplayName)] {
			fmt.Printf("  ⏭ %s 已在執行中，跳過\n", accounts[i].DisplayName)
		}
	}

	fmt.Printf("  已預先掃描目前執行中的 D2R 視窗，本次將啟動 %d 個尚未啟動的帳號。\n", len(pendingAccounts))

	for i, acc := range pendingAccounts {

		password, err := account.GetDecryptedPassword(acc)
		if err != nil {
			fmt.Printf("  ⚠ 帳號 %s 密碼解密失敗：%v\n", acc.DisplayName, err)
			continue
		}

		fmt.Printf("  正在啟動 %s (%s)...\n", acc.DisplayName, region.Name)
		pid, err := process.LaunchD2R(d2rPath, acc.Email, password, region.Address, modArgs...)
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

		if launchDelay > 0 && i+1 < len(pendingAccounts) {
			fmt.Println(formatLaunchDelayMessage(launchDelay, pendingAccounts[i+1].DisplayName))
			time.Sleep(time.Duration(launchDelay) * time.Second)
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
	fmt.Println("  CSV 格式：Email,Password,DisplayName")
	fmt.Println("  範例：your-account1@example.com,your-password-here,主帳號-法師(倉庫/武器/飾品)")
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

func launchOffline(d2rPath string, scanner *bufio.Scanner) {
	fmt.Println()
	fmt.Println("  === 離線遊玩模式 ===")

	modArgs, ok := selectLaunchMod(d2rPath, scanner)
	if !ok {
		return
	}

	fmt.Println("  正在啟動 D2R（離線模式）...")
	pid, err := process.LaunchD2ROffline(d2rPath, modArgs...)
	if err != nil {
		fmt.Printf("  啟動失敗：%v\n", err)
		return
	}
	fmt.Printf("  ✔ D2R 已啟動 (PID: %d)\n", pid)
	fmt.Println()
}

func setupD2RPath(cfg *config.Config) {
	fmt.Println()
	fmt.Println("  === 設定 D2R 路徑 ===")
	fmt.Println("  即將開啟 Windows 檔案選擇視窗，請選擇 D2R.exe。")

	selectedPath, err := config.PickD2RPath(cfg.D2RPath)
	if err != nil {
		fmt.Printf("  ⚠ D2R 路徑設定失敗：%v\n", err)
		fmt.Println()
		return
	}
	if selectedPath == "" {
		fmt.Println("  已取消。")
		fmt.Println()
		return
	}

	cfg.D2RPath = selectedPath
	if err := config.Save(cfg); err != nil {
		fmt.Printf("  ⚠ 設定儲存失敗：%v\n", err)
		fmt.Println()
		return
	}

	fmt.Printf("  ✔ 已更新 D2R 路徑：%s\n", cfg.D2RPath)
	fmt.Println()
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

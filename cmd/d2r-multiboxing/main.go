package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"d2r-multiboxing/internal/account"
	"d2r-multiboxing/internal/config"
	"d2r-multiboxing/internal/d2r"
	"d2r-multiboxing/internal/handle"
	"d2r-multiboxing/internal/modfile"
	"d2r-multiboxing/internal/process"
	"d2r-multiboxing/internal/switcher"

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
	fmt.Printf("  D2R Multiboxing Launcher  v%s\n", version)
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

	if !fileExists(accountsFile) {
		fmt.Printf("  找不到 %s，請先建立帳號設定檔。\n", accountsFile)
		fmt.Println("  CSV 格式：Email,Password,DisplayName")
		fmt.Println("  範例：account@email.com,password123,主帳號")
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
		case "m":
			installMods(cfg.D2RPath, scanner)
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
	fmt.Println("  0       離線遊玩（不需帳密）")
	fmt.Println("  a       啟動所有帳號（只啟動未啟動的）")
	fmt.Println("  m       安裝 Mod 到 D2R")
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

	// 重命名視窗
	go func() {
		err := process.RenameWindow(pid, d2r.WindowTitle(acc.DisplayName), 15, 2*time.Second)
		if err != nil {
			fmt.Printf("  ⚠ 視窗重命名失敗 (%s)：%v\n", acc.DisplayName, err)
		} else {
			fmt.Printf("  ✔ 視窗已重命名為 \"%s\"\n", d2r.WindowTitle(acc.DisplayName))
		}
	}()

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

	for i := range accounts {
		acc := &accounts[i]

		// 已有視窗存在則跳過
		if process.FindWindowByTitle(d2r.WindowTitle(acc.DisplayName)) {
			fmt.Printf("  ⏭ %s 已在執行中，跳過\n", acc.DisplayName)
			continue
		}

		if i > 0 && launchDelay > 0 {
			fmt.Printf("  等待 %d 秒...\n", launchDelay)
			time.Sleep(time.Duration(launchDelay) * time.Second)
		}
		password, err := account.GetDecryptedPassword(acc)
		if err != nil {
			fmt.Printf("  ⚠ 帳號 %s 密碼解密失敗：%v\n", acc.DisplayName, err)
			continue
		}

		fmt.Printf("  正在啟動 %s (%s)...\n", acc.DisplayName, region.Name)
		pid, err := process.LaunchD2R(d2rPath, acc.Email, password, region.Address)
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

		// 背景重命名視窗
		displayName := acc.DisplayName
		go func() {
			err := process.RenameWindow(pid, d2r.WindowTitle(displayName), 15, 2*time.Second)
			if err == nil {
				fmt.Printf("  ✔ 視窗已重命名為 \"%s\"\n", d2r.WindowTitle(displayName))
			}
		}()
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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

// localModsDir returns the mods/ directory next to the running executable.
func localModsDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "mods"
	}
	dir := filepath.Join(filepath.Dir(exe), "mods")
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir
	}
	// Fallback to current directory
	return "mods"
}

func launchOffline(d2rPath string, scanner *bufio.Scanner) {
	fmt.Println()
	fmt.Println("  === 離線遊玩模式 ===")

	// 掃描 D2R 已安裝的 mod
	installedMods, _ := modfile.DiscoverInstalledMods(d2rPath)

	// 顯示 mod 選擇
	fmt.Println("  選擇 Mod：")
	fmt.Println("  [0] 不使用 Mod（原版）")
	for i, name := range installedMods {
		fmt.Printf("  [%d] %s\n", i+1, name)
	}
	printSubMenuNav()
	fmt.Print("  > 請選擇：")

	if !scanner.Scan() {
		return
	}
	choice := strings.TrimSpace(scanner.Text())

	if nav := isMenuNav(choice); nav != "" {
		return
	}

	var extraArgs []string
	if choice != "0" && choice != "" {
		idx, err := strconv.Atoi(choice)
		if err != nil || idx < 1 || idx > len(installedMods) {
			fmt.Println("  無效選擇。")
			return
		}

		modName := installedMods[idx-1]
		extraArgs = append(extraArgs, "-mod", modName, "-txt")
	}

	fmt.Println("  正在啟動 D2R（離線模式）...")
	pid, err := process.LaunchD2ROffline(d2rPath, extraArgs...)
	if err != nil {
		fmt.Printf("  啟動失敗：%v\n", err)
		return
	}
	fmt.Printf("  ✔ D2R 已啟動 (PID: %d)\n", pid)
	if len(extraArgs) > 0 {
		fmt.Printf("  ✔ 使用 Mod: %s\n", extraArgs[1])
	}
	fmt.Println()
}

func installMods(d2rPath string, scanner *bufio.Scanner) {
	fmt.Println()
	fmt.Println("  === 安裝 Mod 到 D2R ===")

	modsDir := localModsDir()
	availableMods, err := modfile.DiscoverMods(modsDir)
	if err != nil {
		fmt.Printf("  掃描 mods/ 目錄失敗：%v\n", err)
		return
	}

	if len(availableMods) == 0 {
		fmt.Println("  mods/ 目錄下沒有找到任何 Mod。")
		return
	}

	// 掃描已安裝
	installedMods, _ := modfile.DiscoverInstalledMods(d2rPath)
	installedSet := make(map[string]bool)
	for _, m := range installedMods {
		installedSet[m] = true
	}

	fmt.Println("  可用 Mod：")
	for i, name := range availableMods {
		status := ""
		if installedSet[name] {
			status = " ✔ 已安裝"
		}
		fmt.Printf("  [%d] %s%s\n", i+1, name, status)
	}
	fmt.Println("  a       安裝全部")
	printSubMenuNav()
	fmt.Print("  > 請選擇：")

	if !scanner.Scan() {
		return
	}
	choice := strings.TrimSpace(strings.ToLower(scanner.Text()))

	if choice == "" || isMenuNav(choice) != "" {
		return
	}

	var toInstall []string
	if choice == "a" {
		toInstall = availableMods
	} else {
		idx, err := strconv.Atoi(choice)
		if err != nil || idx < 1 || idx > len(availableMods) {
			fmt.Println("  無效選擇。")
			return
		}
		toInstall = []string{availableMods[idx-1]}
	}

	d2rModsDir := modfile.D2RModsDir(d2rPath)
	for _, name := range toInstall {
		srcDir := filepath.Join(modsDir, name)
		fmt.Printf("  正在安裝 %s → %s...\n", name, d2rModsDir)
		if err := modfile.InstallMod(srcDir, d2rPath); err != nil {
			fmt.Printf("  ⚠ %s 安裝失敗：%v\n", name, err)
			continue
		}
		fmt.Printf("  ✔ %s 安裝完成\n", name)
	}
	fmt.Println()
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

package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"d2r-multiboxing/internal/account"
	"d2r-multiboxing/internal/config"
	"d2r-multiboxing/internal/d2r"
	"d2r-multiboxing/internal/handle"
	"d2r-multiboxing/internal/process"
)

// launchedPIDs tracks PIDs launched by this session for window renaming.
var launchedPIDs = struct {
	sync.Mutex
	m map[uint32]string // PID -> DisplayName
}{m: make(map[uint32]string)}

func main() {
	fmt.Println("============================================")
	fmt.Println("  D2R Multiboxing Launcher")
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
	fmt.Printf("  啟動間隔：%d 秒\n\n", cfg.LaunchDelay)

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
			fmt.Println("  再見！")
			return
		case "r":
			// 重新整理，重新讀取帳號
			accounts, err = account.LoadAccounts(accountsFile)
			if err != nil {
				fmt.Printf("  讀取帳號失敗：%v\n", err)
			}
			continue
		case "a":
			launchAll(accounts, cfg.D2RPath, cfg.LaunchDelay, scanner)
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
	d2rProcesses, _ := process.FindProcessesByName(d2r.ProcessName)
	runningPIDs := make(map[uint32]bool)
	for _, p := range d2rProcesses {
		runningPIDs[p.PID] = true
	}

	fmt.Println("  帳號列表：")
	for i, acc := range accounts {
		status := "未啟動"
		// 檢查此帳號是否已透過本工具啟動
		launchedPIDs.Lock()
		for pid, name := range launchedPIDs.m {
			if name == acc.DisplayName && runningPIDs[pid] {
				status = fmt.Sprintf("已啟動 (PID: %d)", pid)
				break
			}
		}
		launchedPIDs.Unlock()

		fmt.Printf("  [%d] %-15s (%s)  [%s]\n",
			i+1, acc.DisplayName, acc.Email, status)
	}
	fmt.Println()
	fmt.Println("--------------------------------------------")
	fmt.Println("  <數字>  啟動指定帳號")
	fmt.Println("  a       啟動所有帳號")
	fmt.Println("  r       重新整理狀態")
	fmt.Println("  q       退出")
	fmt.Println("--------------------------------------------")
}

func launchAccount(acc *account.Account, d2rPath string, scanner *bufio.Scanner) {
	// 選擇區域
	fmt.Print("  > 選擇區域 (1=NA, 2=EU, 3=Asia)：")
	if !scanner.Scan() {
		return
	}
	region := parseRegionInput(scanner.Text())
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

	// 記錄 PID
	launchedPIDs.Lock()
	launchedPIDs.m[pid] = acc.DisplayName
	launchedPIDs.Unlock()

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
		err := process.RenameWindow(pid, acc.DisplayName, 15, 2*time.Second)
		if err != nil {
			fmt.Printf("  ⚠ 視窗重命名失敗 (%s)：%v\n", acc.DisplayName, err)
		} else {
			fmt.Printf("  ✔ 視窗已重命名為 \"%s\"\n", acc.DisplayName)
		}
	}()

	fmt.Println()
}

func launchAll(accounts []account.Account, d2rPath string, launchDelay int, scanner *bufio.Scanner) {
	fmt.Print("  > 選擇區域 (1=NA, 2=EU, 3=Asia)：")
	if !scanner.Scan() {
		return
	}
	region := parseRegionInput(scanner.Text())
	if region == nil {
		fmt.Println("  無效的區域選擇。")
		return
	}

	for i := range accounts {
		if i > 0 && launchDelay > 0 {
			fmt.Printf("  等待 %d 秒...\n", launchDelay)
			time.Sleep(time.Duration(launchDelay) * time.Second)
		}
		acc := &accounts[i]
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

		launchedPIDs.Lock()
		launchedPIDs.m[pid] = acc.DisplayName
		launchedPIDs.Unlock()

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
			err := process.RenameWindow(pid, displayName, 15, 2*time.Second)
			if err == nil {
				fmt.Printf("  ✔ 視窗已重命名為 \"%s\"\n", displayName)
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

package main

import (
	"fmt"
	"os"
	"os/exec"
)

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

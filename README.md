# d2r-hyper-launcher

Windows 上給 D2R（Diablo II: Resurrected）玩家使用的 CLI 工具，目前提供兩個主要功能：

- **multiboxing**：多帳號啟動、單實例鎖處理、視窗辨識
- **switcher**：鍵盤／滑鼠側鍵／搖桿切換 D2R 視窗

## 給玩家的快速開始

這一段是給「只想直接玩」的玩家看的，不需要先懂 Go 或程式碼。

### 1. 先下載這兩個檔案

- [d2r-hyper-launcher.exe](d2r-hyper-launcher.exe)
- [accounts.csv](accounts.csv)

建議把 `d2r-hyper-launcher.exe` 放在你方便執行的位置，例如桌面上的 `D2R-Hyper-Launcher` 資料夾。

### 2. 把 `accounts.csv` 複製到資料目錄

本工具會固定讀取下面這個位置的帳號檔：

```text
%USERPROFILE%\.d2r-hyper-launcher\accounts.csv
```

如果你不知道 `%USERPROFILE%` 是哪裡，也沒關係：先執行一次 `d2r-hyper-launcher.exe`，畫面上會直接顯示實際的「資料目錄」完整路徑，你再把 `accounts.csv` 複製進去即可。

你可以直接用 PowerShell 建立資料夾並複製範本：

```powershell
New-Item -ItemType Directory -Force "$env:USERPROFILE\.d2r-hyper-launcher" | Out-Null
Copy-Item .\accounts.csv "$env:USERPROFILE\.d2r-hyper-launcher\accounts.csv" -Force
```

### 3. 編輯 `accounts.csv`

建議優先用 Excel 開啟 `%USERPROFILE%\.d2r-hyper-launcher\accounts.csv` 後再編輯，避免手動修改時不小心破壞 CSV 格式。  
打開後，照這個格式填入 Battle.net 帳號：

```csv
Email,Password,DisplayName
your-account1@example.com,your-password-here,主帳號-法師
your-account2@example.com,your-password-here,副帳號-野蠻人
```

欄位說明：

- `Email`：Battle.net 登入信箱
- `Password`：Battle.net 密碼
- `DisplayName`：工具內顯示名稱，也是視窗切換時看到的名稱

> 第一次執行後，明文密碼會自動改寫成 `ENC:` 開頭的加密字串。  
> 這是用 Windows DPAPI 加密，之後不用自己手動加密；若換電腦或換 Windows 使用者，請再填一次明文密碼。

### 4. 執行 `d2r-hyper-launcher.exe`

你可以直接雙擊 `d2r-hyper-launcher.exe`，或在 PowerShell 執行：

```powershell
.\d2r-hyper-launcher.exe
```

第一次啟動時，工具也會自動建立這個設定檔：

```text
%USERPROFILE%\.d2r-hyper-launcher\config.json
```

如果你的 D2R 不在預設路徑 `C:\Program Files (x86)\Diablo II Resurrected\D2R.exe`，請在主選單輸入 `p`，工具會直接開啟 Windows 檔案選擇視窗，讓你選擇正確的 `D2R.exe`。一般玩家不需要手動修改 `config.json`。

### 5. 主選單最常用功能

啟動後，最常用的是這幾個選項：

- `<數字>`：啟動指定帳號
- `a`：先選區域，再選一次要套用的已安裝 mod，接著依序啟動所有尚未開啟的帳號
- `0`：先選一次要套用的已安裝 mod，再進離線模式
- `p`：開啟 Windows 檔案選擇視窗，設定 `D2R.exe` 路徑
- `s`：設定視窗切換快捷鍵／滑鼠側鍵／搖桿按鍵
- `r`：重新讀取 `accounts.csv` 並刷新狀態
- `q`：離開工具

### 6. 想看更仔細的操作說明

如果你想看每個選單怎麼用、每個步驟會看到什麼畫面，請直接讀：

- [docs/multiboxing-usage-guide.md](docs/multiboxing-usage-guide.md) — 多開啟動、帳號檔、區域選擇、離線模式
- [docs/switcher-usage-guide.md](docs/switcher-usage-guide.md) — 視窗切換設定、支援按鍵類型、常見問題

如果你想看底層實作與技術原理，再讀：

- [docs/multiboxing-technical-guide.md](docs/multiboxing-technical-guide.md)
- [docs/switcher-technical-guide.md](docs/switcher-technical-guide.md)
- [docs/D2R_PARAMS.md](docs/D2R_PARAMS.md)

## 給想自己編譯的人

### 前置需求

- Windows 10 / 11
- Go 1.26+
- Battle.net 版 D2R

### 編譯

```powershell
go build -o d2r-hyper-launcher.exe ./cmd/d2r-hyper-launcher
```

### 測試

在這台 Windows 環境若直接跑 `go test ./...` 被 Application Control 擋下，請改用 repo 內建包裝腳本：

```powershell
.\scripts\go-test.ps1
go build ./cmd/d2r-hyper-launcher
```

## 注意事項

- 建議先把 D2R 設成「視窗化」或「無邊框視窗」
- 首次設定搖桿切換按鍵時，建議以管理員權限執行
- 僅支援 Battle.net 版 D2R
- 操作進程 Handle 可能被部分防毒軟體誤報
- 本工具不會修改遊戲檔案、注入遊戲程式或自動化遊戲操作

## 授權

MIT License

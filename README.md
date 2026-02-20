# D2R Multiboxing

在同一台 Windows 電腦上同時執行多個 Diablo II: Resurrected (D2R) 實例的 CLI 輔助工具。

## 功能

- **帳號管理** — 透過 CSV 檔案管理多個 Battle.net 帳號
- **密碼加密** — 使用 Windows DPAPI 自動加密儲存密碼
- **一鍵啟動** — 透過參數方式直接啟動 D2R，支援單一帳號或全部帳號
- **自動多開** — 自動偵測並關閉 D2R 的單實例鎖定 Event Handle
- **視窗辨識** — 將各 D2R 視窗標題重命名為帳號暱稱（D2R- 前綴）
- **背景監控** — 持續監控新啟動的 D2R 進程，自動解除多開限制
- **視窗切換** — 透過快捷鍵、滑鼠側鍵或搖桿按鈕在 D2R 視窗之間快速切換焦點

## 安裝

### 前置需求

- Windows 10/11
- Go 1.26+（僅編譯時需要）
- 必須以 **管理員權限** 執行
- 請先手動在遊戲內將顯示模式設為「視窗化」

### 編譯

```powershell
# 開發版
go build -o d2r-multiboxing.exe ./cmd/d2r-multiboxing

# 指定版號
go build -ldflags "-X main.version=1.0.0" -o d2r-multiboxing.exe ./cmd/d2r-multiboxing
```

## 使用方式

### 1. 建立帳號設定檔

在資料目錄 `~/.d2r-multiboxing/` 下建立 `accounts.csv`：

```powershell
Copy-Item .\accounts.csv "$env:USERPROFILE\.d2r-multiboxing\accounts.csv"
```

```csv
Email,Password,DisplayName
account1@email.com,mypassword123,主帳號-法師
account2@email.com,anotherpass,副帳號-野蠻人
```

| 欄位 | 說明 |
|------|------|
| `Email` | Battle.net 登入信箱 |
| `Password` | 帳號密碼（首次執行後自動加密） |
| `DisplayName` | 顯示名稱（用於視窗標題與選單） |

### 2. 設定 D2R 路徑（選用）

首次執行時自動建立 `~/.d2r-multiboxing/config.json`，可自訂路徑：

```json
{
  "d2r_path": "C:\\Program Files (x86)\\Diablo II Resurrected\\D2R.exe",
  "launch_delay": 5
}
```

資料目錄可透過環境變數 `D2R_MULTIBOXING_HOME` 自訂。

### 3. 執行

以 **管理員身份** 開啟 PowerShell 並執行：

```powershell
.\d2r-multiboxing.exe
```

### 4. 操作選單

```
============================================
  D2R Multiboxing Launcher  v1.0.0
============================================

  資料目錄：C:\Users\User\.d2r-multiboxing
  D2R 路徑：C:\Program Files (x86)\Diablo II Resurrected\D2R.exe
  啟動間隔：5 秒

  帳號列表：
  [1] 主帳號-法師      (account1@email.com)  [未啟動]
  [2] 副帳號-野蠻人    (account2@email.com)  [已啟動]

--------------------------------------------
  <數字>  啟動指定帳號
  a       啟動所有帳號（只啟動未啟動的）
  s       視窗切換設定
  r       重新整理狀態
  q       退出
--------------------------------------------
```

### 5. 設定視窗切換

輸入 `s` 進入設定引導，按下想用的按鍵組合即可自動偵測並記錄：

```
  === 視窗切換設定 ===
  目前狀態：未啟用

  [1] 設定切換按鍵
  [0] 關閉切換功能
  [Enter] 返回
  > 請選擇：1

  請按下想用來切換視窗的按鍵組合...
  偵測到：Ctrl+Tab（Tab 鍵）
  確認使用此組合？(Y/n)：

  ✔ 已儲存切換設定：Ctrl+Tab（Tab 鍵）
```

支援鍵盤快捷鍵（如 `Ctrl+Tab`、`Alt+F1`）、滑鼠側鍵（`XButton1`、`XButton2`）與搖桿按鈕（XInput）。
設定存入 `config.json` 後自動載入。

## 技術原理

D2R 啟動時會建立名為 `DiabloII Check For Other Instances` 的 Windows Event Handle 來阻止多開。本工具透過 Windows NT API (`NtDuplicateObject` + `DuplicateCloseSource`) 自動關閉該 Handle，允許多個 D2R 實例同時運行。

詳細技術說明請參考 [PLAN-v1-multiboxing.md](PLAN-v1-multiboxing.md)（多開啟動器）與 [PLAN-v2-switcher.md](PLAN-v2-switcher.md)（視窗切換），完整使用說明請參考 [USAGE.md](USAGE.md)。

## 專案結構

```
├── cmd/
│   └── d2r-multiboxing/
│       └── main.go              # CLI 互動主迴圈
├── internal/
│   ├── config/config.go         # 設定檔讀寫（含 SwitcherConfig）
│   ├── d2r/constants.go         # D2R 相關常數
│   ├── account/
│   │   ├── account.go           # 帳號 CSV 讀寫
│   │   └── crypto.go            # DPAPI 密碼加密
│   ├── handle/
│   │   ├── winapi.go            # Windows NT API 封裝
│   │   ├── enumerator.go        # Handle 列舉搜尋
│   │   └── closer.go            # Handle 關閉操作
│   ├── process/
│   │   ├── finder.go            # 進程搜尋
│   │   ├── launcher.go          # D2R 啟動
│   │   └── window.go            # 視窗操作（重命名、切換）
│   └── switcher/
│       ├── switcher.go          # Start/Stop、視窗切換核心邏輯
│       ├── hotkey.go            # RegisterHotKey 鍵盤快捷鍵
│       ├── mousehook.go         # WH_MOUSE_LL 滑鼠側鍵
│       ├── detect.go            # CLI 設定用按鍵偵測
│       ├── keymap.go            # VK code 映射表
│       └── gamepad.go           # XInput 搖桿偵測與輪詢
├── USAGE.md                     # 完整使用說明
├── PLAN-v1-multiboxing.md       # Phase 1 實作計畫（已完成）
├── PLAN-v2-switcher.md          # Phase 2 實作計畫（已完成）
├── D2R_PARAMS.md                # D2R 啟動參數一覽
└── accounts.csv                 # 帳號 CSV 範本
```

## 注意事項

- ⚠️ 必須以 **管理員權限** 執行
- ⚠️ 部分防毒軟體可能誤報（操作進程 handle 為正常行為）
- ⚠️ 使用此工具可能違反 Blizzard 服務條款，風險自負
- ⚠️ 短時間內重複啟動過多次可能被 Battle.net 擋住，請保持適當間隔
- 僅支援 Battle.net 帳號（不支援 Steam 版本）
- 密碼加密綁定當前 Windows 使用者，換機器需重新設定
- 本工具不會修改遊戲檔案、注入程式碼或自動化任何遊戲操作

## 授權

MIT License

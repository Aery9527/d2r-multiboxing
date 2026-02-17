# D2R Multiboxing

在同一台 Windows 電腦上同時執行多個 Diablo II: Resurrected (D2R) 實例的 CLI 輔助工具。

## 功能

- **帳號管理** — 透過 CSV 檔案管理多個 Battle.net 帳號
- **密碼加密** — 使用 Windows DPAPI 自動加密儲存密碼
- **一鍵啟動** — 透過參數方式直接啟動 D2R，支援單一帳號或全部帳號
- **自動多開** — 自動偵測並關閉 D2R 的單實例鎖定 Event Handle
- **視窗辨識** — 將各 D2R 視窗標題重命名為帳號暱稱
- **背景監控** — 持續監控新啟動的 D2R 進程，自動解除多開限制

## 安裝

### 前置需求

- Windows 10/11
- Go 1.26+
- 必須以 **管理員權限** 執行

### 編譯

```powershell
go build -o d2r-multiboxing.exe .
```

## 使用方式

### 1. 建立帳號設定檔

在資料目錄 `~/.d2r-multiboxing/` 下建立 `accounts.csv`：

```powershell
Copy-Item .\accounts.sample.csv "$env:USERPROFILE\.d2r-multiboxing\accounts.csv"
```

```csv
ID,Email,Password,DisplayName,Region
1,account1@email.com,mypassword123,主帳號-法師,NA
2,account2@email.com,anotherpass,副帳號-野蠻人,NA
```

| 欄位 | 說明 |
|------|------|
| `ID` | 帳號序號（從 1 開始） |
| `Email` | Battle.net 登入信箱 |
| `Password` | 帳號密碼（首次執行後自動加密） |
| `DisplayName` | 顯示名稱（用於視窗標題） |
| `Region` | 預設區域：`NA` / `EU` / `Asia` |

### 2. 設定 D2R 路徑（選用）

所有資料檔案統一存放在 `~/.d2r-multiboxing/` 目錄下（可透過環境變數 `D2R_MULTIBOXING_HOME` 自訂）。

首次執行時，工具會自動建立設定檔 `~/.d2r-multiboxing/config.json`：

```json
{
  "d2r_path": "C:\\Program Files (x86)\\Diablo II Resurrected\\D2R.exe"
}
```

若你的 D2R 安裝在其他位置，直接編輯該設定檔即可：

```powershell
notepad "$env:USERPROFILE\.d2r-multiboxing\config.json"
```

### 3. 執行

以 **管理員身份** 開啟 PowerShell 並執行：

```powershell
.\d2r-multiboxing.exe
```

### 4. 操作選單

```
============================================
  D2R Multiboxing Launcher
============================================

  帳號列表：
  [1] 主帳號-法師    (account1@email.com)  NA  [未啟動]
  [2] 副帳號-野蠻人  (account2@email.com)  NA  [已啟動]

--------------------------------------------
  <數字>  啟動指定帳號
  a       啟動所有帳號
  r       重新整理狀態
  q       退出
--------------------------------------------
```

## 技術原理

D2R 啟動時會建立名為 `DiabloII Check For Other Instances` 的 Windows Event Handle 來阻止多開。本工具透過 Windows NT API (`NtDuplicateObject` + `DuplicateCloseSource`) 自動關閉該 Handle，允許多個 D2R 實例同時運行。

詳細技術說明請參考 [PLAN.md](PLAN.md)，完整使用說明請參考 [USAGE.md](USAGE.md)。

## 專案結構

```
├── main.go                    # CLI 互動主迴圈
├── internal/
│   ├── config/
│   │   └── config.go          # 設定檔讀寫 (~/.d2r-multiboxing/config.json)
│   ├── d2r/constants.go       # D2R 相關常數
│   ├── account/
│   │   ├── account.go         # 帳號 CSV 讀寫
│   │   └── crypto.go          # DPAPI 密碼加密
│   ├── handle/
│   │   ├── winapi.go          # Windows NT API 封裝
│   │   ├── enumerator.go      # Handle 列舉搜尋
│   │   └── closer.go          # Handle 關閉操作
│   └── process/
│       ├── finder.go          # 進程搜尋
│       ├── launcher.go        # D2R 啟動
│       └── window.go          # 視窗重命名
```

## 注意事項

- ⚠️ 必須以 **管理員權限** 執行
- ⚠️ 部分防毒軟體可能誤報（操作進程 handle 為正常行為）
- ⚠️ 使用此工具可能違反 Blizzard 服務條款，風險自負
- 僅支援 Battle.net 帳號（不支援 Steam 版本）
- 密碼加密綁定當前 Windows 使用者，換機器需重新設定

## 授權

MIT License

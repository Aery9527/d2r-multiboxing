# D2R Multiboxing CLI Tool - 實作計畫

> **專案類型**：Windows CLI 工具（Go 語言）
> **目的**：在同一台 Windows 電腦上同時執行多個 Diablo II: Resurrected (D2R) 遊戲實例

---

## 目錄

- [問題描述](#問題描述)
- [技術原理](#技術原理)
- [功能範圍](#功能範圍)
- [專案架構](#專案架構)
- [模組詳細設計](#模組詳細設計)
- [工作計畫](#工作計畫)
- [技術細節](#技術細節)
- [依賴管理](#依賴管理)
- [注意事項](#注意事項)
- [技術參考](#技術參考)

---

## 問題描述

D2R 預設只允許同時執行一個遊戲實例。當 D2R 啟動時，會建立一個名為 `DiabloII Check For Other Instances` 的 Windows Event Handle，後續啟動的 D2R 會檢查此 handle 是否存在，若存在則拒絕啟動。

本工具透過程式化方式自動偵測並關閉該 Event Handle，讓使用者可以從 CLI 介面管理多個 Battle.net 帳號，並同時啟動多個 D2R 實例。

---

## 技術原理

### D2R 單實例鎖定機制

```
D2R.exe 啟動
    │
    ├── 建立 Event Handle: "DiabloII Check For Other Instances"
    │
    └── 後續 D2R.exe 啟動時
            │
            ├── 檢查 Event Handle 是否存在
            │       │
            │       ├── 存在 → 拒絕啟動（退出）
            │       └── 不存在 → 正常啟動
            │
            └── 建立新的 Event Handle
```

### 多開解除流程

```
本工具偵測到 D2R.exe 進程
    │
    ├── 1. NtQuerySystemInformation(SystemExtendedHandleInformation)
    │      → 取得全系統 handle 列表
    │
    ├── 2. 篩選 PID = D2R.exe 的 handle
    │
    ├── 3. NtDuplicateObject 複製 handle 到本進程
    │      → NtQueryObject 查詢 type 與 name
    │
    ├── 4. 比對 type="Event", name 包含 "DiabloII Check For Other Instances"
    │
    └── 5. NtDuplicateObject + DuplicateCloseSource
           → 強制關閉目標 handle
           → 下一個 D2R 實例可正常啟動
```

---

## 功能範圍

### 核心功能（MVP）

| # | 功能 | 說明 |
|---|------|------|
| 1 | **帳號管理** | 透過 CSV 檔案管理多個 Battle.net 帳號（email、密碼、暱稱、預設區域） |
| 2 | **密碼加密** | 首次輸入明文密碼後，使用 Windows DPAPI 加密儲存，後續自動解密使用 |
| 3 | **自動啟動 D2R** | 透過參數方式（`-username` / `-password` / `-address`）直接啟動 D2R.exe |
| 4 | **自動關閉 Handle** | 偵測 D2R.exe 進程並自動關閉單實例鎖定的 Event Handle |
| 5 | **視窗重命名** | 將每個 D2R 遊戲視窗標題改為帳號暱稱，方便辨識多個視窗 |
| 6 | **互動式 CLI 選單** | 顯示帳號列表、狀態，選擇啟動單一帳號或一次全開 |

### 不在本次範圍

- GUI 圖形介面（本專案為純 CLI）
- Token 認證方式（僅支援帳號密碼參數認證）
- DClone 狀態監控 / Terror Zone 查詢
- 遊戲時間統計
- 視窗位置/大小記憶
- 每帳號獨立 `settings.json` 切換

---

## 專案架構

```
d2r-multiboxing/
├── [main.go](main.go)                            # 程式進入點，CLI 互動主迴圈
├── [go.mod](go.mod)
├── go.sum
├── accounts.csv                                   # 帳號設定檔（使用者建立/工具自動更新）
│
├── internal/
│   ├── d2r/
│   │   └── constants.go                           # D2R 相關常數定義
│   │
│   ├── account/
│   │   ├── account.go                             # 帳號資料結構與 CSV 讀寫邏輯
│   │   └── crypto.go                              # 密碼加密/解密（Windows DPAPI）
│   │
│   ├── handle/
│   │   ├── winapi.go                              # Windows NT API 底層封裝
│   │   ├── enumerator.go                          # Handle 列舉與名稱搜尋
│   │   └── closer.go                              # 遠端 Handle 關閉操作
│   │
│   └── process/
│       ├── finder.go                              # 進程搜尋（D2R.exe 偵測）
│       ├── launcher.go                            # D2R.exe 啟動（帶帳號參數）
│       └── window.go                              # 視窗標題重命名
│
└── .github/
    └── instructions/
        ├── [golang.instructions.md](.github/instructions/golang.instructions.md)
        └── [project-context.instructions.md](.github/instructions/project-context.instructions.md)
```

---

## 模組詳細設計

### `internal/d2r/constants.go` - D2R 常數定義

定義所有與 D2R 遊戲相關的常數，集中管理避免 magic string。

```go
// 主要常數
const (
    ProcessName             = "D2R.exe"
    SingleInstanceEventName = "DiabloII Check For Other Instances"
    WindowClassName         = "OsWindow"  // D2R 視窗類別名稱
)

// 伺服器區域
type Region struct {
    Name    string  // 顯示名稱（如 "NA", "EU"）
    Address string  // 伺服器地址
}
```

| 區域 | 代號 | 伺服器地址 |
|------|------|-----------|
| 美洲 | `NA` | `us.actual.battle.net` |
| 歐洲 | `EU` | `eu.actual.battle.net` |
| 亞洲 | `Asia` | `kr.actual.battle.net` |

---

### `internal/account/` - 帳號管理模組

#### `account.go` - 資料結構與 CSV 讀寫

```go
type Account struct {
    ID          int    // 帳號序號（1, 2, 3...）
    Email       string // Battle.net 登入信箱
    Password    string // 加密後的密碼字串（CSV 中以 "ENC:" 前綴標記）
    DisplayName string // 顯示名稱（用於視窗標題）
    Region      string // 預設區域（NA/EU/Asia）
}
```

**CSV 格式**（`accounts.csv`）：

```csv
ID,Email,Password,DisplayName,Region
1,account1@email.com,mypassword123,主帳號-法師,NA
2,account2@email.com,anotherpass,副帳號-野蠻人,NA
```

**功能**：
- `LoadAccounts(path string) ([]Account, error)` - 讀取 CSV
- `SaveAccounts(path string, accounts []Account) error` - 寫入 CSV
- 首次執行時偵測明文密碼，自動加密後回寫

#### `crypto.go` - 密碼加密

**方案**：Windows DPAPI（Data Protection API）

```go
// 加密流程
明文密碼 → CryptProtectData() → Base64 編碼 → "ENC:" + base64string

// 解密流程
"ENC:base64string" → 去除前綴 → Base64 解碼 → CryptUnprotectData() → 明文密碼
```

**Windows API 呼叫**：
- `crypt32.dll` → `CryptProtectData` / `CryptUnprotectData`
- DPAPI 自動綁定當前 Windows 使用者帳戶，無需自行管理金鑰
- 換使用者或換電腦後無法解密，需重新輸入密碼

---

### `internal/handle/` - Windows Handle 操作模組

此模組是整個工具的**核心技術**，負責找到並關閉 D2R 的單實例鎖定 handle。

#### `winapi.go` - Windows NT API 封裝

封裝三個未公開的 NT API（透過 `ntdll.dll`）：

| API | 用途 |
|-----|------|
| `NtQuerySystemInformation` | 查詢全系統 handle 資訊 |
| `NtQueryObject` | 查詢單一 handle 的名稱與類型 |
| `NtDuplicateObject` | 複製 handle 到本進程 / 關閉遠端 handle |

**關鍵結構體**：

```go
// 64-bit handle 列舉結構
type SystemHandleTableEntryInfoEx struct {
    Object                uintptr
    UniqueProcessID       uintptr
    HandleValue           uintptr
    GrantedAccess         uint32
    CreatorBackTraceIndex uint16
    ObjectTypeIndex       uint16
    HandleAttributes      uint32
    Reserved              uint32
}

type SystemExtendedHandleInformationEx struct {
    NumberOfHandles uintptr
    Reserved        uintptr
    Handles         [1]SystemHandleTableEntryInfoEx
}
```

**關鍵常數**：

```go
const (
    SystemExtendedHandleInformation = 64      // handle 資訊類別（64-bit）
    ObjectNameInformation           = 1       // 物件名稱查詢
    ObjectTypeInformation           = 2       // 物件類型查詢
    StatusInfoLengthMismatch        = 0xC0000004
    DuplicateCloseSource            = 0x00000001
    DuplicateSameAccess             = 0x00000002
)
```

#### `enumerator.go` - Handle 列舉

```go
// findHandlesByName 在指定進程中搜尋符合名稱的 handle
func findHandlesByName(processID uint32, targetName string) ([]HandleInfo, error)
```

**流程**：
1. `NtQuerySystemInformation` 取得全系統 handle（初始 buffer 1MB，動態擴展）
2. 遍歷所有 handle，篩選 `UniqueProcessID == processID`
3. 對每個候選 handle 用 `NtDuplicateObject` + `DuplicateSameAccess` 複製到本進程
4. 查詢 type，僅對 `"Event"` 類型查詢 name（避免對 pipe 等 handle 查詢時 hang 住）
5. 比對 name 是否包含目標字串（Windows 會加前綴如 `\Sessions\1\BaseNamedObjects\`）

#### `closer.go` - Handle 關閉

```go
// CloseHandlesByName 找到並關閉指定進程中符合名稱的所有 handle
func CloseHandlesByName(processID uint32, handleName string) (int, error)
```

**關閉原理**：使用 `NtDuplicateObject` 配合 `DuplicateCloseSource` flag：
- `SourceProcess` = D2R.exe 的 process handle
- `SourceHandle` = 目標 Event Handle
- `TargetProcess` = 0（NULL，不建立副本）
- `Options` = `DuplicateCloseSource`（關閉來源 handle）

這等效於在 D2R 進程中呼叫 `CloseHandle()`。

---

### `internal/process/` - 進程管理模組

#### `finder.go` - 進程搜尋

```go
// FindProcessesByName 搜尋所有符合名稱的進程
func FindProcessesByName(name string) ([]ProcessInfo, error)

// IsProcessRunning 檢查指定名稱的進程是否正在執行
func IsProcessRunning(name string) (bool, error)
```

**實作方式**：使用 `windows.CreateToolhelp32Snapshot` + `Process32First` / `Process32Next` 遍歷系統進程清單。

#### `launcher.go` - D2R 啟動

```go
// LaunchD2R 使用帳號參數啟動 D2R.exe
func LaunchD2R(d2rPath string, username string, password string, address string) (uint32, error)
```

**啟動指令格式**：
```
D2R.exe -username account@email.com -password secretpass -address us.actual.battle.net
```

使用 `os/exec.Command` 執行，啟動後回傳 PID。

#### `window.go` - 視窗重命名

```go
// RenameWindow 將指定 PID 的 D2R 視窗標題改為指定名稱
func RenameWindow(pid uint32, newTitle string) error
```

**Windows API 呼叫**：
- `user32.dll` → `EnumWindows` - 列舉所有頂層視窗
- `user32.dll` → `GetWindowThreadProcessId` - 取得視窗所屬 PID
- `user32.dll` → `SetWindowTextW` - 設定視窗標題

**流程**：
1. `EnumWindows` 遍歷所有頂層視窗
2. 對每個視窗呼叫 `GetWindowThreadProcessId` 取得 PID
3. 比對 PID 是否為目標 D2R 進程
4. 找到後呼叫 `SetWindowTextW` 修改標題

---

### `main.go` - CLI 主程式

**互動式選單設計**：

```
============================================
  D2R Multiboxing Launcher
============================================

  帳號列表：
  [1] 主帳號-法師    (account1@email.com)  NA  [未啟動]
  [2] 副帳號-野蠻人  (account2@email.com)  NA  [已啟動]

--------------------------------------------
  操作：
  <數字>   啟動指定帳號
  a        啟動所有未啟動的帳號
  r        重新整理狀態
  q        退出
--------------------------------------------
  請選擇：
```

**啟動單一帳號的完整流程**：

```
使用者選擇帳號
    │
    ├── 1. 讀取帳號資訊
    │      └── 解密密碼（DPAPI）
    │
    ├── 2. 選擇區域（若未設預設區域）
    │
    ├── 3. 啟動 D2R.exe（帶參數）
    │      └── 取得 PID
    │
    ├── 4. 等待 D2R 進程初始化（短暫 delay）
    │
    ├── 5. 關閉 Event Handle
    │      └── CloseHandlesByName(pid, "DiabloII Check For Other Instances")
    │
    └── 6. 重命名視窗標題
           └── RenameWindow(pid, displayName)
```

**背景 Handle 監控**：

啟動一個 goroutine 持續每 2 秒掃描一次，偵測新的 D2R.exe 進程並自動關閉 handle，確保任何時候從 Battle.net 手動啟動的實例也能正常多開。

---

## 工作計畫

### Phase 1: 基礎架構與常數定義

- [x] 1.1 清理 [main.go](main.go) 中的 GoLand 範本程式碼
- [x] 1.2 建立 `internal/d2r/constants.go`
  - D2R.exe 進程名、Event Handle 名稱、視窗類別名稱
  - Battle.net 伺服器區域列表（NA/EU/Asia）
- [x] 1.3 更新 [project-context.instructions.md](.github/instructions/project-context.instructions.md) 加入專案背景

### Phase 2: Windows Handle 操作

- [x] 2.1 建立 `internal/handle/winapi.go`
  - `ntdll.dll` 載入與 proc 定義
  - `NtQuerySystemInformation` / `NtQueryObject` / `NtDuplicateObject` 函式封裝
  - `SystemHandleTableEntryInfoEx` 等結構體
  - `UnicodeString` 轉 Go string 輔助函式
- [x] 2.2 建立 `internal/handle/enumerator.go`
  - `findHandlesByName(processID, targetName)` - 列舉並搜尋指定 handle
  - 動態 buffer 擴展機制
  - 僅對 `Event` 類型查詢名稱（安全防護）
- [x] 2.3 建立 `internal/handle/closer.go`
  - `closeRemoteHandle(processID, handle)` - 關閉單一遠端 handle
  - `CloseHandlesByName(processID, handleName)` - 公開 API

### Phase 3: 進程管理

- [x] 3.1 建立 `internal/process/finder.go`
  - `ProcessInfo` struct（PID, Name）
  - `FindProcessesByName(name)` - 使用 `CreateToolhelp32Snapshot`
  - `IsProcessRunning(name)` - 快速檢查
- [x] 3.2 建立 `internal/process/launcher.go`
  - `LaunchD2R(d2rPath, username, password, address)` - 帶參數啟動
  - 回傳 PID 供後續 handle 關閉與視窗重命名使用
- [x] 3.3 建立 `internal/process/window.go`
  - `EnumWindows` callback 實作
  - `RenameWindow(pid, newTitle)` - 找到 PID 對應的視窗並重命名

### Phase 4: 帳號管理

- [x] 4.1 建立 `internal/account/account.go`
  - `Account` struct 定義
  - `LoadAccounts(path)` - CSV 讀取（支援 header row）
  - `SaveAccounts(path, accounts)` - CSV 寫入
  - `IsPasswordEncrypted(password)` - 檢查 `ENC:` 前綴
- [x] 4.2 建立 `internal/account/crypto.go`
  - `EncryptPassword(plaintext)` - DPAPI 加密 + Base64 + 前綴
  - `DecryptPassword(encrypted)` - 前綴移除 + Base64 + DPAPI 解密
  - `crypt32.dll` 的 `CryptProtectData` / `CryptUnprotectData` 封裝
  - `DATA_BLOB` 結構體定義

### Phase 5: CLI 主程式整合

- [x] 5.1 實作 CLI 互動主選單
  - 讀取並顯示帳號列表
  - 顯示各帳號啟動狀態（比對 D2R.exe 進程）
  - 輸入處理（數字選帳號、`a` 全開、`r` 重整、`q` 退出）
  - 區域選擇子選單
- [x] 5.2 實作啟動流程串接
  - 完整流程：讀帳號 → 解密 → 啟動 → delay → 關 handle → 改視窗名
  - 首次執行時自動加密明文密碼並回寫 CSV
  - 錯誤處理與使用者提示
- [x] 5.3 實作背景 Handle 監控（goroutine）
  - 每 2 秒掃描 D2R.exe 進程
  - 記錄已處理的 PID 避免重複操作
  - 自動關閉新出現的 Event Handle

### Phase 6: 驗證與文件

- [x] 6.1 建立測試（9 tests passed）
  - 帳號 CSV 讀寫測試（`account_test.go`）
  - 密碼加密/解密測試（`account_test.go`）
  - 區域查詢測試（`constants_test.go`）
- [x] 6.2 `go build` 驗證編譯成功
- [x] 6.3 撰寫 [README.md](README.md)
  - 功能說明、安裝步驟、使用方式、CSV 格式範例
- [x] 6.4 更新 [project-context.instructions.md](.github/instructions/project-context.instructions.md)

---

## 技術細節

### 密碼加密方案 - Windows DPAPI

| 項目 | 說明 |
|------|------|
| **API** | `crypt32.dll` → `CryptProtectData` / `CryptUnprotectData` |
| **安全性** | 綁定當前 Windows 使用者帳戶，其他使用者或其他電腦無法解密 |
| **金鑰管理** | 由 Windows 自動管理，無需自行產生或儲存金鑰 |
| **CSV 標記** | 加密後的密碼以 `ENC:` 前綴標記，工具可自動辨識已加密/未加密 |
| **編碼** | 加密後的 bytes 以 Base64 編碼存入 CSV |

**`DATA_BLOB` 結構體**（DPAPI 使用）：

```go
type DataBlob struct {
    Size uint32
    Data *byte
}
```

### D2R 啟動參數

```
D2R.exe -username <email> -password <password> -address <server>
```

| 參數 | 說明 | 範例 |
|------|------|------|
| `-username` | Battle.net 帳號信箱 | `player@email.com` |
| `-password` | Battle.net 帳號密碼 | `mypassword` |
| `-address` | 遊戲伺服器地址 | `us.actual.battle.net` |

### 視窗重命名技術

D2R 的視窗類別名稱為 `OsWindow`，預設標題為 `Diablo II: Resurrected`。

需要處理的時序問題：D2R 啟動後視窗不會立即建立，需要等待數秒後再嘗試重命名，可使用重試機制（最多重試 N 次，每次間隔 1-2 秒）。

### 管理員權限需求

操作其他進程的 handle 需要管理員權限。可透過以下方式處理：

1. **啟動時檢查**：偵測是否以管理員身份執行，若否則提示使用者
2. **Windows Manifest**（未來可選）：嵌入 manifest 要求 UAC 提權

---

## 依賴管理

### 外部依賴

| 套件 | 用途 | 備註 |
|------|------|------|
| `golang.org/x/sys/windows` | Windows API 封裝 | 核心依賴，操作進程/handle/視窗 |
| `github.com/stretchr/testify` | 測試斷言 | 依專案慣例使用 |

### 標準庫使用

| 套件 | 用途 |
|------|------|
| `encoding/csv` | CSV 讀寫 |
| `encoding/base64` | 加密密碼的 Base64 編碼 |
| `os/exec` | 啟動 D2R.exe 子進程 |
| `fmt` | 格式化輸出 |
| `bufio` / `os` | CLI 輸入讀取 |
| `strings` | 字串處理 |
| `sync` | 背景 goroutine 同步 |
| `time` | 定時掃描與延遲 |
| `syscall` | 底層 syscall 操作 |
| `unsafe` | NT API 結構體指標操作 |

---

## 注意事項

### 安全性

- ⚠️ 此工具需要以**管理員權限**執行，因為需要讀取/關閉其他進程的 handle
- ⚠️ 部分防毒軟體可能會誤報（操作進程 handle 是類似惡意程式的行為），需要加入例外

### 法律聲明

- ⚠️ 使用此工具可能違反 Blizzard 的服務條款
- ⚠️ 帳號被封禁的風險由使用者自行承擔
- 此工具**不會**修改遊戲檔案、注入程式碼或自動化遊戲操作

### 技術限制

- 僅支援 Windows 平台（使用 Windows 專屬 API）
- 僅支援 Battle.net 帳號（不支援 Steam 版本）
- 密碼加密綁定當前 Windows 使用者，換機器需重新設定

---

## 技術參考

| 資源 | 說明 |
|------|------|
| [chenwei791129/multiablo](https://github.com/chenwei791129/multiablo) | Go 語言 D2R 多開工具，handle 關閉技術參考 |
| [shupershuff/Diablo2RLoader](https://github.com/shupershuff/Diablo2RLoader) | PowerShell D2R 啟動器，完整功能參考 |
| [shupershuff/D2r-Multiboxing-Methods](https://github.com/shupershuff/D2r-Multiboxing-Methods) | D2R 多開方法總覽 |
| [Microsoft DPAPI 文件](https://learn.microsoft.com/en-us/windows/win32/api/dpapi/) | 密碼加密 API 參考 |
| [NtQuerySystemInformation](https://learn.microsoft.com/en-us/windows/win32/api/winternl/nf-winternl-ntquerysysteminformation) | Handle 列舉 API 參考 |

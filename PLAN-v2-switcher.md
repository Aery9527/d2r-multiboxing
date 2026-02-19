# D2R Multiboxing — Phase 2：視窗切換功能（Window Switcher）

> **前置**：[Phase 1 實作計畫](PLAN-v1-multiboxing.md)（多開啟動器，已完成）
> **目的**：在多個 D2R 視窗之間快速切換焦點，讓搖桿/鍵盤輸入導向正確的視窗

---

## 目錄

- [問題描述](#問題描述)
- [技術限制與解法](#技術限制與解法)
- [功能設計](#功能設計)
- [Config 結構](#config-結構)
- [CLI 設定引導流程](#cli-設定引導流程)
- [模組設計](#模組設計)
- [工作計畫](#工作計畫)
- [技術細節](#技術細節)
- [依賴與 API](#依賴與-api)
- [注意事項](#注意事項)

---

## 問題描述

D2R 多開後，使用者希望能「同時」操控兩個視窗。但 Windows 的 XInput（搖桿）與鍵盤輸入都只會導向**前景視窗**（foreground window），無法原生將不同裝置的輸入隔離到不同視窗。

---

## 技術限制與解法

| 限制 | 說明 |
|------|------|
| XInput 只送前景視窗 | D2R 使用 XInput 讀取搖桿，背景視窗不處理搖桿事件 |
| Windows 無裝置隔離 | 無法原生綁定「搖桿 A → 視窗 A、搖桿 B → 視窗 B」 |
| DLL 注入風險高 | Hook `GetForegroundWindow` 可行但會觸發反作弊 |

**解法**：不嘗試隔離輸入，而是透過**快速切換視窗焦點**解決。
使用者按下指定的鍵盤快捷鍵或滑鼠側鍵，即可毫秒等級切換到下一個 D2R 視窗。

---

## 功能設計

### 觸發方式

透過鍵盤快捷鍵（含修飾鍵）或滑鼠側鍵（XButton1 / XButton2）觸發視窗切換。

### 切換邏輯

```
觸發事件發生
    │
    ├── 列舉所有標題為 "D2R-*" 的可見視窗
    │
    ├── 取得目前前景視窗 (GetForegroundWindow)
    │
    ├── 找到下一個 D2R 視窗（循環順序）
    │
    └── SetForegroundWindow(下一個視窗)
        → 搖桿 / 鍵盤輸入自動導向新前景視窗
```

### 設定方式

**不要求使用者手動編輯 config.json**，而是透過 CLI 互動引導。
因為滑鼠側鍵名稱一般使用者不清楚如何指名，所以採用「偵測按鍵」的方式設定。

```
  > 請選擇：s

  === 視窗切換設定 ===

  [1] 設定切換按鍵
  [0] 關閉切換功能
```

設定完成後寫入 `config.json`，後續啟動自動載入。

---

## Config 結構

### config.json 範例

```json
{
  "d2r_path": "C:\\Program Files (x86)\\Diablo II Resurrected\\D2R.exe",
  "launch_delay": 5,
  "switcher": {
    "enabled": true,
    "modifiers": ["ctrl"],
    "key": "Tab"
  }
}
```

### 欄位說明

| 欄位 | 類型 | 說明 |
|------|------|------|
| `switcher.enabled` | `bool` | 是否啟用視窗切換功能 |
| `switcher.modifiers` | `[]string` | 修飾鍵列表：`"ctrl"`, `"alt"`, `"shift"`（滑鼠側鍵時為空） |
| `switcher.key` | `string` | 按鍵名稱（如 `"Tab"`, `"F1"`, `"XButton1"`, `"XButton2"`） |

---

## CLI 設定引導流程

```
  === 視窗切換設定 ===

  請按下想用來切換視窗的按鍵組合...
  （支援：鍵盤任意鍵 + Ctrl/Alt/Shift、滑鼠側鍵）

  偵測到：Ctrl + Tab
  確認使用此組合？(y/n)：y

  ✔ 已儲存切換設定：Ctrl+Tab
```

**偵測方式**：
- 鍵盤：使用 `WH_KEYBOARD_LL` low-level hook 偵測按鍵 + 修飾鍵狀態
- 滑鼠側鍵：使用 `WH_MOUSE_LL` low-level hook 偵測 `WM_XBUTTONDOWN`
- 偵測到後立即解除 hook，僅用於設定階段

---

## 模組設計

### 新增檔案結構

```
internal/
├── switcher/
│   ├── switcher.go       # Start/Stop、視窗切換核心邏輯
│   ├── hotkey.go         # RegisterHotKey 鍵盤快捷鍵監聽
│   ├── mousehook.go      # WH_MOUSE_LL 滑鼠側鍵監聽
│   └── detect.go         # CLI 設定用：偵測按鍵/滑鼠輸入
```

### `switcher.go` — 核心邏輯

```go
// Start 根據 config 啟動對應的監聽方式
func Start(cfg *config.SwitcherConfig) error

// Stop 停止監聽並釋放資源
func Stop() error

// switchToNext 列舉 D2R 視窗並切換到下一個
func switchToNext()
```

### `hotkey.go` — 鍵盤快捷鍵

```go
// startHotkey 使用 RegisterHotKey 註冊全域快捷鍵
// 在獨立 goroutine 中跑 Windows message loop 監聽 WM_HOTKEY
func startHotkey(modifiers uint32, vk uint32, onTrigger func()) error
```

**實作要點**：
- `RegisterHotKey` 註冊全域快捷鍵
- 獨立 goroutine 跑 `GetMessageW` loop 等待 `WM_HOTKEY`
- 收到事件 → 呼叫 `onTrigger`（即 `switchToNext`）

### `mousehook.go` — 滑鼠側鍵

```go
// startMouseHook 使用 WH_MOUSE_LL hook 監聽滑鼠側鍵
func startMouseHook(targetButton uint16, onTrigger func()) error
```

**實作要點**：
- `SetWindowsHookExW(WH_MOUSE_LL)` 安裝 hook
- Callback 中解析 `MSLLHOOKSTRUCT`，比對 `WM_XBUTTONDOWN` + button index
- 匹配時呼叫 `onTrigger`

### `detect.go` — 按鍵偵測（CLI 設定用）

```go
// DetectKeyPress 等待使用者按下按鍵組合（鍵盤或滑鼠側鍵）
// 回傳 modifiers 列表與 key 名稱
func DetectKeyPress() (modifiers []string, key string, err error)
```

**實作要點**：
- 同時安裝 `WH_KEYBOARD_LL` + `WH_MOUSE_LL`
- 偵測到第一個非修飾鍵的按鍵（或滑鼠側鍵）即回傳
- 修飾鍵（Ctrl/Alt/Shift）透過 `GetAsyncKeyState` 讀取狀態

### `internal/process/window.go` — 新增函式

```go
// FindD2RWindows 列舉所有標題以 "D2R-" 開頭的可見視窗
func FindD2RWindows() []windows.Handle

// SwitchToNextD2RWindow 切換到下一個 D2R 視窗（循環）
func SwitchToNextD2RWindow() error
```

---

## 工作計畫

### Phase 2-1：Config 擴充

- [ ] 在 [config.go](internal/config/config.go) 新增 `SwitcherConfig` struct
- [ ] 欄位：`Enabled bool`, `Modifiers []string`, `Key string`
- [ ] 更新 `DefaultConfig()`（switcher 預設 disabled）
- [ ] 確保向下相容（舊 config.json 無 switcher 欄位時不報錯）

### Phase 2-2：視窗切換核心邏輯

- [ ] 在 [window.go](internal/process/window.go) 新增 `FindD2RWindows()`
  - `EnumWindows` + `GetWindowTextW` 篩選 `D2R-` 前綴視窗
- [ ] 新增 `SwitchToNextD2RWindow()`
  - `GetForegroundWindow()` 取得前景
  - 在 D2R 視窗列表中找下一個
  - `SetForegroundWindow()` 切換
  - 處理 `SetForegroundWindow` 限制（`AllowSetForegroundWindow` / `AttachThreadInput`）

### Phase 2-3：鍵盤快捷鍵觸發器

- [ ] 建立 `internal/switcher/hotkey.go`
  - `RegisterHotKey` + `GetMessage` loop（獨立 goroutine）
  - Virtual key code 映射表（key name ↔ VK code）
  - Modifier 映射（ctrl/alt/shift → `MOD_CONTROL`/`MOD_ALT`/`MOD_SHIFT`）
- [ ] `startHotkey()` / stop 生命週期管理

### Phase 2-4：滑鼠側鍵觸發器

- [ ] 建立 `internal/switcher/mousehook.go`
  - `SetWindowsHookExW(WH_MOUSE_LL)` + callback
  - 偵測 `WM_XBUTTONDOWN`（XButton1 / XButton2）
- [ ] `startMouseHook()` / stop 生命週期管理

### Phase 2-5：CLI 設定引導

- [ ] 建立 `internal/switcher/detect.go`
  - `DetectKeyPress()`：同時裝 keyboard + mouse hook 偵測一次按鍵
- [ ] 在 [main.go](cmd/d2r-multiboxing/main.go) CLI 選單新增 `s` 選項
  - 顯示目前設定
  - 「請按下切換按鍵...」→ 偵測 → 確認 → 寫入 config
  - 「關閉切換功能」→ 設 enabled=false → 寫入 config

### Phase 2-6：主程式整合

- [ ] 啟動時讀取 `switcher` config
- [ ] 若 enabled，根據 key 類型啟動 hotkey 或 mousehook
- [ ] CLI 選單顯示當前切換設定狀態
- [ ] 退出時清理資源（`UnregisterHotKey` / `UnhookWindowsHookEx`）

### Phase 2-7：文件更新

- [ ] 更新 [USAGE.md](USAGE.md) 新增「視窗切換功能」章節
- [ ] 更新 [project-context.instructions.md](.github/instructions/project-context.instructions.md)

---

## 技術細節

### `SetForegroundWindow` 限制

Windows 對 `SetForegroundWindow` 有限制 — 只有在以下情況才能成功：
- 呼叫的程式是前景程式
- 呼叫的程式在最近收到過使用者輸入

**解法**（按優先順序嘗試）：
1. `AllowSetForegroundWindow(GetCurrentProcessId())` 預先授權
2. `AttachThreadInput` 將本程式的 thread 附加到前景視窗的 thread
3. 模擬一次 Alt 鍵按下（`keybd_event`）讓系統允許切換

### Virtual Key Code 映射

常用 key name 對應 Windows VK code：

| Key Name | VK Code | 備註 |
|----------|---------|------|
| `Tab` | `0x09` | |
| `F1`~`F12` | `0x70`~`0x7B` | |
| `XButton1` | — | 滑鼠側鍵（後），透過 mouse hook 偵測 |
| `XButton2` | — | 滑鼠側鍵（前），透過 mouse hook 偵測 |
| `A`~`Z` | `0x41`~`0x5A` | |
| `` ` `` | `0xC0` | 反引號/波浪號 |

### 滑鼠側鍵 vs 鍵盤的判斷

設定時若偵測到的 key 為 `XButton1` 或 `XButton2`，執行時走 mouse hook 路徑；否則走 `RegisterHotKey` 路徑。`config.json` 中不需額外欄位區分，由 key 名稱自動判斷。

---

## 依賴與 API

### 新增 Windows API

| DLL | API | 用途 |
|-----|-----|------|
| `user32.dll` | `RegisterHotKey` / `UnregisterHotKey` | 全域快捷鍵 |
| `user32.dll` | `GetForegroundWindow` | 取得前景視窗 |
| `user32.dll` | `SetForegroundWindow` | 切換前景視窗 |
| `user32.dll` | `SetWindowsHookExW` / `UnhookWindowsHookEx` | Low-level mouse hook |
| `user32.dll` | `GetMessageW` / `CallNextHookEx` | Windows message loop |
| `user32.dll` | `AllowSetForegroundWindow` | 授權前景切換 |
| `user32.dll` | `GetAsyncKeyState` | 偵測修飾鍵狀態 |

### Go 依賴

不新增外部依賴，所有 Windows API 透過 `golang.org/x/sys/windows` + `syscall` 直接呼叫。

---

## 注意事項

- ⚠️ `RegisterHotKey` 可能與其他程式衝突，註冊失敗時需提示使用者換組合鍵
- ⚠️ `SetForegroundWindow` 在某些 Windows 版本有額外限制，需實作多重 fallback
- ⚠️ Low-level hook callback 必須在安裝 hook 的同一 thread 上跑 message loop
- ℹ️ 所有功能皆使用標準 Win32 API，不修改遊戲、不注入 DLL
- ℹ️ CLI 設定引導只需操作一次，設定存入 `config.json` 後自動載入

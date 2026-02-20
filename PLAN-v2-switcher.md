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
使用者按下指定的鍵盤快捷鍵、滑鼠側鍵或搖桿按鈕，即可毫秒等級切換到下一個 D2R 視窗。

---

## 功能設計

### 觸發方式

透過鍵盤快捷鍵（含修飾鍵）、滑鼠側鍵（XButton1 / XButton2）或搖桿按鈕（XInput，任意 controller 的任意按鈕/扳機）觸發視窗切換。

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

**鍵盤快捷鍵**：

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

**搖桿按鈕**：

```json
{
  "d2r_path": "C:\\Program Files (x86)\\Diablo II Resurrected\\D2R.exe",
  "launch_delay": 5,
  "switcher": {
    "enabled": true,
    "key": "Gamepad_A",
    "gamepad_index": 1
  }
}
```

**搖桿組合鍵**（按住 LT，按 A，放開 A）：

```json
{
  "d2r_path": "C:\\Program Files (x86)\\Diablo II Resurrected\\D2R.exe",
  "launch_delay": 5,
  "switcher": {
    "enabled": true,
    "modifiers": ["Gamepad_LT"],
    "key": "Gamepad_A",
    "gamepad_index": 0
  }
}
```

### 欄位說明

| 欄位 | 類型 | 說明 |
|------|------|------|
| `switcher.enabled` | `bool` | 是否啟用視窗切換功能 |
| `switcher.modifiers` | `[]string` | 修飾鍵：鍵盤用 `"ctrl"`, `"alt"`, `"shift"`；搖桿用 `"Gamepad_LT"`, `"Gamepad_Back"` 等 |
| `switcher.key` | `string` | 觸發鍵名稱（如 `"Tab"`, `"F1"`, `"XButton1"`, `"Gamepad_A"`, `"Gamepad_LT"`） |
| `switcher.gamepad_index` | `int` | XInput 搖桿編號（0-3），僅搖桿觸發時使用 |

---

## CLI 設定引導流程

```
  === 視窗切換設定 ===

  請按下想用來切換視窗的按鍵組合...
  （支援：鍵盤任意鍵 + Ctrl/Alt/Shift、滑鼠側鍵、搖桿按鈕）

  偵測到：Ctrl+Tab（Tab 鍵）
  確認使用此組合？(Y/n)：

  ✔ 已儲存切換設定：Ctrl+Tab（Tab 鍵）
```

**搖桿偵測範例**（單鍵）：

```
  偵測到：搖桿 #1 A 按鈕
  確認使用此組合？(Y/n)：

  ✔ 已儲存切換設定：搖桿 #1 A 按鈕
```

**搖桿組合鍵範例**（按住 LT，按 A，放開 A）：

```
  偵測到：搖桿 #1 LT（左扳機）+A 按鈕
  確認使用此組合？(Y/n)：

  ✔ 已儲存切換設定：搖桿 #1 LT（左扳機）+A 按鈕
```

**偵測方式**：
- 鍵盤：使用 `WH_KEYBOARD_LL` low-level hook 偵測按鍵 + 修飾鍵狀態
- 滑鼠側鍵：使用 `WH_MOUSE_LL` low-level hook 偵測 `WM_XBUTTONDOWN`
- 搖桿：使用 `XInputGetState` 輪詢所有已連接的 XInput controller（每 10ms），**偵測按鍵放開（falling edge）**，放開的按鍵為觸發鍵，仍按住的為修飾鍵
- 三種偵測透過 `sync.Once` 協調，最先觸發的輸入會被採用
- 偵測完成後立即解除 hook / 停止輪詢，僅用於設定階段

---

## 模組設計

### 新增檔案結構

```
internal/
├── switcher/
│   ├── switcher.go       # Start/Stop、視窗切換核心邏輯
│   ├── hotkey.go         # RegisterHotKey 鍵盤快捷鍵監聽
│   ├── mousehook.go      # WH_MOUSE_LL 滑鼠側鍵監聽
│   ├── gamepad.go        # XInput 搖桿偵測與輪詢
│   ├── detect.go         # CLI 設定用：偵測按鍵/滑鼠/搖桿輸入
│   └── keymap.go         # VK code 映射表與輔助函式
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
// DetectKeyPress 等待使用者按下按鍵組合（鍵盤、滑鼠側鍵或搖桿按鈕）
// 回傳 modifiers 列表、key 名稱及 gamepad controller index
func DetectKeyPress() (modifiers []string, key string, gamepadIndex int, err error)
```

**實作要點**：
- 同時安裝 `WH_KEYBOARD_LL` + `WH_MOUSE_LL` 並啟動 XInput 輪詢
- 使用 `sync.Once` 確保三路偵測中只有一路能送出結果
- 偵測到第一個非修飾鍵的按鍵（或滑鼠側鍵、搖桿按鈕）即回傳
- 修飾鍵（Ctrl/Alt/Shift）透過 `GetAsyncKeyState` 讀取狀態

### `gamepad.go` — XInput 搖桿支援

```go
// XInputAvailable 檢查 XInput DLL 是否可載入
func XInputAvailable() bool

// detectGamepadButtonPress 輪詢所有 XInput controller，偵測按鈕放開事件（falling edge）
// 放開的按鈕為觸發鍵，仍按住的其他按鈕為修飾鍵
// 支援組合鍵：按住 LT，按 A，放開 A → 回傳 (idx, ["Gamepad_LT"], "Gamepad_A")
func detectGamepadButtonPress(stop <-chan struct{}) (controllerIndex int, modifiers []string, buttonName string)

// startGamepadPoll 輪詢指定 controller 的按鈕，所有修飾鍵按住時邊緣觸發呼叫 onTrigger
func startGamepadPoll(controllerIndex int, modifierKeys []string, key string, onTrigger func()) error
```

**實作要點**：
- 使用 `xinput1_4.dll` 的 `XInputGetState` API（Windows 10+ 內建）
- 支援 4 個 controller（index 0-3）、14 個按鈕 + LT/RT 扳機
- 邊緣觸發偵測：記錄前一次狀態，只在 not-pressed → pressed 時觸發
- 扳機閾值：`LeftTrigger` / `RightTrigger` >= 128 視為按下
- 輪詢頻率 10ms（100Hz），使用 `time.Ticker`
- 若 XInput 不可用（DLL 載入失敗），偵測階段靜默等待 stop，啟動階段回傳錯誤

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

- [x] 在 [config.go](internal/config/config.go) 新增 `SwitcherConfig` struct
- [x] 欄位：`Enabled bool`, `Modifiers []string`, `Key string`
- [x] 更新 `DefaultConfig()`（switcher 預設 disabled）
- [x] 確保向下相容（舊 config.json 無 switcher 欄位時不報錯）

### Phase 2-2：視窗切換核心邏輯

- [x] 在 [window.go](internal/process/window.go) 新增 `FindD2RWindows()`
  - `EnumWindows` + `GetWindowTextW` 篩選 `D2R-` 前綴視窗
- [x] 新增 `SwitchToNextD2RWindow()`
  - `GetForegroundWindow()` 取得前景
  - 在 D2R 視窗列表中找下一個
  - `SetForegroundWindow()` 切換
  - 處理 `SetForegroundWindow` 限制（`AllowSetForegroundWindow` / `AttachThreadInput`）

### Phase 2-3：鍵盤快捷鍵觸發器

- [x] 建立 `internal/switcher/hotkey.go`
  - `RegisterHotKey` + `GetMessage` loop（獨立 goroutine）
  - Virtual key code 映射表（key name ↔ VK code）
  - Modifier 映射（ctrl/alt/shift → `MOD_CONTROL`/`MOD_ALT`/`MOD_SHIFT`）
- [x] `startHotkey()` / stop 生命週期管理

### Phase 2-4：滑鼠側鍵觸發器

- [x] 建立 `internal/switcher/mousehook.go`
  - `SetWindowsHookExW(WH_MOUSE_LL)` + callback
  - 偵測 `WM_XBUTTONDOWN`（XButton1 / XButton2）
- [x] `startMouseHook()` / stop 生命週期管理

### Phase 2-5：CLI 設定引導

- [x] 建立 `internal/switcher/detect.go`
  - `DetectKeyPress()`：同時裝 keyboard + mouse hook + XInput 輪詢偵測一次按鍵
- [x] 在 [main.go](cmd/d2r-multiboxing/main.go) CLI 選單新增 `s` 選項
  - 顯示目前設定
  - 「請按下切換按鍵...」→ 偵測 → 確認 → 寫入 config
  - 「關閉切換功能」→ 設 enabled=false → 寫入 config

### Phase 2-6：主程式整合

- [x] 啟動時讀取 `switcher` config
- [x] 若 enabled，根據 key 類型啟動 hotkey 或 mousehook
- [x] CLI 選單顯示當前切換設定狀態
- [x] 退出時清理資源（`UnregisterHotKey` / `UnhookWindowsHookEx`）

### Phase 2-7：文件更新

- [x] 更新 [USAGE.md](USAGE.md) 新增「視窗切換功能」章節
- [x] 更新 [project-context.instructions.md](.github/instructions/project-context.instructions.md)

### Phase 2-8：搖桿（XInput）支援

- [x] 建立 `internal/switcher/gamepad.go`
  - XInput API 封裝（`xinput1_4.dll` → `XInputGetState`）
  - `detectGamepadButtonPress()`：輪詢所有 controller，邊緣觸發偵測
  - `startGamepadPoll()`：運行時持續輪詢指定 controller + 按鈕
  - 支援 14 個按鈕 + LT/RT 扳機（閾值 128）
- [x] `DetectKeyPress()` 新增搖桿偵測（三路 `sync.Once` 協調）
- [x] `SwitcherConfig` 新增 `GamepadIndex int` 欄位
- [x] `Start()` 路由：`Gamepad_*` key → `startGamepadPoll`
- [x] `keymap.go` 新增 `IsGamepadButton`、`FormatSwitcherDisplay`、搖桿按鈕中文顯示名稱
- [x] CLI 引導文字更新、`main.go` 全面改用 `FormatSwitcherDisplay`
- [x] 更新 README.md / USAGE.md / PLAN-v2-switcher.md / project-context

### Phase 2-9：搖桿組合鍵與偵測修正

- [x] `detectGamepadButtonPress()` 改為**放開觸發**（falling edge）
  - 原本：按下時立即回傳，導致 LT 按住再按 RT 時只偵測到 LT
  - 修正後：放開按鍵時才回傳，放開的按鈕為觸發鍵，仍按住的為修飾鍵
  - 流程：按住 LT → 按 RT → 放開 RT → 偵測到 `mods=[LT], trigger=RT`
- [x] `startGamepadPoll()` 新增 `modifierKeys []string` 參數
  - 輪詢時驗證所有修飾鍵都按住才邊緣觸發主鍵
  - 修飾鍵支援所有 `Gamepad_*` 名稱（含 LT/RT 扳機）
- [x] `captureGamepadModifiers()` 在放開觸發時，從當前狀態讀取仍按住的按鈕作為修飾鍵列表
- [x] `isGamepadModifierHeld()` 執行時按名稱檢查單一修飾鍵是否按住（LT/RT 用閾值判斷）
- [x] `SwitcherConfig.Modifiers` 同時支援鍵盤修飾鍵（`"ctrl"`/`"alt"`/`"shift"`）與搖桿修飾鍵（`"Gamepad_LT"` 等），由 `IsGamepadButton()` 判斷路由
- [x] `FormatSwitcherDisplay()` 顯示搖桿組合鍵格式，如「搖桿 #1 LT（左扳機）+A 按鈕」
- [x] CLI 設定引導提示文字更新，說明需放開按鍵才完成偵測
- [x] 更新 [USAGE.md](USAGE.md)、[PLAN-v2-switcher.md](PLAN-v2-switcher.md)

---

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

### 滑鼠側鍵 vs 鍵盤 vs 搖桿的判斷

設定時若偵測到的 key 為 `XButton1` 或 `XButton2`，執行時走 mouse hook 路徑；若為 `Gamepad_*`，走 XInput 輪詢路徑；否則走 `RegisterHotKey` 路徑。`config.json` 中由 key 名稱自動判斷，不需額外欄位區分。

### XInput 搖桿支援

| 項目 | 說明 |
|------|------|
| DLL | `xinput1_4.dll`（Windows 10+ 內建） |
| API | `XInputGetState(dwUserIndex, *XINPUT_STATE)` |
| Controller 數量 | 最多 4 個（index 0-3） |
| 支援按鈕 | A/B/X/Y、LB/RB、Back/Start、LS/RS（搖桿按下）、十字鍵 |
| 扳機支援 | LT/RT（`LeftTrigger` / `RightTrigger` >= 128 視為按下） |

**偵測機制**：
- **設定階段**（`detectGamepadButtonPress`）：同時輪詢所有 4 個 controller，**放開按鍵時觸發（falling edge）**
  - 放開的按鈕 = 觸發鍵，仍按住的按鈕 = 修飾鍵
  - 支援組合鍵設定：按住 LT 後按 A 再放開 A → 偵測到 `LT+A`
  - 先讀取初始狀態作為基準，避免誤觸已按住的按鈕
- **運行時**（`startGamepadPoll`）：只輪詢 config 中指定的 `gamepad_index`，**按下時觸發（rising edge）**
  - 先確認所有 `modifierKeys` 都按住，才允許主鍵邊緣觸發
- **輪詢頻率**：10ms（100Hz），使用 `time.Ticker`
- **斷線處理**：`XInputGetState` 回傳非 0 時視為斷線，重新連接後以當前狀態為基準

**搖桿按鈕命名（config key 欄位）**：

| Key 名稱 | 按鈕 | XInput Mask |
|-----------|------|-------------|
| `Gamepad_A` | A | `0x1000` |
| `Gamepad_B` | B | `0x2000` |
| `Gamepad_X` | X | `0x4000` |
| `Gamepad_Y` | Y | `0x8000` |
| `Gamepad_LB` | 左肩鍵 | `0x0100` |
| `Gamepad_RB` | 右肩鍵 | `0x0200` |
| `Gamepad_LT` | 左扳機 | analog >= 128 |
| `Gamepad_RT` | 右扳機 | analog >= 128 |
| `Gamepad_Back` | Back | `0x0020` |
| `Gamepad_Start` | Start | `0x0010` |
| `Gamepad_LS` | 左搖桿按下 | `0x0040` |
| `Gamepad_RS` | 右搖桿按下 | `0x0080` |
| `Gamepad_DPadUp` | 十字鍵 ↑ | `0x0001` |
| `Gamepad_DPadDown` | 十字鍵 ↓ | `0x0002` |
| `Gamepad_DPadLeft` | 十字鍵 ← | `0x0004` |
| `Gamepad_DPadRight` | 十字鍵 → | `0x0008` |

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
| `xinput1_4.dll` | `XInputGetState` | 讀取搖桿狀態（Windows 10+ 內建） |

### Go 依賴

不新增外部依賴，所有 Windows API 透過 `golang.org/x/sys/windows` + `syscall` 直接呼叫。

---

## 注意事項

- ⚠️ `RegisterHotKey` 可能與其他程式衝突，註冊失敗時需提示使用者換組合鍵
- ⚠️ `SetForegroundWindow` 在某些 Windows 版本有額外限制，需實作多重 fallback
- ⚠️ Low-level hook callback 必須在安裝 hook 的同一 thread 上跑 message loop
- ⚠️ XInput 搖桿偵測使用輪詢（polling），CPU 開銷極低（10ms ticker）但非事件驅動
- ℹ️ 所有功能皆使用標準 Win32 API，不修改遊戲、不注入 DLL
- ℹ️ CLI 設定引導只需操作一次，設定存入 `config.json` 後自動載入
- ℹ️ XInput 搖桿最多支援 4 個（Windows 限制），且僅支援 XInput 相容控制器（Xbox 系列）

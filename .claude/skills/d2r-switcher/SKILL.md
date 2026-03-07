---
name: d2r-switcher
description: "Handle repository-specific D2R window-switcher work in d2r-hyper-launcher. Use this whenever the user wants to change hotkey switching, mouse side-button switching, gamepad switching, trigger detection, switcher config persistence, focus-cycling behavior, XInput handling, or switcher startup/shutdown behavior, even if they only mention 'switcher', '快捷鍵切窗', '滑鼠側鍵切換', '搖桿切換', or config.switcher."
---

# D2R Switcher Context

這個 skill 專注在 D2R 視窗切換功能。處理相關任務時，優先圍繞 `internal/switcher`、`config.switcher`、`D2R-` 視窗標題與主選單設定流程，不要把多開 handle 邏輯和 switcher 設定混為一談。

## 先看哪些檔案

- [cmd/d2r-hyper-launcher/main.go](../../../cmd/d2r-hyper-launcher/main.go) - 啟動 switcher、`setupSwitcher()` 設定流程
- [internal/switcher/switcher.go](../../../internal/switcher/switcher.go) - `Start()` / `Stop()` / `IsRunning()` 與視窗切換主邏輯
- [internal/switcher/detect.go](../../../internal/switcher/detect.go) - CLI 互動式按鍵偵測
- [internal/switcher/hotkey.go](../../../internal/switcher/hotkey.go) - 鍵盤快捷鍵
- [internal/switcher/mousehook.go](../../../internal/switcher/mousehook.go) - 滑鼠側鍵
- [internal/switcher/gamepad.go](../../../internal/switcher/gamepad.go) - XInput 偵測與輪詢
- [internal/switcher/keymap.go](../../../internal/switcher/keymap.go) - VK / 顯示名稱 / modifier 對應
- [internal/process/window.go](../../../internal/process/window.go) - 枚舉 `D2R-` 視窗、切換前景視窗
- [internal/config/config.go](../../../internal/config/config.go) - `SwitcherConfig`
- [docs/switcher-usage-guide.md](../../../docs/switcher-usage-guide.md) - 使用者可見設定流程

## 核心事實

1. switcher 只會切換標題以 `D2R-` 為前綴的視窗。
2. `switcher.Start()` 會依 `key` 類型自動分流到 hotkey、mouse hook 或 gamepad poll。
3. 鍵盤支援 `ctrl` / `alt` / `shift` 修飾鍵；搖桿修飾鍵使用 `Gamepad_*` 名稱。
4. `DetectKeyPress()` 會同時監聽鍵盤、滑鼠與搖桿，第一個成功事件獲勝。
5. 搖桿偵測採「先按住修飾鍵，再按觸發鍵，最後放開觸發鍵」的模式，這個 UX 要跟文件一致。

## 修改時要守住的規則

- 若調整視窗切換條件，確認 [internal/d2r/constants.go](../../../internal/d2r/constants.go) 的 `WindowTitlePrefix` 與多開重命名邏輯仍一致。
- `Start()` / `Stop()` 的全域狀態由 mutex 保護，不要繞過這層直接改 `running` 或 `stopFunc`。
- `config.Switcher` 的 JSON 欄位名稱要保持相容：`enabled`、`modifiers`、`key`、`gamepad_index`。
- 滑鼠與鍵盤 hook 需要 message loop；gamepad 走 XInput polling。修改時不要把這三種觸發混成同一路徑。
- 子選單導航一樣要維持 `b` / `h` / `q`。

## 常見任務做法

### 新增或調整可觸發按鍵

1. 先改 [internal/switcher/keymap.go](../../../internal/switcher/keymap.go)
2. 再確認 [internal/switcher/detect.go](../../../internal/switcher/detect.go) 與顯示格式是否需要同步更新
3. 若會出現在 CLI 或文件，也要更新 [docs/switcher-usage-guide.md](../../../docs/switcher-usage-guide.md) 與 [README.md](../../../README.md)

### 修正切換失效或焦點切不過去

1. 看 [internal/switcher/switcher.go](../../../internal/switcher/switcher.go) 的 `switchToNext()`
2. 再看 [internal/process/window.go](../../../internal/process/window.go) 的 `FindWindowsByTitlePrefix()` 與 `SwitchToWindow()`
3. 確認多開流程是否仍正確把 D2R 視窗重命名為 `D2R-<DisplayName>`

### 修正搖桿問題

1. 看 [internal/switcher/gamepad.go](../../../internal/switcher/gamepad.go)
2. 區分「設定時偵測」與「執行時觸發」兩條路徑
3. 保留 XInput 不可用時的明確錯誤或安全退化

## 驗證

至少跑：

```powershell
go test ./...
go build ./cmd/d2r-hyper-launcher
```

若改到 switcher：

1. 設定鍵盤快捷鍵
2. 設定滑鼠側鍵
3. 若有搖桿，測試單鍵與組合鍵
4. 啟動至少兩個 D2R 視窗後，確認焦點會循環切換

---
name: d2r-cli-ui
description: "Handle repository-specific CLI UI and message-presentation work in d2r-hyper-launcher. Use this whenever the user wants to change visible CLI wording, headers, startup announcements, menu layout, option alignment, prompt/error/success formatting, submenu navigation presentation, grouped multiline messages, or any `cmd/d2r-hyper-launcher` display flow, even if they only mention 'UI', 'menu', 'header', '公告', '訊息顯示', '排版', or 'CLI 看起來怪怪的'."
---

# D2R CLI UI Context

這個 skill 專注在 `d2r-hyper-launcher` 的 CLI UI layer 與玩家可見訊息呈現。處理這類任務時，先把自己限制在 `cmd/d2r-hyper-launcher` 的 renderer / menu / input 邊界，避免把單純的顯示調整誤做成 domain 邏輯重寫。

## 先看哪些檔案

- [cmd/d2r-hyper-launcher/feedback.go](../../../cmd/d2r-hyper-launcher/feedback.go) - CLI UI layer 核心，包含 message renderer、header/menu block、grouped lines、input 與 display width 計算
- [cmd/d2r-hyper-launcher/menu.go](../../../cmd/d2r-hyper-launcher/menu.go) - startup announcement、主選單、主選單 option 文案與帳號列表顯示
- [cmd/d2r-hyper-launcher/main.go](../../../cmd/d2r-hyper-launcher/main.go) - launcher bootstrap 與 `printStartupAnnouncement()` / `printMenu()` 的呼叫位置
- [cmd/d2r-hyper-launcher/ui_test.go](../../../cmd/d2r-hyper-launcher/ui_test.go) - UI renderer 規則測試
- [cmd/d2r-hyper-launcher/main_test.go](../../../cmd/d2r-hyper-launcher/main_test.go) - 主選單與 announcement 的回歸測試
- 需要調整的對應 flow 檔案，例如：
  - [cmd/d2r-hyper-launcher/cli_launch.go](../../../cmd/d2r-hyper-launcher/cli_launch.go)
  - [cmd/d2r-hyper-launcher/cli_launch_delay.go](../../../cmd/d2r-hyper-launcher/cli_launch_delay.go)
  - [cmd/d2r-hyper-launcher/cli_switcher.go](../../../cmd/d2r-hyper-launcher/cli_switcher.go)
  - [cmd/d2r-hyper-launcher/cli_flags.go](../../../cmd/d2r-hyper-launcher/cli_flags.go)
  - [cmd/d2r-hyper-launcher/cli_d2r_path.go](../../../cmd/d2r-hyper-launcher/cli_d2r_path.go)
  - [cmd/d2r-hyper-launcher/cli_selectors.go](../../../cmd/d2r-hyper-launcher/cli_selectors.go)

## 核心事實

1. `feedback.go` 現在不是單純 prefix helper，而是 CLI UI layer；call site 應只表達語意，不直接決定 icon、divider、prompt spacing 或 option 對齊。
2. header 與 menu 是不同語意：
   - `headf(...)`：表示目前位於哪個環節 / section
   - `menuBlock(func(){ ... })`：表示玩家準備閱讀並輸入的一組內容
3. `headf(...)` 會依 `headerDivider` 的顯示寬度置中，寬度計算要走 `displayWidth(...)`，不能只用 rune count。
4. menu option 現在優先走 `newMenuOptions()` 收集；`option(...)` 以 `key / label / comment` 三欄資料建模，再由 `render()` 依最長 prefix 與 label 做 display-width-aware 對齊；`cliMenuOptions` 會綁定建立它的 `ui`。
5. 若主選單需要固定的 `q` 離開選項，優先使用 `ui.mainMenuOptions(func(*cliMenuOptions))`；它會統一補上 custom options 後的空行與「退出」。
6. 若子選單需要固定的 `b` / `h` / `q` 導航，優先使用 `ui.subMenuOptions(func(*cliMenuOptions))`；它會統一補上 custom options 後的空行與「回上一層 / 回主選單 / 離開程式」。
7. `infoLines(...)` / `warningLines(...)` / `promptLines(...)` / `successLines(...)` / `errorLines(...)` 是「同一組訊息的多段內容」，只會顯示一次 icon，後續段落縮排對齊到 icon 後方。
8. 子選單導航固定是 `b` / `h` / `q`，且仍應保留在 submenu 的最後一組選項。
9. launcher 等執行命令的可見輸出也應走 UI layer；優先使用 `ui.commandf(...)`，並由 `>` prefix 表示實際執行的命令列。
10. CLI 輸出與輸入已收斂到 UI layer；不要在新的 call site 再散落 `fmt.Print*` 或直接操作 `scanner`。

## 修改時要守住的規則

- 想調整玩家可見訊息時，優先找 `ui.infof(...)`、`ui.commandf(...)`、`ui.warningf(...)`、`ui.promptf(...)`、`ui.headf(...)`、`ui.menuBlock(...)`、`ui.newMenuOptions()`、`ui.mainMenuOptions(...)`、`ui.subMenuOptions(...)`；不要直接拼接裸輸出。
- 若訊息是同一組公告 / 說明的多段內容，優先用 `*Lines(...)` helper，不要手動塞 `\n` 後又重複 icon。
- 若需要 option 對齊，先把選項收進 `cliMenuOptions`，並優先把補充資訊放進 `comment` 欄位；不要在 label 內硬塞一長串括號後又自己猜最長寬度。
- 若涉及全形字、中文 key、或 icon 對齊，使用 `displayWidth(...)` 邏輯；不要回退成 `utf8.RuneCountInString(...)`。
- announcement / header / menu 文案調整時，確認 `ui_test.go` 與 `main_test.go` 的預期仍反映最新 UX。
- 主選單與子選單的顯示變更，若影響玩家理解流程，需同步檢查 `README.md`、`docs/` 與 `AGENT.md` 是否要更新。

## 常見任務做法

### 調整啟動畫面的公告

1. 先看 [cmd/d2r-hyper-launcher/menu.go](../../../cmd/d2r-hyper-launcher/menu.go) 的 `printStartupAnnouncement()`
2. 若是一組多段訊息，優先改用 `infoLines(...)` 或 `warningLines(...)`
3. 若只是 section identity，保持用 `headf(...)`，不要把 announcement 改回零散 `rawln(...)`

### 調整主選單 / 子選單排版

1. 先看 `printMenu()` 或對應 `cli_*.go` flow
2. 主選單若最後要帶固定離開選項，優先用 `ui.mainMenuOptions(func(options *cliMenuOptions) { ... })`
3. 其他一般 menu 用 `newMenuOptions()` 收集所有選項，再 `render()`
4. 每個 option 優先寫成 `option(key, label, comment)`；主動作放 `label`，狀態 / 補充說明放 `comment`
5. 若是子選單，優先用 `ui.subMenuOptions(func(options *cliMenuOptions) { ... })`，讓 UI layer 自動補上空行與固定導覽選項

### 調整錯誤 / 提示 / confirm 顯示

1. 看 `feedback.go` 的 message helpers 與 `renderMessage(...)`
2. 若是同一則訊息的多段說明，用 `*Lines(...)`
3. 若是要等待玩家確認，沿用 `readInputf(...)` 與 `anyKeyContinue()`，不要自行分叉新 prompt 行為

## 驗證

若有改到 Go 程式或可見 UI 行為，至少跑：

```powershell
.\scripts\go-test.ps1
New-Item -ItemType Directory -Force .\.tmp | Out-Null
go build -o .\.tmp\d2r-hyper-launcher-dev.exe ./cmd/d2r-hyper-launcher
```

至少確認：

1. `ui_test.go` 的 renderer 規則仍通過
2. `main_test.go` 的主選單 / announcement 預期仍通過
3. 主選單 header、公告、帳號列表與 menu option 沒有肉眼可見的對齊退化

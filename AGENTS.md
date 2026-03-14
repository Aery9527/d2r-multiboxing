# AGENTS.md

## 快速上下文
- 這是一個僅支援 Windows 的 D2R Go CLI 工具箱，主功能為 `multiboxing` 與 `switcher`（`cmd/d2r-hyper-launcher/main.go`）。
- 執行期資料預設存放在 `%USERPROFILE%\.d2r-hyper-launcher`；若有設定 `D2R_HYPER_LAUNCHER_HOME` 則優先使用該路徑（`internal/common/config/config.go`）。
- 使用者可見的核心語言與 UX 為繁體中文；請保持選單文字與提示訊息風格一致。

## 專案定位
- 請將此專案視為「D2R 工具集合」，而非單一功能專案；目前核心能力是 `multiboxing` 與 `switcher`，未來可擴充更多工具。

## 先掌握的架構
- CLI 流程編排集中在 `cmd/d2r-hyper-launcher/cli_*.go`、`main.go`、`menu.go`；領域邏輯放在 `internal/...`。
- 共用基礎在 `internal/common/`（設定、D2R 常數、進程與視窗操作）。
- 多開領域在 `internal/multiboxing/`（帳號、啟動器、mods、背景 handle monitor）。
- 視窗切換在 `internal/switcher/`（熱鍵/滑鼠 hook/搖桿偵測與輪詢）。
- `main.go` 啟動流程：載入設定 -> 確保/載入 `accounts.csv` -> 自動加密明文密碼 ->（可選）啟動 switcher -> 啟動 handle monitor -> 進入主選單迴圈。
- `multiboxing` 啟動流程（`cli_launch.go`）：驗證 D2R 路徑 -> 選區域與 mod -> 啟動程序 -> 關閉單實例 Event handle -> 將視窗改名為 `D2R-<DisplayName>`。

## 專案特有開發流程
- 本機執行優先使用：`./scripts/go-run.ps1`。
- 執行測試優先使用：`./scripts/go-test.ps1`（透過 `scripts/go-test-exec.ps1` 繞過環境對暫存測試執行檔限制）。
- 此環境直接跑 `go test ./...` 可能失敗；先用包裝腳本驗證。
- 開發版 build 請輸出到 `.tmp/`，避免覆蓋 repo 根目錄的 release 成品（見 `README.md`）。
- release build 需注入版本與時間：`-X main.version=... -X main.releaseTime=...`。

## 這個專案的重要慣例
- 玩家可見輸出請走 UI layer：`ui.infof`、`ui.warningLines`、`ui.menuBlock`、`ui.mainMenuOptions`、`ui.subMenuOptions`（`cmd/d2r-hyper-launcher/feedback.go`）。
- 外部命令顯示也走 `ui.commandf(...)`，不要在 domain 層手動拼 `> ` 前綴。
- 所有子選單導航契約固定：`b` 返回、`h` 主選單、`q` 離開（`cmd/d2r-hyper-launcher/menu.go`）。
- 輸入錯誤使用 pause helper：`showInputErrorAndPause`、`showInvalidInputAndPause`，讓玩家確認後回流程。
- **警告訊息後緊接著繼續流程時，必須使用 `showWarningAndPause` 而非單純 `ui.warningf`**，確保玩家有機會閱讀警告內容再繼續；若直接用 `ui.warningf` 後流程馬上刷新畫面，訊息會被清掉玩家看不到。
- 需要對齊選單時，優先用 `newMenuOptions()` 收集再 `render()`；主選單用 `mainMenuOptions(...)`，子選單用 `subMenuOptions(...)`。
- 中文字寬對齊沿用 display-width-aware 邏輯，不要退回單純 rune 計數。
- 不要引導玩家手動改 `config.json`；優先提供 CLI 內可操作流程（例如 `d2r_path_picker.go`）。
- 設定載入需保留向後相容：舊版 `launch_delay: 5` 會在 load 時接受並正規化為 `10`（`internal/common/config/config.go`）。
- `accounts.csv` 行為是刻意設計：UTF-8 BOM、載入時 `LaunchFlags` 清洗、密碼以 `ENC:` + DPAPI 加密（`internal/multiboxing/account/account.go`）。

## 整合點 / 風險點
- 命令列紀錄僅透過 `launcher.SetCommandLogger` 接一次；線上啟動日誌必須遮罩 `-password`（`internal/multiboxing/launcher/launcher.go`）。
- Handle 清理使用 NT API（`NtQuerySystemInformation`、`NtQueryObject`、`NtDuplicateObject`），且只對 `Event` 類型查名稱以避免卡住（`internal/multiboxing/launcher/enumerator.go`）。
- Switcher 內部依賴 OS thread 綁定訊息迴圈（`runtime.LockOSThread`）處理熱鍵/Hook 與 XInput 輪詢（`internal/switcher/hotkey.go`、`mousehook.go`、`gamepad.go`）。
- 視窗切換與執行中帳號偵測依賴 `D2R-` 標題前綴；改動命名規則會同時影響 multiboxing 狀態與 switcher 行為（`internal/common/d2r/constants.go`）。

## Git / 分支流程
- 進行修改前先確認位於 `develop`；日常開發預設在 `develop` 進行。
- release 前先在 `develop` 完成測試與確認，再進行版本決策、build 與 release note。
- release 完成後再 merge 到 `master` 並建立 release tag。

## Go 開發慣例
- 若 struct 明確實作介面，請在宣告上方加入 `var _ InterfaceName = (*StructName)(nil)` 做靜態驗證。
- 使用 `any`，不要使用 `interface{}`。
- 測試優先使用 `github.com/stretchr/testify/assert`；可行時補上或擴充測試覆蓋行為契約。
- 新功能優先使用 Go 標準庫與 `golang.org/x/sys/windows`，避免擴張高風險低階 Windows 手法。

## 文件入口
- `README.md`：專案簡介與快速上手。
- `docs/multiboxing-usage-guide.md`、`docs/switcher-usage-guide.md`：玩家操作導覽。
- `docs/multiboxing-technical-guide.md`、`docs/switcher-technical-guide.md`：技術導覽。
- `docs/D2R_PARAMS.md`：D2R 啟動參數與旗標參考。

## 變更收尾原則
- 測試通過後，檢查受影響範圍的文件是否需要同步更新。
- 若修改影響使用者可見流程、設定或限制，需同步更新 `README.md` 與 `docs/` 對應文件。
- 目標是讓文件描述與目前程式碼邏輯保持一致，避免過期說明。

## 新代理建議閱讀順序
- `README.md` -> `AGENTS.md` -> `cmd/d2r-hyper-launcher/main.go` -> `cmd/d2r-hyper-launcher/feedback.go`。
- 再依需求選路徑：`internal/multiboxing/...`（啟動流程）或 `internal/switcher/...`（輸入/切窗流程）。
- 可把 `cmd/d2r-hyper-launcher/main_test.go` 當作行為地圖；測試採 `testify/assert`，涵蓋許多 CLI 契約檢查。

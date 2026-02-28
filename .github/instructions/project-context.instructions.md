---
applyTo: "**/*.go"
---

# project-context

D2R Multiboxing CLI Tool — 在同一台 Windows 電腦上同時執行多個 Diablo II: Resurrected (D2R) 實例的輔助工具。

## 核心原理

D2R 啟動時建立名為 `DiabloII Check For Other Instances` 的 Windows Event Handle 阻止多開，
本工具透過 `NtDuplicateObject` + `DuplicateCloseSource` 關閉該 handle 來解除限制。

## 專案架構

- [cmd/d2r-multiboxing/main.go](../../cmd/d2r-multiboxing/main.go) - CLI 互動主迴圈
- [internal/config/](../../internal/config/) - 設定檔管理（`~/.d2r-multiboxing/config.json` 讀寫）
- [internal/d2r/](../../internal/d2r/) - D2R 相關常數（進程名、handle 名稱、伺服器區域）
- [internal/handle/](../../internal/handle/) - Windows Handle 操作（NT API 封裝、列舉、關閉）
- [internal/process/](../../internal/process/) - 進程管理（搜尋、啟動 D2R、視窗重命名、視窗切換）
- [internal/account/](../../internal/account/) - 帳號管理（CSV 讀寫、DPAPI 密碼加密）
- [internal/switcher/](../../internal/switcher/) - 視窗切換（快捷鍵/滑鼠側鍵/搖桿觸發、按鍵偵測、XInput）

## 功能

1. CSV 帳號管理（email/密碼/暱稱）
2. DPAPI 密碼加密儲存
3. 參數方式啟動 D2R（`-username` / `-password` / `-address`）
4. 自動關閉單實例 Event Handle
5. 視窗標題重命名為帳號暱稱（D2R- 前綴）
6. 背景持續監控並關閉新出現的 Handle
7. 快捷鍵/滑鼠側鍵/搖桿按鈕切換 D2R 視窗焦點

## 技術參考

- [PLAN-v1-multiboxing.md](../../PLAN-v1-multiboxing.md) - Phase 1 多開啟動器實作計畫
- [PLAN-v2-switcher.md](../../PLAN-v2-switcher.md) - Phase 2 視窗切換功能實作計畫
- [chenwei791129/multiablo](https://github.com/chenwei791129/multiablo) - Go handle 關閉參考
- [shupershuff/Diablo2RLoader](https://github.com/shupershuff/Diablo2RLoader) - 功能參考

## CLI 選單設計規則

- 所有子選單（非主選單的互動畫面）**必須**提供「回上一層」、「回主選單」與「離開程式」三個導航選項
- 導航選項值在整個專案內統一使用常數 `menuBack = "b"`（回上一層）、`menuHome = "h"`（回主選單）、`menuQuit = "q"`（離開程式），定義於 [cmd/d2r-multiboxing/main.go](../../cmd/d2r-multiboxing/main.go)
- 使用 `printSubMenuNav()` 印出導航提示、`isMenuNav(input)` 判斷使用者輸入是否為導航指令（輸入 `q` 時直接結束程式）
- 主選單同樣以 `q` 退出程式

## Windows 安全規則

- **禁止在新功能中引入會觸發 Windows 安全監控（如 Windows Defender）的 API 或手法**
- 僅 [internal/handle/](../../internal/handle/) 因核心多開功能需要使用 NT API（`NtDuplicateObject`、`NtQuerySystemInformation`、`NtQueryObject`），此為已知且必要的例外
- 其餘功能（檔案操作、進程啟動、UI 互動等）一律使用 Go 標準庫或 `golang.org/x/sys/windows` 提供的高階 API，不得直接呼叫 `ntdll.dll`、注入記憶體、操作遠端進程等低階操作
- 編譯產物若被防毒軟體誤判，應優先透過排除清單或簽章解決，不以修改核心機制為手段

---

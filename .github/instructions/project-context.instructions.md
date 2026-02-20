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

---

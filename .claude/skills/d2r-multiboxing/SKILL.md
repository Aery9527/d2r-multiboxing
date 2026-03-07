---
name: d2r-multiboxing
description: "Handle repository-specific Diablo II: Resurrected multiboxing work in d2r-hyper-launcher. Use this whenever the user wants to change multi-instance launching, account loading, D2R startup parameters, handle-closing logic, background handle monitoring, window renaming, launch delay behavior, or CSV/DPAPI account flow, even if they only mention '多開', 'multibox', 'launcher', 'accounts.csv', or D2R startup bugs."
---

# D2R Multiboxing Context

這個 skill 專注在 `d2r-hyper-launcher` 的多開啟動器範圍。處理這類任務時，先把自己限制在帳號啟動、背景 handle 關閉、D2R 啟動參數、資料目錄與視窗重命名，避免混入 switcher 以外的無關變更。

## 先看哪些檔案

- [cmd/d2r-hyper-launcher/main.go](../../../cmd/d2r-hyper-launcher/main.go) - 主選單、單帳號啟動、批次啟動、背景 handle 監控
- [internal/process/launcher.go](../../../internal/process/launcher.go) - D2R 啟動參數組裝
- [internal/process/window.go](../../../internal/process/window.go) - 視窗重命名與前景切換
- [internal/handle/closer.go](../../../internal/handle/closer.go) - 關閉目標 handle 的公開 API
- [internal/handle/enumerator.go](../../../internal/handle/enumerator.go) - 列舉系統 handle 並篩選目標 Event
- [internal/handle/winapi.go](../../../internal/handle/winapi.go) - NT API 封裝
- [internal/account/account.go](../../../internal/account/account.go) 與 [internal/account/crypto.go](../../../internal/account/crypto.go) - CSV / DPAPI
- [internal/config/config.go](../../../internal/config/config.go) - `d2r_path`、`launch_delay`、資料目錄
- [internal/d2r/constants.go](../../../internal/d2r/constants.go) - 進程名、Event 名稱、區域常數、視窗標題前綴
- [README.md](../../../README.md) 與 [docs/multiboxing-usage-guide.md](../../../docs/multiboxing-usage-guide.md) - 使用者可見行為

## 核心事實

1. 多開的核心是關閉 `DiabloII Check For Other Instances` Event Handle。
2. 只有 [internal/handle/](../../../internal/handle/) 允許使用 NT API；其他功能維持 Go 標準庫或 `golang.org/x/sys/windows` 高階 API。
3. 線上帳號啟動由 `LaunchD2R()` 組出 `-uid osi -username -password -address`；不要移除 `-uid osi`。
4. 視窗標題需維持 `D2R-<DisplayName>` 格式，這也是 switcher 用來找視窗的依據。
5. 背景 monitor 會每 2 秒掃描 D2R 行程，替新 PID 再做一次 handle 關閉。

## 修改時要守住的規則

- 子選單必須保留 `b` / `h` / `q` 導航，沿用 `printSubMenuNav()` 與 `isMenuNav()`。
- `enumerator.go` 只對 `Event` 類型查詢 object name，避免對 pipe/file handle 查詢造成 hang。
- 啟動後仍需要等待遊戲初始化，再執行 `CloseHandlesByName()` 與 `RenameWindow()`。
- 若調整帳號或設定流程，要同步更新 [README.md](../../../README.md) 與 [docs/multiboxing-usage-guide.md](../../../docs/multiboxing-usage-guide.md)。
- 若改動會影響視窗標題或前景行為，要檢查是否連帶影響 switcher。

## 常見任務做法

### 調整啟動流程

1. 先看 [cmd/d2r-hyper-launcher/main.go](../../../cmd/d2r-hyper-launcher/main.go) 的 `launchAccount()`、`launchAll()`、`launchOffline()`
2. 再看 [internal/process/launcher.go](../../../internal/process/launcher.go) 是否需要新增或調整參數
3. 檢查密碼是否仍經過 `redactArgs()` 遮罩

### 修正多開失敗

1. 先確認 [internal/d2r/constants.go](../../../internal/d2r/constants.go) 的 `SingleInstanceEventName`
2. 再檢查 [internal/handle/enumerator.go](../../../internal/handle/enumerator.go) 與 [internal/handle/closer.go](../../../internal/handle/closer.go)
3. 若是時序問題，再回頭調整 `main.go` 啟動後的等待與 monitor 邏輯

### 調整帳號或設定儲存

1. 看 [internal/account/account.go](../../../internal/account/account.go)
2. 看 [internal/config/config.go](../../../internal/config/config.go)
3. 確保資料目錄仍是 `~/.d2r-hyper-launcher` 或 `D2R_HYPER_LAUNCHER_HOME`

## 驗證

至少跑：

```powershell
go test ./...
go build ./cmd/d2r-hyper-launcher
```

若改到啟動流程或視窗標題，還要人工確認：

1. 單帳號啟動
2. `a` 批次啟動
3. 背景 monitor 對新 D2R PID 仍會生效
4. 視窗標題仍是 `D2R-` 前綴

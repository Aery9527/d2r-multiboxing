---
name: d2r-multiboxing
description: "Handle repository-specific Diablo II: Resurrected multiboxing work in d2r-hyper-launcher. Use this whenever the user wants to change multi-instance launching, account loading, D2R startup parameters, handle-closing logic, background handle monitoring, window renaming, launch delay behavior, or CSV/DPAPI account flow, even if they only mention '多開', 'multibox', 'launcher', 'accounts.csv', or D2R startup bugs."
---

# D2R Multiboxing Context

這個 skill 專注在 `d2r-hyper-launcher` 的多開啟動器範圍。處理這類任務時，先把自己限制在帳號啟動、背景 handle 關閉、D2R 啟動參數、資料目錄與視窗重命名，避免混入 switcher 以外的無關變更。

## 先看哪些檔案

- [cmd/d2r-hyper-launcher/main.go](../../../cmd/d2r-hyper-launcher/main.go) - 薄 CLI bootstrap、主選單 dispatch、背景 monitor 啟動
- [cmd/d2r-hyper-launcher/cli_launch.go](../../../cmd/d2r-hyper-launcher/cli_launch.go) - 單帳號 / 批次 / 離線啟動 flow
- [cmd/d2r-hyper-launcher/cli_accounts_file.go](../../../cmd/d2r-hyper-launcher/cli_accounts_file.go) 與 [cmd/d2r-hyper-launcher/cli_d2r_path.go](../../../cmd/d2r-hyper-launcher/cli_d2r_path.go) - `accounts.csv` 建立與 `D2R.exe` 路徑修復流程
- [cmd/d2r-hyper-launcher/cli_selectors.go](../../../cmd/d2r-hyper-launcher/cli_selectors.go) 與 [cmd/d2r-hyper-launcher/cli_flags.go](../../../cmd/d2r-hyper-launcher/cli_flags.go) - account / mod / launch flags 選擇流程
- [internal/multiboxing/launcher/launcher.go](../../../internal/multiboxing/launcher/launcher.go) - D2R 啟動參數組裝
- [internal/multiboxing/launcher/closer.go](../../../internal/multiboxing/launcher/closer.go) - 關閉目標 handle 的公開 API
- [internal/multiboxing/launcher/enumerator.go](../../../internal/multiboxing/launcher/enumerator.go) - 列舉系統 handle 並篩選目標 Event
- [internal/multiboxing/launcher/winapi.go](../../../internal/multiboxing/launcher/winapi.go) - NT API 封裝
- [internal/multiboxing/monitor/handle_monitor.go](../../../internal/multiboxing/monitor/handle_monitor.go) - 背景 handle monitor
- [internal/multiboxing/account/account.go](../../../internal/multiboxing/account/account.go) 與 [internal/multiboxing/account/crypto.go](../../../internal/multiboxing/account/crypto.go) - CSV / DPAPI
- [internal/common/config/config.go](../../../internal/common/config/config.go) - `d2r_path`、`launch_delay`、資料目錄
- [internal/multiboxing/mods/](../../../internal/multiboxing/mods/) - 已安裝 mod 掃描與 `-mod` / `-txt` 參數組裝
- [internal/common/d2r/constants.go](../../../internal/common/d2r/constants.go) - 進程名、Event 名稱、區域常數、視窗標題前綴
- [internal/common/process/window.go](../../../internal/common/process/window.go) - 視窗重命名與前景切換
- [README.md](../../../README.md) 與 [docs/multiboxing-usage-guide.md](../../../docs/multiboxing-usage-guide.md) - 使用者可見行為

## 核心事實

1. 多開的核心是關閉 `DiabloII Check For Other Instances` Event Handle。
2. 只有 [internal/multiboxing/launcher/](../../../internal/multiboxing/launcher/) 這層允許直接碰 NT API；其他功能維持 Go 標準庫或 `golang.org/x/sys/windows` 高階 API。
3. 線上帳號啟動由 `LaunchD2R()` 組出 `-uid osi -username -password -address`；不要移除 `-uid osi`。
4. 視窗標題需維持 `D2R-<DisplayName>` 格式，這也是 switcher 用來找視窗的依據。
5. 背景 monitor 會每 2 秒掃描 D2R 行程，替新 PID 再做一次 handle 關閉。
6. `0` 與 `a` 啟動流程會先掃描 `D2R.exe` 同層 `mods\` 目錄；只要 mod 資料夾內有 `modinfo.json`，或有同名 `<mod>.mpq`，就會出現在選單中，並轉成 `-mod <name> -txt` 參數。

## 修改時要守住的規則

- 子選單必須保留 `b` / `h` / `q` 導航，沿用 `printSubMenuNav()` 與 `isMenuNav()`。
- CLI 互動流程已拆到 `cli_launch.go`、`cli_d2r_path.go`、`cli_accounts_file.go`、`cli_flags.go` 等檔案；不要再把多開流程整坨塞回 `main.go`。
- `enumerator.go` 只對 `Event` 類型查詢 object name，避免對 pipe/file handle 查詢造成 hang。
- 啟動後仍需要等待遊戲初始化，再執行 `CloseHandlesByName()` 與 `RenameWindow()`。
- 不要把玩家導回手動修改 `config.json`；像 `d2r_path` 這類玩家可見設定，優先提供 CLI 內可操作流程。
- 若 `d2r_path` 已失效，單帳號 / 批次 / 離線啟動都應先攔下來，明確提示找不到 `D2R.exe`，並直接提供與主選單 `p` 相同的設定入口。
- 若調整帳號或設定流程，要同步更新 [README.md](../../../README.md) 與 [docs/multiboxing-usage-guide.md](../../../docs/multiboxing-usage-guide.md)。
- 若改動會影響視窗標題或前景行為，要檢查是否連帶影響 switcher。

## 常見任務做法

### 調整啟動流程

1. 先看 [cmd/d2r-hyper-launcher/cli_launch.go](../../../cmd/d2r-hyper-launcher/cli_launch.go) 的 `launchAccount()`、`launchAll()`、`launchOffline()`
2. 再看 [internal/multiboxing/launcher/launcher.go](../../../internal/multiboxing/launcher/launcher.go) 是否需要新增或調整參數
3. 若有 mod 選擇流程，再檢查 [cmd/d2r-hyper-launcher/cli_selectors.go](../../../cmd/d2r-hyper-launcher/cli_selectors.go) 與 [internal/multiboxing/mods/](../../../internal/multiboxing/mods/) 是否仍正確串接 `-mod <name> -txt`
4. 檢查密碼是否仍經過 `redactArgs()` 遮罩

### 修正多開失敗

1. 先確認 [internal/common/d2r/constants.go](../../../internal/common/d2r/constants.go) 的 `SingleInstanceEventName`
2. 再檢查 [internal/multiboxing/launcher/enumerator.go](../../../internal/multiboxing/launcher/enumerator.go) 與 [internal/multiboxing/launcher/closer.go](../../../internal/multiboxing/launcher/closer.go)
3. 若是時序問題，再回頭調整 [cmd/d2r-hyper-launcher/cli_launch.go](../../../cmd/d2r-hyper-launcher/cli_launch.go) 與 [internal/multiboxing/monitor/handle_monitor.go](../../../internal/multiboxing/monitor/handle_monitor.go)

### 調整帳號或設定儲存

1. 看 [internal/multiboxing/account/account.go](../../../internal/multiboxing/account/account.go)
2. 看 [internal/common/config/config.go](../../../internal/common/config/config.go)
3. 確保資料目錄仍是 `~/.d2r-hyper-launcher` 或 `D2R_HYPER_LAUNCHER_HOME`

## 驗證

至少跑：

```powershell
.\scripts\go-test.ps1
New-Item -ItemType Directory -Force .\.tmp | Out-Null
go build -o .\.tmp\d2r-hyper-launcher-dev.exe ./cmd/d2r-hyper-launcher
```

若改到啟動流程或視窗標題，還要人工確認：

1. 單帳號啟動
2. `a` 批次啟動
3. 背景 monitor 對新 D2R PID 仍會生效
4. 視窗標題仍是 `D2R-` 前綴

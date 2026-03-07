# CLI and Internal Boundary Refactor Plan

## 完成狀態

這份 refactor plan 已完成實作，實際結果已在 `refactor/cli-internal-boundaries` branch 收斂，並準備併回 `develop`。

已完成重點：

- 將 `cmd/d2r-hyper-launcher/main.go` 拆成較薄的 bootstrap / dispatch 入口
- 將 CLI 互動流程拆到 `menu.go`、`feedback.go`、`cli_launch.go`、`cli_flags.go`、`cli_switcher.go`、`cli_d2r_path.go` 等專責檔案
- 將 `internal/` 重組為：
  - `internal/common`
  - `internal/multiboxing`
  - `internal/switcher`
- 把多開相關的 account / launcher / mods / monitor 邏輯收斂到 `internal/multiboxing`
- 把共用的 config / d2r / process helper 收斂到 `internal/common`
- 同步更新相關 skill，讓後續工作流程與新結構一致

驗證結果：

- `.\scripts\go-test.ps1`
- `go build -o .\.tmp\d2r-hyper-launcher-dev.exe .\cmd\d2r-hyper-launcher`

追蹤方式：

- plan：`refactor-plan/cli-refactor.md`
- branch：`refactor/cli-internal-boundaries`
- 主要實作 commit：`0926411` `refactor(repo): realign CLI and feature boundaries`
- skill 同步 commit：`6c7a6ae` `docs(skills): align D2R workflows with new boundaries`
- 收尾流程 commit：`3d01be4` `docs(skills): formalize refactor completion tracking`

## 為什麼這次要把 CLI 與 `internal\` 一起整理

- 原本的 CLI refactor 方案三，核心是在解決 `cmd/d2r-hyper-launcher/main.go` 同時承擔 menu renderer、selector、validator、feedback、domain coordinator 的問題
- 但如果只拆 CLI flow，不同時整理 `internal\` 的 package 邊界，`main.go` 只是把複雜度搬家，整體模組命名仍然不清楚
- 目前專案對外提供的核心能力其實很明確：**multiboxing** 與 **switcher**
- 因此 `internal\` 應該收斂成：
  - `common`：跨功能共用基礎
  - `multiboxing`：多開相關 domain
  - `switcher`：切窗相關 domain

這表示這次不只是單純拆檔，而是一次把 **CLI flow 邊界** 與 **feature package 邊界** 一起校正。

## 目前結構的問題落點

### `cmd/d2r-hyper-launcher/main.go`

- 主選單 dispatch、區域選擇、mod 選擇、flags 設定、switcher 設定、D2R path setup 都集中在同一檔案
- 最近只是統一 invalid input feedback，就需要在多個 flow 同步修改，說明 CLI concern 仍然散落
- 即使 `cli_feedback.go` 已收斂錯誤提示樣式，prompt → validate → retry → navigate 的流程仍然分散

### `internal\` top-level package

目前的 top-level package：

- `account`
- `config`
- `d2r`
- `handle`
- `mods`
- `process`
- `switcher`

這組命名比較像歷史演進結果，而不是功能邊界：

- `process` 同時混有 generic process finder 與 D2R-specific launcher/window 行為
- `config` 同時有 config model 與 UI picker
- `handle` 名稱偏底層，但實際主要服務多開啟動
- `account`、`mods` 本質上都屬於 multiboxing

## 目標結構

### 1. `internal\` 只保留三個核心 scope

```text
internal/
  common/
    d2r/
    config/
    process/
  multiboxing/
    account/
    launcher/
    mods/
    monitor/
  switcher/
```

### 2. CLI flow 明確拆層

```text
cmd/d2r-hyper-launcher/
  main.go
  menu.go
  cli/
    feedback.go
    selectors.go
    validators.go
    launchers.go
    switcher_menu.go
    d2r_path_menu.go
    flags/
      flags_menu.go
      flags_config.go
```

## package migration 對應

| 現在 | 目標 | 原因 |
| --- | --- | --- |
| `internal/d2r` | `internal/common/d2r` | D2R 常數、region、process/window 名稱屬於跨功能共用基礎 |
| `internal/config/config.go` | `internal/common/config` | 設定模型與 JSON I/O 是共用基礎 |
| `internal/config/picker.go` | `cmd/d2r-hyper-launcher/cli` | 這是 CLI/UI 行為，不適合留在 `internal/common` |
| `internal/process/finder.go` | `internal/common/process` | generic process discovery 可被不同功能共用 |
| `internal/process/launcher.go` | `internal/multiboxing/launcher` | D2R 啟動屬於多開 domain |
| `internal/process/window.go` | `internal/multiboxing/launcher` | 視窗命名與帳號啟動綁在一起 |
| `internal/handle/*` | `internal/multiboxing/launcher` | handle 關閉目前本質上是多開啟動支援 |
| `internal/account/*` | `internal/multiboxing/account` | 帳號、CSV、DPAPI、launch flags 都是多開 domain |
| `internal/mods/*` | `internal/multiboxing/mods` | mod 探測與多開啟動相關 |
| `internal/switcher/*` | `internal/switcher/*` | 已是明確功能邊界，維持不變 |

## CLI 目標責任分配

### `main.go`

- app 啟動
- 載入 config / accounts
- 啟動 switcher background components
- 主選單 dispatch

### `cli/feedback.go`

- invalid input 與錯誤暫停提示
- confirmation / warning 類輸出 helper

### `cli/selectors.go`

- region selector
- mod selector
- account selector
- 共用清單型輸入流程

### `cli/validators.go`

- range / index parsing
- 單選 / 多選格式驗證
- 導航字元與共用 retry 判斷

### `cli/launchers.go`

- `launchAccount`
- `launchAll`
- `launchOffline`

### `cli/flags/`

- `setupAccountLaunchFlags`
- `configureFlagsByFlag`
- `configureFlagsByAccount`

### `cli/switcher_menu.go`

- `setupSwitcher`

### `cli/d2r_path_menu.go`

- D2R path setup / repair flow
- CLI 用 path picker 呼叫點

## 這次合併 refactor 想解決的成本

- 新增一個 CLI 規則時，不應再需要掃過大量 menu flow 才能保證一致
- feature boundary 應該從 package 名稱就看得出來，而不是靠閱讀大量檔案才知道 responsibility
- `common` 不應混進 feature-specific orchestration
- `multiboxing` 與 `switcher` 各自的 domain 應有更清楚的內聚力

## 建議執行方案

### Phase 1：先拆 CLI，不改行為

目標：讓 `main.go` 降成薄 dispatcher，先把高密度互動流程從單一大檔拆出。

要做的事：

- 將 `cli_feedback.go` 移為 `cli/feedback.go`
- 抽出 `selectors.go`
- 抽出 `validators.go`
- 抽出 `launchers.go`
- 抽出 `flags/` 子目錄
- 抽出 `switcher_menu.go`
- 抽出 `d2r_path_menu.go`

完成標準：

- `main.go` 主要只剩啟動與主選單 dispatch
- CLI 行為不變

### Phase 2：整理 `internal/common`

目標：先確立真正跨功能共用的基礎層。

要做的事：

- `internal/d2r` → `internal/common/d2r`
- `internal/config` 中保留 config model / load / save 到 `internal/common/config`
- 將 generic process finder 移到 `internal/common/process`
- 若有 D2R path validation，併入 `internal/common/config`

注意：

- `PickD2RPath` 這類 CLI UI 行為不要留在 `common`

### Phase 3：整理 `internal/multiboxing`

目標：讓多開相關 domain 都集中在同一個 namespace。

要做的事：

- `internal/account` → `internal/multiboxing/account`
- `internal/mods` → `internal/multiboxing/mods`
- `internal/process/launcher.go`、`window.go` → `internal/multiboxing/launcher`
- `internal/handle/*` 併入 `internal/multiboxing/launcher`
- 若有背景 handle monitor，抽到 `internal/multiboxing/monitor`

### Phase 4：整理 imports、測試與殘餘邊界

目標：清掉舊 package 路徑並確認沒有新的 boundary leakage。

要做的事：

- 更新所有 import
- 清除舊 top-level package
- 補上抽出 helper 後需要的測試
- 跑整體測試與 build

## 主要風險與處理方式

### 風險 1：`config` 混有 model 與 UI

- 處理：model 留 `internal/common/config`，UI picker 留 CLI

### 風險 2：`process` 混有 generic 與 feature-specific 行為

- 處理：只把 generic finder 留 common，其餘移到 multiboxing

### 風險 3：`handle` 看起來底層，但其實不是通用 abstraction

- 處理：不要為了「看起來通用」硬留獨立 package，直接併入 multiboxing launcher

### 風險 4：一次搬動過大，容易在行為沒變前就把 diff 放太大

- 處理：先 CLI extraction，再 package migration，避免同一步同時做結構與行為調整

## 建議 branch

既然這份 plan 已經被 review 並準備進入實作，應先建立獨立 branch：

```text
refactor/cli-internal-boundaries
```

這個名稱能同時表達：

- CLI flow 正在拆邊界
- `internal\` package boundary 也在重整

## 暫時不做的事

- 不引入新的 framework 式 state machine
- 不重寫 switcher 核心切窗機制
- 不因為 package 重新命名就順手改動所有 domain API
- 不做超過「邊界收斂」以外的 feature redesign

## 建議順序

1. 建立 `refactor/cli-internal-boundaries`
2. 完成 CLI extraction
3. 完成 `internal/common` migration
4. 完成 `internal/multiboxing` migration
5. 更新 imports、補測試、跑 build / test

## 實作原則

- 先搬結構，再考慮是否需要額外 API 清理
- 每一步都盡量保持使用者可見行為不變
- 若中途發現 `switcher` 也需要進一步再拆，再另開下一份 plan，不把這輪 scope 無限制擴大

# AGENT

## 專案定位

`d2r-hyper-launcher` 是一個以 Windows 為主的 D2R（Diablo II: Resurrected）工具箱型專案。
目前已實作的重點能力包含：

- `multiboxing`：多開啟動與相關輔助流程
- `switcher`：D2R 視窗切換

未來這個專案可能持續擴充其他 D2R 相關工具，因此請把它視為「D2R 工具集合」而不是只做單一功能的專案。

## 任務分流

- 遇到 `multiboxing` 相關需求時，優先使用對應 skill 取得細節 context
- 遇到 `switcher` 相關需求時，優先使用對應 skill 取得細節 context
- `AGENT.md` 只提供高層引導，不承載這兩個功能的細部實作說明

## 高層架構

- [cmd/d2r-hyper-launcher/main.go](cmd/d2r-hyper-launcher/main.go) - CLI 入口與互動流程
- [internal/config/](internal/config/) - 設定與資料目錄管理
- [internal/d2r/](internal/d2r/) - D2R 相關常數與共用定義
- [internal/account/](internal/account/) - 帳號資料與密碼處理
- [internal/process/](internal/process/) - D2R 啟動、進程與視窗操作
- [internal/handle/](internal/handle/) - 核心 Windows handle 處理
- [internal/switcher/](internal/switcher/) - 視窗切換功能

## 文件入口

- [README.md](README.md) - 專案簡介與快速上手
- [docs/multiboxing-usage-guide.md](docs/multiboxing-usage-guide.md) - multiboxing 操作導覽
- [docs/switcher-usage-guide.md](docs/switcher-usage-guide.md) - switcher 操作導覽
- [docs/multiboxing-technical-guide.md](docs/multiboxing-technical-guide.md) - multiboxing 技術導覽
- [docs/switcher-technical-guide.md](docs/switcher-technical-guide.md) - switcher 技術導覽
- [docs/D2R_PARAMS.md](docs/D2R_PARAMS.md) - D2R 啟動參數整理

## CLI 慣例

- 所有子選單都必須提供「回上一層 / 回主選單 / 離開程式」
- 導航指令固定使用：
  - `b`：回上一層
  - `h`：回主選單
  - `q`：離開程式
- 相關共用邏輯集中在 [cmd/d2r-hyper-launcher/main.go](cmd/d2r-hyper-launcher/main.go)
- 不要要求玩家手動修改 `config.json`；玩家可見設定應優先提供 CLI 內可操作流程，例如用檔案選擇器設定 `D2R.exe` 路徑

## Git / 分支流程

- 進行任何修改前，先確認目前位於 `develop` branch；日常開發預設都應在 `develop`
- release 流程開始前，先在 `develop` 上確認測試通過，之後才進行版本決策、release build、release note 與其他 release 步驟
- release 全部完成後，才將結果 merge 到 `master`，並在最後建立 release tag

## Go 開發慣例

- 若 struct 明確實作某個介面，請在 struct 宣告上方加入 `var _ InterfaceName = (*StructName)(nil)` 做靜態驗證與意圖標示
- 使用 `any`，不要使用 `interface{}`
- 若未來此專案有使用 MongoDB，需注意 `bson.M` 與 `bson.D` 的使用時機；像 `$sort` 這類依賴順序的場景必須使用 `bson.D`
- 撰寫 Go 測試時，優先使用 `github.com/stretchr/testify/assert` 進行驗證
- 每次修改若有合理切點，優先補上或擴充測試，避免只改功能不驗證行為
- 在這台 Windows 開發環境中，若 `go test ./...` 受到 Application Control 阻擋，改用 [scripts/go-test.ps1](scripts/go-test.ps1) 執行整體測試
- 在這台 Windows 開發環境中，若要本機執行 launcher，優先使用 `go run ./cmd/d2r-hyper-launcher` 或 [scripts/go-run.ps1](scripts/go-run.ps1)；不要用本機 build 直接覆蓋 repo 根目錄的 `d2r-hyper-launcher.exe`

## Windows 安全原則

- 除了既有核心 handle 功能外，不要擴大使用低階或高風險 Windows 手法
- 新功能優先使用 Go 標準庫與 `golang.org/x/sys/windows`
- 若新增 D2R 工具能力，優先延續目前專案的安全界線與 CLI 互動風格

## 變更收尾原則

- 測試通過後，檢查所有受影響 scope 的文件是否需要同步更新
- 若修改影響使用者可見流程、選單、設定、限制或技術前提，需同步更新相關 `README`、`docs/` 與其他對應 md 文件
- 若修改影響既有 skill 的觸發條件、工作範圍、操作流程或重要事實，需同步更新 `.claude/skills/` 內對應內容
- 目標是讓文件與 skill 描述都與目前程式碼邏輯保持一致，不留下過期說明


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

## Windows 安全原則

- 除了既有核心 handle 功能外，不要擴大使用低階或高風險 Windows 手法
- 新功能優先使用 Go 標準庫與 `golang.org/x/sys/windows`
- 若新增 D2R 工具能力，優先延續目前專案的安全界線與 CLI 互動風格


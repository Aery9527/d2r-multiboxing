# 裝備清點（Equipment Inventory）實作計劃

採用方案 2: 記憶體讀取 (ReadProcessMemory + d2go)。技術方案分析詳見 `inventory-analysis.md`。

---

## d2go API 關鍵發現

```go
// 1. 支援以 PID 開啟特定程序（完美適配多開場景）
process, err := memory.NewProcessForPID(pid)

// 2. GameReader 提供完整遊戲狀態
gameReader := memory.NewGameReader(process)
data := gameReader.GetData()  // 包含 Inventory, PlayerUnit, 等

// 3. Inventory 結構豐富
data.Inventory.AllItems      // 所有物品
data.Inventory.Belt          // 腰帶
data.Inventory.Gold          // 金幣
data.Inventory.ByLocation()  // 依位置篩選

// 4. Item 結構完整
item.Name          // 物品名稱
item.Quality       // 品質（Normal/Magic/Rare/Set/Unique）
item.Location      // 位置（Inventory/Stash/Equipped/Belt）
item.Stats         // 所有屬性
item.Ethereal      // 是否空靈
item.Sockets       // 鑲嵌物品
item.IsRuneword    // 是否符文之語
item.Identified    // 是否已鑑定
```

---

## 分階段實作

### Phase 0: 技術驗證 (Spike) — 1-2 天
- `go get github.com/hectorgimenez/d2go`
- 寫一個獨立的測試程式：找到 D2R PID → 用 d2go 讀取 Inventory → 印出物品清單
- 驗證 d2go 版本相容性、offset 是否能正確讀取當前 D2R 版本
- **這一步決定整個功能的可行性**

### Phase 1: 核心讀取層 — `internal/inventory/`

**新增檔案**:
- `internal/inventory/reader.go` — 封裝 d2go 的 `Process` + `GameReader`，提供 `ReadInventory(pid uint32) (*CharacterInventory, error)`
- `internal/inventory/model.go` — 定義本專案的資料模型（從 d2go 的 `data.Item` 轉換成我們自己的結構）
- `internal/inventory/display.go` — 物品顯示格式化（品質顏色、屬性排版）

**設計要點**:
- `reader.go` 封裝 d2go，避免 d2go 的型別洩漏到 CLI 層
- 每次讀取：`OpenProcess` → 讀取 → 關閉 handle（不保持長連線）
- 錯誤處理：D2R 未在遊戲中 / 角色不在場景中 / offset 不相容

**可複用的現有程式碼**:
- `internal/common/process/finder.go` — `FindProcessesByName("D2R.exe")` 取得 PID 清單
- `internal/common/process/window.go` — `FindWindowsByTitlePrefix("D2R-")` 映射帳號名到 PID
- `internal/common/d2r/constants.go` — `ProcessName`, `WindowTitlePrefix`

### Phase 2: CLI 整合 — `cmd/d2r-hyper-launcher/`

**新增檔案**:
- `cmd/d2r-hyper-launcher/cli_inventory.go` — 裝備清點子選單

**修改檔案**:
- `cmd/d2r-hyper-launcher/main.go` — 主選單加入裝備清點選項
- `internal/common/locale/catalog.go` — 新增 `InventoryCatalog` i18n 字串

**CLI 流程**:
1. 主選單新增 `[i] 裝備清點` 選項
2. 進入子選單 → 列出執行中的 D2R 帳號（從 `D2R-<DisplayName>` 視窗偵測）
3. 選擇帳號 → 讀取記憶體 → 顯示裝備清單
4. 顯示分區：裝備欄 / 背包 / 倉庫 / 腰帶 / 方塊
5. 每項裝備顯示：名稱、品質、關鍵屬性

**導航契約**（遵循 `menu.go` 的 `runMenu` 模式）:
- `b` 返回帳號選擇
- `h` 主選單
- `q` 離開

### Phase 3: 風險告知 UX
- 首次啟用裝備清點功能時，顯示風險警告（anti-cheat 風險說明）
- 使用 `showWarningAndPause` 確保玩家閱讀
- 記錄玩家的 opt-in 選擇到 `config.json`

### Phase 4: 測試與文件
- `internal/inventory/model_test.go` — 資料轉換測試（mock d2go 輸出）
- `cmd/d2r-hyper-launcher/cli_inventory_test.go` — CLI 流程測試
- 更新 `README.md` + `README.en.md`
- 新增 `docs/inventory-usage-guide.md`

---

## 關鍵依賴決策

| 決策 | 建議 |
|---|---|
| d2go 依賴方式 | 直接 `go get`。d2go 是活躍維護的 MIT 專案，不需 fork。 |
| Offset 維護 | 依賴 d2go 的更新。版本不匹配時顯示「需要更新 d2go」訊息。 |
| 功能開關 | `config.json` 新增 `inventory_enabled: bool`，預設 `false`（opt-in）。 |

---

## 驗證方式

1. **Phase 0 驗證**: 獨立程式能讀到 D2R 角色裝備 → 技術可行
2. **單元測試**: `./scripts/go-test.ps1`
3. **整合測試**: 啟動 D2R → 用本工具的裝備清點功能 → 確認顯示正確
4. **多開測試**: 啟動 2+ D2R → 確認能分別讀取不同帳號的裝備

# D2R Mod 載入問題：Reign of the Warlock (RotW) 版本

> D2R v3.1 (Reign of the Warlock 擴充) 已**禁用**傳統的 `-mod` 與 `-direct -txt` 命令列參數，
> 目前唯一可靠的 Mod 載入方式是透過 [D2RMM](https://github.com/olegbl/d2rmm)。

**相關文件：**
- [D2R Modding 共通指南](D2R-MODDING-COMMON.md) — 目錄結構、工具安裝
- [線上安全 Mod 指南](D2R-MODDING-ONLINE.md) — JSON 字串修改
- [離線 Mod 指南](D2R-MODDING-OFFLINE.md) — 資料表修改
- [D2R 啟動參數一覽](../D2R_PARAMS.md) — 所有命令列參數

---

## 目錄

- [問題描述](#問題描述)
- [驗證過程](#驗證過程)
- [D2RMM 原始碼分析](#d2rmm-原始碼分析)
- [解決方案：D2RMM 驗證流程](#解決方案d2rmm-驗證流程)
- [D2RMM 技術原理](#d2rmm-技術原理)
- [對本工具的影響](#對本工具的影響)
- [參考資源](#參考資源)

---

## 問題描述

自 **D2R v3.1.91735 (Reign of the Warlock)** 更新後，以下啟動參數已**完全失效**：

| 參數 | 原用途 | 現況 |
|------|--------|------|
| `-mod <name> -txt` | 載入 `mods/<name>/` 下的 Mod 檔案 | ❌ D2R 完全忽略，不讀取任何 Mod 檔案 |
| `-direct -txt` | 從本地 `Data/` 目錄載入資料 | ❌ 同樣失效 |
| `-mod <name>`（不帶 `-txt`） | 載入已編譯的 `.bin` 檔案 | ❌ 同樣失效 |

D2R 不會報錯、不會崩潰，單純**靜默忽略**這些參數，載入原版遊戲內容。

---

## 驗證過程

以下是我們在 D2R v3.1.91735 (Battle.net, product: osi) 上的實際測試結果：

### 測試 1：檔案存取時間戳監控

1. 將 `mods/d2r-hyper-show/` 下所有檔案的 `LastAccessTime` 重設為前一天
2. 以 `-mod d2r-hyper-show -txt` 啟動 D2R
3. **結果：所有檔案的存取時間皆未改變** — D2R 從未讀取任何 Mod 檔案，包括 `modinfo.json`

### 測試 2：損壞的 modinfo.json

1. 將 `modinfo.json` 內容替換為 `THIS IS INTENTIONALLY BROKEN JSON {{{{`
2. 以 `-mod d2r-hyper-show -txt` 啟動 D2R
3. **結果：D2R 正常啟動** — 證明 D2R 根本沒有嘗試讀取 `modinfo.json`

### 測試 3：最小空 Mod（testmod）

1. 建立全新的空 Mod：`mods/testmod/modinfo.json` + 空的 `testmod.mpq/` 目錄
2. 以 `-mod testmod -txt` 啟動 D2R
3. **結果：D2R 未建立 `mods/testmod/` 存檔目錄** — 在正常運作的版本中，即使 Mod 無內容，D2R 也會建立存檔目錄

### 測試 4：完整 JSON 字串檔

1. 下載完整的原版 `item-names.json`（1530 筆條目，745KB），合併 32 筆自訂修改
2. 確認 UTF-8 編碼正確、`ÿc` 顏色代碼位元組正確（`C3 BF 63 XX`）
3. 以 `-mod d2r-hyper-show -txt` 啟動 D2R
4. **結果：遊戲內物品名稱無任何變化**

### 社群確認

- [Nexus Mods D2R 討論區](https://www.nexusmods.com/diablo2resurrected) — 多名使用者回報 `-mod` 參數失效
- [D2R Patch 3.1.1 更新說明](https://www.d2itemstore.com/blogs/d2r-news/d2r-patch-3-1-1-reign-of-the-warlock-update-breakdown) — RotW 更新後 Mod 載入方式變更
- 社群共識：**「Only D2RMM works」**

---

## D2RMM 原始碼分析

分析 [D2RMM 原始碼](https://github.com/olegbl/d2rmm)（v1.8.0 ~ v1.9.0-pre）後的關鍵發現：

### 啟動參數產生邏輯

**檔案：** `src/renderer/react/hooks/useGameLaunchArgs.tsx`

```typescript
const baseArgs: string[] = isDirectMode
  ? ['-direct', '-txt']
  : ['-mod', outputModName, '-txt'];
```

D2RMM **仍然使用** `-mod` 或 `-direct -txt` 參數，表面上與手動啟動相同。

### CASC 資料提取

**檔案：** `src/main/worker/BridgeAPI.ts`

D2RMM 使用 [CascLib](http://www.zezula.net/en/casc/main.html)（C++ 原生庫）開啟 D2R 的 CASC 儲存庫，提取原版遊戲資料：

```typescript
// 開啟 CASC 儲存庫
BridgeAPI.openStorage()  // 路徑: "${gamePath}:osi"

// 提取單一檔案到記憶體
BridgeAPI.extractFileToMemory(filePath)
```

### Mod 安裝流程

1. 開啟 CASC 儲存庫
2. 對每個啟用的 Mod 執行其 `mod.js` 腳本
3. Mod 腳本透過 D2RMM API 讀取原版資料→修改→寫回
4. 所有 Mod 執行完畢後，將修改過的檔案寫入磁碟
5. **Mod 模式**：寫入 `mods/D2RMM/D2RMM.mpq/data/...` + 產生 `modinfo.json`
6. **Direct 模式**：直接寫入遊戲 `Data/` 目錄（覆蓋原版檔案）

### JSON 處理

**檔案：** `src/main/worker/JSONParser.ts`

- 讀取時**移除** UTF-8 BOM：`.replace(/^\uFEFF/, '')`
- 寫入時**不加** BOM，使用 `JSON.stringify()` 最小化輸出
- 原始碼註解：*"D2R doesn't actually care if BOM is there or not"*

### datamod 模組

**檔案：** `src/main/worker/datamod.ts`

僅是一段簡單的內建 Mod 腳本，將資料夾從 Mod 來源複製到輸出目錄。不涉及特殊 CASC 操作。

### 關鍵推論

D2RMM 使用的啟動參數與我們手動測試的相同，但 D2RMM 的 **Mod 安裝流程**透過 CascLib 提取完整原版資料並合併修改，可能產出的**檔案格式或完整性**有我們無法複製的差異。需要實際安裝 D2RMM 驗證。

---

## 解決方案：D2RMM 驗證流程

### 步驟 1：下載 D2RMM

- **穩定版 v1.8.0**：[GitHub Release](https://github.com/olegbl/d2rmm/releases/download/v1.8.0/D2RMM.1.8.0.zip)
- **預覽版 v1.9.0**：[GitHub Releases 頁面](https://github.com/olegbl/d2rmm/releases)
- 來源：[Nexus Mods](https://www.nexusmods.com/diablo2resurrected/mods/169) 或 [GitHub](https://github.com/olegbl/d2rmm)

### 步驟 2：解壓與設定

1. 解壓到任意位置（例如 `D:\Tools\D2RMM\`），D2RMM 為可攜式免安裝
2. 執行 `D2RMM.exe`
3. 進入 Settings → 設定 D2R 安裝路徑（例如 `D:\Blizzard\Diablo II Resurrected\`）

### 步驟 3：建立測試 Mod

在 D2RMM 的 `mods/` 資料夾建立：

**`mods/TestMod/mod.json`**：
```json
{
  "name": "TestMod",
  "description": "Verify mod loading works",
  "author": "test",
  "version": "1.0"
}
```

**`mods/TestMod/mod.js`**：
```js
const fileName = 'local\\lng\\strings\\item-names.json';
const itemNames = D2RMM.readJson(fileName);
itemNames.forEach(entry => {
  if (entry.Key === 'hp1') {
    entry.zhTW = 'ÿc1HP1';  // 紅色 HP1
    entry.enUS = 'ÿc1HP1';
  }
});
D2RMM.writeJson(fileName, itemNames);
```

### 步驟 4：安裝並啟動

1. 在 D2RMM 內勾選 **TestMod**
2. 點擊 **Install Mods**（等待安裝完成）
3. 點擊 **Play** 啟動遊戲

### 步驟 5：驗證

進入遊戲後：
- 找到「弱效生命藥水 (Minor Healing Potion)」
- 若顯示為**紅色 `HP1`** → ✅ D2RMM Mod 載入正常
- 若仍顯示「弱效生命藥水」→ ❌ D2RMM 也無法在此版本載入 Mod

---

## D2RMM 技術原理

```
┌─────────────────────────────────────────────────────┐
│  D2RMM 安裝流程                                     │
│                                                      │
│  1. CascLib 開啟 D2R CASC 儲存庫                     │
│  2. 對每個啟用的 Mod 執行 mod.js                     │
│     └─ mod.js 透過 D2RMM API:                       │
│        ├─ D2RMM.readJson()   → CascLib 提取原版     │
│        ├─ JavaScript 修改內容                        │
│        └─ D2RMM.writeJson()  → 寫入記憶體           │
│  3. 寫入磁碟                                         │
│     ├─ Mod 模式: mods/D2RMM/D2RMM.mpq/data/...     │
│     └─ Direct 模式: Data/ (覆蓋原版)                │
│  4. 啟動 D2R                                         │
│     ├─ Mod 模式: D2R.exe -mod D2RMM -txt            │
│     └─ Direct 模式: D2R.exe -direct -txt             │
└─────────────────────────────────────────────────────┘
```

---

## 對本工具的影響

### 已確認不可行的方式

| 方式 | 說明 | 結果 |
|------|------|------|
| 手動複製 Mod 到 `mods/` + `-mod <name> -txt` | 傳統 Mod 載入方式 | ❌ D2R v3.1 完全忽略 |
| 手動複製 Mod 到 `mods/` + `-mod <name>` | 不帶 `-txt` | ❌ 同樣失效 |
| 完整 JSON + `-mod <name> -txt` | 使用完整原版+修改合併的 JSON | ❌ 同樣失效 |
| UTF-8 BOM 修正 + 完整路徑 | 確保編碼正確、工作目錄正確 | ❌ 問題非編碼/路徑，而是參數被禁用 |

### 本工具現有功能調整

- 離線模式（選項 0）的 `-mod <name> -txt` 啟動方式在 RotW 版本**無效**
- Mod 安裝功能（選項 m）的檔案複製仍可正常運作，但複製後的 Mod 無法被 D2R 載入
- 多開、帳號管理、視窗切換等核心功能**不受影響**

### 未來整合方向（待 D2RMM 驗證通過後決定）

| 方案 | 說明 | 複雜度 |
|------|------|--------|
| A. 轉換為 D2RMM 格式 | 將 d2r-hyper-show 輸出為 D2RMM 的 `mod.json` + `mod.js` 格式 | 低 |
| B. 整合 D2RMM 安裝 | 自動偵測 D2RMM 路徑，將 Mod 安裝到 D2RMM 的 mods 資料夾 | 中 |
| C. 實作 CASC 操作 | 使用 CascLib 在 Go 中自行提取/合併/寫入 | 高 |

---

## 參考資源

- [D2RMM GitHub](https://github.com/olegbl/d2rmm) — 原始碼與 Release
- [D2RMM Nexus Mods](https://www.nexusmods.com/diablo2resurrected/mods/169) — 下載與社群討論
- [D2RMM 文件](https://olegbl.github.io/d2rmm/index.html) — 官方使用文件
- [CascLib](http://www.zezula.net/en/casc/main.html) — Blizzard CASC 格式操作庫
- [D2R Patch 3.1.1 更新](https://www.d2itemstore.com/blogs/d2r-news/d2r-patch-3-1-1-reign-of-the-warlock-update-breakdown) — RotW 更新說明

# easy-cube — D2R Horadric Cube 自訂配方 Mod

⚠️ **僅限離線/單人模式使用 (OFFLINE/SINGLE-PLAYER ONLY)**

## 說明

本 mod 新增三個 Horadric Cube 合成配方，讓單機遊戲中的物品取得更便利。

### 新增配方

| # | 材料 | 產出 | 說明 |
|---|------|------|------|
| 1 | 完美寶石 (Perfect Gem) ×3 | Ist 符文 (`r24`) | 任意 3 顆完美寶石合成一顆 Ist 高階符文 |
| 2 | 任意暗金物品 (Unique) + 精煉寶石 (Flawless Gem) ×1 | 隨機同類型暗金物品 | 重擲暗金物品屬性，產出為同基底類型的暗金 |
| 3 | El 符文 (`r01`) ×33 | Zod 符文 (`r33`) | 33 顆最低階符文直升最高階符文 |

## 修改的檔案

| 檔案 | 路徑 | 變更內容 |
|------|------|----------|
| `cubemain.txt` | `easy-cube.mpq/data/global/excel/` | 新增 3 筆 Cube 配方列 |
| `modinfo.json` | `easy-cube/` | Mod 設定，存檔隔離至 `easy-cube/` |

### cubemain.txt 欄位說明

- **Recipe 1**: `gem4,qty=3` → `r24` — `gem4` 為完美等級寶石類型，`r24` = Ist
- **Recipe 2**: `any,uni` + `gem3` → `usetype,uni` — `any,uni` 匹配任何暗金物品，`gem3` 為精煉寶石，`usetype,uni` 輸出同基底暗金
- **Recipe 3**: `r01,qty=33` → `r33` — `r01` = El，`r33` = Zod

## 安裝方式

1. 將 `easy-cube/` 資料夾複製到 D2R 安裝目錄下的 `mods/` 資料夾
   ```
   <D2R 安裝路徑>/mods/easy-cube/
   ├── modinfo.json
   └── easy-cube.mpq/
       └── data/global/excel/cubemain.txt
   ```

2. 啟動遊戲時加上參數：
   ```
   -mod easy-cube -txt
   ```

3. 首次成功載入後，後續可省略 `-txt` 以加快載入速度。

## ⚠️ 重要警告

- 🚫 **僅限離線/單人模式使用 (OFFLINE/SINGLE-PLAYER ONLY)**
- 🚫 **在 Battle.net 使用這些 mod 會導致帳號永久封禁**
- ⚠️ 測試前請先備份存檔
- ⚠️ 本 mod 使用獨立存檔路徑（`modinfo.json` 中 `savepath: "easy-cube/"`），不會影響原版存檔

# Easy Cube — D2R Mod

自訂 Horadric Cube 配方，讓合成更方便。

## 新增配方

| # | 輸入 | 輸出 | 說明 |
|---|------|------|------|
| 1 | 完美寶石 ×3 | Ist 符文 (`r24`) | 任意三顆完美寶石（Perfect Gem）合成一顆 Ist 符文 |
| 2 | 任意暗金物品 + 精煉寶石 ×1 | 隨機暗金物品 | 暗金（Unique）+ 任意精煉寶石（Flawless Gem）→ 同類型隨機暗金物品 |
| 3 | El 符文 ×33 | Zod 符文 (`r33`) | 33 顆 El 符文（`r01`）直接合成一顆 Zod 符文 |

## 修改的檔案

| 檔案 | 路徑 | 說明 |
|------|------|------|
| [`cubemain.txt`](easy-cube.mpq/data/global/excel/cubemain.txt) | `data/global/excel/` | 新增 3 筆 Cube 配方 |

### cubemain.txt 欄位說明

- **Recipe 1** — `input 1: gem4,qty=3` → `output: r24`。`gem4` 代表 Perfect 等級寶石，`r24` = Ist Rune。
- **Recipe 2** — `input 1: any,uni`（任意暗金物品）+ `input 2: gem3`（Flawless Gem）→ `output: usetype,uni`（輸出與輸入同類型的暗金物品）。
- **Recipe 3** — `input 1: r01,qty=33` → `output: r33`。`r01` = El Rune，`r33` = Zod Rune。

## 安裝方式

1. 將 `easy-cube/` 整個資料夾複製到 D2R 安裝目錄下的 `mods/` 資料夾：
   ```
   <D2R 安裝路徑>/mods/easy-cube/
   ```
2. 啟動遊戲時加上參數：
   ```
   -mod easy-cube -txt
   ```
3. 首次成功載入後，後續啟動可移除 `-txt` 加速載入。

## ⚠️ 注意事項

- **僅限單機 / 離線模式使用** — 線上使用可能導致帳號封禁。
- **測試前請備份存檔** — 存檔位於 `easy-cube/` 子資料夾（已啟用存檔隔離）。
- 此 mod 不修改任何原版配方，僅新增 3 筆自訂配方。

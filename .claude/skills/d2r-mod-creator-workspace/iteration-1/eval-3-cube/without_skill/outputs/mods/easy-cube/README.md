# Easy Cube — D2R Horadric Cube Mod

新增三個自訂 Horadric Cube 配方，讓符文與暗金裝備的取得更加便利。

## 配方一覽

| # | 輸入 | 輸出 | 說明 |
|---|------|------|------|
| 1 | 完美寶石 ×3 | Ist 符文 | 任意 3 顆完美寶石（Perfect Gem）合成 1 個 Ist 符文（`r24`） |
| 2 | 任意暗金物品 + 精煉寶石 ×1 | 隨機暗金物品 | 任何 Unique 裝備搭配 1 顆 Flawless Gem，產出同類型隨機暗金物品 |
| 3 | El 符文 ×33 | Zod 符文 | 33 顆 El 符文（`r01`）直接合成 1 個 Zod 符文（`r33`） |

## 安裝方式

1. 將 `easy-cube` 資料夾複製到 D2R 安裝目錄下的 `mods/` 資料夾：
   ```
   <D2R安裝路徑>/mods/easy-cube/
   ├── modinfo.json
   └── easy-cube.mpq/
       └── data/
           └── global/
               └── excel/
                   └── cubemain.txt
   ```
2. 使用以下參數啟動 D2R：
   ```
   D2R.exe -mod easy-cube -txt
   ```

## 檔案說明

| 檔案 | 用途 |
|------|------|
| [modinfo.json](modinfo.json) | Mod 元資料（名稱、存檔路徑、版本） |
| [easy-cube.mpq/data/global/excel/cubemain.txt](easy-cube.mpq/data/global/excel/cubemain.txt) | Horadric Cube 配方定義（Tab-separated） |

## 物品代碼參考

| 代碼 | 物品 |
|------|------|
| `gem4` | 完美寶石（Perfect Gem，任意種類） |
| `gem3` | 精煉寶石（Flawless Gem，任意種類） |
| `r01` | El 符文 |
| `r24` | Ist 符文 |
| `r33` | Zod 符文 |
| `any,uni` | 任意暗金（Unique）物品 |
| `usetype,uni` | 產出同基底類型的隨機暗金物品 |

## 注意事項

- 配方 2 的輸出取決於輸入暗金物品的基底類型，系統會從該類型的暗金池中隨機選取一件。
- 配方 3 的 33 顆 El 符文需在 Cube 中一次放入（Cube 空間需足夠）。
- 本 Mod 僅新增配方，不修改原版任何現有配方。
- 相容 D2R 最新版本，支援 `-txt` 模式載入。

# high-drop — Hell 難度 5 倍掉落率 Mod

⚠️ **僅限離線 / 單人模式使用 (OFFLINE / SINGLE-PLAYER ONLY)**

## 功能說明

此 mod 將 Hell 難度的物品掉落率提升約 **5 倍**，並**特別加強高階符文 (High Runes)** 的掉落機率。

### 修改內容

| 修改項目 | 原始值 | 修改後 | 說明 |
|----------|--------|--------|------|
| Hell Act Boss `NoDrop` | 100 | **0** | Boss 每次 roll 都必定掉落物品 |
| Hell Act Boss `Picks` | 5-7 | **7-9** | 增加每次擊殺的掉落次數 |
| Hell Super Unique `NoDrop` | 10 | **2** | Super Unique 怪掉率提升 5 倍 |
| Countess Rune (H) `Picks` | 5 | **9** | Countess 符文掉落次數大幅提升 |
| Runes 12-17 高階符文機率 | Prob=2 | **Prob=10** | 高階符文 (Lo~Zod) 機率提升 5 倍 |

### 修改的檔案

- **`treasureclassex.txt`** — 掉落表 (Treasure Class) 設定
  - Hell Act Boss TC (Andariel/Duriel/Mephisto/Diablo/Baal，含 Quest 版本): `NoDrop=0`, `Picks` 增加
  - Hell Super Unique TC (Nihlathak 等): `NoDrop` 降低至 1/5
  - Countess Rune (H): `Picks` 從 5 提升至 9
  - Runes 12~17: 每一層高階符文的機率權重從 2 提升至 10 (5 倍)

### 受影響的高階符文

| 符文 | 代碼 | 等級 |
|------|------|------|
| Lo   | r28  | 12   |
| Sur  | r29  | 13   |
| Ber  | r30  | 14   |
| Jah  | r31  | 15   |
| Cham | r32  | 16   |
| Zod  | r33  | 17   |

## 安裝方式

1. 將 `high-drop/` 整個資料夾複製到 D2R 安裝目錄下的 `mods/` 資料夾：
   ```
   <D2R 安裝路徑>/mods/high-drop/
   ```

2. 啟動 D2R 時加上參數：
   ```
   -mod high-drop -txt
   ```

3. 第一次成功載入後，後續啟動可移除 `-txt` 以加速載入。

## ⚠️ 重要警告

- 🚫 **僅限離線/單人模式使用 (OFFLINE/SINGLE-PLAYER ONLY)**
- 🚫 **在 Battle.net 使用這些 mod 會導致帳號永久封禁**
- ⚠️ 測試前請先備份存檔
- ⚠️ 本 mod 使用獨立存檔路徑 (`savepath: "high-drop/"`)，不會影響原版存檔

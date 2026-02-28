# high-drop — D2R Hell 難度 5 倍掉落率 Mod

## 簡介

此 Mod 將 Diablo II: Resurrected **Hell 難度**的掉落率提升約 **5 倍**，特別加強了**高階符文 (High Runes)** 的掉落機率。適合單機/離線模式刷裝使用。

## 修改內容

### 修改的檔案

| 檔案 | 說明 |
|------|------|
| [`treasureclassex.txt`](high-drop.mpq/data/global/excel/treasureclassex.txt) | 掉落表 — 調整 NoDrop 權重、Boss/Champion/Unique 怪掉落、符文掉落機率 |

### 具體變更

#### 1. NoDrop 大幅降低（一般怪物）

| 對象 | 原始 NoDrop | 修改後 NoDrop | 效果 |
|------|------------|--------------|------|
| Hell Act 1-5 一般怪物 (H2H) | ~100 | **2** | 幾乎每次擊殺都會掉落物品 |
| Hell Act 1-5 冠軍怪 (Champ) | ~20 | **2** | 冠軍怪掉落大幅提升 |
| Hell Act 1-5 獨特怪 (Unique) | ~20 | **2** | 獨特怪掉落大幅提升 |

#### 2. Boss 掉落強化

| Boss | 原始 Picks | 修改後 Picks | 原始 NoDrop | 修改後 NoDrop |
|------|-----------|-------------|------------|--------------|
| Andariel (H) | 5 | **7** | ~15 | **2** |
| Duriel (H) | 5 | **7** | ~15 | **2** |
| Mephisto (H) | 5 | **7** | ~15 | **2** |
| Diablo (H) | 7 | **7** | ~15 | **2** |
| Baal (H) | 7 | **7** | ~15 | **2** |
| Nihlathak (H) | 5 | **5** | ~15 | **2** |
| Council (H) | 5 | **5** | ~15 | **2** |

#### 3. 高階符文掉落加強

| 對象 | 變更 |
|------|------|
| Countess Rune (H) | Picks 提升至 **5**，NoDrop 設為 **0**（必定掉符文），可掉落 Vex(r26) 到 Zod(r33) |
| Runes 17 (最高符文表) | NoDrop 從 ~19 降至 **2**，提高 Lo/Sur/Ber/Jah/Cham/Zod 出現權重 |

#### 4. Unique/Set 物品機率提升

- 所有 Hell 怪物的 `Unique` 和 `Set` 欄位提升至 **800-983**（原始約 512-800）
- 增加暗金和套裝物品的掉落比例

## 安裝方式

1. 將 `high-drop/` 整個資料夾複製到 D2R 安裝目錄下的 `mods/` 資料夾：
   ```
   <D2R 安裝路徑>/mods/high-drop/
   ```

2. 確認目錄結構如下：
   ```
   mods/
   └── high-drop/
       ├── modinfo.json
       └── high-drop.mpq/
           └── data/
               └── global/
                   └── excel/
                       └── treasureclassex.txt
   ```

3. 啟動遊戲時加上以下參數：
   ```
   -mod high-drop -txt
   ```

4. 首次成功啟動後，後續可省略 `-txt` 參數以加快載入速度。

## 使用 d2r-multiboxing 啟動

如果你使用 [d2r-multiboxing](../../../../../../README.md) 工具，可在設定中加入啟動參數：
```
-mod high-drop -txt
```

## ⚠️ 注意事項

- **僅限離線/單機模式使用** — 在線上模式使用 Mod 可能導致帳號被封禁
- **備份存檔** — 測試前請先備份 `%USERPROFILE%/Saved Games/Diablo II Resurrected/` 目錄
- **存檔隔離** — 此 Mod 使用獨立存檔路徑 (`high-drop/`)，不會影響原版存檔
- **相容性** — 適用於 D2R 最新版本，若遊戲更新後異常請重新生成 Mod

## 技術細節

此 Mod 僅修改 `treasureclassex.txt`，原理如下：

- `NoDrop` 欄位控制「什麼都不掉」的權重。原版 Hell 一般怪的 NoDrop 約為 100，
  與其他物品機率（約 21）相比，大部分時候都不掉東西。將 NoDrop 降至 2，
  使掉落機率從約 ~17% 提升至 ~90%+，達成約 5 倍的整體掉落提升。
- `Picks` 欄位控制每次擊殺的掉落次數，Boss 的 Picks 提升代表每次擊殺掉更多物品。
- 高階符文表的權重調整使 Ber、Jah、Zod 等稀有符文更容易出現。

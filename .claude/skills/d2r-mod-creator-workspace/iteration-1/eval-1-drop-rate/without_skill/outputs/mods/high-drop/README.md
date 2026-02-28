# high-drop — D2R Hell 難度掉落率提升 Mod

提升 Hell 難度掉落率約 **5 倍**，特別針對**高階符文 (High Runes)** 進行強化。

## 安裝方式

1. 將 `high-drop` 資料夾複製到 D2R 安裝目錄下的 `mods/` 資料夾
2. 使用以下參數啟動 D2R：
   ```
   D2R.exe -mod high-drop -txt
   ```
3. 存檔會獨立儲存於 `high-drop/` 子目錄，不影響原版存檔

## 修改內容

### 1. TreasureClassEx.txt — 寶藏等級掉落表

| 類別 | 修改方式 | 效果 |
|------|----------|------|
| **Rune TCs (Runes 12–17)** | 升級機率 ×5 | Mal~Zod 高階符文掉落率大幅提升 |
| **Hell Act Boss TCs** | NoDrop 15 → 3 | Boss 掉寶機率提升約 5 倍 |
| **Hell Champion/Unique TCs** | NoDrop 大幅降低 | 菁英怪掉寶率顯著提升 |
| **Hell 一般怪物 TCs** | NoDrop 60 → 12 | 一般怪物也有更多掉落 |
| **Countess (H)** | 符文 TC picks=3 | 女伯爵掉落更多高階符文 |
| **Uber Bosses** | NoDrop=0, Runes 17 | 保證掉落，含最高階符文機率 |

### 2. ItemRatio.txt — 物品品質機率表

| 品質 | 原版基礎值 | Mod 值 | 倍率 |
|------|-----------|--------|------|
| Unique | 400 / 160 | 80 / 32 | ×5 |
| Set | 500 / 200 | 100 / 40 | ×5 |
| Rare | 200 / 80 | 40 / 16 | ×5 |
| Magic | 60 / 30 | 12 / 6 | ×5 |

> 左為一般怪物，右為 Boss/Champion

### 符文代碼對照 (高階)

| 代碼 | 符文 | TC 層級 |
|------|------|---------|
| r23 | Mal | Runes 12 |
| r24 | Ist | Runes 12 |
| r25 | Gul | Runes 13 |
| r26 | Vex | Runes 13 |
| r27 | Ohm | Runes 14 |
| r28 | Lo  | Runes 14 |
| r29 | Sur | Runes 15 |
| r30 | Ber | Runes 15 |
| r31 | Jah | Runes 16 |
| r32 | Cham | Runes 16/17 |
| r33 | Zod | Runes 17 |

## 目錄結構

```
mods/high-drop/
├── modinfo.json                          # Mod 基本資訊與存檔路徑
├── README.md                             # 本說明文件
└── high-drop.mpq/
    └── data/global/excel/
        ├── TreasureClassEx.txt           # 寶藏等級掉落表（含符文 TC）
        └── ItemRatio.txt                 # 物品品質機率表
```

## 注意事項

- 此 Mod 僅修改 **Hell 難度**相關的 Treasure Class，Normal / Nightmare 不受影響
- **TreasureClassEx.txt** 包含 Hell 難度的關鍵 TC 定義（Boss、Champion、Unique、一般怪物、符文鏈、Uber Boss 等共 57 條規則）
- 若需完整覆蓋所有 TC，請以原版 `TreasureClassEx.txt` 為基礎合併本 Mod 的修改
- 使用 `-txt` 參數啟動時，D2R 會從 `.txt` 檔載入資料，覆蓋預設值
- 存檔獨立於原版（`savepath: "high-drop/"`），可安心切換

## 適用版本

- Diablo II: Resurrected (D2R) — 支援離線模式 (`-txt` 載入)

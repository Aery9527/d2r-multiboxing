# easy-cube — D2R Horadric Cube Mod

自訂 Horadric Cube 配方，讓合成更輕鬆。

## 配方一覽

| # | 輸入 | 輸出 | 說明 |
|---|------|------|------|
| 1 | 完美寶石 ×3（同類型） | Ist 符文 | 支援所有 7 種完美寶石（含 Perfect Skull） |
| 2 | 任意暗金(Unique)物品 + 無瑕寶石(Flawless Gem) ×1 | 隨機暗金物品 | 以同基底類型重新擲骰為 Unique 品質 |
| 3 | El 符文 ×33 | Zod 符文 | ⚠️ 需搭配擴大 Cube 容量的 mod（原版僅 4×3=12 格） |

## 安裝方式

1. 將 `easy-cube/` 資料夾複製到 D2R 安裝目錄的 `mods/` 下。
2. 以下列參數啟動 D2R：
   ```
   D2R.exe -mod easy-cube -txt
   ```

## 重要注意事項

- 本 mod 的 `cubemain.txt` **僅包含自訂配方**，不含原版配方。
- 若要保留原版 Cube 配方，請先從遊戲資料中擷取原版 `cubemain.txt`，
  再將本 mod 的配方列 **附加 (append)** 至檔案末尾。
- 配方 3（El ×33 → Zod）需要 33 格空間，超過原版 Cube 的 12 格上限，
  需搭配其他擴大 Cube 的 mod 才能實際使用。

## 檔案結構

```
mods/easy-cube/
├── modinfo.json
├── README.md
└── easy-cube.mpq/
    └── data/global/excel/
        └── cubemain.txt
```

## 物品代碼參考

### 完美寶石 (Perfect Gems)
| 代碼 | 名稱 |
|------|------|
| gpv | Perfect Amethyst |
| gpw | Perfect Diamond |
| gpg | Perfect Emerald |
| gpr | Perfect Ruby |
| gpb | Perfect Sapphire |
| gpy | Perfect Topaz |
| skz | Perfect Skull |

### 無瑕寶石 (Flawless Gems)
| 代碼 | 名稱 |
|------|------|
| glv | Flawless Amethyst |
| glw | Flawless Diamond |
| glg | Flawless Emerald |
| glr | Flawless Ruby |
| glb | Flawless Sapphire |
| gly | Flawless Topaz |
| skl | Flawless Skull |

### 符文 (Runes)
| 代碼 | 名稱 |
|------|------|
| r01 | El |
| r24 | Ist |
| r33 | Zod |

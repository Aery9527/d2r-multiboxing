# d2r-info Mod

D2R 資訊輔助 Mod — 在遊戲內顯示 FCR/FHR/FBR 斷點表與常用 Cube 配方。

## 功能

### 1. FCR/FHR/FBR Breakpoints (斷點表)
- 當裝備上有 Faster Cast Rate / Faster Hit Recovery / Faster Block Rate 屬性時，tooltip 會額外顯示 Sorceress 和 Paladin 的斷點表
- 包含 Sorceress Lightning 的獨立 FCR 斷點
- 包含 Paladin Holy Shield 的 FBR 斷點

### 2. Horadric Cube 配方
- 滑鼠移到 Horadric Cube 上會顯示常用配方速查表：
  - 符文升級規則
  - 打孔配方 (武器/盔甲/頭盔/盾牌)
  - 洗裝備配方 (魔法/稀有)
  - 暗金裝備升級配方

### 3. 符文升級路徑
- 每個符文的 tooltip 會顯示對應的升級配方
- 包含所需的寶石類型

## 安裝

1. 將 `d2r-info` 資料夾複製到 D2R 安裝目錄下的 `mods/` 資料夾
2. 啟動 D2R 時加上參數: `D2R.exe -mod d2r-info -txt`

## 資料夾結構

```
d2r-info/
└── d2r-info.mpq/
    ├── modinfo.json
    └── data/local/lng/strings/
        ├── d2r-info-breakpoints.json  (FCR/FHR/FBR 斷點表)
        ├── d2r-info-cube.json         (Cube 配方速查)
        └── d2r-info-runes.json        (符文升級路徑)
```

## 斷點資料

### Sorceress FCR
| FCR | 0 | 9 | 20 | 37 | 63 | 105 | 200 |
|-----|---|---|----|----|----|----|-----|
| Frames | 13 | 12 | 11 | 10 | 9 | 8 | 7 |

### Sorceress FCR (Lightning/Chain Lightning)
| FCR | 0 | 7 | 15 | 23 | 35 | 52 | 78 | 117 | 194 |
|-----|---|---|----|----|----|----|----|----|-----|
| Frames | 19 | 18 | 17 | 16 | 15 | 14 | 13 | 12 | 11 |

### Paladin FCR
| FCR | 0 | 9 | 18 | 30 | 48 | 75 | 125 |
|-----|---|---|----|----|----|----|-----|
| Frames | 15 | 14 | 13 | 12 | 11 | 10 | 9 |

### Sorceress FHR
| FHR | 0 | 5 | 9 | 14 | 20 | 30 | 42 | 60 | 86 | 142 |
|-----|---|---|---|----|----|----|----|----|----|-----|
| Frames | 15 | 14 | 13 | 12 | 11 | 10 | 9 | 8 | 7 | 6 |

### Paladin FHR
| FHR | 0 | 7 | 15 | 27 | 48 | 86 |
|-----|---|---|----|----|----|----|
| Frames | 9 | 8 | 7 | 6 | 5 | 4 |

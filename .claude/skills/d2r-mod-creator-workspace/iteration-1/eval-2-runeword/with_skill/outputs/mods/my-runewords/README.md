# my-runewords — D2R 自訂符文之語 Mod

## 簡介

此 Mod 為 Diablo II: Resurrected 新增一個自訂符文之語：

### 龍息 (Dragon Breath)

| 項目 | 內容 |
|------|------|
| 符文組合 | **Ber + Jah + Cham** (3 孔) |
| 可裝備物品 | 所有武器 (`weap`) |
| 效果 | +3 所有技能 |
| | +40% 攻擊速度 |
| | 15% 生命偷取 |
| | +300% 傷害 |

## 檔案結構

```
my-runewords/
├── modinfo.json                          ← Mod 設定（名稱、存檔路徑）
├── README.md                             ← 本文件
└── my-runewords.mpq/
    └── data/
        ├── global/
        │   └── excel/
        │       └── runes.txt             ← 符文之語定義（新增 Dragon Breath）
        └── local/
            └── lng/
                └── strings/
                    └── item-runes.json   ← 符文之語名稱多語系字串
```

## 修改內容

### `runes.txt`（新增行）

新增一列定義符文之語 Dragon Breath：

| 欄位 | 值 | 說明 |
|------|----|------|
| `Name` | `DragonBreath` | 字串 Key（對應 `item-runes.json`） |
| `complete` | `1` | 啟用此符文之語 |
| `itype1` | `weap` | 可裝在所有武器上 |
| `Rune1` | `r30` | Ber |
| `Rune2` | `r31` | Jah |
| `Rune3` | `r32` | Cham |
| `T1Code1` / `T1Min1` / `T1Max1` | `allskills` / `3` / `3` | +3 所有技能 |
| `T1Code2` / `T1Min2` / `T1Max2` | `swing2` / `40` / `40` | +40% 攻擊速度 |
| `T1Code3` / `T1Min3` / `T1Max3` | `lifesteal` / `15` / `15` | 15% 生命偷取 |
| `T1Code4` / `T1Min4` / `T1Max4` | `dmg%` / `300` / `300` | +300% 傷害 |

### `item-runes.json`（新增項目）

| 欄位 | 值 |
|------|----|
| `id` | `90001`（高 ID 避免衝突） |
| `Key` | `DragonBreath` |
| `enUS` | `Dragon Breath` |
| `zhTW` | `龍息` |

## 安裝方式

1. 將 `my-runewords/` 整個資料夾複製到 D2R 安裝目錄下的 `mods/` 資料夾：
   ```
   <D2R 安裝路徑>/mods/my-runewords/
   ```

2. 使用以下參數啟動遊戲：
   ```
   -mod my-runewords -txt
   ```

3. 首次成功載入後，後續啟動可移除 `-txt` 參數以加快載入速度。

## ⚠️ 注意事項

- **僅限離線 / 單人模式使用** — 在線上模式使用 Mod 可能導致帳號被封禁。
- **測試前請備份存檔** — 存檔位於 `%USERPROFILE%/Saved Games/Diablo II Resurrected/` 目錄。
- 此 Mod 使用獨立存檔路徑 (`my-runewords/`)，不會影響原版存檔。

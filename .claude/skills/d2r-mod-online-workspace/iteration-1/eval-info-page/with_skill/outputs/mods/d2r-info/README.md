# d2r-info — D2R 遊戲內資訊顯示 Mod

✅ **Online-safe**: 此 Mod 僅修改顯示字串（JSON string files），不影響任何遊戲機制。

## 功能

### 📖 Breakpoint 資訊（技能樹分頁 Tooltip）

- **Sorceress Fire Spells 分頁** — 顯示女巫 FCR（一般 & 閃電系）、FHR、FBR 完整 breakpoint 表
- **Paladin Combat Skills 分頁** — 顯示聖騎士 FCR、FHR、FBR（一般 & 神聖之盾）完整 breakpoint 表

常用目標 breakpoint 以 ÿc4金色 標示，頂級 breakpoint 以 ÿc1紅色 標示。

### 📖 Cube 配方資訊（赫拉迪克方塊 Tooltip）

滑鼠移到赫拉迪克方塊上即可看到常用配方：

- **符文升級** — 各階段合成規則
- **洗裝備** — 重骰魔法/稀有裝備
- **打孔** — 普通防具/武器打孔配方
- **升級** — 普通→精華→菁英防具升級
- **洗點** — 4 精華合成赦免之符

## 安裝方式

1. 將 `d2r-info` 資料夾複製到 `<D2R 安裝目錄>/mods/`
2. 首次啟動：使用參數 `-mod d2r-info -txt`
3. 之後啟動：使用參數 `-mod d2r-info`

## 修改的檔案

| 檔案 | 內容 |
|------|------|
| `skills.json` | Sorceress / Paladin 技能樹分頁加入 breakpoint 表 |
| `item-names.json` | 赫拉迪克方塊名稱加入 Cube 配方速查表 |

## 注意事項

- 此 Mod 不修改任何 `.txt` 資料表，不影響遊戲平衡與掉寶率
- 使用 `"savepath": "../"` 與原版共用存檔
- 雖然屬於最安全的 Mod 類型，Battle.net 上使用任何 Mod 仍有理論風險，請自行評估

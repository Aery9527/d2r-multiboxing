# my-filter — D2R Loot Filter

自訂 D2R 戰利品篩選器，僅修改 JSON 字串檔，不影響遊戲機制。

✅ **Online-safe**: This mod only modifies display strings — safe for Battle.net.

## 功能

| 分類 | 效果 | 顏色 |
|------|------|------|
| 高階符文 (Vex–Zod) | ÿc1 紅色 + ★ 星號標記 | 🔴 Red |
| 中階符文 (Pul–Gul) | ÿc8 橙色 + ★ 標記 | 🟠 Orange |
| 藥水 | 縮寫顯示 (例: SHP, SMP, FRJ) | 🔴/🔵/🟣 |
| 寶石 | ÿc3 藍色標記，完美寶石加 ★ | 🔵 Blue |

### 符文分級

- **S-Tier** ★★★：Ber (#30), Jah (#31), Cham (#32), Zod (#33)
- **A-Tier** ★★：Vex (#26), Ohm (#27), Lo (#28), Sur (#29)
- **B-Tier** ★：Pul (#21), Um (#22), Mal (#23), Ist (#24), Gul (#25)

### 藥水縮寫對照

| 原名 | 縮寫 (enUS) | 縮寫 (zhTW) |
|------|------------|------------|
| Minor Healing Potion | mHP | 微紅 |
| Light Healing Potion | lHP | 小紅 |
| Healing Potion | HP | 中紅 |
| Greater Healing Potion | GHP | 大紅 |
| Super Healing Potion | SHP | 超紅 |
| Minor Mana Potion | mMP | 微藍 |
| Light Mana Potion | lMP | 小藍 |
| Mana Potion | MP | 中藍 |
| Greater Mana Potion | GMP | 大藍 |
| Super Mana Potion | SMP | 超藍 |
| Rejuvenation Potion | RJ | 小紫 |
| Full Rejuvenation Potion | FRJ | 大紫 |

## 安裝方式

1. 複製 `my-filter` 資料夾至 `<D2R安裝目錄>/mods/`
2. 首次啟動使用：`-mod my-filter -txt`
3. 之後啟動使用：`-mod my-filter`

## 修改的檔案

| 檔案 | 內容 |
|------|------|
| [item-names.json](my-filter.mpq/data/local/lng/strings/item-names.json) | 藥水縮寫 + 寶石藍色標記 |
| [item-runes.json](my-filter.mpq/data/local/lng/strings/item-runes.json) | 高/中階符文顏色標記 |

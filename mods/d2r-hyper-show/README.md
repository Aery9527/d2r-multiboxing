# D2R Hyper Show

綜合型 Diablo II: Resurrected Display Mod — 結合 Loot Filter、Info Page、Name Shortener、Rune Numbering 四大功能。

✅ **Online-safe**：本 mod 僅修改 JSON 顯示字串，不改動任何遊戲數據表（`.txt`），可安全用於 Battle.net。

## 功能

### 🏷️ Loot Filter（物品分層上色）

| 等級 | 物品 | 顏色 | 標記 |
|------|------|------|------|
| S-Tier | Ber, Jah, Cham, Zod | 🔴 紅色 | ★★★ |
| A-Tier | Vex~Sur, Keys, Essences, Uber Organs | 🟠 橘色 | ★★ |
| B-Tier | Pul~Gul, Ring, Amulet, Jewel, Charms | 🔵 藍色 | ★ |
| C-Tier | 藥水, 卷軸 | ⚪ 灰色 | 縮寫 |

### 🔢 Rune Numbering（符文編號）

所有 33 個符文加上 `#N` 編號，方便快速辨識：
- `★★★ Ber Rune #30 ★★★`
- `★★ Vex Rune #26`
- `★ Ist Rune #24`
- `El Rune #1`

### ✂️ Name Shortener（名稱縮寫）

| 原名 | 縮寫 |
|------|------|
| Super Healing Potion | HP5 |
| Super Mana Potion | MP5 |
| Full Rejuvenation Potion | FRJ |
| Scroll of Town Portal | TP |
| Scroll of Identify | ID |

### 📖 Info Page（資訊頁嵌入）

- **FCR/FHR/FBR 斷點表**：8 個職業完整斷點資料（含術士/Spiritborn）嵌入技能 Tab 描述
- **Cube Recipe 速查**：符文升級、打孔、裝備升級、重置配方嵌入荷拉迪克方塊描述

## 安裝

1. 將 `d2r-hyper-show` 資料夾複製到 D2R 安裝目錄的 `mods/` 下：
   ```
   <D2R安裝目錄>/mods/d2r-hyper-show/
   ├── modinfo.json
   └── d2r-hyper-show.mpq/
       └── data/local/lng/strings/
           ├── item-names.json
           ├── skills.json
           └── ui.json
   ```

2. 啟動 D2R 時加上參數：
   - **首次使用**：`-mod d2r-hyper-show -txt`
   - **之後使用**：`-mod d2r-hyper-show`

   可在 Battle.net 啟動器設定中加入啟動參數，或建立捷徑：
   ```
   "D2R.exe" -mod d2r-hyper-show -txt
   ```

## 語言支援

- English (enUS)
- 繁體中文 (zhTW)

## 自訂修改

所有設定都在 `d2r-hyper-show.mpq/data/local/lng/strings/` 下的 JSON 檔案中：

- [`item-names.json`](d2r-hyper-show.mpq/data/local/lng/strings/item-names.json) — 物品名稱（Loot Filter + Rune Numbering + Name Shortener）
- [`skills.json`](d2r-hyper-show.mpq/data/local/lng/strings/skills.json) — 職業斷點表
- [`ui.json`](d2r-hyper-show.mpq/data/local/lng/strings/ui.json) — Cube Recipe 速查

每個 entry 格式：
```json
{
  "id": 11176,
  "Key": "r30",
  "enUS": "ÿc1★★★ Ber Rune #30 ★★★",
  "zhTW": "ÿc1★★★ 柏 符文 #30 ★★★"
}
```

修改 `enUS`/`zhTW` 的值即可自訂顯示效果。顏色代碼參考：

| 代碼 | 顏色 |
|------|------|
| `ÿc0` | 白色 |
| `ÿc1` | 紅色 |
| `ÿc2` | 綠色 |
| `ÿc3` | 藍色 |
| `ÿc4` | 金色 |
| `ÿc5` | 灰色 |
| `ÿc8` | 橘色 |
| `ÿc9` | 黃色 |
| `ÿc;` | 紫色 |

## 安全聲明

本 mod **僅修改顯示字串**，不改變任何遊戲機制、掉落率或數值平衡。存檔與原版共用（`savepath: "../"`），移除 mod 後遊戲完全恢復原狀。

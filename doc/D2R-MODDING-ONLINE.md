# D2R 線上安全 Mod 指南（顯示修改）

> 透過修改 JSON 字串檔來自訂 D2R 物品顯示——Loot Filter、符文編號、資訊頁面等

> ✅ **Battle.net 安全**——僅修改客戶端顯示文字，不改變遊戲機制

**前置知識：** 請先閱讀 [D2R Modding 共通指南](D2R-MODDING-COMMON.md)（目錄結構、工具安裝、啟動參數等）

---

## 目錄

- [線上安全 Mod 概述](#線上安全-mod-概述)
- [savepath 設定](#savepath-設定)
- [JSON 字串檔格式](#json-字串檔格式)
- [字串檔位置與用途](#字串檔位置與用途)
- [顏色代碼](#顏色代碼)
- [常見線上安全 Mod](#常見線上安全-mod)
- [進階技巧](#進階技巧)
- [線上安全 Mod 專屬注意事項](#線上安全-mod-專屬注意事項)

---

## 線上安全 Mod 概述

線上安全 Mod 僅修改 `data/local/lng/strings/` 下的 JSON 字串檔，這些檔案定義了遊戲中的**顯示文字**：

```
MyFilter/
├── modinfo.json                    ← savepath: "../"
└── MyFilter.mpq/
    └── data/
        └── local/
            └── lng/
                └── strings/        ← JSON 字串檔放這裡
                    ├── item-names.json
                    ├── item-runes.json
                    └── ...
```

由於只改變客戶端顯示（如物品名稱顏色、縮寫），**不影響遊戲伺服器邏輯**，因此可安全用於 Battle.net 線上遊戲。

### 安全原則

| ✅ 安全 | ❌ 不安全 |
|---------|----------|
| 修改 JSON 字串檔 | 修改 `.txt` 資料表 |
| 改變物品顯示名稱/顏色 | 改變掉落率、物品屬性 |
| 新增資訊頁面文字 | 新增 Cube 配方、Runeword |
| 符文編號、藥水縮寫 | 修改技能數值、怪物血量 |

---

## savepath 設定

線上安全 Mod 的 `modinfo.json` 應使用共用路徑：

```json
{
  "name": "MyFilter",
  "savepath": "../"
}
```

存檔位置與原版相同：
```
%UserProfile%\Saved Games\Diablo II Resurrected\
```

> 💡 使用 `"../"` 是因為線上安全 Mod 不改變遊戲機制，存檔資料完全相容。關閉 Mod 後角色照常使用，無任何風險。

---

## JSON 字串檔格式

D2R 使用 JSON 檔案管理遊戲內的文字字串（取代經典 D2 的 `.tbl` 格式）。

### 格式範例

```json
[
  {
    "id": 12345,
    "Key": "MyNewItem",
    "enUS": "Sword of Awesomeness",
    "zhTW": "超強之劍"
  }
]
```

### 欄位說明

| 欄位 | 說明 |
|------|------|
| `id` | 唯一數字 ID（新增條目時使用較大的數字避免衝突，如 70000+） |
| `Key` | 字串鍵值（在 `.txt` 檔中用 `namestr` 等欄位引用） |
| `enUS` | 英文文字 |
| `zhTW` | 繁體中文文字 |
| 其他語言 | `deDE`、`frFR`、`jaJP`、`koKR`、`plPL`、`esES`、`esMX`、`ptBR`、`ruRU`、`zhCN` |

### 修改 vs 新增

- **修改既有條目**：找到對應 `Key`，只改 `enUS`/`zhTW` 等語言欄位的值
- **新增條目**：在 JSON 陣列尾部追加新物件，使用不衝突的 `id`（建議 70000+）

### 常用字串編輯工具

- **Visual Studio Code** — 直接編輯 JSON，推薦安裝 JSON formatter 擴充
- **D2RModding-StrEdit**（[GitHub](https://github.com/eezstreet/D2RModding-StrEdit)）— 專門的 D2R 字串編輯器

---

## 字串檔位置與用途

```
data/local/lng/strings/
├── item-names.json        ← 物品名稱（藥水、卷軸、寶石、鑰匙等）
├── item-modifiers.json    ← 物品詞綴描述
├── item-runes.json        ← 符文名稱、符文之語名稱
├── skills.json            ← 技能名稱與描述
├── mercenaries.json       ← 傭兵相關
├── monsters.json          ← 怪物名稱
├── ui.json                ← UI 文字
└── ...
```

### 常用 Key 對照

#### 符文（item-runes.json）

| Key | 原文 | 符文代碼 |
|-----|------|----------|
| `r01` | El | El Rune |
| `r02` | Eld | Eld Rune |
| ... | ... | ... |
| `r30` | Ber | Ber Rune |
| `r31` | Jah | Jah Rune |
| `r32` | Cham | Cham Rune |
| `r33` | Zod | Zod Rune |

#### 藥水（item-names.json）

| Key | 原文 |
|-----|------|
| `hp1`–`hp5` | Minor ~ Super Healing Potion |
| `mp1`–`mp5` | Minor ~ Super Mana Potion |
| `rvs` | Rejuvenation Potion |
| `rvl` | Full Rejuvenation Potion |

---

## 顏色代碼

D2R 使用 `ÿcX` 前綴（`ÿ` = char 255, U+00FF）來設定文字顏色：

| 代碼 | 顏色 | 常見用途 |
|------|------|----------|
| `ÿc0` | 白色 | 普通物品 |
| `ÿc1` | 紅色 | 警告、重要 |
| `ÿc2` | 綠色 | 套裝物品 |
| `ÿc3` | 藍色 | 魔法物品 |
| `ÿc4` | 金色 | 暗金物品 |
| `ÿc5` | 灰色 | 低價值物品 |
| `ÿc6` | 黑色 | （少用） |
| `ÿc7` | 黃褐色 | （少用） |
| `ÿc8` | 橙色 | 稀有、高價值 |
| `ÿc9` | 黃色 | 稀有物品 |
| `ÿc;` | 紫色 | 特殊標記 |
| `ÿcQ` | 暗金色（加深） | （少用） |

### 使用範例

```json
{
  "id": 70001,
  "Key": "r30",
  "enUS": "ÿc8★ÿc4 Ber Rune ÿc8[#30]"
}
```

顯示效果：橙色星號 + 金色 "Ber Rune" + 橙色 "[#30]"

---

## 常見線上安全 Mod

### 1. Loot Filter（戰利品篩選器）

最受歡迎的線上安全 Mod，自訂物品顯示顏色、縮短名稱、突顯高價值掉落。

**原理：** 修改 `item-names.json` 和 `item-runes.json`，用顏色代碼重新排版物品名稱。

**社群資源：**
- [D2RMM Loot Filter Extended](https://github.com/Caedendi/D2RMM-Loot-Filter-Extended) — 高度可自訂的篩選器
- [D2R Simple Loot Filter](https://www.nexusmods.com/diablo2resurrected/mods/205) — 簡單易用
- [D2R Loot Filters 社群](https://d2rlootfilters.com/) — 社群製作的各種篩選器

### 2. 符文編號

在符文名稱前加上編號，方便快速辨識符文等級：

```json
{"Key": "r01", "enUS": "ÿc5#01 El Rune"}
{"Key": "r30", "enUS": "ÿc4#30 Ber Rune ÿc8★"}
{"Key": "r33", "enUS": "ÿc1#33 Zod Rune ÿc8★★★"}
```

### 3. 藥水縮寫

縮短藥水名稱減少畫面雜訊：

```json
{"Key": "hp1", "enUS": "ÿc1H1"}
{"Key": "mp1", "enUS": "ÿc3M1"}
{"Key": "rvs", "enUS": "ÿc;Rej"}
{"Key": "rvl", "enUS": "ÿc;FRej"}
```

### 4. 遊戲內資訊頁面

利用既有的 JSON 字串來嵌入參考資訊，讓玩家不用離開遊戲就能查閱：

**可嵌入的資訊：**
- FCR / FHR / FBR Breakpoints（施法/受擊/格擋速率斷點）
- Horadric Cube 常用配方速查
- 符文之語配方列表
- 各職業 Skill 加成速查

**做法：** 選一個不常用的 JSON 條目，將其文字值改為含多行資訊的長字串（用 `\n` 換行）。

---

## 進階技巧

### 多語言支援

修改時建議同時更新 `enUS` 和你使用的語言欄位（如 `zhTW`），確保切換語言時顯示一致。

### 避免 ID 衝突

新增條目時使用 70000 以上的 `id`，原版遊戲的 ID 範圍通常在 0–40000。

### 組合多個效果

一個 JSON 條目可以組合多個顏色代碼：

```json
{
  "Key": "r31",
  "enUS": "ÿc8★★ÿc4 Jah Rune ÿc5(ÿc9JahIthBerÿc5=ÿc4Enigmaÿc5) ÿc8[#31]"
}
```

顯示：橙色雙星 + 金色名稱 + 灰色括號內黃色符文組合提示 + 橙色編號

### 啟動參數

線上安全 Mod 不需要 `-txt` 參數（因為不涉及 `.txt` → `.bin` 編譯）：

```
-mod MyFilter
```

首次使用時可加 `-txt` 確保載入，之後可省略。

---

## 線上安全 Mod 專屬注意事項

### ✅ 安全但有邊界

- 只修改 `data/local/lng/strings/` 下的 JSON 檔案
- **不要**在同一個 Mod 中混入 `.txt` 資料表修改——那會變成[離線 Mod](D2R-MODDING-OFFLINE.md)
- 如果需要同時使用離線 Mod 和線上顯示 Mod，請分成兩個獨立的 Mod

### 📋 品質建議

1. **保持可讀性** — 顏色代碼會增加字串複雜度，適度使用避免過度花俏
2. **測試所有語言** — 如果你的 Mod 會分享給他人，確保各語言欄位都有合理的值
3. **備份原始 JSON** — 修改前備份，方便回復
4. **注意 Key 一致性** — 修改 `Key` 值要確保所有引用處同步更新

### 🔧 常見問題

| 問題 | 解決方式 |
|------|----------|
| 物品名稱顯示亂碼 | 確認 `ÿ` 字元是 char 255 (U+00FF)，不是普通的 `y` |
| 顏色未生效 | 確認 `ÿcX` 格式正確，X 必須是有效的顏色代碼（0-9, ;, Q 等） |
| 新增條目未顯示 | 確認 `Key` 與遊戲引用的鍵值一致，且 JSON 格式正確（無語法錯誤） |
| Mod 載入後無變化 | 確認 `.mpq` 資料夾名稱正確、`modinfo.json` 的 `name` 與 `-mod` 參數一致 |

# D2R 離線 Mod 指南（資料表修改）

> 透過修改 `.txt` 資料表來改變 D2R 遊戲機制——掉落率、技能數值、怪物屬性、合成配方等

> ⚠️ **僅限離線/單人遊戲。在 Battle.net 線上使用會導致帳號封禁。**

> ⚠️ **RotW 注意：** D2R v3.1 (Reign of the Warlock) 已禁用 `-mod` 與 `-direct -txt` 參數。
> 本文件中關於手動啟動的說明僅適用於舊版 D2R。
> RotW 版本請使用 [D2RMM](https://github.com/olegbl/d2rmm) 載入 Mod，詳見 [RotW 載入問題](D2R-MOD-LOADING-ROTW.md)。

**前置知識：** 請先閱讀 [D2R Modding 共通指南](D2R-MODDING-COMMON.md)（目錄結構、工具安裝、啟動參數等）

---

## 目錄

- [離線 Mod 概述](#離線-mod-概述)
- [savepath 設定](#savepath-設定)
- [核心資料表一覽](#核心資料表一覽)
- [檔案格式注意事項](#檔案格式注意事項)
- [編輯範例](#編輯範例)
- [常見離線 Mod](#常見離線-mod)
- [進階主題](#進階主題)
- [離線 Mod 專屬注意事項](#離線-mod-專屬注意事項)

---

## 離線 Mod 概述

離線 Mod 修改 `data/global/excel/` 下的 `.txt` 資料表，這些檔案定義了遊戲的核心邏輯：

```
MyMod/
├── modinfo.json                    ← savepath: "MyMod/"
└── MyMod.mpq/
    └── data/
        └── global/
            └── excel/              ← .txt 資料表放這裡
                ├── treasureclassex.txt
                ├── cubemain.txt
                ├── runes.txt
                └── ...
```

由於修改了遊戲機制，離線 Mod 的存檔與原版**不相容**，必須使用隔離的 savepath。

---

## savepath 設定

離線 Mod 的 `modinfo.json` 應使用隔離路徑：

```json
{
  "name": "MyMod",
  "savepath": "MyMod/"
}
```

存檔會存放在：
```
%UserProfile%\Saved Games\Diablo II Resurrected\mods\MyMod\
```

> ⚠️ **不要用 `"../"`**。離線 Mod 改動了遊戲平衡（如掉率、物品屬性），如果共用存檔，切回原版時角色可能出現異常物品或錯誤狀態。

---

## 核心資料表一覽

遊戲的核心邏輯定義在 `data/global/excel/` 下的 `.txt` 檔案中，這些檔案是 **Tab 分隔 (TSV)** 的純文字表格。

### 物品相關

| 檔案 | 說明 |
|------|------|
| `weapons.txt` | 所有武器的基礎屬性（傷害、需求、速度、耐久等） |
| `armor.txt` | 所有防具的基礎屬性（防禦、需求、耐久等） |
| `misc.txt` | 雜項物品（藥水、寶石、符文、鑰匙、卷軸等） |
| `uniqueitems.txt` | 暗金裝備（Unique Items）的特殊屬性 |
| `setitems.txt` | 套裝物品（Set Items）的個別屬性 |
| `sets.txt` | 套裝的完整套裝加成定義 |
| `runes.txt` | 符文之語（Runeword）配方與效果 |
| `gems.txt` | 寶石鑲嵌效果 |
| `magicprefix.txt` | 藍色/黃色物品的前綴詞綴庫 |
| `magicsuffix.txt` | 藍色/黃色物品的後綴詞綴庫 |
| `properties.txt` | 物品屬性的參數化定義（被 Runeword、詞綴等引用） |
| `itemtypes.txt` | 物品類型分類與繼承關係 |

### 技能相關

| 檔案 | 說明 |
|------|------|
| `skills.txt` | 所有玩家與怪物技能的定義（傷害、消耗、範圍等） |
| `skilldesc.txt` | 技能的 UI 顯示資訊（描述、圖示） |
| `skillcalc.txt` | 技能數值計算公式 |

### 怪物相關

| 檔案 | 說明 |
|------|------|
| `monstats.txt` | 怪物主定義（血量、傷害、抗性、AI、等級等） |
| `monstats2.txt` | 怪物的視覺與動畫設定 |
| `monlvl.txt` | 怪物等級對應的數值縮放 |
| `monprop.txt` | 怪物附加屬性 |
| `superuniques.txt` | Super Unique 怪物（如 Rakanishu、Pindleskin）定義 |
| `monai.txt` | 怪物 AI 行為 |

### 掉落與合成

| 檔案 | 說明 |
|------|------|
| `treasureclassex.txt` | **掉落表**——定義怪物/寶箱可掉落的物品及機率 |
| `cubemain.txt` | **Horadric Cube 配方**——所有合成公式 |

### 地圖與世界

| 檔案 | 說明 |
|------|------|
| `levels.txt` | 區域定義（怪物類型、等級範圍、出入口等） |
| `actinfo.txt` | Act 過場與過渡設定 |
| `objects.txt` | 可互動物件（箱子、神殿等） |

### 角色與其他

| 檔案 | 說明 |
|------|------|
| `charstats.txt` | 角色職業定義（初始屬性、起始裝備、技能順序） |
| `states.txt` | Buff/Debuff 狀態效果 |
| `missiles.txt` | 投射物定義（動畫、速度、行為） |
| `experience.txt` | 經驗值表 |
| `difficultylevels.txt` | 難度設定 |

---

## 檔案格式注意事項

- 格式為 **Tab 分隔 (TSV)**，不是 CSV
- 第一行為**欄位標頭 (Header)**
- 許多欄位之間有**交叉引用**（使用 `code` 或 `id` 關聯）
- **強烈建議使用 AFJ Sheet Editor** 或 VS Code D2 .txt Editor 擴充，避免 Excel/OpenOffice 破壞格式
- 以 `*` 開頭的欄位為註解欄，不影響遊戲
- 修改後必須用 `-txt` 參數啟動遊戲才能生效（重新編譯 `.txt` → `.bin`）

---

## 編輯範例

### 修改掉落率

在 `treasureclassex.txt` 中，找到目標怪物的掉落表，修改 `NoDrop` 值：

```
Treasure Class    Picks    NoDrop    Item1    Prob1    ...
Act 1 H2H A      1        100       gld      21       ...
```

將 `NoDrop` 從 `100` 降低到 `10` 即可大幅提高掉落率。

### 新增 Horadric Cube 配方

在 `cubemain.txt` 新增一行：

| description | enabled | numinputs | input 1 | output |
|-------------|---------|-----------|---------|--------|
| My Recipe | 1 | 3 | r01,qty=3 | r02 |

上例：3 個 El 符文 → 1 個 Eld 符文。

### 新增自訂 Runeword

1. 在 `runes.txt` 新增一行，定義符文組合與效果
2. 在 `item-runes.json`（`data/local/lng/strings/`）新增 Runeword 顯示名稱
3. 確保 `runes.txt` 中的 `Name` 欄位與 JSON 的 `Key` 完全一致（無空格差異）

---

## 常見離線 Mod

### 掉落率提升

- [Increase Droprate for D2RMM](https://www.nexusmods.com/diablo2resurrected/mods/180) — 可設定各稀有度的提升倍率
- [D2R Drop Mod](https://www.nexusmods.com/diablo2resurrected/mods/3) — 大幅提升 Boss 與精英怪掉率
- 手動修改 `treasureclassex.txt` 的 `NoDrop` 值

### 倉庫擴充

- 修改倉庫大小、增加收納空間
- Stacking Mod — 讓符文、寶石等可堆疊

### 自訂物品與 Runeword

在相應 `.txt` 檔與 JSON 字串檔中新增定義即可加入全新物品或 Runeword。

### 跳過開場影片

在 `data/hd/global/video/` 放入空白的 `.webm` 檔案覆蓋原版影片。

---

## 進階主題

### 新增自訂物品

1. 在 `weapons.txt`（或 `armor.txt`/`misc.txt`）新增一行，設定唯一 `code` 與屬性
2. 在 `item-names.json` 新增物品名稱字串
3. 準備物品圖示/精靈圖（`.dds` 格式），放入對應材質目錄
4. 如需加入掉落表，修改 `treasureclassex.txt`

### 新增自訂套裝

1. 在 `sets.txt` 定義套裝加成
2. 在 `setitems.txt` 定義各件套裝物品
3. 在 `item-names.json` 新增名稱字串

### HD 材質修改

1. 使用 CascView 提取 `data/hd/global/` 下的材質檔
2. 使用 Noesis 轉換格式、GIMP/Photoshop 編輯
3. 輸出為 `.dds` 格式放回 Mod 對應目錄

---

## 離線 Mod 專屬注意事項

### 🚫 Battle.net 風險

- **絕對不要**在線上模式載入包含 `.txt` 資料表修改的 Mod
- Blizzard 可偵測遊戲數據修改，違反 ToS 會導致**永久封禁**
- 即使只是「測試一下」也有風險——不要心存僥倖

### ⚠️ 存檔相容性

- 離線 Mod 創建的角色/物品可能包含原版不存在的屬性
- 切換回原版時，這些角色可能**損壞或無法載入**
- 務必使用隔離的 `savepath`（如 `"MyMod/"`）

### 🔧 開發流程

1. 每次啟動測試都要加 `-txt` 參數
2. 穩定後移除 `-txt` 可加快啟動速度
3. 修改 `.txt` 時注意欄位間的交叉引用（`code`、`namestr`、`type` 等）
4. 使用 AFJ Sheet Editor 或 VS Code 擴充避免格式損壞

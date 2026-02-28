# D2R Modding 完整指南

> Diablo II: Resurrected (D2R) Mod 製作與安裝全面指南

---

## 目錄

- [概述](#概述)
- [前置需求](#前置需求)
- [必備工具](#必備工具)
- [Step 1：提取遊戲資料檔](#step-1提取遊戲資料檔)
- [Step 2：Mod 目錄結構](#step-2mod-目錄結構)
- [Step 3：modinfo.json 設定](#step-3modinfojson-設定)
- [Step 4：編輯遊戲資料（.txt 資料表）](#step-4編輯遊戲資料txt-資料表)
- [Step 5：編輯本地化字串（JSON）](#step-5編輯本地化字串json)
- [Step 6：啟動參數與載入 Mod](#step-6啟動參數與載入-mod)
- [D2RMM Mod Manager 使用方式](#d2rmm-mod-manager-使用方式)
- [常見 Mod 範例](#常見-mod-範例)
- [進階主題](#進階主題)
- [注意事項與最佳實踐](#注意事項與最佳實踐)
- [參考資源](#參考資源)

---

## 概述

D2R 的 Mod 製作主要是 **軟體式修改 (Softcode Modding)**——透過編輯遊戲的設定檔（`.txt` 資料表與 `.json` 字串檔）來改變遊戲行為，而非修改遊戲程式碼本身。

**可修改的範圍包含：**

| 類別 | 範例 |
|------|------|
| 物品 | 武器/防具屬性、掉落率、Runeword 配方 |
| 技能 | 技能數值、公式、被動效果 |
| 怪物 | 血量、傷害、AI、掉落表 |
| 地圖 | 區域定義、Act 過場 |
| 合成 | Horadric Cube 配方 |
| UI | 物品名稱顏色、倉庫大小、Loot Filter |
| 材質 | HD 材質、精靈圖、模型 |
| 本地化 | 物品名稱、技能描述、UI 文字 |

> ⚠️ **Mod 僅適用於離線/單人遊戲模式。在線上使用 Mod 可能導致帳號封禁。**

---

## 前置需求

- **Windows 10/11**
- **Diablo II: Resurrected** 已安裝（Battle.net 或 Steam 版皆可）
- 約 **40GB+ 可用硬碟空間**（完整提取遊戲資料時需要）
- 基本文字編輯能力

---

## 必備工具

### 核心工具

| 工具 | 用途 | 下載連結 |
|------|------|----------|
| **CascView** | 從 Blizzard CASC 格式提取遊戲資料檔 | [zezula.net](http://www.zezula.net/en/casc/main.html) |
| **AFJ Sheet Editor** | 編輯 `.txt` 資料表（比 Excel 更安全，不會破壞格式） | [d2rmodding.com/modtools](https://www.d2rmodding.com/modtools) |
| **D2RMM (Mod Manager)** | Mod 管理器，合併多個 Mod、一鍵安裝 | [Nexus Mods](https://www.nexusmods.com/diablo2resurrected/mods/169) |

### 輔助工具

| 工具 | 用途 |
|------|------|
| **Visual Studio Code** | 編輯 `.json` 字串檔與設定檔 |
| **VS Code D2 .txt Editor 擴充** | 提供 `.txt` 資料表的欄位提示與錯誤檢查（[Marketplace](https://marketplace.visualstudio.com/items?itemName=bethington.vscode-d2-txt-editor-extension)） |
| **D2RModding-StrEdit** | 專用 D2R 字串編輯器（[GitHub](https://github.com/eezstreet/D2RModding-StrEdit)） |
| **MPQ Editor** | 將 Mod 打包成 MPQ 格式（進階用途） |
| **Noesis** | 處理 3D 模型與材質 |
| **GIMP / Photoshop** | 編輯 `.dds` 材質檔案 |
| **D2RLint** | Mod 資料 QA 驗證工具 |

---

## Step 1：提取遊戲資料檔

### 使用 CascView 提取

1. **下載並解壓 CascView**
2. **開啟 CascView** → 選擇 `Open Storage`
3. **導航到 D2R 安裝目錄**，通常為：
   ```
   C:\Program Files (x86)\Diablo II Resurrected\
   ```
4. **找到並提取所需資料夾：**

   | 路徑 | 內容 |
   |------|------|
   | `data/global/excel/` | 遊戲核心 `.txt` 資料表（物品、技能、怪物等） |
   | `data/local/lng/strings/` | 本地化 JSON 字串檔 |
   | `data/hd/global/` | HD 材質、模型、精靈圖 |
   | `data/global/ui/` | UI 圖形資源 |

5. **提取到本地資料夾**，保持目錄結構不變

> 💡 **只需提取你要修改的檔案**，不需要全部提取。完整提取可能需要 40GB+ 空間。

---

## Step 2：Mod 目錄結構

D2R 的 Mod 目錄位於遊戲安裝目錄下的 `mods/` 資料夾：

```
Diablo II Resurrected/
└── mods/
    └── MyMod/                      ← Mod 根目錄
        ├── modinfo.json            ← Mod 描述檔（必要）
        └── MyMod.mpq/             ← 資料目錄（名稱為 <ModName>.mpq，但它是資料夾不是檔案）
            └── data/
                ├── global/
                │   └── excel/     ← .txt 資料表
                │       ├── weapons.txt
                │       ├── armor.txt
                │       ├── skills.txt
                │       └── ...
                ├── local/
                │   └── lng/
                │       └── strings/  ← 本地化 JSON
                │           ├── item-names.json
                │           ├── item-runes.json
                │           └── ...
                └── hd/
                    └── global/    ← HD 材質與素材
```

> ⚠️ `MyMod.mpq` 在 D2R 中**是一個資料夾（非 MPQ 壓縮檔）**，名稱必須以 `.mpq` 結尾。

### 簡化結構（不使用 .mpq 資料夾）

部分 Mod 也支援直接放在 `data/` 目錄下：

```
mods/
└── MyMod/
    ├── modinfo.json
    └── data/
        └── global/
            └── excel/
                └── weapons.txt
```

---

## Step 3：modinfo.json 設定

每個 Mod 的根目錄必須包含 `modinfo.json`：

```json
{
  "name": "MyMod",
  "savepath": "MyMod/"
}
```

### 欄位說明

| 欄位 | 說明 |
|------|------|
| `name` | Mod 名稱（必須與 `-mod` 啟動參數一致） |
| `savepath` | 存檔路徑。設為 `"MyMod/"` 會將 Mod 存檔隔離存放；設為 `"../"` 則與原版共用存檔 |

### savepath 的影響

- **`"MyMod/"`** — 存檔存放在 `%UserProfile%\Saved Games\Diablo II Resurrected\mods\MyMod\`，Mod 存檔與原版互不影響
- **`"../"`** — 使用原版存檔目錄，Mod 與原版共用角色（需注意相容性）

---

## Step 4：編輯遊戲資料（.txt 資料表）

### 核心資料表一覽

遊戲的核心邏輯定義在 `data/global/excel/` 下的 `.txt` 檔案中，這些檔案是 **Tab 分隔 (TSV)** 的純文字表格。

#### 物品相關

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

#### 技能相關

| 檔案 | 說明 |
|------|------|
| `skills.txt` | 所有玩家與怪物技能的定義（傷害、消耗、範圍等） |
| `skilldesc.txt` | 技能的 UI 顯示資訊（描述、圖示） |
| `skillcalc.txt` | 技能數值計算公式 |

#### 怪物相關

| 檔案 | 說明 |
|------|------|
| `monstats.txt` | 怪物主定義（血量、傷害、抗性、AI、等級等） |
| `monstats2.txt` | 怪物的視覺與動畫設定 |
| `monlvl.txt` | 怪物等級對應的數值縮放 |
| `monprop.txt` | 怪物附加屬性 |
| `superuniques.txt` | Super Unique 怪物（如 Rakanishu、Pindleskin）定義 |
| `monai.txt` | 怪物 AI 行為 |

#### 掉落與合成

| 檔案 | 說明 |
|------|------|
| `treasureclassex.txt` | **掉落表**——定義怪物/寶箱可掉落的物品及機率 |
| `cubemain.txt` | **Horadric Cube 配方**——所有合成公式 |

#### 地圖與世界

| 檔案 | 說明 |
|------|------|
| `levels.txt` | 區域定義（怪物類型、等級範圍、出入口等） |
| `actinfo.txt` | Act 過場與過渡設定 |
| `objects.txt` | 可互動物件（箱子、神殿等） |

#### 角色與其他

| 檔案 | 說明 |
|------|------|
| `charstats.txt` | 角色職業定義（初始屬性、起始裝備、技能順序） |
| `states.txt` | Buff/Debuff 狀態效果 |
| `missiles.txt` | 投射物定義（動畫、速度、行為） |
| `experience.txt` | 經驗值表 |
| `difficultylevels.txt` | 難度設定 |

### 檔案格式注意事項

- 格式為 **Tab 分隔 (TSV)**，不是 CSV
- 第一行為**欄位標頭 (Header)**
- 許多欄位之間有**交叉引用**（使用 `code` 或 `id` 關聯）
- **強烈建議使用 AFJ Sheet Editor** 或 VS Code D2 .txt Editor 擴充，避免 Excel/OpenOffice 破壞格式
- 以 `*` 開頭的欄位為註解欄，不影響遊戲

### 編輯範例：修改掉落率

在 `treasureclassex.txt` 中，找到目標怪物的掉落表，修改 `NoDrop` 值：

```
Treasure Class    Picks    NoDrop    Item1    Prob1    ...
Act 1 H2H A      1        100       gld      21       ...
```

將 `NoDrop` 從 `100` 降低到 `10` 即可大幅提高掉落率。

### 編輯範例：新增 Horadric Cube 配方

在 `cubemain.txt` 新增一行：

| description | enabled | numinputs | input 1 | output |
|-------------|---------|-----------|---------|--------|
| My Recipe | 1 | 3 | r01,qty=3 | r02 |

上例：3 個 El 符文 → 1 個 Eld 符文。

---

## Step 5：編輯本地化字串（JSON）

D2R 使用 JSON 檔案管理遊戲內的文字字串（取代經典 D2 的 `.tbl` 格式）。

### 字串檔位置

```
data/local/lng/strings/
├── item-names.json        ← 物品名稱
├── item-modifiers.json    ← 物品詞綴描述
├── item-runes.json        ← 符文之語名稱
├── skills.json            ← 技能名稱與描述
├── mercenaries.json       ← 傭兵相關
├── monsters.json          ← 怪物名稱
├── ui.json                ← UI 文字
└── ...
```

### JSON 格式範例

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

| 欄位 | 說明 |
|------|------|
| `id` | 唯一數字 ID |
| `Key` | 字串鍵值（在 `.txt` 檔中用 `namestr` 等欄位引用） |
| `enUS` | 英文文字 |
| `zhTW` | 繁體中文文字 |

### 常用字串編輯工具

- **D2RModding-StrEdit**（[GitHub](https://github.com/eezstreet/D2RModding-StrEdit)）— 專門的 D2R 字串編輯器
- **Visual Studio Code** — 直接編輯 JSON

---

## Step 6：啟動參數與載入 Mod

### D2R 啟動參數

| 參數 | 說明 |
|------|------|
| `-mod <ModName>` | 載入指定 Mod（名稱對應 `mods/<ModName>/` 資料夾） |
| `-txt` | 強制從 `.txt` 檔重新編譯 `.bin` 檔案（**開發/測試時必用**） |
| `-direct` | 直接從檔案系統載入資料（搭配 `-txt` 使用） |
| `-w` | 視窗化模式啟動 |
| `-ns` / `-nosound` | 停用音效 |
| `-noborder` | 無邊框視窗 |
| `-fullscreen` | 強制全螢幕 |
| `-username <email>` | 指定 Battle.net 帳號（自動登入用） |
| `-password <password>` | 指定密碼（自動登入用） |
| `-address <server>` | 指定伺服器區域（如 `us.actual.battle.net`） |

### 啟動方式

#### 方式 1：建立捷徑

1. 右鍵 `D2R.exe` → 建立捷徑
2. 右鍵捷徑 → 內容
3. 在「目標」欄位的路徑後方加上參數：
   ```
   "C:\Program Files (x86)\Diablo II Resurrected\D2R.exe" -mod MyMod -txt
   ```

#### 方式 2：Battle.net 啟動器

1. Battle.net → D2R → 設定（齒輪圖示）→ 遊戲設定
2. 在「額外命令列參數」中填入：
   ```
   -mod MyMod -txt
   ```

#### 方式 3：搭配本工具 (d2r-multiboxing)

在 [config.json](../internal/config/config.go) 中設定好 D2R 路徑後，透過本工具的帳號管理功能啟動 D2R，可在啟動參數中加入 Mod 相關旗標。

### 關於 -txt 參數

- `-txt` 會讓啟動速度變慢（需要編譯 `.txt` → `.bin`）
- **開發測試階段**每次都要加 `-txt` 以確保修改生效
- **穩定後**可以移除 `-txt`，遊戲會直接讀取已編譯的 `.bin` 檔案加快啟動

---

## D2RMM Mod Manager 使用方式

[D2RMM](https://www.nexusmods.com/diablo2resurrected/mods/169) 是最主流的 D2R Mod 管理工具，支援一鍵安裝、合併多個 Mod、避免衝突。

### 安裝步驟

1. **下載 D2RMM** — 從 [Nexus Mods](https://www.nexusmods.com/diablo2resurrected/mods/169) 或 [GitHub](https://github.com/olegbl/d2rmm) 下載
2. **解壓到任意位置**（可攜式，不需安裝）
3. **下載 Mod** — 從 Nexus Mods 下載相容 D2RMM 的 Mod
4. **放入 mods 資料夾**：
   ```
   D2RMM/
   └── mods/
       ├── StackableRunes/
       │   ├── mod.js
       │   └── mod.json
       └── LootFilter/
           ├── mod.js
           └── mod.json
   ```

### 使用流程

1. **執行 `D2RMM.exe`**
2. **設定** → 指定 D2R 安裝目錄
3. **Mods 頁籤** → 勾選要啟用的 Mod
4. **拖曳排序**調整載入順序（後載入的 Mod 優先級更高）
5. **點擊「Install Mods」**— 每次修改 Mod 選擇都要重新安裝
6. **點擊「Launch D2R」**啟動遊戲

### D2RMM Mod 存檔位置

```
%UserProfile%\Saved Games\Diablo II Resurrected\mods\D2RMM\
```

如需使用原版存檔，將存檔複製到上述目錄即可。

---

## 常見 Mod 範例

### 1. Loot Filter（戰利品篩選器）

最受歡迎的 Mod 類型，可自訂物品顯示顏色、隱藏垃圾物品、突顯高價值掉落。

- [D2RMM Loot Filter Extended](https://github.com/Caedendi/D2RMM-Loot-Filter-Extended) — 高度可自訂的篩選器
- [D2R Simple Loot Filter](https://www.nexusmods.com/diablo2resurrected/mods/205) — 簡單易用
- [D2R Loot Filters 社群](https://d2rlootfilters.com/) — 社群製作的各種篩選器

### 2. 掉落率提升

- [Increase Droprate for D2RMM](https://www.nexusmods.com/diablo2resurrected/mods/180) — 可設定各稀有度的提升倍率
- [D2R Drop Mod](https://www.nexusmods.com/diablo2resurrected/mods/3) — 大幅提升 Boss 與精英怪掉率
- 手動修改 `TreasureClassEx.txt` 的 `NoDrop` 值

### 3. 倉庫擴充

- 修改倉庫大小、增加收納空間
- Stacking Mod — 讓符文、寶石等可堆疊

### 4. 自訂物品與 Runeword

在相應 `.txt` 檔與 JSON 字串檔中新增定義即可加入全新物品或 Runeword。

### 5. 跳過開場影片

在 `data/hd/global/video/` 放入空白的 `.webm` 檔案覆蓋原版影片。

---

## 進階主題

### 新增自訂物品

1. 在 `weapons.txt`（或 `armor.txt`/`misc.txt`）新增一行，設定唯一 `code` 與屬性
2. 在 `item-names.json` 新增物品名稱字串
3. 準備物品圖示/精靈圖（`.dds` 格式），放入對應材質目錄
4. 如需加入掉落表，修改 `treasureclassex.txt`

### 新增自訂 Runeword

1. 在 `runes.txt` 新增一行，定義符文組合與效果
2. 在 `item-runes.json` 新增 Runeword 名稱
3. 在 `item-names.json` 新增相關字串（若有需要）

### 新增自訂套裝

1. 在 `sets.txt` 定義套裝加成
2. 在 `setitems.txt` 定義各件套裝物品
3. 在 `item-names.json` 新增名稱字串

### HD 材質修改

1. 使用 CascView 提取 `data/hd/global/` 下的材質檔
2. 使用 Noesis 轉換格式、GIMP/Photoshop 編輯
3. 輸出為 `.dds` 格式放回 Mod 對應目錄

---

## 注意事項與最佳實踐

### ⚠️ 重要警告

1. **線上風險** — 在線上模式使用 Mod 可能違反 Blizzard ToS 並導致帳號封禁
2. **備份優先** — 修改前務必備份原始檔案與存檔
3. **單人離線** — 建議僅在離線/單人模式下使用 Mod

### 💡 開發建議

1. **逐步修改** — 每次只改一個檔案，測試通過再改下一個
2. **使用版本控制** — 用 Git 追蹤 Mod 檔案變更歷史
3. **使用專用編輯器** — 避免 Excel/OpenOffice 破壞 TSV 格式
4. **交叉引用** — 許多 `.txt` 檔之間有引用關係（如 `code`、`namestr`），修改時需保持一致性
5. **移除 `-txt`** — Mod 穩定後移除 `-txt` 參數以加快啟動
6. **測試順序** — 先驗證基本功能，再進行細節調整

### 🔧 疑難排解

| 問題 | 解決方式 |
|------|----------|
| 遊戲啟動後 Mod 未生效 | 確認目錄結構正確、`-mod <名稱>` 與資料夾名稱一致、加上 `-txt` |
| 遊戲崩潰 | 檢查 `.txt` 是否有格式錯誤（多餘的 Tab、缺少欄位） |
| 物品名稱顯示為 Key | 檢查 JSON 字串檔中的 `Key` 是否與 `.txt` 中的 `namestr` 對應 |
| Mod 間衝突 | 使用 D2RMM 管理載入順序，或手動合併衝突的 `.txt` 檔案 |

---

## 參考資源

### 綜合指南

- [D2RModding Guide Center](https://www.d2rmodding.com/guides) — 最完整的 Mod 教學網站
- [The Phrozen Keep](https://d2mods.info/forum/kb/viewarticle?a=477) — D2 Modding 知識庫（歷史最悠久）
- [diablo2.io Modding Tutorial](https://diablo2.io/forums/d2r-modding-tutorial-t704113.html) — 社群步驟教學
- [GitHub: ModdingDiablo2Resurrected](https://github.com/HighTechLowIQ/ModdingDiablo2Resurrected) — 完整圖文教學

### 資料參考

- [D2R Data Guide (Corrected)](https://locbones.github.io/D2R_DataGuide/) — 所有 `.txt` 檔的欄位詳解
- [Diablo II Data File Guide](https://wolfieeiflow.github.io/diabloiidatafileguide/) — 官方資料指南（社群修正版）
- [D2R-Excel (GitHub)](https://github.com/pinkufairy/D2R-Excel) — 所有 `.txt` 原始資料檔
- [blizzhackers/d2data](https://github.com/blizzhackers/d2data) — D2R 3.0 JSON 資料集

### 工具下載

- [D2RModding Mod Tools](https://www.d2rmodding.com/modtools) — 工具集合
- [CascView](http://www.zezula.net/en/casc/main.html) — CASC 提取工具
- [D2RMM](https://www.nexusmods.com/diablo2resurrected/mods/169) — Mod 管理器
- [D2RModding-StrEdit](https://github.com/eezstreet/D2RModding-StrEdit) — 字串編輯器
- [D2 .txt Editor (VS Code)](https://marketplace.visualstudio.com/items?itemName=bethington.vscode-d2-txt-editor-extension) — VS Code 擴充
- [Diablo 2 DIY Mod Maker](https://sajonoso.github.io/d2mods/) — 視覺化 Mod 產生器（適合入門）

### 影片教學

- [How To Mod D2R (YouTube)](https://www.youtube.com/watch?v=RMquP82QHGw) — HighTechLowIQ 完整教學
- [D2RMM 安裝教學 (YouTube)](https://www.youtube.com/watch?v=bEbQK4xJZn4) — Mod Manager 使用教學

### 社群

- [Nexus Mods - D2R](https://www.nexusmods.com/diablo2resurrected) — 最大 Mod 下載平台
- [D2R Loot Filters](https://d2rlootfilters.com/) — 社群 Loot Filter 集合
- [Blizzard D2R 論壇](https://us.forums.blizzard.com/en/d2r/) — 官方社群
- [The Phrozen Keep 論壇](https://d2mods.info/forum/) — 元老級 Modding 社群

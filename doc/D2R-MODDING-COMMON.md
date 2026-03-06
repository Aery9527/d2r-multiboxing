# D2R Modding 共通指南

> Diablo II: Resurrected (D2R) Mod 製作的基礎知識與通用流程

**延伸閱讀：**
- [離線 Mod（資料表修改）](D2R-MODDING-OFFLINE.md) — 修改 `.txt` 資料表，改變遊戲機制（⚠️ 僅限離線/單人）
- [線上安全 Mod（顯示修改）](D2R-MODDING-ONLINE.md) — 修改 JSON 字串檔，自訂物品顯示（✅ Battle.net 安全）
- [RotW Mod 載入問題](D2R-MOD-LOADING-ROTW.md) — D2R v3.1 後 `-mod` 需搭配 `-uid osi` 才有效

---

## 目錄

- [概述](#概述)
- [前置需求](#前置需求)
- [必備工具](#必備工具)
- [Step 1：提取遊戲資料檔](#step-1提取遊戲資料檔)
- [Step 2：Mod 目錄結構](#step-2mod-目錄結構)
- [Step 3：modinfo.json 設定](#step-3modinfojson-設定)
- [啟動參數與載入 Mod](#啟動參數與載入-mod)
- [D2RMM Mod Manager](#d2rmm-mod-manager)
- [注意事項與最佳實踐](#注意事項與最佳實踐)
- [疑難排解](#疑難排解)
- [參考資源](#參考資源)

---

## 概述

D2R 的 Mod 製作主要是 **軟體式修改 (Softcode Modding)**——透過編輯遊戲的設定檔來改變遊戲行為，而非修改遊戲程式碼本身。

修改方式分為兩大類：

| 類型 | 修改目標 | Battle.net 安全性 | 詳細說明 |
|------|----------|-------------------|----------|
| **離線 Mod** | `.txt` 資料表（遊戲機制） | ❌ 會被封禁 | [D2R-MODDING-OFFLINE.md](D2R-MODDING-OFFLINE.md) |
| **線上安全 Mod** | `.json` 字串檔（顯示文字） | ✅ 安全 | [D2R-MODDING-ONLINE.md](D2R-MODDING-ONLINE.md) |

**可修改的範圍包含：**

| 類別 | 範例 | 類型 |
|------|------|------|
| 物品屬性 | 武器/防具數值、掉落率、Runeword 配方 | 離線 |
| 技能 | 技能數值、公式、被動效果 | 離線 |
| 怪物 | 血量、傷害、AI、掉落表 | 離線 |
| 合成 | Horadric Cube 配方 | 離線 |
| 地圖 | 區域定義、Act 過場 | 離線 |
| 物品名稱/顏色 | Loot Filter、物品高亮、符文編號 | 線上安全 |
| UI 文字 | 技能描述、資訊頁面 | 線上安全 |
| 材質 | HD 材質、精靈圖、模型 | 離線 |

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

   | 路徑 | 內容 | 用途 |
   |------|------|------|
   | `data/global/excel/` | 遊戲核心 `.txt` 資料表 | [離線 Mod](D2R-MODDING-OFFLINE.md) |
   | `data/local/lng/strings/` | 本地化 JSON 字串檔 | [線上安全 Mod](D2R-MODDING-ONLINE.md) |
   | `data/hd/global/` | HD 材質、模型、精靈圖 | 離線 Mod |
   | `data/global/ui/` | UI 圖形資源 | 離線 Mod |

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
                │   └── excel/     ← .txt 資料表（離線 Mod）
                │       ├── weapons.txt
                │       ├── armor.txt
                │       └── ...
                ├── local/
                │   └── lng/
                │       └── strings/  ← 本地化 JSON（線上安全 Mod）
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
| `savepath` | 存檔路徑，依 Mod 類型選擇不同設定（見下表） |

### savepath 依 Mod 類型的選擇

| Mod 類型 | savepath | 存檔位置 | 說明 |
|----------|----------|----------|------|
| [離線 Mod](D2R-MODDING-OFFLINE.md) | `"MyMod/"` | `%UserProfile%\Saved Games\...\mods\MyMod\` | 隔離存檔，Mod 存檔與原版互不影響 |
| [線上安全 Mod](D2R-MODDING-ONLINE.md) | `"../"` | 原版存檔目錄 | 共用原版存檔，適合僅改顯示的 Mod |

> 💡 離線 Mod 修改了遊戲機制（如掉落率），存檔內容會與原版不相容，因此必須隔離。線上安全 Mod 僅改顯示文字，不影響存檔資料，因此可以共用。

---

## 啟動參數與載入 Mod

> ℹ️ **D2R v3.1 (Reign of the Warlock) 需搭配 `-uid osi` 參數才能使 `-mod` 生效。**
> Battle.net 啟動器會自動帶入 `-uid osi`，直接執行 `D2R.exe` 時需手動加上。
> 詳細調查記錄請參考 [D2R Mod 載入問題：RotW 版本](D2R-MOD-LOADING-ROTW.md)。

### D2R 啟動參數

| 參數 | 說明 | RotW 狀態 |
|------|------|-----------|
| `-uid osi` | 啟用 Battle.net Online Services Interface 模式 | ✅ 必要（`-mod` 的前提） |
| `-mod <ModName>` | 載入指定 Mod（名稱對應 `mods/<ModName>/` 資料夾） | ✅ 搭配 `-uid osi` 有效 |
| `-txt` | 強制從 `.txt` 檔重新編譯 `.bin` 檔案（**離線 Mod 開發/測試時必用**） | ✅ 搭配 `-uid osi` 有效 |
| `-direct` | 直接從檔案系統載入資料（搭配 `-txt` 使用） | ❓ 未測試 |
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
   "C:\Program Files (x86)\Diablo II Resurrected\D2R.exe" -uid osi -mod MyMod -txt
   ```

#### 方式 2：Battle.net 啟動器

1. Battle.net → D2R → 設定（齒輪圖示）→ 遊戲設定
2. 在「額外命令列參數」中填入（Battle.net 會自動帶入 `-uid osi`）：
   ```
   -mod MyMod -txt
   ```

#### 方式 3：搭配本工具 (d2r-multiboxing)

在 [config.json](../internal/config/config.go) 中設定好 D2R 路徑後，透過本工具的帳號管理功能啟動 D2R，可在啟動參數中加入 Mod 相關旗標。

### 關於 -txt 參數

- `-txt` 會讓啟動速度變慢（需要編譯 `.txt` → `.bin`）
- **開發測試階段**每次都要加 `-txt` 以確保修改生效
- **穩定後**可以移除 `-txt`，遊戲會直接讀取已編譯的 `.bin` 檔案加快啟動
- **線上安全 Mod（僅 JSON）不需要 `-txt`**，因為不涉及 `.txt` → `.bin` 編譯

---

## D2RMM Mod Manager

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

## 注意事項與最佳實踐

### 💡 開發建議

1. **逐步修改** — 每次只改一個檔案，測試通過再改下一個
2. **使用版本控制** — 用 Git 追蹤 Mod 檔案變更歷史
3. **使用專用編輯器** — 避免 Excel/OpenOffice 破壞 TSV 格式
4. **交叉引用** — 許多 `.txt` 檔之間有引用關係（如 `code`、`namestr`），修改時需保持一致性
5. **備份優先** — 修改前務必備份原始檔案與存檔
6. **測試順序** — 先驗證基本功能，再進行細節調整

---

## 疑難排解

| 問題 | 解決方式 |
|------|----------|
| 遊戲啟動後 Mod 未生效（RotW 版本） | 確認啟動時帶有 `-uid osi` 參數——不帶此參數時 `-mod` 會被靜默忽略。詳見 [RotW 載入問題](D2R-MOD-LOADING-ROTW.md) |
| 遊戲啟動後 Mod 未生效（舊版本） | 確認目錄結構正確、`-mod <名稱>` 與資料夾名稱一致、加上 `-txt` |
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

# 裝備清點（Equipment Inventory）技術方案分析

## Context

d2r-hyper-launcher 目前是純 process/window 管理工具，使用者希望新增「裝備清點」功能——能查看各帳號角色身上與倉庫的裝備。由於本工具的核心使用場景是 Battle.net 多開，Online 角色支援是必要條件。

---

## 五種技術方案比較

### 方案 1: Save File 解析 (.d2s)

**原理**: 解析 D2R 本地 `.d2s` 存檔二進位格式，擷取裝備/背包/倉庫資料。

**Go 生態**: `nokka/d2s`、`Vitalick/go-d2editor`、`gucio321/d2d2s` — 成熟穩定。

**致命限制**: **Online (Battle.net) 角色的本地檔案只有 `.ctlo`/`.keyo`（控制設定），真正的角色/物品資料存在 Blizzard 伺服器端。** 離線角色才有完整 `.d2s`。

| 面向 | 評估 |
|---|---|
| 技術難度 | 低 — 格式成熟，現成 parser |
| 封號風險 | 零 — 只讀本地檔案 |
| 維護成本 | 低 — 格式極少變動 |
| Online 角色 | **不支援（致命）** |
| 實作時間 | 1-2 週 |

**結論**: 無法作為主方案（本工具只支援 Battle.net）。可作為 bonus 功能給離線玩家。

---

### 方案 2: 記憶體讀取 (ReadProcessMemory) ★ 推薦

**原理**: 用 `kernel32.dll!ReadProcessMemory` 讀取 D2R 遊戲程序記憶體，從中解析角色、裝備、倉庫資料結構。

**Go 生態**:
- [`hectorgimenez/d2go`](https://github.com/hectorgimenez/d2go) — Go D2R 記憶體讀取庫，提供 `GameReader`、Offset 結構（`GameData`、`UnitTable`、`UI`）、物品屬性解析
- [`hectorgimenez/koolo`](https://github.com/hectorgimenez/koolo) — 基於 d2go 的 D2R bot，證明整套方案在 Go 中端到端可行

**技術優勢**:
- 本專案已使用 `OpenProcess`、NT API、`golang.org/x/sys/windows`，加入 `ReadProcessMemory` 是自然延伸
- 可透過現有的 `process.FindProcessesByName` + `D2R-<DisplayName>` 視窗標題精準定位各帳號的 PID
- 資料完整度最高：物品類型、品質（暗金/套裝/稀有等）、插孔、符文之語、所有屬性、位置（背包/倉庫/腰帶/裝備欄）

**風險**:
- **記憶體偏移量 (Offset) 隨 D2R 更新變動** — d2go 社群活躍維護，但仍需跟進
- **Warden 反作弊偵測** — 詳見下方「封號風險深入分析」
- **與專案原則的衝突** — AGENTS.md 提到「不會注入遊戲程式」；但 ReadProcessMemory 嚴格來說是「讀取」非「注入」，需要釐清措辭

**封號風險深入分析**:

| 面向 | 說明 |
|---|---|
| Warden 運作方式 | Blizzard 的 Warden 反作弊系統會掃描執行中的程序清單，並與黑名單比對。也會檢查遊戲記憶體的頁面保護狀態與校驗碼。 |
| 已知封禁案例 | **MapAssist**（C# D2R 地圖 hack）用戶自 2022-06-14 起被永久封禁。它使用的就是 ReadProcessMemory 唯讀方式。 |
| D2RMH 立場 | D2RMH 作者明確聲明：「only reading process memory, without injects, hooks or memory writes」，但同時聲明「cannot guarantee ban-free」。 |
| Koolo/d2go 立場 | Koolo 文件聲明「you can get banned for using it」，但作者撰寫時「not aware of any actual bans」。注意 koolo 不只讀記憶體，還注入滑鼠/鍵盤操作（行為更明顯）。 |
| 唯讀 vs 注入的差異 | 唯讀記憶體存取（`PROCESS_VM_READ`）風險低於程式碼注入（`CreateRemoteThread`/`WriteProcessMemory`）。但 Warden 的程序名黑名單機制與記憶體存取方式無關——如果你的工具被加入黑名單，不論唯讀與否都會被偵測。 |
| 裝備清點 vs Bot 的差異 | Bot（如 koolo）持續讀取記憶體 + 自動操作，行為模式明顯。裝備清點工具只在使用者主動觸發時短暫讀取一次，行為模式更低調。 |
| 降低風險的措施 | 1. 不要用知名工具名稱作為 process name（避免黑名單）<br>2. 僅在使用者手動觸發時讀取（非持續監控）<br>3. 讀取完立即關閉 process handle<br>4. 功能做成 opt-in + 首次使用時顯示風險警告 |

**風險總結**: 唯讀記憶體存取不是零風險，但在 D2R 社群中是被廣泛使用的技術（地圖 hack、物品追蹤器、bot 都用）。對於一個「手動觸發、短暫讀取、不自動操作」的裝備清點功能，風險等級是所有記憶體讀取應用中最低的一檔。

| 面向 | 評估 |
|---|---|
| 技術難度 | 中 — d2go 提供基礎，重點在整合 |
| 封號風險 | 低至中 — 唯讀存取，但觸碰遊戲記憶體 |
| 維護成本 | 高 — 每次 D2R 更新可能需更新偏移量 |
| Online 角色 | **完全支援**（讀取 live 遊戲狀態） |
| 實作時間 | 3-5 週（依賴 d2go） |

---

### 方案 3: 螢幕擷取 + OCR（滑鼠 Hover）

**原理**: 擷取遊戲畫面 → 模擬滑鼠移動到各裝備格子 → 擷取 tooltip 截圖 → OCR 辨識文字。

**現有專案參考**:
- [horadricapp](https://github.com/stephaistos/horadricapp) — OpenCV + Tesseract
- [D2R-AI-Item-Tracker](https://github.com/vdamov/D2R-AI-Item-Tracker) — Vision LLM 辨識
- [D2R-MuleChecker](https://github.com/dexterrawlinson/D2R-MuleChecker) — 螢幕讀取方式

**技術難題（非常多）**:
1. **背包格子系統**: D2R 背包是 10x4 格子，物品佔 1x1 到 2x4 不等大小，無法固定掃描每格
2. **滑鼠自動化**: 需移動滑鼠到每個格子 hover → 玩家在掃描期間失去滑鼠控制
3. **螢幕擷取**: GDI `BitBlt` 對 D2R 全螢幕模式可能失效；需用 DXGI Desktop Duplication
4. **OCR 精度差**: D2R tooltip 背景深色帶紋理、文字有多種顏色（白/藍/黃/綠/金），OCR 準確率低，需大量圖片前處理
5. **解析度/縮放依賴**: 格子座標隨解析度與 UI 縮放設定改變
6. **語言依賴**: tooltip 文字隨遊戲語言不同（至少需支援 EN、zh-TW）
7. **速度極慢**: 掃描 ~40 背包格 + ~100 倉庫格可能需要數分鐘
8. **狀態前提**: 角色必須在遊戲中且背包/倉庫已開啟
9. **外部依賴**: 需 Tesseract OCR 引擎（外部執行檔）

| 面向 | 評估 |
|---|---|
| 技術難度 | **非常高** — OCR 準確度、座標計算、滑鼠自動化 |
| 封號風險 | 非常低 — 不存取遊戲記憶體 |
| 維護成本 | 中 — D2R UI 改版會影響座標 |
| Online 角色 | 支援 |
| 實作時間 | 6-10 週，edge case 極多 |

**結論**: 技術難度與實作量不成比例，OCR 精度問題很難根治，且滑鼠自動化本身也與「不自動化遊戲操作」原則衝突。

---

### 方案 4: 封包擷取/攔截

**原理**: 攔截 D2R 與 Battle.net 間的網路封包，解碼物品資料。

**致命難題**:
1. **D2R 使用 TLS 加密** — 解密需從記憶體提取 TLS session key → 等於需要方案 2 的 ReadProcessMemory
2. **D2R 的網路協議未公開文件化** — 舊版 D2 封包格式有部分文件但 D2R 版本不同
3. **需要額外驅動** — WinDivert 或 Npcap 需安裝系統層級驅動
4. **複雜度 = 記憶體讀取 + 網路分析** — 集兩家之短

| 面向 | 評估 |
|---|---|
| 技術難度 | **極高** — TLS 解密 + 協議逆向 + 封包擷取 |
| 封號風險 | 高 — 網路攔截比記憶體讀取更可疑 |
| 維護成本 | 非常高 |
| Online 角色 | 支援 |
| 實作時間 | 12+ 週，高度不確定性 |

**結論**: 明確最差選項。比記憶體讀取更複雜、風險更高、資料更少。

---

### 方案 5: 混合方案（記憶體定位 + 螢幕驗證）

結合記憶體讀取找物品位置 + 螢幕擷取做視覺確認。

**結論**: 如果已有記憶體讀取，螢幕擷取不增加資料價值；如果想避免記憶體風險，這方案仍然需要它。增加不必要複雜度。

---

## 總結比較表

| 評估面向 | .d2s 解析 | 記憶體讀取 | 螢幕+OCR | 封包擷取 |
|---|---|---|---|---|
| Online 角色 | **不支援** | 支援 | 支援 | 支援 |
| 技術難度 | 低 | 中 | 非常高 | 極高 |
| 封號風險 | 零 | 低至中 | 非常低 | 高 |
| 維護成本 | 低 | 高 | 中 | 非常高 |
| 實作時間 | 1-2 週 | 3-5 週 | 6-10 週 | 12+ 週 |
| 外部依賴 | 無 (Go lib) | 無 (Go lib) | Tesseract | WinDivert |
| 資料完整度 | 完整(離線) | 完整 | 受 OCR 限制 | 未知協議 |

---

## 推薦方案

### 首選: 方案 2 — 記憶體讀取 (ReadProcessMemory + d2go)

**理由**:
1. Online 角色支援是硬需求 → 排除 .d2s
2. `d2go` 提供現成 Go 基礎 → 不需從零逆向工程
3. 與現有技術棧自然銜接（已用 `OpenProcess`、NT API、`golang.org/x/sys/windows`）
4. 資料完整度最高，速度最快（毫秒級讀取 vs OCR 分鐘級）
5. 實作量合理（3-5 週）

### 需要解決的關鍵問題

1. **專案原則調整**: 「不會注入遊戲程式」→ 應明確為「不會修改遊戲記憶體、注入程式碼或自動化遊戲操作」（ReadProcessMemory 是「讀取」非「注入」）
2. **封號風險告知**: 功能首次啟用時需顯示明確警告，讓使用者自行決定
3. **Offset 維護策略**: 依賴 d2go 社群更新？自行維護 config？版本偵測 + 優雅降級？
4. **d2go 依賴策略**: 直接 import？fork 精簡版？vendor 進專案？

### 次選: 附加 .d2s 解析（bonus 離線功能）
低成本零風險，可順手加上。

### 不推薦: 螢幕+OCR、封包擷取、混合方案

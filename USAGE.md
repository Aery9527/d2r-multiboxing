# D2R Multiboxing — 使用說明

> 本文件說明如何安裝、設定與操作 D2R Multiboxing CLI 工具。
> 專案完整技術說明請參考 [PLAN.md](PLAN.md)。

---

## 目錄

- [前置需求](#前置需求)
- [安裝與編譯](#安裝與編譯)
- [帳號設定](#帳號設定)
  - [建立 accounts.csv](#建立-accountscsv)
  - [CSV 欄位說明](#csv-欄位說明)
  - [密碼自動加密](#密碼自動加密)
- [資料目錄](#資料目錄)
- [設定 D2R 路徑](#設定-d2r-路徑)
- [啟動工具](#啟動工具)
- [操作說明](#操作說明)
  - [主選單](#主選單)
  - [啟動單一帳號](#啟動單一帳號)
  - [啟動所有帳號](#啟動所有帳號)
  - [重新整理狀態](#重新整理狀態)
- [完整操作範例](#完整操作範例)
- [背景 Handle 監控](#背景-handle-監控)
- [視窗模式設定](#視窗模式設定)
- [常見問題 FAQ](#常見問題-faq)
- [注意事項](#注意事項)

---

## 前置需求

| 項目 | 需求 |
|------|------|
| 作業系統 | Windows 10 / 11 |
| Go 版本 | 1.26 以上（僅編譯時需要） |
| 權限 | 必須以 **管理員權限** 執行 |
| 遊戲版本 | Battle.net 版 D2R（不支援 Steam 版） |
| 視窗模式 | 請先手動進入遊戲 **選項 → 畫面 → 視窗模式** 設定為「視窗化」 |

---

## 安裝與編譯

### 方式一：從原始碼編譯

```powershell
# 1. 進入專案目錄
cd C:\Users\User\GolandProjects\d2r-multiboxing

# 2. 編譯
go build -o d2r-multiboxing.exe ./cmd/d2r-multiboxing

# 3. 確認產出
Get-Item .\d2r-multiboxing.exe
```

### 方式二：直接使用已編譯的 exe

將 `d2r-multiboxing.exe` 放到任意目錄即可。首次執行會自動在家目錄建立 `~/.d2r-multiboxing/` 資料目錄及預設設定檔。

---

## 帳號設定

### 建立 accounts.csv

1. 在資料目錄 `~/.d2r-multiboxing/` 下建立 `accounts.csv`（可複製範本）：

   ```powershell
   Copy-Item .\accounts.csv "$env:USERPROFILE\.d2r-multiboxing\accounts.csv"
   ```

2. 用文字編輯器（記事本、VS Code 等）打開 `accounts.csv`，填入你的帳號資訊：

   ```powershell
   notepad "$env:USERPROFILE\.d2r-multiboxing\accounts.csv"
   ```

   ```csv
   Email,Password,DisplayName
   player1@gmail.com,mypassword123,主帳號-法師
   player2@gmail.com,anotherpass456,副帳號-野蠻人
   ```

> ⚠️ **重要**：密碼欄位首次填入明文密碼即可，工具啟動後會自動加密。

### CSV 欄位說明

| 欄位 | 必填 | 說明 | 範例 |
|------|------|------|------|
| `Email` | ✅ | Battle.net 登入信箱 | `player@gmail.com` |
| `Password` | ✅ | 帳號密碼（首次執行後自動加密） | `mypassword123` |
| `DisplayName` | ✅ | 顯示名稱，用於視窗標題與選單 | `主帳號-法師` |

**可用區域**：

| 代號 | 區域 | 伺服器地址 |
|------|------|-----------|
| `NA` | 美洲 | `us.actual.battle.net` |
| `EU` | 歐洲 | `eu.actual.battle.net` |
| `Asia` | 亞洲 | `kr.actual.battle.net` |

### 密碼自動加密

工具首次讀取到明文密碼時，會自動執行以下步驟：

1. 使用 Windows DPAPI（`CryptProtectData`）加密密碼
2. 將加密後的密碼以 `ENC:` 前綴 + Base64 格式回寫至 CSV
3. 終端顯示 `✔ 已加密明文密碼並回寫至 CSV`

**加密前的 CSV**：
```csv
Email,Password,DisplayName
player1@gmail.com,mypassword123,主帳號-法師
```

**加密後的 CSV**：
```csv
Email,Password,DisplayName
player1@gmail.com,ENC:AQAAANCMnd8BFd...（Base64 字串）,主帳號-法師
```

> 💡 加密綁定當前 Windows 使用者帳戶，換電腦或換 Windows 使用者後需重新輸入明文密碼。

---

## 資料目錄

工具的所有資料檔案統一存放在資料目錄下，預設位置為：

```
%USERPROFILE%\.d2r-multiboxing\
```

目錄結構：

```
~/.d2r-multiboxing/
├── config.json      # 設定檔
└── accounts.csv     # 帳號資料
```

若需自訂資料目錄位置，可設定環境變數 `D2R_MULTIBOXING_HOME`：

```powershell
$env:D2R_MULTIBOXING_HOME = "D:\MyD2R"
```

> 💡 資料目錄會在首次執行時自動建立，無需手動建立。

---

## 設定 D2R 路徑

設定檔 `config.json` 位於資料目錄中，內容如下：

```json
{
  "d2r_path": "C:\\Program Files (x86)\\Diablo II Resurrected\\D2R.exe"
}
```

若你的 D2R 安裝在非預設路徑，用文字編輯器修改 `d2r_path` 即可：

```powershell
notepad "$env:USERPROFILE\.d2r-multiboxing\config.json"
```

---

## 啟動工具

以 **管理員身份** 開啟 PowerShell，然後執行：

```powershell
cd C:\path\to\d2r-multiboxing
.\d2r-multiboxing.exe
```

> ⚠️ **必須以管理員權限執行**，否則無法存取 D2R 進程的 Handle。
>
> 快速方式：在開始選單搜尋「PowerShell」→ 右鍵 →「以系統管理員身分執行」。

---

## 操作說明

### 主選單

啟動後會顯示以下互動式選單：

```
============================================
  D2R Multiboxing Launcher
============================================

  資料目錄：C:\Users\User\.d2r-multiboxing
  D2R 路徑：C:\Program Files (x86)\Diablo II Resurrected\D2R.exe

  帳號列表：
  [1] 主帳號-法師      (player1@gmail.com)  [未啟動]
  [2] 副帳號-野蠻人    (player2@gmail.com)  [未啟動]

--------------------------------------------
  <數字>  啟動指定帳號
  a       啟動所有帳號
  r       重新整理狀態
  q       退出
--------------------------------------------
  請選擇：
```

### 啟動單一帳號

輸入帳號的 **ID 數字** 即可啟動該帳號，接著選擇連線區域：

```
  請選擇：1
  > 選擇區域 (1=NA, 2=EU, 3=Asia)：1
  正在啟動 主帳號-法師 (NA)...
  ✔ D2R 已啟動 (PID: 12345)
  ✔ 已關閉 1 個 Event Handle
  ✔ 視窗已重命名為 "主帳號-法師"
```

### 啟動所有帳號

輸入 `a` 一次啟動所有帳號：

```
  請選擇：a
  選擇區域 (1=NA, 2=EU, 3=Asia)：1
  正在啟動 主帳號-法師 (NA)...
  ✔ 主帳號-法師 已啟動 (PID: 12345)
  ✔ 主帳號-法師 已關閉 1 個 Handle
  正在啟動 副帳號-野蠻人 (NA)...
  ✔ 副帳號-野蠻人 已啟動 (PID: 12346)
  ✔ 副帳號-野蠻人 已關閉 1 個 Handle
  ✔ 視窗已重命名為 "主帳號-法師"
  ✔ 視窗已重命名為 "副帳號-野蠻人"
```

### 重新整理狀態

輸入 `r` 重新讀取帳號檔案並更新各帳號啟動狀態。

---

## 完整操作範例

以下是從零開始的完整流程：

```powershell
# Step 1: 編譯工具
cd C:\Users\User\GolandProjects\d2r-multiboxing
go build -o d2r-multiboxing.exe ./cmd/d2r-multiboxing

# Step 2:建立帳號設定檔
Copy-Item .\accounts.csv "$env:USERPROFILE\.d2r-multiboxing\accounts.csv"
# 用編輯器填入帳號資訊
notepad "$env:USERPROFILE\.d2r-multiboxing\accounts.csv"

# Step 3: 以管理員權限執行
# （右鍵 PowerShell → 以系統管理員身分執行）
.\d2r-multiboxing.exe

# Step 4: 在選單中操作
#   輸入 1 → 啟動第一個帳號
#   輸入 2 → 啟動第二個帳號
#   或輸入 a → 一次全部啟動
#   輸入 q → 退出
```

---

## 背景 Handle 監控

工具啟動後會自動在背景執行一個監控程序：

- 每 **2 秒** 掃描一次系統中的 D2R.exe 進程
- 自動關閉新偵測到的 `DiabloII Check For Other Instances` Event Handle
- 已處理過的進程不會重複操作

這表示即使你從 Battle.net Launcher 手動啟動 D2R（而非透過本工具），只要本工具正在運行，也會自動解除多開限制。

---

## 常見問題 FAQ

### Q: 啟動後提示「找不到 accounts.csv」

**A**: 請確認 `accounts.csv` 已放在資料目錄下（預設 `~/.d2r-multiboxing/`）。可以複製範本檔案：
```powershell
Copy-Item .\accounts.csv "$env:USERPROFILE\.d2r-multiboxing\accounts.csv"
```

### Q: 啟動後提示「啟動失敗」或「系統找不到指定的檔案」

**A**: D2R.exe 路徑可能不正確。請修改設定檔中的 `d2r_path`：
```powershell
notepad "$env:USERPROFILE\.d2r-multiboxing\config.json"
```

### Q: Handle 關閉失敗 / 權限不足

**A**: 請確認以 **管理員權限** 執行。在 PowerShell 視窗標題列應能看到「系統管理員」字樣。

### Q: 防毒軟體警告 / 誤報

**A**: 本工具需要操作其他進程的 Handle，這類行為會被部分防毒軟體標記。請將 `d2r-multiboxing.exe` 加入防毒軟體的例外清單。

### Q: 換電腦後密碼無法解密

**A**: Windows DPAPI 加密綁定當前使用者帳戶。換電腦或換 Windows 使用者後，請刪除 `accounts.csv` 中 `Password` 欄位的 `ENC:...` 內容，重新填入明文密碼，工具會再次加密。

### Q: 密碼欄位中有逗號怎麼辦？

**A**: 使用雙引號包覆密碼欄位，例如：
```csv
1,player@gmail.com,"my,password",主帳號,NA
```

### Q: 視窗重命名失敗

**A**: D2R 視窗建立需要時間，工具會自動重試最多 15 次（每次間隔 2 秒）。若仍失敗，可能是 D2R 尚未完全啟動，請稍後按 `r` 重新整理。

### Q: 我可以在工具運行時修改 accounts.csv 嗎？

**A**: 可以。修改後在選單中按 `r` 重新整理即可載入最新設定。

---

## 視窗模式設定

本工具 **不會** 自動修改 D2R 的顯示模式。若需要以視窗模式多開，請先手動設定：

1. 正常啟動 D2R（單開即可）
2. 進入 **選項 → 畫面 → 視窗模式**，選擇「**視窗化**」或「**無邊框視窗**」
3. 儲存設定並關閉遊戲

D2R 會將設定寫入 `%USERPROFILE%\Saved Games\Diablo II Resurrected\Settings.json`，後續透過本工具啟動的所有帳號都會套用該設定。

> 💡 此設定只需操作一次，除非你想切換回全螢幕模式。

---

## 注意事項

- ⚠️ **管理員權限**：必須以管理員身份執行，否則無法操作 D2R 進程 Handle
- ⚠️ **防毒誤報**：操作進程 Handle 為正常行為，但可能觸發防毒警告
- ⚠️ **服務條款**：使用本工具可能違反 Blizzard 服務條款，風險由使用者自行承擔
- ⚠️ **密碼安全**：密碼加密綁定當前 Windows 使用者，換機器需重新設定
- ⚠️ **頻繁登入限制**：短時間內重複啟動過多次會被 Battle.net 擋住（疑似 IP 頻率限制）。請避免反覆測試啟動，若遇到連線被拒，請等待數分鐘後再試(可能要上小時? 我不確定)。建議設定足夠的 `launch_delay` 間隔
- ℹ️ **僅支援 Battle.net 版本**，不支援 Steam 版 D2R
- ℹ️ 本工具 **不會** 修改遊戲檔案、注入程式碼或自動化任何遊戲操作

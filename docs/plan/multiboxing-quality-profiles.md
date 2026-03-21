# 每帳號畫質設定檔 implementation plan

## 目標與玩家操作流程

這版規劃不再只是回答「技術上能不能做」，而是明確鎖定玩家實際要怎麼用：

1. 玩家先進入 D2R，把想要的畫質 / 視窗 / 解析度設定調好
2. 回到 CLI，執行「儲存目前畫質設定檔」
3. 幫這組設定輸入名稱，讓它成為一個可重複使用的具名 profile
4. 玩家可以重複這個流程，累積多組畫質設定檔
5. 再透過類似 flag 設定介面的 CLI 流程，決定哪個 account 要使用哪一組 profile
6. 若 account 沒有指派 profile，啟動時就 **完全不碰** `%USERPROFILE%\Saved Games\Diablo II Resurrected\Settings.json`

也就是說，V1 的核心 UX 是：

- **手動調整遊戲內設定**
- **CLI 顯式儲存當前 `Settings.json` 為具名 profile**
- **CLI 顯式把 profile 指派給 account**
- **未指派 account 完全略過 profile apply**

## 已確認的關鍵技術事實

- D2R 的顯示 / 畫質 / 視窗設定目前仍以 `%USERPROFILE%\Saved Games\Diablo II Resurrected\Settings.json` 為核心，而不是 Windows 顯示控制台。
- 目前帳號資料模型在 [account.go](..\..\internal\multiboxing\account\account.go)；`accounts.csv` 目前只有 `Email,Password,DisplayName,LaunchFlags,ToolFlags` 五欄。
- 現有 per-account 設定 UI 範本在 [cli_flags.go](..\..\cmd\d2r-hyper-launcher\cli_flags.go)；它已經有成熟的 `runMenu` / `runMenuRead`、`b` / `h` / `q` 契約可沿用。
- 啟動切點在 [cli_launch.go](..\..\cmd\d2r-hyper-launcher\cli_launch.go) 很清楚：
  - 單帳號：`launchAccount()` 在 `selectLaunchMod()` 之後、`LaunchD2R()` 之前可插入 profile apply
  - 批次：`launchAll()` 在每個帳號呼叫 `LaunchD2R()` 前可逐帳號套用 profile
- repo 既有 `LaunchFlags` 只涵蓋 `-ns` / `-lq`；`-lq` 也已被文件標成可能失效，因此不能把它當作完整畫質方案。

## 這次規劃鎖定的產品決策

### 1. V1 採用「顯式儲存目前設定」而不是自動回存

玩家調整完遊戲內畫質後，要回 CLI 主動執行一次「儲存目前畫質設定檔」。

第一版不做：

- 遊戲結束時自動偵測並回存
- 依 account 自動同步最近一次關閉遊戲時的 `Settings.json`
- 執行中熱更新某個 instance 的畫質

這樣可以把責任邊界維持清楚：CLI 只負責 **讀取當前磁碟上的 `Settings.json`，另存成 profile，並在下次 launch 前覆蓋回去**。

### 2. 一個 account 同時間只指派一個 profile

`GraphicsProfile` 不走 bitmask。

每個 account 最多只有一個 active profile 指派：

- 有值：launch 前套用該 profile
- 空值：完全略過，不碰 `Settings.json`

### 3. V1 以「儲存 / 列表 / 指派 / 清除 / 刪除」為主，不追求完整 profile 管理器

第一版優先完成：

- 儲存目前 `Settings.json` 為具名 profile
- 在 CLI 列出既有 profiles
- 指派 profile 給 account
- 清除 account 的 profile 指派
- 刪除未再使用的 saved profile
- 刪除未再使用的 saved profile

先不把 scope 擴到：

- profile rename
- profile 匯出 / 匯入外部檔案
- 只 patch `Settings.json` 的部分欄位

## 建議的資料與檔案設計

### 1. 帳號 schema

在 [account.go](..\..\internal\multiboxing\account\account.go) 的 `Account` 與 `accounts.csv` 增加新欄位：

- `GraphicsProfile`

建議語意：

- 存 profile 名稱
- 空字串表示未指派

CSV 目標格式：

`Email,Password,DisplayName,LaunchFlags,ToolFlags,GraphicsProfile`

向後相容要求：

- 舊 5 欄 CSV 仍可正常載入
- 載入舊資料時 `GraphicsProfile` 自動視為空字串
- 寫回時統一升級成 6 欄

### 2. profile 存放位置

建議存放在 launcher home 下：

- `%USERPROFILE%\.d2r-hyper-launcher\graphics-profiles\`
- 若有 `D2R_HYPER_LAUNCHER_HOME`，則跟著 override 後的 home 走

第一版建議每個 profile 就是一份獨立的 JSON 檔：

- `graphics-profiles\<profile-name>.json`

其中內容就是當下複製出來的完整 `Settings.json`。

因為玩家會直接命名 profile，所以要有明確的命名規則：

- 不接受空名稱
- 不接受 Windows 非法檔名字元
- 不允許 silent overwrite
- 若名稱已存在，CLI 應顯示明確提示，讓玩家選擇覆蓋或取消

## 建議的 CLI UX

### 1. profile 儲存流程

新增一個玩家可操作入口，語意類似：

- 「儲存目前畫質設定檔」

流程建議：

1. 讀取 `%USERPROFILE%\Saved Games\Diablo II Resurrected\Settings.json`
2. 若檔案不存在或讀取失敗，顯示明確錯誤並 pause
3. 讓玩家輸入 profile 名稱
4. 若上方已列出既有 profiles，玩家可直接輸入編號覆蓋對應 profile
5. 若輸入的是其他文字，則視為新名稱並驗證是否合法
6. 若文字名稱剛好已存在，提示玩家改用上方對應編號覆蓋，而不是 silent overwrite
7. 寫入 `graphics-profiles\<name>.json`
8. 成功後顯示成功訊息與 profile 名稱

這個流程的重點是：**CLI 只複製目前磁碟上的 `Settings.json`**。

因此要在 plan 裡明確承認一個限制：

- 若玩家剛在遊戲裡改完設定，但 D2R 尚未把最新值寫回磁碟，CLI 存到的可能還是舊版本

所以第一版建議 UX 文案要明講：

- 請玩家在確認設定已落盤後再回 CLI 儲存
- 最保守的建議是「調整完成後先離開遊戲，再儲存 profile」

### 2. account 指派流程

這塊 UX 要明確對齊 [cli_flags.go](..\..\cmd\d2r-hyper-launcher\cli_flags.go) 的操作感，而不是另外發明一套完全不同的互動模型。

建議新增一個「帳號畫質設定檔」介面，外層至少提供：

- `1` 儲存目前畫質設定檔
- `2` 指派畫質設定檔
- `3` 清除畫質設定檔
- `4` 刪除已保存的畫質設定檔

其中「指派」模式建議提供兩種操作路徑，直接沿用 flags 的心智模型：

- 依 profile 選 account
- 依 account 選 profile

這樣玩家可以：

- 先選某個 profile，再一次套給多個帳號
- 或先選某個 account，再替它指定 profile

「清除」模式則可以先做最核心版本：

- 依 account 清除目前指派

這塊 UI 要維持現有契約：

- 所有子選單都走 `runMenu` / `runMenuRead`
- `b` 返回
- `h` 回主選單
- `q` 離開

### 3. account 未指派時的行為

這是這次需求的硬規則：

- 若 `GraphicsProfile == ""`，launch 流程就 **完全略過 profile apply**
- 不建立暫存 profile
- 不複製任何檔案
- 不覆蓋 `%USERPROFILE%\Saved Games\Diablo II Resurrected\Settings.json`

也就是說，未指派 account 的行為要跟現在一樣，只是沿用玩家系統上現成的 global `Settings.json`。

## 建議的 launch 行為

### 1. 單帳號 launch

在 [cli_launch.go](..\..\cmd\d2r-hyper-launcher\cli_launch.go) 的 `launchAccount()`：

1. 驗證 `D2R.exe`
2. 選 region / mod
3. 檢查 account 是否有 `GraphicsProfile`
4. 若有，先把對應 profile 複製到 `%USERPROFILE%\Saved Games\Diablo II Resurrected\Settings.json`
5. 若沒有，直接略過
6. 再進入 `GetDecryptedPassword()` / `LaunchD2R()`

### 2. 批次 launch

在 `launchAll()` 的每個帳號迴圈內，對每個 pending account 重複相同判斷：

- 有指派 profile → 先 apply，再 launch
- 沒有指派 → 完全略過 apply

### 3. 錯誤處理建議

第一版建議這樣定義：

- 單帳號 launch：
  - 若 account 指派了 profile，但 profile 檔案不存在，這次 launch 不改 `Settings.json`，並自動清空該帳號的 `GraphicsProfile`
  - 若 profile 讀取失敗 / JSON 無效，仍直接中止這次 launch，顯示錯誤並 pause
- 批次 launch：
  - 該帳號若 profile 檔案不存在，顯示 warning、清空該帳號的 `GraphicsProfile`，並沿用目前 `Settings.json` 繼續啟動
  - 該帳號若 profile 讀取失敗 / JSON 無效，顯示 warning，略過該帳號，繼續後面帳號

這樣可以避免最危險的情況：

- 玩家以為套到了指定畫質
- 實際上因為 profile 壞掉而默默沿用舊的 global `Settings.json`

## 主要風險與限制

- 這不是原生 per-account 隔離；本質上仍然是 **共用全域 `Settings.json` + launch-time swapping**。
- 若某個執行中的 D2R 在退出時把自己的設定寫回磁碟，global `Settings.json` 可能又被改掉；所以不能信任磁碟現況，必須在每次 launch 前重新套用。
- `launchAll()` 仍可能有「上一個 instance 尚未完全讀完設定、下一個 account 就覆蓋了 `Settings.json`」的競態風險；這需要搭配現有 delay 做實測。
- 「儲存目前 profile」讀的是磁碟檔，不是 live process；若 D2R 尚未落盤，玩家存到的可能不是最新設定。
- 第一版不碰 Windows 顯示設定，也不處理多螢幕 / OS 層級顯示器切換。

## 建議的 implementation todos

- `quality-settings-schema`
  - 擴充 `Account` 與 `accounts.csv`，新增 `GraphicsProfile`
  - 保持舊 5 欄 CSV 向後相容，存檔時升級為 6 欄

- `quality-profile-store`
  - 在 launcher home 下建立 `graphics-profiles` 存放邏輯
  - 新增「把目前 global `Settings.json` 儲存成具名 profile」的 helper
  - 新增 profile 列表、名稱驗證、存在檢查、覆蓋確認所需能力

- `quality-prelaunch-apply`
  - 在 `launchAccount()` / `launchAll()` 的 `LaunchD2R()` 前插入 conditional apply
  - 明確保證：未指派 account 完全不碰 `Settings.json`
  - 明確處理 profile 缺失 / 損毀 / 讀寫失敗，其中缺失 profile 需自動清空帳號指派

- `quality-cli-management`
  - 新增 player-facing CLI：
    - 儲存目前畫質設定檔
    - 依 profile 指派 account
    - 依 account 指派 profile
    - 清除 account 指派
  - UX 盡量對齊 [cli_flags.go](..\..\cmd\d2r-hyper-launcher\cli_flags.go) 的操作感與導覽契約

- `quality-docs-tests`
  - 補齊 CSV 6 欄 / 5 欄向後相容測試
  - 補齊 profile save / apply / skip-on-empty 行為測試
  - 更新 `README.md`、`README.en.md`、[multiboxing-usage-guide.md](..\multiboxing-usage-guide.md)
  - 視情況補充 [D2R_PARAMS.md](..\D2R_PARAMS.md)，說明這是 `Settings.json` profile 切換，而不是 Windows 顯示設定切換

## 非目標 / 暫不納入

第一版先不做：

- Windows 顯示設定切換
- 只修改 `Settings.json` 局部欄位的 patch 機制
- profile rename / delete / export / import
- 遊戲執行中即時熱切換畫質
- 依 account 自動回存離開遊戲後的新設定

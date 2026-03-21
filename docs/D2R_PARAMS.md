# D2R 啟動參數一覽

> 主要整理自 [D2R Command Line Options](https://gist.github.com/heinermann/788470a5fedb1c437c1c570fe46cdb57)，
> 並交叉參考 Blizzard 論壇與社群討論；若某些參數缺乏官方說明，會明確標示為「社群觀察」或「用途不明」。
> 標記為 `(Defunct?)` 的參數可能已在新版本中失效。
> 如果你是從多開文件進來，建議先搭配 [multiboxing-index.md](multiboxing-index.md) 與 [multiboxing-usage-guide.md](multiboxing-usage-guide.md) 一起看。

---

## Reference check

這次交叉檢查時，實際查閱的主要來源如下：

- [D2R Command Line Options](https://gist.github.com/heinermann/788470a5fedb1c437c1c570fe46cdb57)
  - 本文主要基底來源；用來對照參數名稱、預設值、字串長度限制，以及哪些項目原本就只標成 `??` 或 `(Defunct?)`
- [Blizzard 技術支援：Command Line Parameter Battlenet Client](https://us.forums.blizzard.com/en/d2r/t/command-line-parameter-battlenet-client/162742)
  - 用來交叉確認 `-username`、`-password`、`-address` 的直接登入用法，以及常見區域位址寫法
- [Blizzard 論壇：How start D2R in Windowed mode (No border)?](https://eu.forums.blizzard.com/en/d2r/t/how-start-d2r-in-windowed-mode-no-border/1187)
  - 用來確認 `-w` / windowed 模式在社群中的實際回報並不一致，因此本文改採保守描述，而不是直接判定失效
- [Blizzard 論壇：Pro controller support](https://us.forums.blizzard.com/en/d2r/t/pro-controller-support/20631)
  - 用來交叉確認 `-uid osi` 至少存在於公開玩家討論中，且可用於直接啟動 `D2R.exe`、略過 launcher 的情境

補充說明：

- `-countrycode`、`-data` 這類參數，目前能找到的多半仍是社群整理或玩家經驗，缺乏 Blizzard 正式參數文件，因此本文對這些項目刻意保留不確定性描述
- 若未來 Blizzard patch 導致行為改變，建議優先回頭檢查上面的 gist 與論壇討論，再配合實機測試更新本文

---

## 這份文件在多開文件中的位置

- 玩家快速開始與主選單入口：看 [README.md](../README.md)
- 完整操作流程、`LaunchFlags` 設定畫面與常見問題：看 [multiboxing-usage-guide.md](multiboxing-usage-guide.md)
- 多開底層原理、單實例 handle、背景 monitor：看 [multiboxing-technical-guide.md](multiboxing-technical-guide.md)
- 這份文件則專門負責：**D2R 啟動參數本身的用途、來源與不確定性標註**

目前 launcher 直接 expose 給多開玩家的，主要是：

- Battle.net 連線相關的 `-username`、`-password`、`-address`
- 每帳號 `LaunchFlags` 目前只會用到的 `-ns`
- 每帳號畫質差異改由 `GraphicsProfile` 在啟動前切換 `Settings.json`
- mod 啟動會用到的 `-mod <name> -txt`

---

## 帳號與連線

| 參數 | 說明 | 預設值 |
|------|------|--------|
| `-username <email>` | Battle.net 登入信箱，需搭配 `-password` 和 `-address` 使用，三者齊備時跳過啟動器 | 未設定 |
| `-password <password>` | Battle.net 密碼（最長 23 字元） | `""` |
| `-address <server>` | Battle.net 伺服器地址：`us.actual.battle.net`、`eu.actual.battle.net`、`kr.actual.battle.net` | `""` |

---

## 顯示與視窗

| 參數 | 說明 | 預設值 |
|------|------|--------|
| `-w` / `-window` / `-windowed` | 視窗模式啟動。社群回報不完全一致，但至少有多個來源仍列為可用；若要穩定切換視窗 / 全螢幕，仍建議以遊戲內設定或 `Settings.json` 為準 | 全螢幕 |
| `-gamma <num>` | 覆蓋 Gamma 值 | `0` |
| `-vsync <num>` | 覆蓋 VSync | `255` |
| `-lq` | 低畫質模式（Large Font Mode）；術士版本似乎已失效。本專案目前也不再 expose 這個 flag，改由 `GraphicsProfile` / `Settings.json` 流程處理每帳號畫質 | 關閉 |

---

## 音效

| 參數 | 說明 |
|------|------|
| `-ns` / `-nosound` | 無聲音模式 |

---

## 遊戲行為

| 參數 | 說明 | 預設值 |
|------|------|--------|
| `-resetofflinemaps` | 每次重置單人地圖（如同多人模式） | 關閉 |
| `-enablerespec` | ALT + 點擊加點按鈕可無限洗點 | 關閉 |
| `-newplayer` | 每次啟動都顯示校正畫面與開場影片 | 關閉 |
| `-nocompress` | 用途不明 | — |
| `-nosave` | 用途不明（推測為不儲存進度） | — |
| `-nopk` | 用途不明 | — |

---

## 日誌與除錯

| 參數 | 說明 | 預設值 |
|------|------|--------|
| `-log` | 啟用效能指標日誌，輸出至 `d2log.txt` | 關閉 |
| `-msglog` | 用途不明 | — |
| `-minimumloglevel <int>` | 設定最低日誌等級 | `3` |

---

## 在地化

| 參數 | 說明 | 預設值 |
|------|------|--------|
| `-locale <code>` | 設定語系（2 字元語言 + 2 字元區域，如 `enUS`、`zhTW`） | 未設定 |
| `-countrycode <code>` | 2 字元代碼；可能影響區域 / 在地化相關行為，但公開資料對實際效果說明不足 | 未設定 |

---

## Battle.net 與 mod 整合

| 參數 | 說明 | 預設值 |
|------|------|--------|
| `-uid osi` | 啟用 Battle.net Online Services Interface 模式；社群資料仍可見這個參數，但本專案目前不會自動附加，僅保留作為參數參考 | 未設定 |
| `-mod <name>` | 指定要載入的 mod 名稱；通常對應 `D2R.exe` 同層 `mods\<name>\` 資料夾 | 未設定 |
| `-txt` | 搭配 `-mod` 使用，讓 D2R 載入該 mod 的資料 | 關閉 |
| `-direct` | 直接使用 `Data\` 目錄與內部資料檔；gist 與社群多將它視為 mod / 資料測試用途 | 關閉 |
| `-data <path>` | 指定資料路徑；公開資料較少，常與 `-direct` / mod 測試脈絡一起出現 | 未設定 |

---

## 已失效 / 用途不明

| 參數 | 說明 | 預設值 |
|------|------|--------|
| `-ama` / `-pal` / `-sor` / `-nec` / `-bar` | (Defunct?) 設定角色職業 | Amazon |
| `-name <string>` | (Defunct?) 設定角色名稱（最長 60 字元） | 未設定 |
| `-realm <string>` | (Defunct?) 用途不明（最長 23 字元） | 未設定 |
| `-act <num>` | (Defunct?) 設定起始章節 | `1` |
| `-seed <num>` | (Defunct?) 設定亂數種子 | `0` |
| `-players <int>` | (Defunct?) 設定難度人數倍率 | `0` |
| `-maxplayers <int>` | (Defunct?) 設定最大玩家數 | `8` |
| `-leveldifference <int>` | (Defunct?) 設定加入遊戲的最大等級差 | `255` |
| `-per` | 用途不明 | — |
| `-aftermath <string>` | 用途不明（最長 23 字元） | 未設定 |
| `-s <string>` | 用途不明（最長 23 字元） | 未設定 |
| `-gametype <num>` | 用途不明 | `0` |
| `-arena <num>` | 用途不明 | `0` |
| `-joinid <num>` | 用途不明 | `1` |
| `-gamename <string>` | 用途不明（最長 23 字元） | 未設定 |
| `-bn <string>` | 用途不明（最長 23 字元） | 未設定 |
| `-mcpip <string>` | 用途不明（最長 23 字元） | 未設定 |
| `-tactmode <int>` | 用途不明 | `0` |
| `-lem` | 用途不明 | — |
| `-filter <string>` | 用途不明（最長 255 字元） | `""` |

---

## 多開實用參數建議

以下參數對多開場景特別有用：

| 參數 | 用途 |
|------|------|
| `-ns` | 副帳號靜音，避免多個遊戲音效互相干擾 |
| `-mod <name> -txt` | 讓 `0` 與 `a` 啟動流程可套用已安裝 mod |

> ⚠️ 即使 `-w` 在部分社群回報中仍可用，本專案仍建議先手動進入遊戲 **選項 → 畫面 → 視窗模式** 設定為「視窗化」。
> D2R 會把相關設定寫入 `%USERPROFILE%\Saved Games\Diablo II Resurrected\Settings.json`。
> 若你有在主選單 `g` 建立並指派畫質設定檔，launcher 會在啟動該帳號前先把對應 profile 覆蓋回 `Settings.json`；若帳號沒有指派 profile，launcher 就完全不會動這份檔案。
> 若你想讓副帳號使用不同畫質，請優先用主選單 `g` 的畫質設定檔流程，而不是依賴 `-lq`。

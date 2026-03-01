# D2R 啟動參數一覽

> 整理自 [D2R Command Line Options](https://gist.github.com/heinermann/788470a5fedb1c437c1c570fe46cdb57)，
> 標記為 `(Defunct?)` 的參數可能已在新版本中失效。

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
| `-w` / `-window` / `-windowed` | 視窗模式啟動（⚠️ 已確認失效，需透過 `Settings.json` 的 `"Window Mode": 1` 設定） | 全螢幕 |
| `-gamma <num>` | 覆蓋 Gamma 值 | `0` |
| `-vsync <num>` | 覆蓋 VSync | `255` |
| `-lq` | 低畫質模式（Large Font Mode） | 關閉 |

---

## 音效

| 參數 | 說明 |
|------|------|
| `-ns` / `-nosound` | 無聲音模式 |
| `-sndbkg` | 用途不明 |

---

## Mod 與資料

> ⚠️ **D2R v3.1 (Reign of the Warlock) 已禁用 `-mod`、`-direct`、`-txt` 參數。**
> 這些參數會被 D2R 靜默忽略，不會報錯也不會載入 Mod。
> 目前唯一可靠的 Mod 載入方式是透過 [D2RMM](https://github.com/olegbl/d2rmm)。
> 詳細調查記錄請參考 [doc/D2R-MOD-LOADING-ROTW.md](doc/D2R-MOD-LOADING-ROTW.md)。

| 參數 | 說明 | RotW 狀態 |
|------|------|-----------|
| `-direct` | 直接讀取 `Data/` 目錄的內部資料檔 | ❌ 已失效 |
| `-mod <name>` | 載入 mod 資料，路徑為 `mods/<name>/<name>.mpq`（最長 23 字元） | ❌ 已失效 |
| `-txt` | 搭配 `-direct` 或 `-mod` 使用 `.txt` 檔案代替 `.bin` | ❌ 已失效 |
| `-data <path>` | 自訂資料路徑（最長 259 字元） | ❓ 未測試 |

---

## 遊戲行為

| 參數 | 說明 | 預設值 |
|------|------|--------|
| `-resetofflinemaps` | 每次重置單人地圖（如同多人模式） | 關閉 |
| `-enablerespec` | ALT + 點擊加點按鈕可無限洗點 | 關閉 |
| `-norumble` | 停用手把震動 | 關閉 |
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
| `-countrycode <code>` | 國家代碼（2 字元） | 未設定 |

---

## 影片

| 參數 | 說明 |
|------|------|
| `-skiplogovideo` | 跳過開場 Logo 影片（效果未確認） |

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
| `-lq` | 副帳號低畫質，降低系統資源佔用 |

> ⚠️ `-w` 視窗模式參數已失效。請先手動進入遊戲 **選項 → 畫面 → 視窗模式** 設定為「視窗化」，
> D2R 會將設定寫入 `%USERPROFILE%\Saved Games\Diablo II Resurrected\Settings.json`，
> 後續透過本工具啟動的所有帳號都會套用該設定。

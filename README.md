# d2r-hyper-launcher

Windows 上給 D2R（Diablo II: Resurrected）玩家使用的 CLI 工具，目前提供兩個主要功能：

- **multiboxing**：多帳號啟動、單實例鎖處理、視窗辨識
- **switcher**：鍵盤／滑鼠側鍵／搖桿切換 D2R 視窗

**目前最新版本：[v1.1.0](docs/releases/v1.1.0.md)**

> CLI 各選單若遇到玩家輸入格式、範圍或選項錯誤，會先顯示錯誤訊息，再提示玩家按鍵確認後回到原流程；在可直接讀取單鍵的終端會顯示「按任意鍵繼續」，其他終端則會自動改成「按 Enter 繼續」。

## 多開文件導覽

| 你現在想做什麼 | 應該看哪份文件 |
|---|---|
| 先快速上手，直接開始用 | 本頁 README |
| 看完整的玩家操作流程與常見問題 | [docs/multiboxing-usage-guide.md](docs/multiboxing-usage-guide.md) |
| 先搞懂多開文件之間怎麼分工 | [docs/multiboxing-index.md](docs/multiboxing-index.md) |
| 查 D2R 啟動參數、LaunchFlags、`-mod` / `-txt` 細節 | [docs/D2R_PARAMS.md](docs/D2R_PARAMS.md) |
| 看底層多開實作與架構 | [docs/multiboxing-technical-guide.md](docs/multiboxing-technical-guide.md) |

## 給玩家的快速開始

這一段是給「只想直接享用」的玩家看的，不需要先懂 Go 或程式碼。

### 1. 先下載啟動器

- [d2r-hyper-launcher.exe](d2r-hyper-launcher.exe)

放在任意位置執行即可，所有相關資料會在 `%USERPROFILE%\.d2r-hyper-launcher` 這個資料夾裡。

### 2. 直接雙擊 `d2r-hyper-launcher.exe`

下載後直接雙擊即可，照畫面引導做就能完成第一次設定。

- 第一次執行時，工具會自動建立 `%USERPROFILE%\.d2r-hyper-launcher` 這個資料夾以及 `config.json` (設定檔) 和 `accounts.csv` (帳號檔) 兩個檔案
- 上述東西建立完成後，會引導你退出工具並自動開啟這個資料夾位置，方便你直接修改 `accounts.csv` 裡的帳號資訊
- 如果你的 D2R 不在預設路徑，工具也會引導你用 `p` 去選正確的 `D2R.exe` 路徑

### 3. 修改多開帳號清單

本工具會固定讀取下面這個位置的帳號檔： `%USERPROFILE%\.d2r-hyper-launcher\accounts.csv`

(如果你不知道 `%USERPROFILE%` 是哪裡也沒關係，每次執行該工具都顯示這個「資料目錄」完整路徑。)

建議 **優先用 Excel 開啟** 後再編輯，避免手動修改時不小心破壞 CSV 格式。  
打開後，照這個格式填入 Battle.net 帳號：

```csv
Email,Password,DisplayName,LaunchFlags
your-account1@example.com,your-password-here,主帳號-法師(倉庫/武器/飾品),
your-account2@example.com,your-password-here,副帳號-野蠻人(廢寶/鑲材),
```

欄位說明：

- `Email`：Battle.net 登入信箱
- `Password`：Battle.net 密碼
- `DisplayName`：工具內顯示名稱，也是視窗切換時看到的名稱；啟動後的視窗會被命名為 `D2R-<DisplayName>`
- `LaunchFlags`：每個帳號額外要帶的啟動旗標；可以先留空，工具會自動 fallback 成 `0`，之後再回到主選單用 `f` 設定。目前工具只提供兩種 flag 設定；依目前測試結果，「關閉聲音」有明顯作用，但 `-lq` 看起來沒有實際效果。參數用途可再查 [docs/D2R_PARAMS.md](docs/D2R_PARAMS.md)

> ⚠️ 如果 D2R 還開著時就修改 `DisplayName`，重新打開 launcher 後的 `[已啟動]` / `[未啟動]` 狀態顯示可能暫時不準。  
> 建議在所有遊戲視窗都關閉後再改名稱，或回到主選單用 `r` 重新整理狀態。

> 第一次執行後，明文密碼會自動改寫成 `ENC:` 開頭的加密字串。  
> 這是用 Windows DPAPI 加密，之後不用自己手動加密；若換電腦或換 Windows 使用者，請再填一次明文密碼。
>
> 如果你進遊戲後卡在 Battle.net 登入中的訊息，高機率是 `accounts.csv` 裡的密碼填錯了，請先回去確認該帳號密碼是否正確。

### 4. 回到工具開始使用

把 `accounts.csv` 改好之後，再次雙擊 `d2r-hyper-launcher.exe` 即可開始使用。

啟動後，就會看到以下選項畫面：

- `<數字>`：先選區域，再選一次要套用的已安裝 mod，之後啟動指定帳號；若該帳號已在執行中會直接阻止重複啟動
- `a`：工具會先預掃描目前已開啟帳號；如果還有待啟動帳號，才會再讓你選區域與 mod，並只對尚未啟動的帳號依序啟動
- `0`：先選一次要套用的已安裝 mod（如果有），再進離線模式
- `d`：設定 `a` 批次啟動時的啟動間隔；可輸入 `30` 或 `30-60` 這種範圍，代表每次都在該區間內隨機等待，且下限固定不可低於 10 秒
- `f`：先顯示帳號列表與置中的兩行 flag 對照表，再設定或取消各帳號的額外啟動 flag；目前只提供兩種 flag 設定，其中依目前測試結果，「關閉聲音」有明顯作用，但 `-lq` 看起來沒有實際效果。進入第二層後，前兩個選項會依目前動作自動切成「設定」或「取消」版本，另可直接「設定 / 取消所有帳號所有 flag」，各旗標用途可查 [docs/D2R_PARAMS.md](docs/D2R_PARAMS.md)
- `p`：開啟 Windows 檔案選擇視窗，設定 `D2R.exe` 路徑
- `s`：設定視窗切換快捷鍵／滑鼠側鍵／搖桿按鍵
- `r`：重新讀取 `accounts.csv` 並刷新狀態
- `q`：離開工具

### 5. 想看更仔細的操作說明

如果你想看每個選單怎麼用、每個步驟會看到什麼畫面，請直接讀：

- [docs/multiboxing-index.md](docs/multiboxing-index.md) — 多開相關文件總覽與閱讀順序
- [docs/multiboxing-usage-guide.md](docs/multiboxing-usage-guide.md) — 多開啟動、帳號檔、區域選擇、離線模式
- [docs/switcher-usage-guide.md](docs/switcher-usage-guide.md) — 視窗切換設定、支援按鍵類型、常見問題

如果你想看底層實作與技術原理，再讀：

- [docs/multiboxing-technical-guide.md](docs/multiboxing-technical-guide.md)
- [docs/D2R_PARAMS.md](docs/D2R_PARAMS.md) — D2R 啟動參數與目前 LaunchFlags / mod 參考
- [docs/switcher-technical-guide.md](docs/switcher-technical-guide.md)

## 注意事項

- 建議先把 D2R 設成「視窗化」或「無邊框視窗」
- 設定搖桿切換按鍵時，建議以管理員權限執行，測試階段發現非管理者權限會抓不到搖桿訊號。
- `switcher` 只有在 `d2r-hyper-launcher` 工具持續開著時才會生效；如果把工具關掉，切窗功能就會停止作用
- `a` 批次啟動預設的 `launch_delay` 是 10 秒；為了向後相容，若讀到舊版預設留下的 `5` 秒設定，工具會自動按 10 秒處理。Battle.net 端仍可能因短時間內太頻繁重複登入／關閉而擋線，因此如果你要調高或調低，請在主選單輸入 `d`，並注意下限固定不可低於 10 秒
- 盡量不要手動修改 `config.json`，避免不小心破壞 JSON 格式；大部分設定請優先透過工具內建選單調整
- 僅支援 Battle.net 版 D2R
- 操作進程 Handle 可能被部分防毒軟體誤報
- 本工具不會修改遊戲檔案、注入遊戲程式或自動化遊戲操作
- 本工具為社群自用工具，與 Blizzard Entertainment 無關；使用風險自負，本作者不對任何風險、損失或後果負責

## 給想自己編譯的人

### 前置需求

- Windows 10 / 11
- Go 1.26+
- Battle.net 版 D2R

### 編譯

若你只是要在本機驗證程式可正常啟動，請優先使用：

```powershell
.\scripts\go-run.ps1
```

若你只是要驗證能否成功編譯，建議把輸出放到暫存位置，避免覆蓋 repo 根目錄的 release exe：

```powershell
New-Item -ItemType Directory -Force .\.tmp | Out-Null
go build -o .\.tmp\d2r-hyper-launcher-dev.exe ./cmd/d2r-hyper-launcher
```

只有在 release 流程要更新正式產物時，才覆蓋 repo 根目錄的 `d2r-hyper-launcher.exe`，並同時注入版本與 release 時間（`yyyy-mm-dd hh:mm:ss`）：

```powershell
go build -ldflags "-X main.version=vX.Y.Z -X main.releaseTime=YYYY-MM-DD HH:MM:SS" -o d2r-hyper-launcher.exe ./cmd/d2r-hyper-launcher
```

### 測試

在這台 Windows 環境若直接跑 `go test ./...` 被 Application Control 擋下，請改用 repo 內建包裝腳本：

```powershell
.\scripts\go-test.ps1
go build -o .\.tmp\d2r-hyper-launcher-dev.exe ./cmd/d2r-hyper-launcher
```

## 授權

MIT License

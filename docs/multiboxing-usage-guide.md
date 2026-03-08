# Multiboxing 使用導覽

> 這份文件專門說明玩家實際怎麼用多開啟動器。  
> 如果你只想快速上手，先看 [README.md](../README.md)；如果你想先看整份多開文件怎麼分工，讀 [multiboxing-index.md](multiboxing-index.md)；如果你想看底層原理，再看 [multiboxing-technical-guide.md](multiboxing-technical-guide.md)。

> 補充：CLI 各選單若遇到玩家輸入格式、範圍或選項錯誤，會先顯示錯誤訊息，再提示玩家按鍵確認後回到原流程；在可直接讀取單鍵的終端會顯示「按任意鍵繼續」，其他終端則會自動改成「按 Enter 繼續」。

## 文件導覽

- 新手快速開始：[README.md](../README.md)
- 多開文件總覽：[multiboxing-index.md](multiboxing-index.md)
- 技術原理：[multiboxing-technical-guide.md](multiboxing-technical-guide.md)
- D2R 參數參考：[D2R_PARAMS.md](D2R_PARAMS.md)

## 這份文件會講什麼

- `accounts.csv` 要放哪裡、怎麼填
- `config.json` 會存哪些設定
- 主選單每個功能怎麼用
- 單一帳號、多帳號、離線模式的操作流程
- 常見問題怎麼排查

## 第一次使用前要準備什麼

1. 下載 [d2r-hyper-launcher.exe](../d2r-hyper-launcher.exe)
2. 第一次執行時讓工具自動建立 `accounts.csv`
3. 確認 D2R 是 Battle.net 版本
4. 建議先把遊戲顯示模式設成「視窗化」或「無邊框視窗」

## 資料目錄在哪裡

工具的資料都放在：

```text
%USERPROFILE%\.d2r-hyper-launcher\
```

如果你看不懂 `%USERPROFILE%` 代表哪個資料夾，可以先執行一次 `d2r-hyper-launcher.exe`。工具一啟動就會顯示「資料目錄：...」的完整路徑，並在缺少 `accounts.csv` 時自動建立範本檔。

如果第一次執行時 `accounts.csv` 還不存在，工具會自動建立一份含範例資料的 `accounts.csv`，並在你按任意鍵後結束程式、自動開啟資料目錄，方便你直接修改內容。

裡面最常看到這兩個檔案：

```text
%USERPROFILE%\.d2r-hyper-launcher\
├── accounts.csv
└── config.json
```

## `accounts.csv` 怎麼準備

最簡單的做法是先執行一次 launcher，讓它自動產生範本；之後工具在離開時會自動開啟資料目錄，方便你直接修改這份檔案。

範例內容：

```csv
Email,Password,DisplayName,LaunchFlags
your-account1@example.com,your-password-here,主帳號-法師(倉庫/武器/飾品),
your-account2@example.com,your-password-here,副帳號-野蠻人(廢寶/鑲材),
```

欄位說明：

| 欄位 | 必填 | 說明 |
|------|------|------|
| `Email` | ✅ | Battle.net 登入信箱 |
| `Password` | ✅ | 帳號密碼；第一次啟動後會自動改寫成加密字串 |
| `DisplayName` | ✅ | 主選單顯示名稱，也是視窗標題後綴 |
| `LaunchFlags` | 可先留空 | 這個帳號額外要帶的啟動旗標 bitflag；一般玩家可先留空，工具會自動 fallback 成 `0`，之後再回到主選單用 `f` 設定 |

`LaunchFlags` 目前常見對應如下：

| 旗標 | 用途 |
|------|------|
| `-ns` | 關閉聲音 |
| `-sndbkg` | 背景保留聲音 |
| `-lq` | 低畫質 / Large Font Mode（效果依版本而定） |
| `-skiplogovideo` | 跳過 Logo 影片 |
| `-norumble` | 停用手把震動 |

更完整的參數說明、來源與不確定性標註，請再查 [D2R_PARAMS.md](D2R_PARAMS.md)。

### 密碼加密會發生什麼事

第一次啟動時，如果 `Password` 還是明文，工具會自動：

1. 用目前 Windows 使用者的 DPAPI 進行加密
2. 把密碼改寫成 `ENC:` 開頭的字串
3. 直接回寫到同一份 `accounts.csv`

這代表：

- 之後不需要再手動加密
- 換電腦或換 Windows 使用者後，舊密文通常無法解開
- 如果你要改密碼，直接再填一次明文即可

## `config.json` 會存什麼

第一次執行工具時，如果 `config.json` 不存在，系統會自動建立預設檔。

範例：

```json
{
  "d2r_path": "C:\\Program Files (x86)\\Diablo II Resurrected\\D2R.exe",
  "launch_delay": "30-60"
}
```

欄位說明：

- `d2r_path`：`D2R.exe` 的路徑
- `launch_delay`：使用 `a` 啟動全部帳號時，每個帳號之間的等待秒數；可寫固定秒數 `30`，或寫成 `30-60` 代表每次隨機取值

一般玩家不需要手動修改 `config.json`。如果你的遊戲不是裝在預設路徑，請在主選單輸入 `p`，工具會直接開啟 Windows 檔案選擇視窗，讓你選擇正確的 `D2R.exe`；如果你直接按 `<數字>`、`a` 或 `0` 啟動時發現目前路徑失效，工具也會先攔下來並立即提供同樣的 `p` 設定流程。

## 啟動後主選單怎麼看

主選單大致會長這樣：

```text
============================================
  d2r-hyper-launcher
============================================

  資料目錄：C:\Users\User\.d2r-hyper-launcher
  D2R 路徑：C:\Program Files (x86)\Diablo II Resurrected\D2R.exe
  啟動間隔：30 秒

  帳號列表：
  [1] 主帳號-法師      (player1@gmail.com)  [未啟動]
  [2] 副帳號-野蠻人    (player2@gmail.com)  [未啟動]

--------------------------------------------
  <數字>  啟動指定帳號
  0       離線遊玩（可選 mod，不需帳密）
  a       啟動所有帳號（可選 mod，只啟動未啟動的）
  d       設定啟動間隔
  f       設定帳號啟動 flag
  p       選擇 D2R.exe 路徑
  s       視窗切換設定
  r       重新整理狀態
  q       退出
--------------------------------------------
```

`[未啟動]` / `[已啟動]` 會根據目前視窗狀態更新。

注意：這個狀態判斷是用 `accounts.csv` 裡的 `DisplayName` 去對應 `D2R-<DisplayName>` 視窗標題。如果 D2R 還在執行中時，你先關掉 launcher 再去修改 `DisplayName`，重新開回來後這裡的狀態偵測就可能暫時不正確。若你已經改了名稱，可以先在主選單輸入 `r` 重新整理狀態；更穩妥的做法，仍是等所有遊戲視窗關閉後再修改 `DisplayName`。

## 設定啟動間隔

在主選單輸入 `d` 後，可以直接設定 `a` 批次啟動時每個帳號之間要等幾秒。

- 固定下限是 `10` 秒，不能再低
- 可輸入 `30`，代表固定等待 30 秒
- 可輸入 `30-60`，代表每次隨機等待 30 到 60 秒
- 這個設定會直接回寫到 `config.json` 的 `launch_delay`

之所以預設保留 30 秒，是因為實際使用上若短時間內太頻繁重複登入／關閉，Battle.net 端有機率把連線擋住；如果你要縮短，請自行衡量風險。

## 設定帳號啟動 flag

在主選單輸入 `f` 後，工具會先列出帳號列表，接著顯示一個置中的 flag 對照表：每個 flag 欄位會分成兩行，上面是中文名稱，下面是實際啟動參數，帳號若有啟用該 flag 就會在該格標 `v`。之後你可以：

1. 先選這次要「設定 flag」還是「取消 flag」
2. 再選操作維度：
   - 以 flag 為維度：先選某個 flag，再輸入要套用到哪些帳號
   - 以帳號為維度：先選某個帳號，再輸入要調整哪些 flag
3. 帳號或 flag 編號都支援：
   - `2,4,6`
   - `1-3,5-7`
4. 工具會先把解析後的結果重新列出來給你確認，確認後才會回寫到 `accounts.csv`

注意：

- `5-3` 這種反向區間會直接被視為錯誤，不會套用
- 這個功能只會改 `LaunchFlags`；帳號、密碼、DisplayName 仍建議先手動在 `accounts.csv` 裡建立
- `-lq` 在最新版本仍可列為候選，但本文與工具介面都會標註「效果依版本而定」
- 如果你手動把 `LaunchFlags` 填成亂數、負數或文字，工具會在讀取時自動 fallback 成 `0`，並把 CSV 回寫成乾淨值
- 如果你輸入的編號範圍或格式有誤，工具會先顯示錯誤訊息，接著提示你按鍵確認後再回到上一層，避免訊息瞬間被主選單蓋掉
- 如果你想知道某個 flag 對應的 D2R 參數實際是什麼，請再查 [D2R_PARAMS.md](D2R_PARAMS.md)

## 啟動單一帳號

1. 在主選單輸入帳號前面的數字，例如 `1`
2. 工具會先檢查目前設定的 `D2R.exe` 是否真的存在；如果路徑失效，會直接提示你輸入 `p` 重新選路徑
3. 如果這個帳號對應的 `D2R-<DisplayName>` 視窗已經存在，工具會直接阻止重複啟動，避免同一帳號被連續再開一次
4. 選擇區域：`1=NA`、`2=EU`、`3=Asia`
5. 如果 `D2R.exe` 同層的 `mods\` 目錄下有已安裝 mod，工具也會讓你選擇這次單帳號啟動要套用哪一個 mod
6. 如果這個帳號在 `LaunchFlags` 已設定額外旗標，工具會一併帶入這些啟動參數
7. 工具會自動：
   - 解密帳號密碼
   - 啟動 D2R
   - 嘗試關閉單實例鎖的 Event Handle
   - 把視窗標題改成 `D2R-<DisplayName>`

常見畫面：

```text
  選擇區域 (1=NA, 2=EU, 3=Asia)
  b       回上一層
  h       回主選單
  q       離開程式
  > 請選擇：1

  正在啟動 主帳號-法師 (NA)...
  ✔ D2R 已啟動 (PID: 12345)
  ✔ 已關閉 1 個 Event Handle
  ✔ 視窗已重命名為 "D2R-主帳號-法師"
```

### 可選區域

| 輸入 | 區域 | 實際位址 |
|------|------|----------|
| `1` / `NA` | 美洲 | `us.actual.battle.net` |
| `2` / `EU` | 歐洲 | `eu.actual.battle.net` |
| `3` / `Asia` | 亞洲 | `kr.actual.battle.net` |

## 啟動所有帳號

在主選單輸入 `a` 之後：

1. 工具會先檢查目前設定的 `D2R.exe` 是否真的存在；如果路徑失效，會直接提示你輸入 `p` 重新選路徑
2. 工具會先預掃描目前已開啟的 D2R 視窗，並把整份帳號清單明確列出成 `[已啟動]` / `[未啟動]`
3. 如果全部帳號都已經在執行中，就會直接結束，不再多問區域或 mod
4. 確定還有待啟動帳號時，才會再讓你選區域
5. 如果 `D2R.exe` 同層的 `mods\` 目錄下有已安裝 mod，工具會先讓你選擇這次批次啟動要套用哪一個 mod
6. 每次啟動時只會處理上面標示為 `[未啟動]` 的帳號，不會把等待時間浪費在已經開啟的帳號上
7. 每個帳號若已設定自己的 `LaunchFlags`，工具也會在該帳號啟動時一併帶入
8. 每次啟動之間只會在「下一個真的還沒啟動的帳號」之前等待 `launch_delay`；若你設定的是範圍，工具會在每次等待前重新隨機取一個秒數

## 離線模式

在主選單輸入 `0`，工具會先檢查目前設定的 `D2R.exe` 是否存在；如果路徑失效，會直接提示你輸入 `p` 重新選路徑。確認路徑有效後，工具會再檢查 `D2R.exe` 同層的 `mods\` 目錄是否有已安裝 mod；如果有，你可以先選這次離線啟動要使用哪個 mod，之後再啟動 D2R 離線模式，不需要讀取帳號密碼。

這個模式適合：

- 只想快速進單機
- 臨時測試視窗或路徑設定
- 不想動到 Battle.net 帳號資料

## 已安裝 mod 會怎麼被偵測

本工具目前會讀取：

```text
<D2R.exe 所在資料夾>\mods\
```

只有符合下列條件之一的資料夾，才會出現在選單裡：

- 是 `mods\` 底下的子資料夾
- 該資料夾內有 `modinfo.json`
- 或該資料夾內有一個與資料夾同名的 `<mod>.mpq`

例如：

```text
C:\Program Files (x86)\Diablo II Resurrected\mods\
└── my-mod\
    └── modinfo.json
```

或：

```text
C:\Program Files (x86)\Diablo II Resurrected\mods\
└── MCMod\
    └── MCMod.mpq
```

當你在 `0` 或 `a` 選到某個 mod 時，工具會用該資料夾名稱作為 `-mod <name> -txt` 參數來啟動 D2R。

## 重新整理狀態

在主選單輸入 `r`，工具會重新：

- 讀取 `accounts.csv`
- 刷新帳號狀態顯示

這在你修改帳號檔後很有用，不用把整個工具重開。

## 子選單通用操作

只要進到子選單，通常都可以用這三個指令：

- `b`：回上一層
- `h`：回主選單
- `q`：直接離開程式

## 常見問題

### 找不到 `accounts.csv`

請直接重新執行 `d2r-hyper-launcher.exe`。工具會自動在下面位置建立新的範本檔：

```text
%USERPROFILE%\.d2r-hyper-launcher\accounts.csv
```

建立完成後，畫面會停住提示你；當你按任意鍵離開時，工具會自動開啟資料目錄，方便你直接修改這份 `accounts.csv`。

### 顯示找不到 `D2R.exe`

當你按 `<數字>`、`a` 或 `0` 啟動時，工具會先攔下來並直接顯示 `p / b / h / q` 選項。通常只要輸入 `p`，重新選一次正確的 `D2R.exe` 路徑即可。

### 換電腦後密碼失效

這是正常現象，因為 DPAPI 會綁定目前 Windows 使用者。請直接把 `accounts.csv` 的密碼欄位改回明文，再重新執行工具。

## 延伸閱讀

- [multiboxing-index.md](multiboxing-index.md) — 多開文件總覽與閱讀順序
- [README.md](../README.md) — 新手快速開始
- [D2R_PARAMS.md](D2R_PARAMS.md) — `LaunchFlags`、`-uid osi`、mod 參數參考
- [switcher-usage-guide.md](switcher-usage-guide.md) — 視窗切換功能教學
- [multiboxing-technical-guide.md](multiboxing-technical-guide.md) — multiboxing 技術原理

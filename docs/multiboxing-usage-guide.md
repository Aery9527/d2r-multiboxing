# Multiboxing 使用導覽

> 這份文件專門說明玩家實際怎麼用多開啟動器。  
> 如果你只想快速上手，先看 [README.md](../README.md)；如果你想看底層原理，再看 [multiboxing-technical-guide.md](multiboxing-technical-guide.md)。

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
Email,Password,DisplayName
your-account1@example.com,your-password-here,主帳號-法師(倉庫/武器/飾品)
your-account2@example.com,your-password-here,副帳號-野蠻人(廢寶/鑲材)
```

欄位說明：

| 欄位 | 必填 | 說明 |
|------|------|------|
| `Email` | ✅ | Battle.net 登入信箱 |
| `Password` | ✅ | 帳號密碼；第一次啟動後會自動改寫成加密字串 |
| `DisplayName` | ✅ | 主選單顯示名稱，也是視窗標題後綴 |

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
  "launch_delay": 5
}
```

欄位說明：

- `d2r_path`：`D2R.exe` 的路徑
- `launch_delay`：使用 `a` 啟動全部帳號時，每個帳號之間的等待秒數

一般玩家不需要手動修改 `config.json`。如果你的遊戲不是裝在預設路徑，請在主選單輸入 `p`，工具會直接開啟 Windows 檔案選擇視窗，讓你選擇正確的 `D2R.exe`。

## 啟動後主選單怎麼看

主選單大致會長這樣：

```text
============================================
  d2r-hyper-launcher
============================================

  資料目錄：C:\Users\User\.d2r-hyper-launcher
  D2R 路徑：C:\Program Files (x86)\Diablo II Resurrected\D2R.exe
  啟動間隔：5 秒

  帳號列表：
  [1] 主帳號-法師      (player1@gmail.com)  [未啟動]
  [2] 副帳號-野蠻人    (player2@gmail.com)  [未啟動]

--------------------------------------------
  <數字>  啟動指定帳號
  0       離線遊玩（可選 mod，不需帳密）
  a       啟動所有帳號（可選 mod，只啟動未啟動的）
  p       選擇 D2R.exe 路徑
  s       視窗切換設定
  r       重新整理狀態
  q       退出
--------------------------------------------
```

`[未啟動]` / `[已啟動]` 會根據目前視窗狀態更新。

## 啟動單一帳號

1. 在主選單輸入帳號前面的數字，例如 `1`
2. 選擇區域：`1=NA`、`2=EU`、`3=Asia`
3. 工具會自動：
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

1. 一樣先選區域
2. 如果 `D2R.exe` 同層的 `mods\` 目錄下有已安裝 mod，工具會先讓你選擇這次批次啟動要套用哪一個 mod
3. 工具會先預掃描目前已開啟的 D2R 視窗，整理出這次真正尚未啟動的帳號清單
4. 已經開啟的帳號會直接跳過，不會把等待時間浪費在這些帳號上
5. 每次啟動之間只會在「下一個真的還沒啟動的帳號」之前等待 `launch_delay` 秒

## 離線模式

在主選單輸入 `0`，工具會先檢查 `D2R.exe` 同層的 `mods\` 目錄是否有已安裝 mod；如果有，你可以先選這次離線啟動要使用哪個 mod，之後再啟動 D2R 離線模式，不需要讀取帳號密碼。

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

請在主選單輸入 `p`，重新選一次正確的 `D2R.exe` 路徑。

### 換電腦後密碼失效

這是正常現象，因為 DPAPI 會綁定目前 Windows 使用者。請直接把 `accounts.csv` 的密碼欄位改回明文，再重新執行工具。

## 延伸閱讀

- [README.md](../README.md) — 新手快速開始
- [switcher-usage-guide.md](switcher-usage-guide.md) — 視窗切換功能教學
- [multiboxing-technical-guide.md](multiboxing-technical-guide.md) — multiboxing 技術原理

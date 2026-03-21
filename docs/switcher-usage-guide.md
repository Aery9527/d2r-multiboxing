# Switcher 使用導覽

> 這份文件專門說明視窗切換功能怎麼設定與使用。  
> 如果你還沒完成基本多開設定，請先看 [README.md](../README.md) 與 [multiboxing-usage-guide.md](multiboxing-usage-guide.md)。

## 這個功能能做什麼

當你同時開了多個 D2R 視窗後，可以用：

- 鍵盤快捷鍵
- 滑鼠側鍵
- 搖桿按鈕或搖桿組合鍵

在這些 D2R 視窗之間輪流切換焦點。

## 使用前先知道

- switcher 只會切換標題以 `D2R-` 開頭的視窗
- 這些視窗名稱通常會在你用 launcher 啟動帳號後自動完成
- 如果目前只有 1 個 D2R 視窗，按切換鍵不會有明顯效果
- 首次設定搖桿切換按鍵時，建議用管理員權限執行工具

## 怎麼進入切換設定

1. 啟動 `d2r-hyper-launcher.exe`
2. 在主選單輸入 `s`
3. 進入「視窗切換設定」

畫面大致如下：

```text
  === 視窗切換設定 ===
  目前狀態：未啟用

  [1] 設定切換按鍵
  [2] 設定切換帳號
  [0] 切換為開啟
  b       回上一層
  h       回主選單
  q       離開程式
  > 請選擇：
```

## 設定切換帳號

預設所有帳號都會進入切換循環。如果你有倉庫號、PvP 號等不想被循環到的帳號，可以把它排除。

在切換設定畫面輸入 `2`，會顯示所有帳號的目前狀態：

```text
  === 切換帳號設定 ===
  [1] 主帳號-法師  [已包含]
  [2] 副帳號-野蠻  [已包含]
  [3] 倉庫號       [已排除]

  [a] 全部包含
  [n] 全部排除
  b  回上一層  h  回主選單  q  離開程式
  > 請選擇：
```

- 輸入帳號編號可以即時切換「已包含」↔「已排除」
- `a` 全部包含（清空排除清單）
- `n` 全部排除（極少用到，但可以臨時停用切換功能中的所有帳號）
- 每次修改都會立即儲存到 `accounts.csv` 並更新切換邏輯，不需重啟工具

> **排除的帳號**：即使目前焦點在被排除的帳號上，按切換鍵仍然會跳到下一個「已包含」帳號；如果所有帳號都被排除，切換鍵不會有任何動作。

### 設定存放位置

帳號的切換設定儲存在 `accounts.csv` 的 `ToolFlags` 欄位（第 5 欄），與 D2R 啟動參數的 `LaunchFlags`（第 4 欄）、畫質設定檔的 `GraphicsProfile`（第 6 欄）、預設登入區域的 `DefaultRegion`（第 7 欄）及預設 mod 的 `DefaultMod`（第 8 欄）各自獨立：

```csv
Email,Password,DisplayName,LaunchFlags,ToolFlags,GraphicsProfile,DefaultRegion,DefaultMod
acc@gmail.com,ENC:xxx,主帳號,0,0,,NA,<vanilla>
acc2@gmail.com,ENC:yyy,倉庫號,0,1,boss-low,EU,sample-mod
```

`ToolFlags = 1` 表示「跳過切換循環」。

## 設定切換按鍵

在切換設定畫面輸入 `1` 後，工具會開始等待你按下想使用的觸發方式。

```text
  請按下想用來切換視窗的按鍵組合...
  （支援：鍵盤任意鍵 + Ctrl/Alt/Shift、滑鼠側鍵、搖桿按鈕）
  （搖桿組合鍵：先按住修飾按鈕，再按觸發按鈕，放開後完成偵測）
  （按 Esc 取消）
```

偵測成功後，工具會顯示結果並詢問你是否要套用：

```text
  偵測到：Ctrl+Tab
  確認使用此組合？(Y/n)：
```

按 `Y` 或直接 Enter 就會儲存；輸入 `n` 則取消這次設定。

## 支援哪些觸發方式

| 類型 | 範例 | 說明 |
|------|------|------|
| 鍵盤快捷鍵 | `Ctrl+Tab`、`Alt+F1` | 任意鍵加上 `Ctrl` / `Alt` / `Shift` |
| 滑鼠側鍵 | `XButton1`、`XButton2` | 常見的前進鍵 / 後退鍵 |
| 搖桿單鍵 | `Gamepad_A`、`Gamepad_LB` | XInput 搖桿上的單一按鈕 |
| 搖桿組合鍵 | `LT+A`、`Back+RB` | 先按住修飾鍵，再按主按鍵 |

## 搖桿組合鍵怎麼按

如果你要設定像 `LT+A` 這類組合，建議照這個順序：

1. 先按住修飾鍵，例如 `LT`
2. 再按主按鍵，例如 `A`
3. 放開按鍵，讓偵測完成

如果你中途不想設定，可以直接按 `Esc` 取消。

## 切換開關狀態

`[0]` 現在是「切換狀態」：

- 如果目前是關閉，且你之前已經設定過切換按鍵，輸入 `0` 就會直接開啟
- 如果目前是開啟，輸入 `0` 就會關閉
- 關閉時只會把 `enabled` 切成 `false`，不會洗掉原本保存的按鍵設定

例如暫時不想用 switcher：

1. 回到主選單輸入 `s`
2. 輸入 `0`
3. 工具會停止切換功能，並把設定存回 `config.json`

畫面會顯示類似：

```text
  ✔ 已關閉切換功能，原設定會保留。
```

若目前還沒有設定過切換按鍵，`[0]` 不會自動幫你開啟；請先使用 `[1] 設定切換按鍵`。

## 設定會存到哪裡

切換設定會寫進：

```text
%USERPROFILE%\.d2r-hyper-launcher\config.json
```

範例：

```json
{
  "d2r_path": "C:\\Program Files (x86)\\Diablo II Resurrected\\D2R.exe",
  "launch_delay": 30,
  "switcher": {
    "enabled": true,
    "modifiers": ["ctrl"],
    "key": "Tab"
  }
}
```

如果是搖桿，還會多一個 `gamepad_index`，用來記錄第幾支控制器。

## 使用時的實際效果

切換器不是跳到指定視窗，而是每按一次就切到「下一個」D2R 視窗。

因此比較適合：

- 固定輪流看兩到三個角色
- 打 BO、開倉庫、開車隊時快速切換
- 用同一組快捷鍵循環多個視窗

## 常見問題

### 按了切換鍵沒反應

先確認這幾件事：

1. 是否真的有 2 個以上的 D2R 視窗
2. 視窗標題是否已經被 launcher 改成 `D2R-<DisplayName>`
3. `config.json` 裡的 `switcher.enabled` 是否為 `true`

### 滑鼠側鍵沒有被偵測到

請確認你的滑鼠側鍵有被 Windows 當成 `XButton1` 或 `XButton2`，而不是被滑鼠廠商驅動改成其他功能。

### 搖桿偵測不到

本工具使用 XInput，請確認：

- 控制器有被 Windows 正常辨識
- 不是只支援 DirectInput 的舊裝置
- 第一次設定時盡量用管理員權限啟動工具

### 想改按鍵

直接回到 `s` 選單，再重新做一次 `設定切換按鍵` 即可，新的設定會覆蓋舊設定。

## 延伸閱讀

- [README.md](../README.md) — 新手快速開始
- [multiboxing-usage-guide.md](multiboxing-usage-guide.md) — 多開啟動操作
- [switcher-technical-guide.md](switcher-technical-guide.md) — switcher 技術原理

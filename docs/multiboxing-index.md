# Multiboxing 文件總覽

> 這份文件是多開相關 Markdown 的導覽頁。  
> 如果你不知道該先讀哪一份，先從這裡開始。

## 多開文件怎麼分工

| 文件 | 適合誰看 | 主要內容 |
|---|---|---|
| [README.md](../README.md) | 第一次接觸工具的玩家 | 快速開始、`accounts.csv` 基本格式、主選單入口 |
| [multiboxing-usage-guide.md](multiboxing-usage-guide.md) | 想照步驟操作的玩家 | 完整操作流程、`f` 選單、單帳號 / 批次 / 離線模式、常見問題 |
| [D2R_PARAMS.md](D2R_PARAMS.md) | 想查啟動參數用途的人 | `LaunchFlags` 對應參數、Battle.net / mod 相關旗標、參數參考與來源 |
| [multiboxing-technical-guide.md](multiboxing-technical-guide.md) | 想看底層原理的開發者 | 單實例鎖、handle 關閉、背景 monitor、視窗重命名、模組分工 |
| [releases/v1.1.0.md](releases/v1.1.0.md) | 想看目前正式版整理內容的人 | `v1.1.0` 的多開相關變更摘要 |

## 建議閱讀順序

### 如果你是玩家

1. 先看 [README.md](../README.md)
2. 再看 [multiboxing-usage-guide.md](multiboxing-usage-guide.md)
3. 要查旗標或 mod 參數時，再看 [D2R_PARAMS.md](D2R_PARAMS.md)

### 如果你是開發者

1. 先看 [README.md](../README.md) 了解玩家會怎麼用
2. 再看 [multiboxing-usage-guide.md](multiboxing-usage-guide.md) 對齊實際操作流程
3. 接著看 [multiboxing-technical-guide.md](multiboxing-technical-guide.md) 理解實作架構
4. 最後用 [D2R_PARAMS.md](D2R_PARAMS.md) 對照啟動參數與目前 exposed 的多開旗標

## 哪些內容算是多開 scope

目前這個專案的 multiboxing 文件，主要涵蓋這幾類行為：

- `accounts.csv` / `config.json` 的準備與讀寫
- Battle.net 帳號直接啟動與區域選擇
- 單帳號、批次、離線模式的啟動流程
- `LaunchFlags` 的 per-account 設定
- mod 掃描與 `-mod <name> -txt` 串接
- 單實例 Event Handle 關閉
- `D2R-<DisplayName>` 視窗命名與已啟動狀態辨識
- 背景 monitor 持續處理新 D2R PID

## 常見查找路徑

| 你要找的內容 | 建議直接看 |
|---|---|
| `accounts.csv` 怎麼填 | [README.md](../README.md) / [multiboxing-usage-guide.md](multiboxing-usage-guide.md) |
| `LaunchFlags` 怎麼設定 | [multiboxing-usage-guide.md](multiboxing-usage-guide.md) |
| `LaunchFlags` 每個 bitflag 對應哪些參數 | [D2R_PARAMS.md](D2R_PARAMS.md) |
| 為什麼能多開 | [multiboxing-technical-guide.md](multiboxing-technical-guide.md) |
| 目前正式版本整理了哪些多開能力 | [releases/v1.1.0.md](releases/v1.1.0.md) |

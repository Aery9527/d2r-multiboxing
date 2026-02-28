# my-runewords — D2R 自訂符文之語 Mod

⚠️ **僅限離線/單人模式使用 (OFFLINE/SINGLE-PLAYER ONLY)**

## 說明

本 mod 新增一個自訂符文之語：

### 龍息 (Dragon Breath)

| 項目 | 內容 |
|------|------|
| 符文組合 | **Ber (30) + Jah (31) + Cham (32)** |
| 可裝備類型 | 所有武器 (`weap`) |
| 效果 | +3 所有技能 (`allskills`) |
| | +40% 攻擊速度 (`swing2`) |
| | +15% 生命偷取 (`lifesteal`) |
| | +300% 增強傷害 (`dmg%`) |

## 修改的檔案

| 檔案 | 說明 |
|------|------|
| [`my-runewords.mpq/data/global/excel/runes.txt`](my-runewords.mpq/data/global/excel/runes.txt) | 新增符文之語定義（一筆新 row） |
| [`my-runewords.mpq/data/local/lng/strings/item-runes.json`](my-runewords.mpq/data/local/lng/strings/item-runes.json) | 新增符文之語名稱字串（英文 + 繁中） |
| [`modinfo.json`](modinfo.json) | Mod 設定（名稱、獨立存檔路徑） |

## 安裝方式

1. 將整個 `my-runewords/` 資料夾複製到 D2R 安裝路徑的 `mods/` 目錄下
   ```
   <D2R 安裝路徑>/mods/my-runewords/
   ```
2. 啟動遊戲時加上參數：
   ```
   -mod my-runewords -txt
   ```
3. 首次成功載入後，後續啟動可移除 `-txt` 以加快載入速度

## ⚠️ 重要警告

- 🚫 **僅限離線/單人模式使用 (OFFLINE/SINGLE-PLAYER ONLY)**
- 🚫 **在 Battle.net 使用這些 mod 會導致帳號永久封禁**
- ⚠️ 測試前請先備份存檔
- ⚠️ 本 mod 使用獨立存檔路徑（`modinfo.json` 中 `savepath: "my-runewords/"`），不會影響原版存檔

# my-runewords

自訂 Diablo II: Resurrected 符文之語 mod。

## 符文之語列表

### 龍息 (Dragon Breath)

| 項目       | 內容                    |
| ---------- | ----------------------- |
| 符文       | Ber + Jah + Cham        |
| 孔數       | 3                       |
| 適用裝備   | 所有武器 (weap)         |

**屬性效果：**

- +3 到所有技能 (`allskills`)
- 40% 增加攻擊速度 (`swing2`)
- 15% 生命偷取 (`lifesteal`)
- +300% 增強傷害 (`dmg%`)

## 安裝方式

1. 將 `my-runewords` 資料夾複製到 D2R 的 `mods` 目錄下
2. 使用啟動參數載入 mod：
   ```
   D2R.exe -mod my-runewords -txt
   ```

## 檔案結構

```
mods/my-runewords/
├── modinfo.json                    # Mod 基本資訊
├── item-runes.json                 # 符文之語 JSON 定義（參考用）
├── README.md                       # 說明文件
└── data/global/excel/
    └── runes.txt                   # D2R 符文之語資料表
```

## 符文編號對照

| 符文 | 編號 |
| ---- | ---- |
| Ber  | r30  |
| Jah  | r31  |
| Cham | r32  |

## 屬性代碼說明

| 代碼       | 說明           |
| ---------- | -------------- |
| allskills  | +X 到所有技能  |
| swing2     | 增加攻擊速度 % |
| lifesteal  | 生命偷取 %     |
| dmg%       | 增強傷害 %     |

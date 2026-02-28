# D2R Online Modding Reference

Display-only modding reference for D2R JSON string files.

## Table of Contents
1. [Color Codes](#color-codes)
2. [Rune Codes & Names](#rune-codes--names)
3. [Common Item String Keys](#common-item-string-keys)
4. [FCR Breakpoints](#fcr-breakpoints-all-classes)
5. [FHR Breakpoints](#fhr-breakpoints-all-classes)
6. [FBR Breakpoints](#fbr-breakpoints-all-classes)
7. [Cube Recipes Quick Reference](#cube-recipes-quick-reference)
8. [JSON String File Locations](#json-string-file-locations)

---

## Color Codes

Use `ÿcX` prefix in string values. `ÿ` = character 255 (Alt+0255).

| Code | Color | Typical Use |
|------|-------|-------------|
| `ÿc0` | White | Normal text, reset color |
| `ÿc1` | Red | Unique/critical items, warnings |
| `ÿc2` | Green | Set items |
| `ÿc3` | Blue | Magic items |
| `ÿc4` | Gold | Unique items, highlights |
| `ÿc5` | Gray | Low-value items, dimmed text |
| `ÿc6` | Black | Hidden/background |
| `ÿc7` | Tan/Light Gold | Set item names |
| `ÿc8` | Orange | Crafted items, mid-tier highlights |
| `ÿc9` | Yellow | Rare items, emphasis |
| `ÿc;` | Purple | Special highlights |
| `ÿcN` | Teal | Info text |
| `ÿcO` | Pink | Feminine highlights |
| `ÿcP` | Light Blue | Cold damage, mana |
| `ÿcQ` | Light Green | Poison, nature |

### Usage Pattern
```json
{
  "id": 11176,
  "Key": "r30",
  "enUS": "ÿc1★★★ Ber Rune #30 ★★★",
  "zhTW": "ÿc1★★★ 柏 符文 #30 ★★★"
}
```

Multiple colors in one string:
```
"ÿc4=== Info ===ÿc0\nÿc9Value:ÿc0 100"
```
Use `\n` for newlines in tooltip text.

---

## Rune Codes & Names

| # | Code | Name (EN) | Name (ZH) | Key |
|---|------|-----------|-----------|-----|
| 1 | r01 | El | 艾爾 | r01 |
| 2 | r02 | Eld | 艾德 | r02 |
| 3 | r03 | Tir | 提爾 | r03 |
| 4 | r04 | Nef | 乃夫 | r04 |
| 5 | r05 | Eth | 乙太 | r05 |
| 6 | r06 | Ith | 伊斯 | r06 |
| 7 | r07 | Tal | 塔爾 | r07 |
| 8 | r08 | Ral | 拉爾 | r08 |
| 9 | r09 | Ort | 乙特 | r09 |
| 10 | r10 | Thul | 乙爾 | r10 |
| 11 | r11 | Amn | 安姆 | r11 |
| 12 | r12 | Sol | 索爾 | r12 |
| 13 | r13 | Shael | 乙爾 | r13 |
| 14 | r14 | Dol | 杜爾 | r14 |
| 15 | r15 | Hel | 海爾 | r15 |
| 16 | r16 | Io | 艾歐 | r16 |
| 17 | r17 | Lum | 盧姆 | r17 |
| 18 | r18 | Ko | 乙歐 | r18 |
| 19 | r19 | Fal | 法爾 | r19 |
| 20 | r20 | Lem | 雷姆 | r20 |
| 21 | r21 | Pul | 普爾 | r21 |
| 22 | r22 | Um | 乙姆 | r22 |
| 23 | r23 | Mal | 乙爾 | r23 |
| 24 | r24 | Ist | 伊斯特 | r24 |
| 25 | r25 | Gul | 古爾 | r25 |
| 26 | r26 | Vex | 乙克斯 | r26 |
| 27 | r27 | Ohm | 歐姆 | r27 |
| 28 | r28 | Lo | 洛 | r28 |
| 29 | r29 | Sur | 瑟 | r29 |
| 30 | r30 | Ber | 柏 | r30 |
| 31 | r31 | Jah | 賈 | r31 |
| 32 | r32 | Cham | 乙安 | r32 |
| 33 | r33 | Zod | 乙得 | r33 |

### Value tiers for loot filter
- **S-Tier** (★★★): Ber(r30), Jah(r31), Cham(r32), Zod(r33)
- **A-Tier** (★★): Vex(r26), Ohm(r27), Lo(r28), Sur(r29)
- **B-Tier** (★): Ist(r24), Gul(r25), Mal(r23), Um(r22), Pul(r21)
- **C-Tier**: Lem(r20) and below

---

## Common Item String Keys

### Potions
| Key | Item (EN) | Item (ZH) |
|-----|-----------|-----------|
| `hp1` | Minor Healing Potion | 小型治療藥水 |
| `hp2` | Light Healing Potion | 輕型治療藥水 |
| `hp3` | Healing Potion | 治療藥水 |
| `hp4` | Greater Healing Potion | 強效治療藥水 |
| `hp5` | Super Healing Potion | 超級治療藥水 |
| `mp1` | Minor Mana Potion | 小型魔法藥水 |
| `mp2` | Light Mana Potion | 輕型魔法藥水 |
| `mp3` | Mana Potion | 魔法藥水 |
| `mp4` | Greater Mana Potion | 強效魔法藥水 |
| `mp5` | Super Mana Potion | 超級魔法藥水 |
| `rvs` | Rejuvenation Potion | 回復藥水 |
| `rvl` | Full Rejuvenation Potion | 完全回復藥水 |

### Scrolls & Keys
| Key | Item (EN) |
|-----|-----------|
| `tsc` | Scroll of Town Portal |
| `isc` | Scroll of Identify |
| `key` | Key |

### Gems (Perfect)
| Key | Gem (EN) |
|-----|----------|
| `gpv` | Perfect Amethyst |
| `gpy` | Perfect Topaz |
| `gpb` | Perfect Sapphire |
| `gpg` | Perfect Emerald |
| `gpr` | Perfect Ruby |
| `gpw` | Perfect Diamond |
| `skc` | Perfect Skull |

### Uber Materials
| Key | Item (EN) | Item (ZH) |
|-----|-----------|-----------|
| `pk1` | Key of Terror | 恐懼之鑰 |
| `pk2` | Key of Hate | 憎恨之鑰 |
| `pk3` | Key of Destruction | 毀滅之鑰 |
| `bet` | Mephisto's Brain | 乙菲斯托之腦 |
| `mbr` | Diablo's Horn | 暗黑破壞神之角 |
| `dhn` | Baal's Eye | 巴爾之眼 |

### Essences
| Key | Item (EN) |
|-----|-----------|
| `tes` | Twisted Essence of Suffering |
| `ceh` | Charged Essense of Hatred |
| `bet` | Burning Essence of Terror |
| `fed` | Festering Essence of Destruction |

### Special Items
| Key | Item (EN) | Item (ZH) |
|-----|-----------|-----------|
| `cm1` | Small Charm | 小型護身符 |
| `cm2` | Large Charm | 大型護身符 |
| `cm3` | Grand Charm | 超大型護身符 |
| `jew` | Jewel | 珠寶 |
| `rin` | Ring | 戒指 |
| `amu` | Amulet | 護身符 |

---

## FCR Breakpoints (All Classes)

### Amazon
| FCR% | Frames |
|------|--------|
| 0 | 19 |
| 7 | 18 |
| 14 | 17 |
| 22 | 16 |
| 32 | 15 |
| 48 | 14 |
| 68 | 13 |
| 99 | 12 |
| 152 | 11 |

### Assassin
| FCR% | Frames |
|------|--------|
| 0 | 16 |
| 8 | 15 |
| 16 | 14 |
| 27 | 13 |
| 42 | 12 |
| 65 | 11 |
| 102 | 10 |
| 174 | 9 |

### Barbarian
| FCR% | Frames |
|------|--------|
| 0 | 13 |
| 9 | 12 |
| 20 | 11 |
| 37 | 10 |
| 63 | 9 |
| 105 | 8 |
| 200 | 7 |

### Druid (Human Form)
| FCR% | Frames |
|------|--------|
| 0 | 18 |
| 4 | 17 |
| 10 | 16 |
| 19 | 15 |
| 30 | 14 |
| 46 | 13 |
| 68 | 12 |
| 99 | 11 |
| 163 | 10 |

### Necromancer (Human Form)
| FCR% | Frames |
|------|--------|
| 0 | 15 |
| 9 | 14 |
| 18 | 13 |
| 30 | 12 |
| 48 | 11 |
| 75 | 10 |
| 125 | 9 |

### Paladin
| FCR% | Frames |
|------|--------|
| 0 | 15 |
| 9 | 14 |
| 18 | 13 |
| 30 | 12 |
| 48 | 11 |
| 75 | 10 |
| 125 | 9 |

### Sorceress
| FCR% | Frames |
|------|--------|
| 0 | 13 |
| 9 | 12 |
| 20 | 11 |
| 37 | 10 |
| 63 | 9 |
| 105 | 8 |
| 200 | 7 |

#### Sorceress Lightning/Chain Lightning
| FCR% | Frames |
|------|--------|
| 0 | 19 |
| 7 | 18 |
| 15 | 17 |
| 23 | 16 |
| 35 | 15 |
| 52 | 14 |
| 78 | 13 |
| 117 | 12 |
| 194 | 11 |

---

## FHR Breakpoints (All Classes)

### Amazon
| FHR% | Frames |
|------|--------|
| 0 | 11 |
| 6 | 10 |
| 13 | 9 |
| 20 | 8 |
| 32 | 7 |
| 52 | 6 |
| 86 | 5 |
| 174 | 4 |
| 600 | 3 |

### Assassin
| FHR% | Frames |
|------|--------|
| 0 | 9 |
| 7 | 8 |
| 15 | 7 |
| 27 | 6 |
| 48 | 5 |
| 86 | 4 |
| 200 | 3 |

### Barbarian
| FHR% | Frames |
|------|--------|
| 0 | 9 |
| 7 | 8 |
| 15 | 7 |
| 27 | 6 |
| 48 | 5 |
| 86 | 4 |
| 200 | 3 |

### Druid (Human Form)
| FHR% | Frames |
|------|--------|
| 0 | 14 |
| 3 | 13 |
| 7 | 12 |
| 13 | 11 |
| 19 | 10 |
| 29 | 9 |
| 42 | 8 |
| 63 | 7 |
| 99 | 6 |
| 174 | 5 |
| 456 | 4 |

### Necromancer (Human Form)
| FHR% | Frames |
|------|--------|
| 0 | 13 |
| 5 | 12 |
| 10 | 11 |
| 16 | 10 |
| 26 | 9 |
| 39 | 8 |
| 56 | 7 |
| 86 | 6 |
| 152 | 5 |
| 377 | 4 |

### Paladin
| FHR% | Frames |
|------|--------|
| 0 | 9 |
| 7 | 8 |
| 15 | 7 |
| 27 | 6 |
| 48 | 5 |
| 86 | 4 |
| 200 | 3 |

### Sorceress
| FHR% | Frames |
|------|--------|
| 0 | 15 |
| 5 | 14 |
| 9 | 13 |
| 14 | 12 |
| 20 | 11 |
| 30 | 10 |
| 42 | 9 |
| 60 | 8 |
| 86 | 7 |
| 142 | 6 |
| 280 | 5 |

---

## FBR Breakpoints (All Classes)

### Amazon (Shield + One-Hand)
| FBR% | Frames |
|------|--------|
| 0 | 5 |
| 13 | 4 |
| 32 | 3 |
| 86 | 2 |
| 600 | 1 |

### Assassin
| FBR% | Frames |
|------|--------|
| 0 | 5 |
| 13 | 4 |
| 32 | 3 |
| 86 | 2 |
| 600 | 1 |

### Barbarian
| FBR% | Frames |
|------|--------|
| 0 | 7 |
| 9 | 6 |
| 20 | 5 |
| 42 | 4 |
| 86 | 3 |
| 280 | 2 |

### Druid (Human Form)
| FBR% | Frames |
|------|--------|
| 0 | 11 |
| 6 | 10 |
| 13 | 9 |
| 20 | 8 |
| 32 | 7 |
| 52 | 6 |
| 86 | 5 |
| 174 | 4 |
| 600 | 3 |

### Necromancer
| FBR% | Frames |
|------|--------|
| 0 | 11 |
| 6 | 10 |
| 13 | 9 |
| 20 | 8 |
| 32 | 7 |
| 52 | 6 |
| 86 | 5 |
| 174 | 4 |
| 600 | 3 |

### Paladin (Normal)
| FBR% | Frames |
|------|--------|
| 0 | 5 |
| 13 | 4 |
| 32 | 3 |
| 86 | 2 |
| 600 | 1 |

### Paladin (Holy Shield)
| FBR% | Frames |
|------|--------|
| 0 | 2 |
| 86 | 1 |

### Sorceress
| FBR% | Frames |
|------|--------|
| 0 | 9 |
| 7 | 8 |
| 15 | 7 |
| 27 | 6 |
| 48 | 5 |
| 86 | 4 |
| 200 | 3 |

---

## Cube Recipes Quick Reference

### Rune Upgrades
| Input | Output | Note |
|-------|--------|------|
| 3× El (r01) | Eld (r02) | |
| 3× Eld (r02) | Tir (r03) | |
| ... | ... | 3 runes = next tier (up to Thul) |
| 3× Thul (r10) + Chipped Topaz | Amn (r11) | Gem required from Amn+ |
| 3× Amn (r11) + Chipped Amethyst | Sol (r12) | |
| 2× Pul (r21) + Flawless Diamond | Um (r22) | 2 runes from Pul+ |
| 2× Um (r22) + Flawless Topaz | Mal (r23) | |
| 2× Mal (r23) + Flawless Amethyst | Ist (r24) | |
| 2× Ist (r24) + Flawless Ruby | Gul (r25) | |
| 2× Gul (r25) + Flawless Emerald | Vex (r26) | |
| 2× Vex (r26) + Flawless Diamond | Ohm (r27) | |
| 2× Ohm (r27) + Flawless Sapphire | Lo (r28) | |
| 2× Lo (r28) + Flawless Topaz | Sur (r29) | |
| 2× Sur (r29) + Flawless Amethyst | Ber (r30) | |
| 2× Ber (r30) + Flawless Ruby | Jah (r31) | |
| 2× Jah (r31) + Flawless Emerald | Cham (r32) | |
| 2× Cham (r32) + Flawless Diamond | Zod (r33) | |

### Common Recipes
| Recipe | Result |
|--------|--------|
| 3× Perfect Gems + Magic Item | Reroll magic item |
| 6× Perfect Skulls + Rare Item | Reroll rare item |
| Perfect Skull + Rare Item + SoJ | Add socket to rare |
| Tal + Thul + Perfect Topaz + Normal Armor | Socket armor |
| Ral + Amn + Perfect Amethyst + Normal Weapon | Socket weapon |
| Ral + Sol + Perfect Emerald + Normal Armor | Upgrade Normal→Exceptional |
| Lum + Pul + Perfect Emerald + Exceptional Armor | Upgrade Exceptional→Elite |
| Ral + Sol + Perfect Emerald + Normal Weapon | Upgrade Normal→Exceptional weapon |

### Token Recipe
| Input | Output |
|-------|--------|
| Twisted Essence + Charged Essence + Burning Essence + Festering Essence | Token of Absolution |

---

## JSON String File Locations

All files go in `data/local/lng/strings/`:

| File | Content | Override Use |
|------|---------|-------------|
| `item-names.json` | Item display names | Loot filter, name shortener |
| `item-runes.json` | Runeword names | Runeword display |
| `item-modifiers.json` | Affix/modifier descriptions | Info tooltips |
| `item-nameaffixes.json` | Magic prefix/suffix names | Display cleanup |
| `skills.json` | Skill names & descriptions | Breakpoint info embedding |
| `monsters.json` | Monster names | Monster display |
| `ui.json` | UI text elements | General UI mods |
| `levels.json` | Area/level names | Zone naming |
| `mercenaries.json` | Mercenary names/descriptions | Merc display |

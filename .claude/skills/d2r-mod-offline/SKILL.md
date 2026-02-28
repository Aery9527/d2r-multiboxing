---
name: d2r-mod-offline
description: "Create Diablo II: Resurrected (D2R) offline/single-player god-mode mods that modify game data tables (.txt files). Change drop rates, create custom runewords, add cube recipes, tweak monster stats, modify skills, adjust character attributes — anything that alters game mechanics. ⚠️ These mods are for OFFLINE/SINGLE-PLAYER ONLY and will get you banned on Battle.net. Use this skill whenever the user wants to modify D2R game data, change drop rates, create custom items or runewords, add cube recipes, tweak monster difficulty, modify skills, adjust character stats, or any D2R modding task that changes game balance — even if they just say 'change D2R drops' or 'make D2R easier'."
---

# D2R Offline God-Mode Mod Creator

You are a D2R modding expert. Your job is to guide users through creating **offline/single-player** mods that modify game data — changing drop rates, items, recipes, monsters, skills, and more.

⚠️ **CRITICAL WARNING**: These mods modify `.txt` data tables that change game mechanics. They are **strictly for offline/single-player use only**. Using them on Battle.net **will result in account bans**.

D2R mods work by overriding the game's data files (`.txt` tables and `.json` string files) placed in a specific directory structure. The game loads these overrides via the `-mod` launch parameter. This is "softcode" modding — you edit configuration, not game code.

## Core Workflow

Follow this sequence for every mod creation request:

### Phase 1: Interview — Understand what the user wants

Ask the user these questions **one at a time** (skip any that are already answered from context):

1. **Mod 名稱 (Mod Name)** — Will be used for the folder name and `modinfo.json`. Suggest a short, lowercase, no-spaces name if the user gives a long one.

2. **Mod 類型 (Mod Type)** — What does the mod change? Common categories:
   - 掉落率修改 (Drop rate changes)
   - 自訂物品 / Runeword (Custom items / runewords)
   - 怪物修改 (Monster tweaks — HP, damage, resistances)
   - 技能修改 (Skill changes — damage, mana cost, synergies)
   - Horadric Cube 配方 (Cube recipes)
   - 角色修改 (Character stats, starting gear)
   - 經驗值 / 難度 (XP rates, difficulty scaling)
   - 綜合修改 (Multiple changes)

3. **具體需求 (Specific details)** — Drill down based on the mod type. Examples:
   - For drop rate: "Which monsters? How much increase? High runes specifically?"
   - For custom items: "Weapon or armor? What stats? What level requirement?"
   - For runewords: "Which rune combination? What item types? What properties?"

4. **輸出目錄 (Output directory)** — Where to generate the mod files. Default: `./mods/<mod-name>/` under the current working directory. Confirm with the user.

5. **存檔隔離 (Save isolation)** — Should mod saves be separate from vanilla saves? Default: yes (separate).

### Phase 2: Generate the mod files

After confirming all details, generate the complete mod structure:

```
<output-dir>/
└── <mod-name>/
    ├── modinfo.json
    └── <mod-name>.mpq/
        └── data/
            ├── global/
            │   └── excel/          ← modified .txt data tables
            │       └── ...
            └── local/
                └── lng/
                    └── strings/    ← modified .json string files (if needed)
                        └── ...
```

#### Always generate:

1. **`modinfo.json`**:
```json
{
  "name": "<mod-name>",
  "savepath": "<mod-name>/"
}
```
If the user chose shared saves, use `"savepath": "../"` instead.

2. **A `README.md`** inside the mod folder with:
   - What the mod does
   - How to install (copy to D2R `mods/` folder)
   - Launch command: `-mod <mod-name> -txt`
   - What files were modified and why

#### Generate based on mod type:

Read the appropriate section from `references/data-tables-reference.md` to understand which `.txt` files to modify and the exact column format.

**Critical rules for .txt files:**
- Files are **Tab-separated (TSV)**, not CSV. Use `\t` between columns.
- The first row is always the **header row** with column names.
- Only include columns that are being changed — but always include the key/identifier columns (like `name`, `code`, etc.) so the game can match rows.
- Leave a comment in the README about which columns were changed and the original vs. new values.
- If adding new rows (e.g., new items), ensure `code` values are unique and don't conflict with existing ones.

**Critical rules for .json string files:**
- Each entry needs `id` (unique integer), `Key` (string key referenced by .txt files), and language fields (`enUS`, `zhTW`, etc.).
- Use high `id` numbers (e.g., 90000+) for custom entries to avoid conflicts.

### Phase 3: Explain and verify

After generating all files:

1. **Summarize** what was created — list every file and briefly explain its purpose.
2. **Provide installation instructions**:
   - Copy the `<mod-name>/` folder to `<D2R install path>/mods/`
   - Launch with `-mod <mod-name> -txt` parameter
   - After first successful launch, can remove `-txt` for faster loading
3. **⚠️ MANDATORY WARNINGS** (always display prominently):
   - 🚫 **僅限離線/單人模式使用 (OFFLINE/SINGLE-PLAYER ONLY)**
   - 🚫 **在 Battle.net 使用這些 mod 會導致帳號永久封禁**
   - ⚠️ 測試前請先備份存檔
   - ⚠️ 建議使用獨立存檔路徑（modinfo.json 中的 savepath）

## Mod Type Specific Guidance

### Drop Rate Mods

The key file is `treasureclassex.txt`. The `NoDrop` column controls how likely "nothing drops":
- Lower `NoDrop` = more drops
- Setting `NoDrop` to `0` = something always drops
- The `Picks` column controls how many items roll per kill

When modifying drop rates, ask:
- Global change or specific monsters/areas?
- Approximate multiplier (2x, 5x, 10x)?
- Should high runes be affected separately?

### Custom Items

Requires modifying `weapons.txt`, `armor.txt`, or `misc.txt` (depending on item type) AND adding name strings to `item-names.json`.

When creating custom items:
- Generate a unique 3-character `code` (check reference to avoid conflicts)
- Set appropriate `level`, `levelreq`, `reqstr`, `reqdex`
- Add the item to a treasure class in `treasureclassex.txt` if it should drop

### Custom Runewords

Modify `runes.txt` and add name strings to `item-runes.json`.

Each runeword row needs:
- `Name` — string key for the runeword name
- `complete` — set to `1`
- `itype1`-`itype6` — which item types accept it (e.g., `weap`, `shie`)
- `Rune1`-`Rune6` — the runes required (e.g., `r01` for El)
- `T1Code1`-`T1Code7` — the properties granted
- `T1Param1`-`T1Param7` — property parameters
- `T1Min1`-`T1Min7`, `T1Max1`-`T1Max7` — min/max values

### Cube Recipes

Modify `cubemain.txt`. Each recipe needs:
- `description` — internal description
- `enabled` — `1` to enable
- `numinputs` — number of input items
- `input 1` through `input 7` — input items (use item codes, `qty=N` for multiples)
- `output` — output item code
- Optionally: `mod 1`-`mod 5` for output properties

### Monster Modifications

Modify `monstats.txt`. Key columns:
- `MinHP`, `MaxHP` — health range per difficulty (normal/nightmare/hell suffixes)
- `ResDm`, `ResMa`, `ResFi`, `ResLi`, `ResCo`, `ResPo` — resistances
- `Level` — monster level
- `AI` — AI behavior reference

### Skill Modifications

Modify `skills.txt`. Key columns:
- `mana`, `minmana` — mana cost
- `EMin`, `EMax` — elemental damage range
- `EDmgSymPerCalc` — synergy damage formula
- `reqlevel` — required character level

## Reference Files

For detailed column specifications of each `.txt` file, read `references/data-tables-reference.md`. It contains the complete column layout for all commonly modded files, accepted values, and cross-reference relationships.

## Language

Respond in the same language the user uses. Default to 繁體中文 (Traditional Chinese) if the user writes in Chinese. Use English for file contents (column names, codes, etc.) since D2R data files are in English.

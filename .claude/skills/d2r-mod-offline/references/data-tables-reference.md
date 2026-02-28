# D2R Data Tables Reference

Quick reference for the most commonly modded `.txt` data table columns and `.json` string file formats.

## .txt File Format Rules

- **Tab-separated (TSV)** — columns separated by `\t`
- First row = column headers
- Columns starting with `*` are comments (ignored by game)
- Empty cells are valid (game uses defaults)
- Files must end with a newline

---

## treasureclassex.txt — Drop Tables

Controls what items drop from monsters, chests, and other sources.

| Column | Type | Description |
|--------|------|-------------|
| `Treasure Class` | string | Unique identifier for this drop table |
| `Picks` | int | Number of items to roll (negative = pick exactly N unique) |
| `Unique` | int | Unique item chance modifier (higher = more likely) |
| `Set` | int | Set item chance modifier |
| `Rare` | int | Rare item chance modifier |
| `Magic` | int | Magic item chance modifier |
| `NoDrop` | int | Weight for "nothing drops" (lower = more drops, 0 = always drop) |
| `Item1`-`Item10` | string | Item code or treasure class name |
| `Prob1`-`Prob10` | int | Probability weight for corresponding item |

### Example: Reduce NoDrop for Act bosses
```
Treasure Class	Picks	NoDrop	Item1	Prob1	Item2	Prob2
Andariel (H)	5	5	gld,mul=2048	21	Act 1 Junk	21
```
Original `NoDrop` was ~100; setting to 5 dramatically increases drops.

---

## weapons.txt — Weapon Definitions

| Column | Type | Description |
|--------|------|-------------|
| `name` | string | Internal weapon name |
| `type` | string | Item type code (references `itemtypes.txt`) |
| `code` | string | **Unique 3-char identifier** (critical for cross-references) |
| `namestr` | string | String key for display name (references JSON strings) |
| `version` | int | 0=classic, 100=expansion |
| `level` | int | Item base level |
| `levelreq` | int | Required character level to equip |
| `reqstr` | int | Required strength |
| `reqdex` | int | Required dexterity |
| `speed` | int | Attack speed modifier |
| `mindam` | int | Minimum physical damage |
| `maxdam` | int | Maximum physical damage |
| `2handmindam` | int | Two-hand minimum damage |
| `2handmaxdam` | int | Two-hand maximum damage |
| `durability` | int | Maximum durability |
| `nodurability` | bool(0/1) | If 1, item has no durability |
| `gemsockets` | int | Maximum number of sockets |
| `cost` | int | Base gold cost |
| `normcode` | string | Normal version code |
| `ubercode` | string | Exceptional version code |
| `ultracode` | string | Elite version code |

---

## armor.txt — Armor Definitions

| Column | Type | Description |
|--------|------|-------------|
| `name` | string | Internal armor name |
| `type` | string | Item type code |
| `code` | string | **Unique 3-char code** |
| `namestr` | string | String key for display name |
| `version` | int | 0=classic, 100=expansion |
| `level` | int | Item base level |
| `levelreq` | int | Required character level |
| `reqstr` | int | Required strength |
| `minac` | int | Minimum defense |
| `maxac` | int | Maximum defense |
| `durability` | int | Maximum durability |
| `gemsockets` | int | Maximum sockets |
| `cost` | int | Base gold cost |
| `normcode` | string | Normal version code |
| `ubercode` | string | Exceptional version code |
| `ultracode` | string | Elite version code |

---

## misc.txt — Miscellaneous Items

Covers potions, gems, runes, keys, scrolls, charms, jewels, etc.

| Column | Type | Description |
|--------|------|-------------|
| `name` | string | Internal name |
| `code` | string | **Unique 3-char code** |
| `namestr` | string | String key for display name |
| `level` | int | Item level |
| `levelreq` | int | Required level |
| `type` | string | Item type |
| `cost` | int | Gold value |
| `stackable` | bool(0/1) | Can be stacked |
| `minstack` | int | Minimum stack size |
| `maxstack` | int | Maximum stack size |

### Rune codes: `r01` (El) through `r33` (Zod)

---

## runes.txt — Runeword Definitions

| Column | Type | Description |
|--------|------|-------------|
| `Name` | string | String key for runeword name |
| `Rune Name` | string | Display name (internal) |
| `complete` | bool(0/1) | 1 = enabled |
| `server` | bool(0/1) | 0 for client-side |
| `itype1`-`itype6` | string | Accepted item types (`weap`, `shie`, `helm`, `tors`, etc.) |
| `etype1`-`etype3` | string | Excluded item types |
| `Rune1`-`Rune6` | string | Rune codes in order (e.g., `r01` = El) |
| `T1Code1`-`T1Code7` | string | Property codes for the runeword |
| `T1Param1`-`T1Param7` | string | Property parameters |
| `T1Min1`-`T1Min7` | int | Property minimum values |
| `T1Max1`-`T1Max7` | int | Property maximum values |

### Common property codes
| Code | Property |
|------|----------|
| `ac` | Defense |
| `ac%` | Enhanced Defense % |
| `dmg%` | Enhanced Damage % |
| `str` | Strength |
| `dex` | Dexterity |
| `vit` | Vitality |
| `enr` | Energy |
| `hp` | Life |
| `mana` | Mana |
| `res-fire` | Fire Resistance |
| `res-ltng` | Lightning Resistance |
| `res-cold` | Cold Resistance |
| `res-pois` | Poison Resistance |
| `res-all` | All Resistances |
| `att` | Attack Rating |
| `att%` | Attack Rating % |
| `swing2` | Increased Attack Speed |
| `move2` | Faster Run/Walk |
| `cast2` | Faster Cast Rate |
| `block2` | Faster Block Rate |
| `hit-skill` | Chance to Cast on Strike |
| `gethit-skill` | Chance to Cast when Struck |
| `skill` | +X to Skill (param = skill ID) |
| `allskills` | +X to All Skills |
| `skilltab` | +X to Skill Tab (param = tab ID) |
| `mag%/lvl` | Magic Find % per Level |
| `gold%/lvl` | Gold Find % per Level |
| `regen-mana` | Mana Regeneration % |
| `abs-fire%` | Fire Absorb % |
| `abs-ltng%` | Lightning Absorb % |
| `abs-cold%` | Cold Absorb % |
| `crush` | Crushing Blow % |
| `deadly` | Deadly Strike % |
| `openwounds` | Open Wounds % |
| `howl` | Hit Causes Monster to Flee % |
| `stupidity` | Hit Blinds Target |
| `knock` | Knockback |
| `noheal` | Prevent Monster Heal |
| `half-freeze` | Half Freeze Duration |
| `nofreeze` | Cannot be Frozen |
| `pierce` | Pierce Attack % |
| `indestruct` | Indestructible |
| `sock` | Number of Sockets |
| `rep-quant` | Replenish Quantity |
| `ease` | Requirements % (negative = reduce) |

---

## cubemain.txt — Cube Recipes

| Column | Type | Description |
|--------|------|-------------|
| `description` | string | Internal description |
| `enabled` | bool(0/1) | 1 = active recipe |
| `ladder` | bool(0/1) | 1 = ladder-only |
| `numinputs` | int | Number of input items |
| `input 1`-`input 7` | string | Input items (code, `"code,qty=N"` for multiples) |
| `output` | string | Output item code (`"usetype"` = keep input base) |
| `lvl` | int | Output item level |
| `mod 1`-`mod 5` | string | Output property codes |
| `mod 1 param`-`mod 5 param` | string | Property parameters |
| `mod 1 min`-`mod 5 min` | int | Min values |
| `mod 1 max`-`mod 5 max` | int | Max values |

### Input format examples
- `r01` — one El rune
- `r01,qty=3` — three El runes
- `weap,uni` — any unique weapon
- `"any,mag"` — any magic item

---

## skills.txt — Skill Definitions

| Column | Type | Description |
|--------|------|-------------|
| `skill` | string | Skill internal name |
| `Id` | int | Unique skill ID |
| `charclass` | string | Class (ama/sor/nec/pal/bar/dru/ass) |
| `skilldesc` | string | References `skilldesc.txt` |
| `srvmissile` | string | Server missile |
| `cltmissile` | string | Client missile |
| `reqlevel` | int | Required level |
| `maxlvl` | int | Maximum skill level |
| `mana` | int | Mana cost at level 1 |
| `minmana` | int | Minimum mana cost |
| `manashift` | int | Mana cost precision (usually 8) |
| `EType` | string | Element type (fire/ltng/cold/pois/mag/phys) |
| `EMin` | int | Element min damage at level 1 |
| `EMax` | int | Element max damage at level 1 |
| `EMinLev1`-`EMinLev5` | int | Element min at level breakpoints |
| `EMaxLev1`-`EMaxLev5` | int | Element max at level breakpoints |
| `EDmgSymPerCalc` | string | Synergy damage formula |
| `passive` | bool(0/1) | Is passive skill |

---

## monstats.txt — Monster Definitions

| Column | Type | Description |
|--------|------|-------------|
| `Id` | string | Unique monster identifier |
| `BaseId` | string | Base monster type |
| `NameStr` | string | String key for display name |
| `Level` | int | Monster level (Normal) |
| `Level(N)` | int | Monster level (Nightmare) |
| `Level(H)` | int | Monster level (Hell) |
| `MinHP` | int | Min HP (Normal) |
| `MaxHP` | int | Max HP (Normal) |
| `MinHP(N)` | int | Min HP (Nightmare) |
| `MaxHP(N)` | int | Max HP (Nightmare) |
| `MinHP(H)` | int | Min HP (Hell) |
| `MaxHP(H)` | int | Max HP (Hell) |
| `AC` | int | Defense (Normal) |
| `AC(N)` | int | Defense (Nightmare) |
| `AC(H)` | int | Defense (Hell) |
| `ResDm` | int | Physical resistance % |
| `ResMa` | int | Magic resistance % |
| `ResFi` | int | Fire resistance % |
| `ResLi` | int | Lightning resistance % |
| `ResCo` | int | Cold resistance % |
| `ResPo` | int | Poison resistance % |
| `ResDm(N)` | int | Physical resistance % (Nightmare) |
| `ResFi(H)` | int | Fire resistance % (Hell) |
| `AI` | string | AI behavior |
| `TreasureClass1`-`TreasureClass4` | string | Drop table per difficulty |

---

## charstats.txt — Character Class Definitions

| Column | Type | Description |
|--------|------|-------------|
| `class` | string | Class name |
| `str` | int | Starting strength |
| `dex` | int | Starting dexterity |
| `int` | int | Starting energy |
| `vit` | int | Starting vitality |
| `hpadd` | int | Life per level |
| `ManaAdd` | int | Mana per level |
| `LifePerVit` | int | Life per vitality point |
| `ManaPerEng` | int | Mana per energy point |
| `StatPerLvl` | int | Stat points per level |
| `SkillsPerLvl` | int | Skill points per level |
| `StartSkill` | string | Starting skill |
| `item1`-`item10` | string | Starting items |

---

## JSON String Files

Located in `data/local/lng/strings/`.

### Format
```json
[
  {
    "id": 90001,
    "Key": "MyCustomItemName",
    "enUS": "Blade of the Phoenix",
    "zhTW": "鳳凰之刃",
    "deDE": "Klinge des Phönix",
    "esES": "Hoja del Fénix",
    "frFR": "Lame du Phénix",
    "itIT": "Lama della Fenice",
    "koKR": "불사조의 검",
    "plPL": "Ostrze Feniksa",
    "esMX": "Hoja del Fénix",
    "jaJP": "フェニックスの刃",
    "ptBR": "Lâmina da Fênix",
    "ruRU": "Клинок Феникса",
    "zhCN": "凤凰之刃"
  }
]
```

### Key String Files
| File | Content |
|------|---------|
| `item-names.json` | Item display names |
| `item-runes.json` | Runeword names |
| `item-modifiers.json` | Item affix/modifier descriptions |
| `skills.json` | Skill names and descriptions |
| `monsters.json` | Monster names |
| `ui.json` | UI text elements |

### Rules
- Use `id` values **90000+** for custom entries to avoid conflicts with base game
- `Key` must match the `namestr` value used in `.txt` files
- At minimum include `enUS`; add `zhTW` for Traditional Chinese support
- The JSON file is an array `[...]` of objects at the root level

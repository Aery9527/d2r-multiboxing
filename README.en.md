# d2r-hyper-launcher

> Language：[繁體中文](README.md) | **English**

A Windows CLI toolkit for D2R (Diablo II: Resurrected) players, currently providing two core features:

- **multiboxing**: multi-account launch, single-instance lock handling, window identification
- **switcher**: switch D2R windows via keyboard / mouse side-buttons / gamepad

**Latest release: [v1.2.0](docs/releases/v1.2.0.md)**

> If you enter an invalid format, out-of-range value, or unsupported option in any CLI menu, the tool displays an error message and waits for you to press a key before returning to the current flow. On terminals that support raw single-key input it shows "press any key to continue"; otherwise it falls back to "press Enter to continue".

## Multiboxing Documentation Index

| What you want to do | Where to look |
|---|---|
| Quick start, just get going | This README |
| Full player workflow and FAQ | [docs/multiboxing-usage-guide.md](docs/multiboxing-usage-guide.md) |
| Understand how the multiboxing docs are organized | [docs/multiboxing-index.md](docs/multiboxing-index.md) |
| D2R launch parameters, LaunchFlags, `-mod` / `-txt` details | [docs/D2R_PARAMS.md](docs/D2R_PARAMS.md) |
| Low-level multiboxing implementation and architecture | [docs/multiboxing-technical-guide.md](docs/multiboxing-technical-guide.md) |

## Quick Start for Players

This section is for players who just want to use the tool — no Go or coding knowledge required.

### 1. Download the launcher

- [d2r-hyper-launcher.exe](d2r-hyper-launcher.exe)

Place it anywhere and run it. All related data will be stored in `%USERPROFILE%\.d2r-hyper-launcher`.

### 2. Double-click `d2r-hyper-launcher.exe`

Just double-click after downloading and follow the on-screen instructions to complete first-run setup.

- On the first run, the tool automatically creates the `%USERPROFILE%\.d2r-hyper-launcher` folder along with `config.json` (settings) and `accounts.csv` (accounts)
- After those files are created, the tool guides you to exit and automatically opens that folder so you can edit `accounts.csv` directly
- If D2R is not at the default path, the tool will guide you to use `p` to select the correct `D2R.exe` path

### 3. Edit the account list

The tool always reads accounts from: `%USERPROFILE%\.d2r-hyper-launcher\accounts.csv`

(If you're not sure where `%USERPROFILE%` is, the tool displays the full data directory path every time it starts.)

It is strongly recommended to **open the file with Excel** before editing to avoid accidentally breaking the CSV format.  
Fill in your Battle.net accounts using this format:

```csv
Email,Password,DisplayName,LaunchFlags,ToolFlags,GraphicsProfile,DefaultRegion,DefaultMod
your-account1@example.com,your-password-here,Main-Sorc(stash/weapons/jewelry),,,boss-low,NA,<vanilla>
your-account2@example.com,your-password-here,Alt-Barb(junk/gems),,,,EU,sample-mod
```

Field descriptions:

- `Email`: Battle.net login email
- `Password`: Battle.net password
- `DisplayName`: the name shown in the tool and used for window switching; after launch the window title is set to `D2R-<DisplayName>`
- `LaunchFlags`: extra **D2R launch flags** per account; can be left blank (the tool defaults to `0`); configure later via `f` in the main menu. The tool now keeps only `-ns` (disable sound). If you want different graphics settings per account, use the main-menu `g` account-default-graphics flow instead. If an old CSV still contains `-lq`, the launcher automatically strips it and rewrites the sanitized value. See [docs/D2R_PARAMS.md](docs/D2R_PARAMS.md) for parameter details.
- `ToolFlags`: per-account **tool-internal settings** (bitmask); can be left blank (defaults to `0`). Currently supported: `1` = exclude this account from the switcher cycle. Configure via `s → [2]` in the main menu — no need to edit manually.
- `GraphicsProfile`: the **named graphics profile** to apply for this account; can be left blank. If it is blank, launching that account will **not touch** `%USERPROFILE%\Saved Games\Diablo II Resurrected\Settings.json` at all. The recommended workflow is to use `g` in the main menu to save the current game settings as a named profile and then assign it to accounts. After that, when you launch a single account or use `a` batch launch, pressing Enter in the graphics menu uses the saved default. If any targeted account still has no valid default graphics configured, the launcher blocks Enter and lists the missing accounts. If you want this launch to leave `Settings.json` untouched, choose `0` manually in the graphics menu.
- `DefaultRegion`: the default region server for this account; can be left blank. The recommended workflow is to assign it via `v` in the main menu. After that, when you launch a single account or use `a` batch launch, pressing Enter in the region menu uses the saved default. If any targeted account still has no default region, the launcher blocks Enter and lists the missing accounts; you can still choose `1` / `2` / `3` manually.
- `DefaultMod`: the default mod for this account; can be left blank. The recommended workflow is to assign it via `m` in the main menu. If you want Enter to default to vanilla, set it to `<vanilla>` (equivalent to choosing "No mod" in the CLI). After that, when you launch a single account or use `a` batch launch, pressing Enter in the mod menu uses the saved default. If any targeted account still has no valid default mod, the launcher blocks Enter and lists the missing accounts. If a saved `DefaultMod` points at a mod that no longer exists in the current D2R install, the launcher clears that account's `DefaultMod` first and then treats it as missing.

> ⚠️ If you change a `DisplayName` while D2R is still running, the `[Running]` / `[Stopped]` status shown after reopening the launcher may be temporarily inaccurate.  
> It is recommended to rename accounts only after all game windows are closed, or use `r` in the main menu to refresh status.

> Plain-text passwords are automatically rewritten to an `ENC:`-prefixed encrypted string after the first run.  
> Encryption uses Windows DPAPI; you do not need to encrypt manually. If you change computers or Windows user accounts, re-enter the plain-text password.
>
> If you get stuck on the Battle.net login screen after entering the game, the most likely cause is a wrong password in `accounts.csv` — please verify the password for that account first.
>
> ⚠️ If an account has not previously switched to another region through the normal Battle.net flow, direct email/password login to that region may show a locked message. Log into that region once through Battle.net normally first, then come back to this tool and it should work as expected.

### 4. Return to the tool and start playing

After editing `accounts.csv`, double-click `d2r-hyper-launcher.exe` again to start.

Once launched, you will see the following options:

- `<number>`: first open the region menu. The prompt shows the target account along with the default region / mod / graphics profile currently prepared for Enter-based launch. If the account already has `DefaultRegion`, you can press Enter to use it, or choose `1` / `2` / `3` to override this launch manually. Then the mod menu appears; if the account already has `DefaultMod`, you can press Enter to use it, or choose `0` / another mod number to override this launch manually. After that the graphics menu appears; if the account already has `GraphicsProfile`, you can press Enter to use it, choose `0` to leave `Settings.json` untouched for this launch, or choose another profile number to override this launch manually. The specified account is then launched; if it is already running, re-launch is blocked
- `a`: the tool pre-scans which accounts are already running; if there are pending accounts, it then asks for region, mod, and graphics. All three menus first list the pending accounts together with the default region / mod / graphics profile currently prepared for Enter-based launch. In the region menu you can press Enter so each pending account uses its own `DefaultRegion`, in the mod menu you can press Enter so each pending account uses its own `DefaultMod`, and in the graphics menu you can press Enter so each pending account uses its own `GraphicsProfile`; you can also choose `1` / `2` / `3`, `0` / another mod number, or `0` / another graphics-profile number once to override the whole batch launch
- `0`: select an installed mod (if any), then launch in offline mode
- `d`: set the launch delay used by `a` batch launch; enter `30` or a range like `30-60` (random wait within that interval each time); minimum is fixed at 10 seconds
- `f`: display the account list and a centered two-line flag reference table, then set or clear extra launch flags per account; currently only the "disable sound" flag remains. For per-account graphics differences, use `g` account default graphics instead. You can still configure flags per account, per flag, or all at once; see [docs/D2R_PARAMS.md](docs/D2R_PARAMS.md) for details
- `g`: manage account default graphics. First adjust graphics / window / resolution in-game, then return to the CLI and save the current `%USERPROFILE%\Saved Games\Diablo II Resurrected\Settings.json` as a named profile; after that you can assign it to accounts, clear account default graphics, or delete saved profiles you no longer need. Unassigned accounts leave `Settings.json` untouched at launch time. After that, when the launch graphics menu appears for a single account or for `a` batch launch, pressing Enter uses the saved defaults. If any targeted account still has no valid default graphics configured, the launcher blocks Enter and lists the missing accounts; you can also choose `0` or another profile number to override this launch manually
- `m`: manage account default mods. Using a flow similar to graphics profiles, you can assign one installed mod per account by mod or by account; if you want an account to default to vanilla on Enter, assign it to "No mod". After that, when the launch mod menu appears for a single account or for `a` batch launch, pressing Enter uses the saved defaults. If any targeted account still has no valid default mod, the launcher blocks Enter and lists the missing accounts; if a saved default mod no longer exists, the launcher clears that account's `DefaultMod` automatically first
- `v`: manage account default login regions. Using a flow similar to graphics profiles, you can assign `NA`, `EU`, or `Asia` per account by region or by account. After that, when the launch region menu appears for a single account or for `a` batch launch, pressing Enter uses the saved defaults. If any targeted account still has no default region, the launcher blocks Enter and lists the missing accounts instead of silently falling back
- `p`: open a Windows file picker to set the `D2R.exe` path
- `s`: configure window-switch hotkey / mouse side-button / gamepad button
- `l`: switch the tool interface language (繁體中文 / English); takes effect immediately and is saved to `config.json`
- `r`: reload `accounts.csv` and refresh status
- `q`: quit the tool

### 5. Want a more detailed walkthrough?

If you want to see how each menu works and what each step looks like, read:

- [docs/multiboxing-index.md](docs/multiboxing-index.md) — multiboxing doc overview and reading order
- [docs/multiboxing-usage-guide.md](docs/multiboxing-usage-guide.md) — batch launch, accounts file, region selection, offline mode
- [docs/switcher-usage-guide.md](docs/switcher-usage-guide.md) — window switch setup, supported input types, FAQ

For low-level implementation and technical details:

- [docs/multiboxing-technical-guide.md](docs/multiboxing-technical-guide.md)
- [docs/D2R_PARAMS.md](docs/D2R_PARAMS.md) — D2R launch parameters and current LaunchFlags / mod reference
- [docs/switcher-technical-guide.md](docs/switcher-technical-guide.md)

## Notes

- It is recommended to set D2R to **Windowed** or **Windowed (Borderless)** mode
- When configuring gamepad switch buttons, run the tool as administrator — testing showed that non-admin permissions may fail to detect gamepad signals
- `switcher` only works while `d2r-hyper-launcher` is running; closing the tool stops window switching
- The default `launch_delay` for `a` batch launch is 10 seconds; for backward compatibility, if the tool reads an old default value of `5` seconds, it automatically treats it as 10. Battle.net may still throttle logins if accounts are launched too rapidly, so if you adjust the delay, use `d` in the main menu and note the minimum is fixed at 10 seconds
- If you want different graphics settings per account, first adjust the settings in-game, then return to the main-menu `g` account-default-graphics flow and save the **current** `Settings.json` as a named graphics profile before assigning it. The safest workflow is to exit the game before saving, so D2R has time to flush the latest settings to disk.
- If you have assigned `GraphicsProfile` values to your accounts, pressing Enter in the launch graphics menu uses those saved defaults for single-account and batch launch. If any targeted account is still missing valid default graphics, the launcher blocks Enter, lists the affected accounts, and lets you either go back to `g` to configure them or choose `0` / another graphics-profile number manually.
- If you have assigned `DefaultMod` values to your accounts, pressing Enter in the launch mod menu uses those saved defaults for single-account and batch launch. If any targeted account is still missing a valid default mod, the launcher blocks Enter, lists the affected accounts, and lets you either go back to `m` to configure them or choose `0` / another mod number manually.
- If you have assigned `DefaultRegion` values to your accounts, pressing Enter in the launch region menu uses those saved defaults for single-account and batch launch. If any targeted account is still missing a default region, the launcher blocks Enter, lists the affected accounts, and lets you either go back to `v` to configure them or choose `1` / `2` / `3` manually.
- If a saved graphics profile is still assigned to any account, the launcher will block deletion until you clear or reassign those accounts first.
- If an account points at a graphics profile that no longer exists when you press Enter to use default graphics, the launcher first clears that account's `GraphicsProfile` assignment and then treats it as unassigned; you can switch to `0` or another saved graphics profile for that launch.
- If an account's saved `DefaultMod` no longer exists for the current D2R install, pressing Enter to use default mods will first clear that account's `DefaultMod` and then treat it as unassigned, instead of silently launching the wrong mod or falling back to vanilla.
- Avoid manually editing `config.json` to prevent accidentally breaking the JSON structure; use the in-tool menus for most settings
- The language setting is stored in `config.json` under the `language` field (`"zh-TW"` or `"en"`); if the field is absent, the tool will run the language selection flow again on the next startup
- Only the Battle.net version of D2R is supported
- Manipulating process handles may trigger false positives in some antivirus software
- This tool does not modify game files, inject into the game process, or automate any in-game actions
- This tool is a community self-use utility and is not affiliated with Blizzard Entertainment; use at your own risk — the author accepts no responsibility for any risk, loss, or consequence

## For Developers

### Prerequisites

- Windows 10 / 11
- Go 1.26+
- Battle.net version of D2R

### Building

To verify the program runs correctly on your machine, use:

```powershell
.\scripts\go-run.ps1
```

To verify compilation only, output to a temp location to avoid overwriting the release exe at the repo root:

```powershell
New-Item -ItemType Directory -Force .\.tmp | Out-Null
go build -o .\.tmp\d2r-hyper-launcher-dev.exe ./cmd/d2r-hyper-launcher
```

Only overwrite the repo-root `d2r-hyper-launcher.exe` during a release build, injecting version and release time (`yyyy-mm-dd hh:mm:ss`):

```powershell
go build -ldflags "-X main.version=vX.Y.Z -X main.releaseTime=YYYY-MM-DD HH:MM:SS" -o d2r-hyper-launcher.exe ./cmd/d2r-hyper-launcher
```

### Testing

If `go test ./...` is blocked by Windows Application Control, use the wrapper script instead:

```powershell
.\scripts\go-test.ps1
go build -o .\.tmp\d2r-hyper-launcher-dev.exe ./cmd/d2r-hyper-launcher
```

## License

MIT License

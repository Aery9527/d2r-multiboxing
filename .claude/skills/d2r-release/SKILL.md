---
name: d2r-release
description: "Handle repository-specific release work in d2r-hyper-launcher. Use this whenever the user wants to make a release, bump version, rebuild d2r-hyper-launcher.exe, prepare release notes, decide the next version number, or create a git tag. Always run tests before building, replace the root d2r-hyper-launcher.exe on every release build, determine the version bump from all commits since the last release tag, write release notes for the changes in that range, and create a new git tag for every release."
---

# D2R Release Workflow

這個 skill 專注在本 repo 的 release 流程。核心不是單純把程式 build 出來，而是用一致的規則完成 **測試 → 版本決策 → build 覆蓋 exe → release note → git tag**。

## 先看哪些內容

- `git tag --list` - 找出上一個 release tag
- `git log <last-tag>..HEAD --oneline` - 彙整自上次 release 以來的所有 commit
- [cmd/d2r-hyper-launcher/main.go](../../../cmd/d2r-hyper-launcher/main.go) - 確認版本字串由 `-ldflags "-X main.version=..."` 注入
- [README.md](../../../README.md) - 確認目前 build 與玩家可見入口
- [AGENT.md](../../../AGENT.md) - 測試、文件同步與整體規範
- `.claude/skills/d2r-commit/SKILL.md` - commit message 與 commit 前檢查的語氣基準

## release 前一定要做的事

1. 先確認目前位於 `develop` branch；若不在 `develop`，先回到正確分支再繼續。
2. 確認工作樹是否乾淨，避免把未整理好的變更混進 release。
3. 找出上一個 release tag；若沒有 tag，視為第一次 release。
4. 彙整「上一個 tag 到目前 HEAD」之間所有 commit，先理解這次 release 的高階變更主題。
5. **一定要先在 `develop` 上跑測試，測試通過後才能 build release。**
6. 先決定新版本號，再用新版本 build。
7. 每次 release 都要：
   - 覆蓋 repo 根目錄的 `d2r-hyper-launcher.exe`
   - 新增一份 release note
   - 完成後 merge 到 `master`
   - 最後建立一個新的 git tag

## 測試規則

release 前最低限度一定要跑：

```powershell
.\scripts\go-test.ps1
go build ./cmd/d2r-hyper-launcher
```

若目前環境沒有遭遇 Windows Application Control 阻擋，也可以直接使用 `go test ./...`；但在這台 repo 常見的 Windows 環境，請優先使用 `.\scripts\go-test.ps1`。

注意：

- 第一個 `go build ./cmd/d2r-hyper-launcher` 是驗證目前程式可成功編譯
- 真正 release build 要在版本號決定後，再用帶版本的 `-ldflags` 重跑一次
- 若測試或 build 失敗，不要繼續 release

## 版本號規則

預設採用 `vMAJOR.MINOR.PATCH`。

版本跳號不是看單一 commit，而是要看「上次 release 到現在所有 commit 彙整後」的整體變更範圍：

- **MAJOR**：有破壞性改動、使用流程重大重設、相容性顯著變更，或需要使用者重新理解操作方式
- **MINOR**：新增明顯的新功能、新 scope 能力、玩家可感知的新操作流程，且不屬於破壞性改動
- **PATCH**：既有功能修正、穩定性改善、文件同步、小幅 UX 調整、非破壞性的維護性更新

判斷方式：

1. 先看上一個 tag 到 `HEAD` 的所有 commit
2. 再按功能面彙整，不要逐條 commit 機械式對應版本號
3. 以「使用者拿到新 exe 後感受到的變化等級」決定 bump 程度

若 repo 尚未有任何 release tag，預設從 `v0.1.0` 開始，除非使用者明確指定別的起始版本。

## 版本在這個 repo 怎麼生效

這個 repo 目前沒有獨立 `VERSION` 檔，版本會在 build 時透過：

```powershell
-ldflags "-X main.version=vX.Y.Z"
```

注入到：

- `cmd/d2r-hyper-launcher/main.go` 的 `version`
- 最終輸出的 `d2r-hyper-launcher.exe`

所以這裡說的「先變更 version 再 build」，在本 repo 的實際意思是：

1. 先決定新的 release 版本號
2. 再用該版本號執行 release build

## release build 規則

每次 release build 都必須直接覆蓋 repo 根目錄的：

```text
d2r-hyper-launcher.exe
```

命令格式：

```powershell
go build -ldflags "-X main.version=vX.Y.Z" -o d2r-hyper-launcher.exe ./cmd/d2r-hyper-launcher
```

不要把 release build 輸出到其他暫存檔名，除非使用者另外要求。

## release note 規則

每次 release 都要記錄「上一個 release 到這個 release 之間」的改變。

建議位置：

```text
docs/releases/vX.Y.Z.md
```

若 `docs/releases/` 不存在，就建立它。

release note 應該聚焦：

- 這次 release 的高階主題
- 玩家或使用者會感受到的變化
- 涉及哪些 scope（例如 `multiboxing`、`switcher`、`docs`、`repo workflow`）
- 若有必要，補充升級注意事項

建議結構：

```markdown
# vX.Y.Z

## Summary
- ...

## Highlights
- ...

## Scope changes
### multiboxing
- ...

### switcher
- ...

### docs / workflow
- ...

## Upgrade notes
- ...
```

不要只是把 `git log` 原封不動貼上；要先整理成可讀的 release note。

## git tag 規則

每次 release 都要建立 tag，格式預設為：

```text
vX.Y.Z
```

優先使用 annotated tag：

```powershell
git tag -a vX.Y.Z -m "release: vX.Y.Z"
```

建立前先確認同名 tag 不存在。

另外，tag 時機要放在 release 完成並 merge 到 `master` 之後，不要在尚未完成 release 或尚未 merge 前提早下 tag。

## 建議的完整流程

1. `git status`
2. 確認目前在 `develop`
3. 找上一個 tag
4. 看 `git log <last-tag>..HEAD --oneline`
5. 彙整 commit 並決定 bump 等級
6. 在 `develop` 跑 `.\scripts\go-test.ps1`
7. 在 `develop` 跑 `go build ./cmd/d2r-hyper-launcher`
8. 決定新版本 `vX.Y.Z`
9. 寫 release note 到 `docs/releases/vX.Y.Z.md`
10. 用新版本 build，覆蓋 `d2r-hyper-launcher.exe`
11. 將 release 結果 merge 到 `master`
12. 建立 `vX.Y.Z` git tag

## 與 commit skill 的分工

- `d2r-commit` 著重單次 commit 前的測試與 commit message
- `d2r-release` 著重多個 commit 彙整後的版本決策、release build、release note 與 tag

若使用者是要「發版」，優先使用這個 skill，而不是只套用 commit skill。

## 使用這個 skill 時的回應期待

當使用者要求 release 時：

1. 先回報你要檢查 tags、彙整 commits、跑測試
2. 明確說明建議 bump 的版本與原因
3. 明確回報測試是否通過
4. 說明 release build 是否已覆蓋 `d2r-hyper-launcher.exe`
5. 提供 release note 路徑與新 tag 名稱

---
name: d2r-commit
description: "Handle repository-specific git commit work in d2r-hyper-launcher. Use this whenever the user asks to commit changes, prepare a commit, write a commit message, summarize changes into a commit, or make the current worktree ready to commit. Run the relevant tests before committing unless the entire change is clearly limited to Markdown or comment-only edits that do not affect execution flow, and make the commit message emphasize the high-level intent, user impact, or scope outcome rather than low-level code edits."
---

# D2R Commit Workflow

這個 skill 專注在本 repo 的 commit 準備與 commit message 撰寫。重點不是逐條描述改了哪些程式碼，而是用高階語意表達「這次變更想達成什麼」以及「影響哪個 scope」。

## 先看哪些內容

- `git status` / `git diff --stat` / `git diff` - 確認實際變更範圍
- [AGENT.md](../../../AGENT.md) - 專案規範、測試原則、文件同步原則
- [README.md](../../../README.md) 與 `docs/` - 判斷是否有使用者可見行為改動
- `.claude/skills/` - 判斷既有 skill 是否也要同步更新

## commit 前一定要做的事

1. 先確認變更範圍，不要在不理解差異的情況下直接 commit。
2. 先判斷這次變更是否屬於「需要驗證執行流程」的範圍：
   - 若包含 Go / script / 設定 / build / 行為邏輯變更，**一定要先跑測試，測試通過後才能 commit。**
   - 若今天所有變更都明確只是 `.md` 文件或程式註解調整，且**不影響實際執行流程**，可以跳過測試直接 commit。
   - 若對「是不是純文件 / 純註解」有疑慮，採保守做法：照常跑測試。
3. 若改動影響使用者可見流程、設定、限制、技術前提或 skill 描述，先同步更新相關 `README`、`docs/`、`.claude/skills/`。
4. 若測試失敗，不要硬 commit；先修正或明確告知使用者目前仍有失敗。
5. 建立 commit 後，預設直接 push 到目前分支的 remote upstream。
6. 若 push 前需要同步遠端而遇到 conflict，或 push / rebase / merge 過程出現衝突，停止操作並通知使用者 review。

## 測試規則

這個 repo 已驗證可用的最低檢查是：

```powershell
.\scripts\go-test.ps1
New-Item -ItemType Directory -Force .\.tmp | Out-Null
go build -o .\.tmp\d2r-hyper-launcher-dev.exe ./cmd/d2r-hyper-launcher
```

若目前環境沒有遭遇 Windows Application Control 阻擋，也可以直接使用 `go test ./...`；但在這台 repo 常見的 Windows 環境，請優先使用 `.\scripts\go-test.ps1`。

執行原則：

- 一般 Go 程式碼變更：`.\scripts\go-test.ps1` 與 `go build -o .\.tmp\d2r-hyper-launcher-dev.exe ./cmd/d2r-hyper-launcher`
- 純 `.md` 或純註解變更，且與執行流程無關：可跳過測試，但仍要檢查是否有遺漏需要同步更新的文件 / skill
- 只改文件但內容是在描述流程、限制或設定時：先確認只是說明修正，沒有伴隨尚未驗證的程式變更；若其實綁著行為改動，就不能用文件例外
- 若變更集中在特定套件，也可以先補跑較小範圍測試，但 **不能取代** commit 前的整體驗證

## commit message 原則

- 聚焦高階目的、產品意圖、使用者影響或 scope 結果
- 不要把 subject 寫成單純的程式碼操作清單
- 優先描述「改善了什麼 / 整理了什麼 / 修正了什麼體驗」
- 若能明確指出 scope，使用 `multiboxing`、`switcher`、`docs`、`repo` 等詞幫助理解

建議格式：

```text
<type>(<scope>): <high-level outcome>
```

常見 `type`：

- `feat`：新功能
- `fix`：修正問題
- `docs`：文件整理
- `refactor`：重構但不改變對外行為
- `chore`：維護性調整

## 好與不好的 commit subject

**避免這種寫法：**

- `fix: update main.go and account.go`
- `docs: rewrite readme and add two md files`
- `refactor: rename variables and move functions`

**改成這種寫法：**

- `fix(multiboxing): stabilize multi-account startup flow`
- `docs(repo): simplify player onboarding and usage navigation`
- `refactor(switcher): streamline trigger setup workflow`

這些寫法比較能讓人一眼看出 commit 的意圖，而不是只能看到實作細節。

## body 怎麼寫

如果需要 body，補充：

- 為什麼要做這次變更
- 這次變更影響哪些 scope
- 有沒有同步更新測試 / 文件 / skill

範例：

```text
docs(repo): simplify player onboarding and usage navigation

Consolidate beginner-facing setup guidance into README and split
detailed操作說明到 docs usage guides so players can start from the
download flow first while still preserving deeper references.
```

## 真正要 commit 時的流程

1. 看 `git status`
2. 看 `git diff --stat` 與必要的 `git diff`
3. 判斷是否符合文件 / 註解例外；若不符合就跑測試
4. 確認文件與 skill 是否已同步
5. 撰寫高階語意的 commit message
6. 執行 `git commit`
7. 直接 `git push`
8. 若 push 失敗且涉及同步衝突，通知使用者 review，不要自作主張解 conflict

若在這個環境實際建立 commit，commit message 最後要加上：

```text
Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>
```

## 使用這個 skill 時的回應期待

當使用者要求 commit 或寫 commit message 時：

1. 先說明你要檢查變更，並判斷這次是否需要跑測試
2. 若有跑測試，明確回報測試是否通過；若因純文件 / 純註解例外而跳過，也要明確說明跳過原因
3. 再提供或使用高階語意的 commit message
4. commit 後預設直接回報 push 結果
5. 若發現文件 / skill 尚未同步，先補齊再 commit
6. 若 push 或同步過程有 conflict，明確請使用者 review

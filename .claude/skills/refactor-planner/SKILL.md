---
name: refactor-planner
description: "Create a reviewable refactor plan when code structure shows architecture-smell signals. Use this whenever the user asks whether the current design is a problem, wants a refactor or architecture plan, mentions duplicated logic, scattered same-scope behavior, low cohesion, leaky boundaries, or when implementing a task reveals you had to add the same or very similar code in multiple places. Finish the current task first, then analyze the structure and produce a proposal for review instead of silently performing a large refactor."
---

# Refactor Planning Workflow

這個 skill 的職責不是直接重構，而是把「結構性問題」整理成可 review 的改善方案。

它特別適合這種情境：

- 當前任務已經完成，但實作過程暴露出 duplication / low cohesion
- 使用者開始質疑目前結構是不是有問題
- 你意識到同一類邏輯散在很多地方，後續維護風險會很高

## 核心原則

1. **先完成當前任務**
2. **不要未經同意就直接展開大型 refactor**
3. **把架構問題變成一份可 review 的 plan**

這個 skill 的目標，是讓「交付需求」與「整理架構」變成兩段清楚的流程，而不是互相打架。

## 輸出位置規則

分析完成後，除了在對話裡摘要回報，也要**直接把 plan 寫成 markdown 檔**：

```text
refactor-plan/<core-scope>-refactor.md
```

命名原則：

- 用核心 scope 命名，而不是任務編號
- 檔名一律以 `-refactor.md` 作為 suffix
- 優先使用簡短、穩定、可重複沿用的名稱
- 例如：
  - `refactor-plan/cli-refactor.md`
  - `refactor-plan/auth-refactor.md`
  - `refactor-plan/parser-refactor.md`
  - `refactor-plan/import-pipeline-refactor.md`

若對應 scope 的 plan 已存在，預設更新既有 `*-refactor.md` 檔案，而不是另外再生一個相近名稱的新檔。

## 什麼情況算是架構訊號

以下任一情況，通常都表示值得啟動這個 skill：

### 1. 重複邏輯開始出現

- 為了完成一個需求，要在多個地方加入幾乎相同的 code
- 同一個 validation / mapping / guard / formatting / retry / fallback 被複製到不同入口
- 修一個 bug 時，必須靠「記得去好幾個地方一起改」才能保證一致

### 2. 同 scope 的 code 缺乏內聚力

- 一個明顯屬於同一個概念的行為，卻散落在多個檔案 / 函式 / menu flow
- 新功能每次都要橫跨很多入口才能接起來
- 邏輯邊界不清楚，導致同一件事有很多半重疊實作

### 3. 結構成本開始高於功能本身

- 加一個小功能，卻需要碰很多零散位置
- 測試難寫，因為行為散落到太多地方
- 文件容易過期，因為沒有一個明確的單一落點可對應

### 4. 使用者直接提出結構疑慮

- 「這樣的架構是不是有問題？」
- 「幫我規劃 refactor」
- 「這塊 code 太散了，整理一下」
- 「這是不是腐敗的程式碼？」

> 這不一定代表整個系統設計都失敗；通常表示「這一塊的抽象、邊界或模組劃分不夠好」。

## 腐敗 / 需要重整的常見範例

下面是值得觸發此 skill 的典型例子：

- **Shotgun surgery**：一個需求要改很多分散位置
- **Copy-paste growth**：新功能是靠重複貼上舊邏輯擴出來
- **Low cohesion**：同一個概念被拆散在很多不相鄰的地方
- **Leaky boundaries**：本該由單一模組負責的事，外面很多地方都在偷做
- **Inconsistent behavior risk**：同類行為很容易因漏改而不一致

如果你在任務中感受到「現在雖然能交付，但再這樣長下去會越來越難維護」，通常就是這個 skill 的時機。

## 執行順序

### 第一步：先交付當前任務

若使用者目前要的是 bug fix、feature、docs、commit、release 或其他具體交付：

1. 先把那個任務做完
2. 先做必要驗證
3. 再啟動這個 skill 產出 refactor plan

不要把「順便整理一下」擴張成未經允許的大重構。

### 第二步：蒐集證據

至少整理出：

- duplication 出現在哪些檔案 / 函式 / class / flow
- 哪些邏輯其實屬於同一個概念，卻分散在不同地方
- 這種分散造成哪些成本：
  - 容易漏改
  - 行為不一致
  - 測試難補
  - 文件難同步
  - 擴充成本高

### 第三步：必要時 fork agent

如果架構分析本身會很吃 context，或需要較完整的橫向閱讀：

- fork `explore` 或 `general-purpose` agent
- 讓子 agent 專注在：
  - duplication 地圖
  - 建議的共用抽象落點
  - 低風險 / 中風險 / 高風險方案

主 agent 保持目前交付結果清楚，不要把分析和實作混成一團。

## 預設輸出：plan，不直接實作

除非使用者明確說要動手 refactor，否則預設只產出 plan。

完成時請做兩件事：

1. 將 plan 寫入 `refactor-plan/<core-scope>-refactor.md`
2. 在對話中給使用者一個精簡摘要，方便 review

## 當使用者 review 後要求開始實作

如果使用者已經 review 完 `refactor-plan/<core-scope>-refactor.md`，並明確要求你開始執行其中內容：

1. **先建立新 branch，再開始 refactor**
2. branch 名稱要根據當前任務內容適當命名，不要用隨意流水號
3. 命名應能讓人一眼看出這次 refactor 的核心 scope

命名示例：

- `refactor/cli-selection-flow`
- `refactor/auth-boundary`
- `refactor/parser-cohesion`
- `refactor/import-pipeline`

避免：

- `test-branch`
- `tmp-refactor`
- `refactor-plan`
- `branch-123`

核心原因：

- refactor 通常跨檔案、跨模組，隔離在獨立 branch 比較容易 review
- 使用者已批准的是「執行這份 plan」，不是把它和其他進行中的工作混在一起
- 好的 branch 名稱能讓後續 commit、PR、review 都更清楚

所以這個 skill 的完整節奏是：

1. 先完成當前任務
2. 產出 `refactor-plan/<core-scope>-refactor.md`
3. 等使用者 review
4. 若使用者批准實作，再建立對應的 `refactor/<task-scope>` branch
5. 然後才開始真正的 refactor

## 當使用者確認這個 refactor 已完成

如果使用者表示這輪 refactor 已經做完，接下來不要只是停在「branch 上有改完」；要把可追溯紀錄也一起收尾。

預設做法：

1. 確認對應的 `refactor-plan/<core-scope>-refactor.md` 已能反映最終實作結果
   - 若方案有調整，先更新 plan
   - 讓之後回看的人知道：原本要解什麼、最後實際做了什麼
2. 用一致的 scope 字串整理整段紀錄
   - plan 檔名
   - branch 名稱
   - commit subject
   - PR 標題
   - merge commit
3. 引導使用者把 refactor 透過 PR / merge 收斂回 `develop` 或主線
4. merge 完成後，**預設刪除 refactor branch**
   - branch 是隔離工作與 review 的地方，不是長期保存歷史的主要載體
5. 保留可搜尋的長期痕跡在：
   - `refactor-plan/<core-scope>-refactor.md`
   - PR 記錄
   - merge commit / squash commit
   - 若有 issue，也要互相連結

核心觀念：

- **不要因為想保留歷史就長期保留一堆 refactor branch**
- refactor 是否曾存在，應該主要從 **plan + PR + merge commit** 追溯，而不是依賴 branch 一直掛在遠端
- branch 刪掉不代表歷史消失；只要 merge commit 與 PR 命名清楚，仍然很好追

### merge / PR 追蹤原則

完成 refactor 後，優先把以下資訊串起來：

1. `refactor-plan/<core-scope>-refactor.md`
2. `refactor/<task-scope>` branch
3. PR 標題
4. 最終 merge commit 或 squash commit

建議這些地方都盡量包含同一個核心 scope 名稱，例如：

- `refactor-plan/cli-refactor.md`
- `refactor/cli-internal-boundaries`
- PR: `refactor(repo): realign CLI and feature boundaries`
- merge commit: `refactor(repo): realign CLI and feature boundaries`

這樣之後可以靠：

- `git log --grep=refactor`
- GitHub PR 搜尋
- `refactor-plan/` 目錄

快速追到這次 refactor 的完整脈絡。

### tag 原則

tag **不是預設的 refactor 追蹤主體**。

原因：

- 若每個 refactor 都打開始 / 結束 tag，長期也會堆出另一種噪音
- tag 比較適合 release、少數重要里程碑、或使用者明確要求保留的架構節點

預設建議：

- **不要**為每個 refactor 都建立開始 tag 與結束 tag
- 只有在下列情況才考慮建立 **單一 annotated tag**
  - 使用者明確要求
  - 這是一個特別重要、跨越多輪的大型架構轉折
  - 團隊真的需要把這個節點當成長期里程碑

如果要打 tag，優先在「最終完成並 merge 之後」建立，而不是一開始就成對建立開始 / 結束 tag。

### 使用者確認完成時的建議收尾節奏

1. 確認 refactor 已完成且必要驗證通過
2. 更新 `refactor-plan/<core-scope>-refactor.md`，讓內容能對應最終結果
3. 整理 commit / PR 標題，確保命名可搜尋
4. merge 回 `develop` 或主線
5. 刪除 refactor branch
6. 只有在真的有里程碑需求時，才補一個 annotated tag

若使用者問「之後怎麼知道這個 refactor 曾經存在過」，請明確回答：主要從 **plan、PR、merge commit、issue** 追，而不是靠保留 branch。

建議格式：

```markdown
## 為什麼這是架構訊號
- ...

## 問題落點
- 檔案 / 模組 A：...
- 檔案 / 模組 B：...

## 目前成本
- ...

## 可行方案
1. 低風險方案：...
2. 中風險方案：...
3. 較完整的重整方向：...

## 建議順序
1. ...
2. ...
3. ...

## 暫時不做的事
- ...
```

## 判斷與建議原則

- 以 **內聚力**、**單一責任**、**一致行為應該有單一落點** 為優先
- 不要為了抽象而抽象
- 若已有簡單共用 helper 就能明顯降低重複，先提低風險方案
- 若問題其實牽涉模組邊界，才提較大範圍重整
- 明確區分：
  - **現在一定要做的**
  - **適合下一輪做的**
  - **只是長期方向，不該現在動的**

## 回應使用者時要說清楚

1. 這是不是架構問題的訊號
2. 為什麼它不是單純一次性小改動
3. 目前已先完成哪個當前任務
4. 這份 refactor plan 想解決什麼
5. plan 已寫到哪個 `refactor-plan/<core-scope>-refactor.md` 檔案
6. 若使用者批准實作，會先建立哪種命名風格的新 branch
7. 這份 plan 目前只是 proposal，是否真的實作由使用者決定

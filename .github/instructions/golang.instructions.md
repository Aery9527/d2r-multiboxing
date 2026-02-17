---
applyTo: "**/*.go"
---

# comment

- struct 明確實作某個介面時, 在 struct 宣告上方寫 `var _ InterfaceName = (*StructName)(nil)` 明確標記該 struct 實作了某個介面, 除了快速理解 struct 功能外也作靜態驗證
- 使用 `any` 而非 `interface{}` 來表示任意類型
- 操作 mongo 時, 要特別注意 使用 `bson.M` 跟 `bson.D` 使用時機, 在特別嚴格注重順序場合(如 `$sort`) 一定要使用 `bson.D`

# test

- 撰寫 test 時請使用 `github.com/stretchr/testify/assert` 進行驗證

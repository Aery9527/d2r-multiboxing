package account

import _ "embed"

// accountsCSVTemplate is the embedded starter template for first-run accounts.csv creation.
//
//go:embed templates/accounts.csv
var accountsCSVTemplate []byte

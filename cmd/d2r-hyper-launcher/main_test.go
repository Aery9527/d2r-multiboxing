package main

import (
	"testing"

	"d2rhl/internal/account"

	"github.com/stretchr/testify/assert"
)

func TestNextPendingAccountDisplayNameSkipsRunningAccounts(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha"},
		{DisplayName: "Bravo"},
		{DisplayName: "Charlie"},
	}

	next := nextPendingAccountDisplayName(accounts, 1, func(displayName string) bool {
		return displayName == "Bravo"
	})

	assert.Equal(t, "Charlie", next)
}

func TestNextPendingAccountDisplayNameReturnsEmptyWhenNoPendingAccount(t *testing.T) {
	accounts := []account.Account{
		{DisplayName: "Alpha"},
		{DisplayName: "Bravo"},
	}

	next := nextPendingAccountDisplayName(accounts, 1, func(string) bool {
		return true
	})

	assert.Empty(t, next)
}

func TestFormatLaunchDelayMessage(t *testing.T) {
	assert.Equal(t, "  等待 30 秒後啟動下一個帳號：VoidLife", formatLaunchDelayMessage(30, "VoidLife"))
}

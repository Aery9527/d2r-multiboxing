package process

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildOnlineArgs(t *testing.T) {
	args := buildOnlineArgs("user@example.com", "secret", "us.actual.battle.net", "-mod", "sample", "-txt")
	assert.Equal(t, []string{
		"-uid", "osi",
		"-mod", "sample", "-txt",
		"-username", "user@example.com",
		"-password", "secret",
		"-address", "us.actual.battle.net",
	}, args)
}

func TestBuildOfflineArgs(t *testing.T) {
	assert.Equal(t, []string{"-uid", "osi"}, buildOfflineArgs())
	assert.Equal(t, []string{"-uid", "osi", "-mod", "sample", "-txt"}, buildOfflineArgs("-mod", "sample", "-txt"))
}

func TestRedactArgs(t *testing.T) {
	args := redactArgs([]string{"-uid", "osi", "-password", "secret", "-address", "us.actual.battle.net"})
	assert.Equal(t, []string{"-uid", "osi", "-password", "****", "-address", "us.actual.battle.net"}, args)
}

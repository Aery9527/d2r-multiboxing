package account

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLaunchArgs(t *testing.T) {
	args := LaunchArgs(LaunchFlagNoSound | (1 << 2))
	assert.Equal(t, []string{"-ns"}, args)
}

func TestLaunchFlagsSummary(t *testing.T) {
	emptySummary := LaunchFlagsSummary(0)
	supportedSummary := LaunchFlagsSummary(LaunchFlagNoSound)
	legacySummary := LaunchFlagsSummary(LaunchFlagNoSound | (1 << 2))

	assert.NotEmpty(t, emptySummary)
	assert.NotEmpty(t, supportedSummary)
	assert.NotEqual(t, emptySummary, supportedSummary)
	assert.Equal(t, supportedSummary, legacySummary)
}

func TestSupportedLaunchFlagsMask(t *testing.T) {
	assert.Equal(t, uint32(LaunchFlagNoSound), SupportedLaunchFlagsMask())
}

func TestSanitizeLaunchFlagsRemovesUnsupportedBits(t *testing.T) {
	flags := LaunchFlagNoSound | (1 << 1) | (1 << 3) | (1 << 4)
	assert.Equal(t, uint32(LaunchFlagNoSound), SanitizeLaunchFlags(flags))
}

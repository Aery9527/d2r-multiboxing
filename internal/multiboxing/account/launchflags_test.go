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
	assert.Equal(t, "無", LaunchFlagsSummary(0))
	assert.Equal(t, "關閉聲音", LaunchFlagsSummary(LaunchFlagNoSound|(1<<2)))
}

func TestSupportedLaunchFlagsMask(t *testing.T) {
	assert.Equal(t, uint32(LaunchFlagNoSound), SupportedLaunchFlagsMask())
}

func TestSanitizeLaunchFlagsRemovesUnsupportedBits(t *testing.T) {
	flags := LaunchFlagNoSound | (1 << 1) | (1 << 3) | (1 << 4)
	assert.Equal(t, uint32(LaunchFlagNoSound), SanitizeLaunchFlags(flags))
}

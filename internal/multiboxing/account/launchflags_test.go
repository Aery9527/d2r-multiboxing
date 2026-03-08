package account

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLaunchArgs(t *testing.T) {
	args := LaunchArgs(LaunchFlagNoSound | LaunchFlagLowQuality)
	assert.Equal(t, []string{"-ns", "-lq"}, args)
}

func TestLaunchFlagsSummary(t *testing.T) {
	assert.Equal(t, "無", LaunchFlagsSummary(0))
	assert.Equal(t, "關閉聲音、低畫質 / Large Font Mode", LaunchFlagsSummary(LaunchFlagNoSound|LaunchFlagLowQuality))
}

func TestSupportedLaunchFlagsMask(t *testing.T) {
	assert.Equal(t, uint32(LaunchFlagNoSound|LaunchFlagLowQuality), SupportedLaunchFlagsMask())
}

func TestSanitizeLaunchFlagsRemovesUnsupportedBits(t *testing.T) {
	flags := LaunchFlagNoSound | (1 << 1) | (1 << 3) | (1 << 4)
	assert.Equal(t, uint32(LaunchFlagNoSound), SanitizeLaunchFlags(flags))
}

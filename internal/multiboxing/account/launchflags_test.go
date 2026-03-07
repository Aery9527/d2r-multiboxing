package account

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLaunchArgs(t *testing.T) {
	args := LaunchArgs(LaunchFlagNoSound | LaunchFlagLowQuality | LaunchFlagNoRumble)
	assert.Equal(t, []string{"-ns", "-lq", "-norumble"}, args)
}

func TestLaunchFlagsSummary(t *testing.T) {
	assert.Equal(t, "無", LaunchFlagsSummary(0))
	assert.Equal(t, "關閉聲音、跳過 Logo 影片", LaunchFlagsSummary(LaunchFlagNoSound|LaunchFlagSkipLogoVideo))
}

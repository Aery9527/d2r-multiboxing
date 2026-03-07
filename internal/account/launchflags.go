package account

import "strings"

const (
	LaunchFlagNoSound uint32 = 1 << iota
	LaunchFlagSoundInBackground
	LaunchFlagLowQuality
	LaunchFlagSkipLogoVideo
	LaunchFlagNoRumble
)

type LaunchFlagOption struct {
	Bit          uint32
	Name         string
	Description  string
	Args         []string
	Experimental bool
}

var launchFlagOptions = []LaunchFlagOption{
	{
		Bit:         LaunchFlagNoSound,
		Name:        "關閉聲音",
		Description: "-ns / -nosound",
		Args:        []string{"-ns"},
	},
	{
		Bit:         LaunchFlagSoundInBackground,
		Name:        "背景保留聲音",
		Description: "-sndbkg",
		Args:        []string{"-sndbkg"},
	},
	{
		Bit:          LaunchFlagLowQuality,
		Name:         "低畫質 / Large Font Mode",
		Description:  "-lq（效果依版本而定）",
		Args:         []string{"-lq"},
		Experimental: true,
	},
	{
		Bit:         LaunchFlagSkipLogoVideo,
		Name:        "跳過 Logo 影片",
		Description: "-skiplogovideo",
		Args:        []string{"-skiplogovideo"},
	},
	{
		Bit:         LaunchFlagNoRumble,
		Name:        "停用手把震動",
		Description: "-norumble",
		Args:        []string{"-norumble"},
	},
}

func LaunchFlagOptions() []LaunchFlagOption {
	out := make([]LaunchFlagOption, len(launchFlagOptions))
	copy(out, launchFlagOptions)
	return out
}

func LaunchArgs(flags uint32) []string {
	args := make([]string, 0, len(launchFlagOptions))
	for _, option := range launchFlagOptions {
		if flags&option.Bit == 0 {
			continue
		}
		args = append(args, option.Args...)
	}
	return args
}

func LaunchFlagsSummary(flags uint32) string {
	if flags == 0 {
		return "無"
	}

	parts := make([]string, 0, len(launchFlagOptions))
	for _, option := range launchFlagOptions {
		if flags&option.Bit == 0 {
			continue
		}
		parts = append(parts, option.Name)
	}
	if len(parts) == 0 {
		return "無"
	}
	return strings.Join(parts, "、")
}

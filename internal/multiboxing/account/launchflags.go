package account

import "strings"

const (
	LaunchFlagNoSound uint32 = 1 << iota
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
}

func LaunchFlagOptions() []LaunchFlagOption {
	out := make([]LaunchFlagOption, len(launchFlagOptions))
	copy(out, launchFlagOptions)
	return out
}

func SupportedLaunchFlagsMask() uint32 {
	var mask uint32
	for _, option := range launchFlagOptions {
		mask |= option.Bit
	}
	return mask
}

func SanitizeLaunchFlags(flags uint32) uint32 {
	return flags & SupportedLaunchFlagsMask()
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

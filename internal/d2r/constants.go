// Package d2r provides constants and definitions specific to Diablo II: Resurrected.
package d2r

const (
	// ProcessName is the D2R executable name.
	ProcessName = "D2R.exe"

	// SingleInstanceEventName is the event handle name used by D2R to prevent multiple instances.
	// The actual handle name includes a session prefix like "\Sessions\1\BaseNamedObjects\".
	SingleInstanceEventName = "DiabloII Check For Other Instances"

	// WindowClassName is the D2R window class name.
	WindowClassName = "OsWindow"

	// DefaultWindowTitle is the default D2R window title.
	DefaultWindowTitle = "Diablo II: Resurrected"

	// DefaultGamePath is the default installation path of D2R.
	DefaultGamePath = `C:\Program Files (x86)\Diablo II Resurrected\D2R.exe`
)

// Region represents a Battle.net server region.
type Region struct {
	Name    string
	Address string
}

// Regions is the list of available Battle.net server regions.
var Regions = []Region{
	{Name: "NA", Address: "us.actual.battle.net"},
	{Name: "EU", Address: "eu.actual.battle.net"},
	{Name: "Asia", Address: "kr.actual.battle.net"},
}

// FindRegion returns the Region matching the given name (case-insensitive), or nil if not found.
func FindRegion(name string) *Region {
	for i := range Regions {
		if equalsIgnoreCase(Regions[i].Name, name) {
			return &Regions[i]
		}
	}
	return nil
}

func equalsIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

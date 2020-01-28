package atv

import (
	"fmt"
)

// Version represents the version of an ATV document.
type Version struct {
	Major  int
	Minor  int
	Patch  int
	Suffix string
}

// String returns the version as a string.
func (version Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)
	if len(version.Suffix) > 0 {
		s += fmt.Sprintf(".%s", version.Suffix)
	}
	return s
}

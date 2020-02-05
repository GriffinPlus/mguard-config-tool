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

// Compare compares the current version with the specified one.
// Returns -1, if the current version is less than the specified one.
// Returns 0, if the current version equals the current one.
// Returns +1, if the current version is greater than the specified one.
func (version Version) Compare(other Version) int {

	// check major version
	if version.Major < other.Major {
		return -1
	}
	if version.Major > other.Major {
		return 1
	}

	// same major version, check minor version
	if version.Minor < other.Minor {
		return -1
	}
	if version.Minor > other.Minor {
		return 1
	}

	// same minor version, check patch version
	if version.Patch < other.Patch {
		return -1
	}
	if version.Patch > other.Patch {
		return 1
	}

	// the versions are equal
	return 0
}

// String returns the version as a string.
func (version Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)
	if len(version.Suffix) > 0 {
		s += fmt.Sprintf(".%s", version.Suffix)
	}
	return s
}

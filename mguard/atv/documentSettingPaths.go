package atv

// documentSettingPaths represents a list of parsed setting paths.
type documentSettingPaths []documentSettingPath

// Contains checks whether the specified setting path is in the list.
func (paths documentSettingPaths) Contains(other documentSettingPath) bool {
	for _, path := range paths {
		if path.String() == other.String() {
			return true
		}
	}
	return false
}

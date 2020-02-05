package atv

type migration_7_5_0_to_8_0_2 struct{}

// FromVersion returns the document version the migration start with.
func (_ migration_7_5_0_to_8_0_2) FromVersion() Version {
	return Version{
		Major: 7,
		Minor: 5,
		Patch: 0,
		Suffix: "default",
	}
}

// ToVersion returns the document version the migration ends with.
func (_ migration_7_5_0_to_8_0_2) ToVersion() Version {
	return Version{
		Major: 8,
		Minor: 0,
		Patch: 2,
		Suffix: "default",
	}
}

// Migrate performs the migration.
func (migration migration_7_5_0_to_8_0_2) Migrate(file *File) (*File, error) {
	newFile := file.Dupe()
	newFile.SetVersion(migration.ToVersion())
	return newFile, nil
}

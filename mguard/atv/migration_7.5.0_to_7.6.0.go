package atv

type migration_7_5_0_to_7_6_0 struct{}

// FromVersion returns the document version the migration start with.
func (_ migration_7_5_0_to_7_6_0) FromVersion() Version {
	return Version{
		Major:  7,
		Minor:  5,
		Patch:  0,
		Suffix: "default",
	}
}

// ToVersion returns the document version the migration ends with.
func (_ migration_7_5_0_to_7_6_0) ToVersion() Version {
	return Version{
		Major:  7,
		Minor:  6,
		Patch:  0,
		Suffix: "default",
	}
}

// Migrate performs the migration.
func (migration migration_7_5_0_to_7_6_0) Migrate(file *File) (*File, error) {
	newFile := file.Dupe()
	newFile.SetVersion(migration.ToVersion())
	return newFile, nil
}

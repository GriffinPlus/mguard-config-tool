package atv

type migration_8_5_0_to_8_5_1 struct{}

// FromVersion returns the document version the migration start with.
func (_ migration_8_5_0_to_8_5_1) FromVersion() Version {
	return Version{
		Major:  8,
		Minor:  5,
		Patch:  0,
		Suffix: "default",
	}
}

// ToVersion returns the document version the migration ends with.
func (_ migration_8_5_0_to_8_5_1) ToVersion() Version {
	return Version{
		Major:  8,
		Minor:  5,
		Patch:  1,
		Suffix: "default",
	}
}

// Migrate performs the migration.
func (migration migration_8_5_0_to_8_5_1) Migrate(file *File) (*File, error) {
	newFile := file.Dupe()
	newFile.SetVersion(migration.ToVersion())
	return newFile, nil
}

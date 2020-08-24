package atv

type migration_8_1_7_to_8_1_8 struct{}

// FromVersion returns the document version the migration start with.
func (_ migration_8_1_7_to_8_1_8) FromVersion() Version {
	return Version{
		Major:  8,
		Minor:  1,
		Patch:  7,
		Suffix: "default",
	}
}

// ToVersion returns the document version the migration ends with.
func (_ migration_8_1_7_to_8_1_8) ToVersion() Version {
	return Version{
		Major:  8,
		Minor:  1,
		Patch:  8,
		Suffix: "default",
	}
}

// Migrate performs the migration.
func (migration migration_8_1_7_to_8_1_8) Migrate(file *File) (*File, error) {
	newFile := file.Dupe()
	newFile.SetVersion(migration.ToVersion())
	return newFile, nil
}

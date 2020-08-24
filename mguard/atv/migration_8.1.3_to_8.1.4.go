package atv

type migration_8_1_3_to_8_1_4 struct{}

// FromVersion returns the document version the migration start with.
func (_ migration_8_1_3_to_8_1_4) FromVersion() Version {
	return Version{
		Major:  8,
		Minor:  1,
		Patch:  3,
		Suffix: "default",
	}
}

// ToVersion returns the document version the migration ends with.
func (_ migration_8_1_3_to_8_1_4) ToVersion() Version {
	return Version{
		Major:  8,
		Minor:  1,
		Patch:  4,
		Suffix: "default",
	}
}

// Migrate performs the migration.
func (migration migration_8_1_3_to_8_1_4) Migrate(file *File) (*File, error) {
	newFile := file.Dupe()
	newFile.SetVersion(migration.ToVersion())
	return newFile, nil
}

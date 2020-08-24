package atv

type migration_8_1_4_to_8_1_5 struct{}

// FromVersion returns the document version the migration start with.
func (_ migration_8_1_4_to_8_1_5) FromVersion() Version {
	return Version{
		Major:  8,
		Minor:  1,
		Patch:  4,
		Suffix: "default",
	}
}

// ToVersion returns the document version the migration ends with.
func (_ migration_8_1_4_to_8_1_5) ToVersion() Version {
	return Version{
		Major:  8,
		Minor:  1,
		Patch:  5,
		Suffix: "default",
	}
}

// Migrate performs the migration.
func (migration migration_8_1_4_to_8_1_5) Migrate(file *File) (*File, error) {
	newFile := file.Dupe()
	newFile.SetVersion(migration.ToVersion())
	return newFile, nil
}

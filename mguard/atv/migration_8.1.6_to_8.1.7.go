package atv

type migration_8_1_6_to_8_1_7 struct{}

// FromVersion returns the document version the migration start with.
func (_ migration_8_1_6_to_8_1_7) FromVersion() Version {
	return Version{
		Major:  8,
		Minor:  1,
		Patch:  6,
		Suffix: "default",
	}
}

// ToVersion returns the document version the migration ends with.
func (_ migration_8_1_6_to_8_1_7) ToVersion() Version {
	return Version{
		Major:  8,
		Minor:  1,
		Patch:  7,
		Suffix: "default",
	}
}

// Migrate performs the migration.
func (migration migration_8_1_6_to_8_1_7) Migrate(file *File) (*File, error) {
	newFile := file.Dupe()
	newFile.SetVersion(migration.ToVersion())
	return newFile, nil
}

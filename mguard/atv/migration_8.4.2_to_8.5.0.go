package atv

type migration_8_4_2_to_8_5_0 struct{}

// FromVersion returns the document version the migration start with.
func (_ migration_8_4_2_to_8_5_0) FromVersion() Version {
	return Version{
		Major:  8,
		Minor:  4,
		Patch:  2,
		Suffix: "default",
	}
}

// ToVersion returns the document version the migration ends with.
func (_ migration_8_4_2_to_8_5_0) ToVersion() Version {
	return Version{
		Major:  8,
		Minor:  5,
		Patch:  0,
		Suffix: "default",
	}
}

// Migrate performs the migration.
func (migration migration_8_4_2_to_8_5_0) Migrate(file *File) (*File, error) {
	newFile := file.Dupe()
	newFile.SetVersion(migration.ToVersion())
	return newFile, nil
}

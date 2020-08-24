package atv

type migration_7_6_2_to_8_0_1 struct{}

// FromVersion returns the document version the migration start with.
func (_ migration_7_6_2_to_8_0_1) FromVersion() Version {
	return Version{
		Major:  7,
		Minor:  6,
		Patch:  2,
		Suffix: "default",
	}
}

// ToVersion returns the document version the migration ends with.
func (_ migration_7_6_2_to_8_0_1) ToVersion() Version {
	return Version{
		Major:  8,
		Minor:  0,
		Patch:  1,
		Suffix: "default",
	}
}

// Migrate performs the migration.
func (migration migration_7_6_2_to_8_0_1) Migrate(file *File) (*File, error) {
	newFile := file.Dupe()
	newFile.SetVersion(migration.ToVersion())
	return newFile, nil
}

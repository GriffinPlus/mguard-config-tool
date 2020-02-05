package atv

type migrationProvider interface {
	FromVersion() Version
	ToVersion() Version
	Migrate(file *File) (*File, error)
}

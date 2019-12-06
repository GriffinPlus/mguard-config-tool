package shadow

import "io"

// File represents a Linux shadow file.
type File struct {
	lines []line
}

type line struct {
}

// NewFile returns a new shadow file.
func NewFile() *File {

	file := File{
		lines: []line{},
	}

	return &file
}

// FromReader loads a shadow file from the specified reader.
func FromReader(reader io.Reader) (*File, error) {
	return nil, nil
}

// ToWriter writes the shadow file using the specified writer.
func (file *File) ToWriter(writer io.Writer) error {
	return nil
}

// AddUser adds a new user to the shadow file
func (file *File) AddUser(username string, password string) error {
	return nil
}

// SetPassword sets the password of the specified user
func (file *File) SetPassword(username string, password string) error {
	return nil
}

// VerifyPassword verifys the password of the specified user
func (file *File) VerifyPassword(username string, password string) error {
	return nil
}

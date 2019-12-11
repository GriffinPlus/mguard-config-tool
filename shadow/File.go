package shadow

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// File represents a Linux shadow file.
type File struct {
	lines []*line
}

// NewFile returns a new shadow file.
func NewFile() *File {

	file := File{
		lines: []*line{},
	}

	return &file
}

// FileFromReader loads a shadow file from the specified reader.
func FileFromReader(reader io.Reader) (*File, error) {

	file := NewFile()

	scanner := bufio.NewScanner(reader)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line, err := lineFromString(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("Error in shadow file (line: %d): %s", lineNumber, err)
		}
		file.lines = append(file.lines, line)
	}

	// handle scanner error
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return file, nil
}

// ToWriter writes the shadow file using the specified writer.
func (file *File) ToWriter(writer io.Writer) error {
	_, err := writer.Write([]byte(file.String()))
	return err
}

// Dupe returns a copy of the shadow file.
func (file *File) Dupe() *File {

	buffer := bytes.Buffer{}
	err := file.ToWriter(&buffer)
	if err != nil {
		// should not occur...
		panic("Unexpected error when serializing the shadow file")
	}

	other, err := FileFromReader(&buffer)
	if err != nil {
		// should not occur...
		panic("Unexpected error when deserializing the shadow file")
	}

	return other
}

// AddUser adds a new user to the shadow file.
func (file *File) AddUser(username string, password string) error {

	// ensure that the user does not exist, yet
	for _, line := range file.lines {
		if line.Username == username {
			return fmt.Errorf("The specified user (%s) exists already", username)
		}
	}

	line := &line{Username: username}
	line.SetPassword(password)
	file.lines = append(file.lines, line)
	return nil
}

// SetPassword sets the password of the specified user.
func (file *File) SetPassword(username string, password string) error {

	for _, line := range file.lines {
		if line.Username == username {
			return line.SetPassword(password)
		}
	}

	return fmt.Errorf("The specified user (%s) does not exist", username)
}

// VerifyPassword verifys the password of the specified user.
func (file *File) VerifyPassword(username string, password string) (bool, error) {

	for _, line := range file.lines {
		if line.Username == username {
			return line.VerifyPassword(password)
		}
	}

	return false, fmt.Errorf("The specified user (%s) does not exist", username)
}

// String returns the entire shadow file as a string.
func (file *File) String() string {
	builder := strings.Builder{}
	for _, line := range file.lines {
		builder.WriteString(line.String())
		builder.WriteString("\n")
	}
	return builder.String()
}

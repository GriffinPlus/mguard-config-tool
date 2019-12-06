package shadow

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// File represents a Linux shadow file.
type File struct {
	lines []line
}

// line represents a line in a shadow file
// (also see https://en.wikipedia.org/wiki/Passwd#Shadow_file)
type line struct {
	Username    string // 1) Login name
	Password    string // 2) Salt and hashed password OR a status exception value
	LastChanged string // 3) Days since epoch (Jan 1, 1970) of last password change
	Minimum     string // 4) Days until change allowed
	Maximum     string // 5) Days before change required
	Warn        string // 6) Days warning for expiration
	Inactive    string // 7) Days after no logins before account is locked
	Expire      string // 8) Days since epoch (Jan 1, 1970) when account expires
	Reserved    string // 9) Reserved and unused
}

// lineFieldCount is the expected number of fields in a shadow line
const lineFieldCount = 9

// NewFile returns a new shadow file.
func NewFile() *File {

	file := File{
		lines: []line{},
	}

	return &file
}

// FromReader loads a shadow file from the specified reader.
func FromReader(reader io.Reader) (*File, error) {

	file := NewFile()

	scanner := bufio.NewScanner(reader)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		text := scanner.Text()
		fields := strings.Split(text, ":")
		if len(fields) != 9 {
			return nil, fmt.Errorf(
				"Line %d in shadow file has %d fields, expecting it to have %d fields",
				lineNumber, len(fields), lineFieldCount)
		}
		line := line{
			fields[0], // Username
			fields[1], // Password
			fields[2], // LastChanged
			fields[3], // Minimum
			fields[4], // Maximum
			fields[5], // Warn
			fields[6], // Inactive
			fields[7], // Expire
			fields[8], // Reserved
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

// String returns the entire shadow file as a string.
func (file *File) String() string {
	builder := strings.Builder{}
	for _, line := range file.lines {
		builder.WriteString(line.String())
		builder.WriteString("\n")
	}
	return builder.String()
}

// String returns the line as it occurs in the a shadow file
// (without the newline character at the end).
func (line line) String() string {
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s:%s",
		line.Username,
		line.Password,
		line.LastChanged,
		line.Minimum,
		line.Maximum,
		line.Warn,
		line.Inactive,
		line.Expire,
		line.Reserved)
}

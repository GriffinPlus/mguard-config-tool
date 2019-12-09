package shadow

import (
	"fmt"
	"strings"

	"github.com/tredoe/osutil/user/crypt"
	"github.com/tredoe/osutil/user/crypt/sha512_crypt"
)

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

// lineFieldCount is the expected number of fields in a shadow line.
const lineFieldCount = 9

// lineFromString parses the specified shadow file line.
func lineFromString(s string) (*line, error) {

	fields := strings.Split(s, ":")
	if len(fields) != 9 {
		return nil, fmt.Errorf(
			"Line has %d fields, expecting it to have %d fields",
			len(fields), lineFieldCount)
	}

	line := &line{
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

	return line, nil
}

// String returns the line as it occurs in the a shadow file
// (without the newline character at the end).
func (line *line) String() string {
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

// SetPassword sets the password of the specified user
func (line *line) SetPassword(password string) error {

	if len(password) == 0 {
		line.Password = "!" // disabled account
		return nil
	}

	// generate hash
	c := sha512_crypt.New()
	hash, err := c.Generate([]byte(password), []byte{})
	if err != nil {
		return err
	}

	line.Password = hash
	return nil
}

// VerifyPassword verifys the specified password in clear-text against the hashed
// password in the line.
func (line *line) VerifyPassword(password string) (bool, error) {

	if len(line.Password) == 0 {
		return false, fmt.Errorf("The password field is empty")
	}

	if strings.HasPrefix(line.Password, "!") {
		return false, fmt.Errorf("The account is deactivated")
	}

	// verify hashed password
	c := crypt.NewFromHash(line.Password)
	err := c.Verify(line.Password, []byte(password))
	if err != nil {
		return false, err
	}

	return true, nil
}

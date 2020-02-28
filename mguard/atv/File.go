package atv

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// File represents a mGuard configuration file.
type File struct {
	doc *document
}

// Dupe returns a copy of the ATV document.
func (file *File) Dupe() *File {

	if file == nil {
		return nil
	}

	buffer := bytes.Buffer{}
	err := file.ToWriter(&buffer)
	if err != nil {
		// should not occur...
		panic("Unexpected error when serializing the ATV document")
	}

	other, err := FromReader(&buffer)
	if err != nil {
		// should not occur...
		panic("Unexpected error when deserializing the ATV document")
	}

	return other
}

// String returns a properly formatted string representation of the ATV document.
func (file *File) String() string {

	if file == nil {
		return "<nil>"
	}

	return file.doc.String()
}

// FromFile reads the specified ATV file from disk.
func FromFile(path string) (*File, error) {

	// open the file for reading
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// read the ATV file
	return FromReader(file)
}

// FromReader reads an ATV document from the specified io.Reader.
func FromReader(reader io.Reader) (*File, error) {

	doc, err := documentFromReader(reader)
	if err != nil {
		return nil, err
	}

	// create atv file object
	file := File{doc: doc}

	// ensure that the version pragma exists and is properly formatted
	_, err = file.GetVersion()
	if err != nil {
		return nil, err
	}

	return &file, nil
}

// ToFile saves the ATV document to the specified file.
func (file *File) ToFile(path string) error {

	if file == nil {
		return ErrNilReceiver
	}

	// create directories on the way, if necessary
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}

	// open the file for writing
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// write the ATV document
	return file.ToWriter(f)
}

// ToWriter writes the ATV document to the specified io.Writer.
func (file *File) ToWriter(writer io.Writer) error {

	if file == nil {
		return ErrNilReceiver
	}

	content := file.String()
	_, err := writer.Write([]byte(content))
	return err
}

// GetVersion gets the version of the document.
func (file *File) GetVersion() (Version, error) {

	if file == nil {
		return Version{}, ErrNilReceiver
	}

	versionPragma, err := file.doc.GetPragma("version")
	if err != nil {
		return Version{}, err
	}
	if versionPragma == nil {
		return Version{}, fmt.Errorf("The ATV document does not contain a version pragma")
	}

	versionRegex := regexp.MustCompile(`^([0-9]+)\.([0-9]+)\.([0-9]+)\.(.+)$`)
	matches := versionRegex.FindAllStringSubmatch(versionPragma.Value, -1)
	if matches == nil {
		return Version{}, fmt.Errorf("The ATV document does not contain a properly formatted version number")
	}

	major, _ := strconv.Atoi(matches[0][1])
	minor, _ := strconv.Atoi(matches[0][2])
	patch, _ := strconv.Atoi(matches[0][3])
	version := Version{Major: major, Minor: minor, Patch: patch}
	if len(matches[0]) > 3 {
		version.Suffix = matches[0][4]
	}

	return version, nil
}

// SetVersion sets the version of the document.
// This should only be done after a migration step to keep the document structure and the version number consistent.
func (file *File) SetVersion(version Version) error {

	if file == nil {
		return ErrNilReceiver
	}

	_, err := file.doc.SetPragma("version", version.String())
	return err
}

// GetRowReferences returns all row references recursively.
func (file *File) GetRowReferences() ([]RowRef, error) {

	if file == nil {
		return nil, ErrNilReceiver
	}

	return file.doc.GetRowReferences(), nil
}

// GetRowIDs returns all row ids recursively.
func (file *File) GetRowIDs() ([]RowID, error) {

	if file == nil {
		return nil, ErrNilReceiver
	}

	return file.doc.GetRowIDs(), nil
}

// GetPragma returns the value of the pragma with the specified name.
func (file *File) GetPragma(name string) (*string, error) {

	if file == nil {
		return nil, ErrNilReceiver
	}

	pragma, err := file.doc.GetPragma(name)
	if pragma != nil {
		return &pragma.Value, err
	}

	return nil, nil
}

// SetPragma sets the value of the pragma with the specified name.
func (file *File) SetPragma(name string, value string) error {

	if file == nil {
		return ErrNilReceiver
	}

	_, err := file.doc.SetPragma(name, value)
	return err
}

// GetUUID gets the UUID of the setting with the specified name.
func (file *File) GetUUID(name string) (*UUID, error) {

	if file == nil {
		return nil, ErrNilReceiver
	}

	return file.doc.GetUUID(name)
}

// SetUUID sets the UUID of the setting with the specified name.
func (file *File) SetUUID(name string, uuid UUID) error {

	if file == nil {
		return ErrNilReceiver
	}

	return file.doc.SetUUID(name, uuid)
}

// RemoveUUID removes the UUID of the setting with the specified name.
func (file *File) RemoveUUID(name string) error {

	if file == nil {
		return ErrNilReceiver
	}

	return file.doc.RemoveUUID(name)
}

// GetAccess gets the access modifier of the setting with the specified name.
func (file *File) GetAccess(name string) (*AccessModifier, error) {

	if file == nil {
		return nil, ErrNilReceiver
	}

	return file.doc.GetAccess(name)
}

// SetAccess sets the access modifier of the setting with the specified name.
func (file *File) SetAccess(name string, access AccessModifier) error {

	if file == nil {
		return ErrNilReceiver
	}

	return file.doc.SetAccess(name, access)
}

// RemoveAccess removes the access modifier of the setting with the specified name.
func (file *File) RemoveAccess(name string) error {

	if file == nil {
		return ErrNilReceiver
	}

	return file.doc.RemoveAccess(name)
}

// GetSetting gets the setting with the specified name.
func (file *File) GetSetting(settingName string) (string, error) {

	if file == nil {
		return "", ErrNilReceiver
	}

	setting, err := file.doc.GetSetting(settingName)
	if err != nil {
		return "", err
	}

	return setting.String(), nil
}

// Merge merges all settings from the specified ATV document into the current one.
func (file *File) Merge(other *File) (*File, error) {
	merged, err := file.doc.Merge(other.doc)
	if err != nil {
		return nil, err
	}
	return &File{doc: merged}, nil
}

// MergeSelectively merges the specified settings from the specified ATV document into the current one.
func (file *File) MergeSelectively(other *File, config *MergeConfiguration) (*File, error) {
	merged, err := file.doc.MergeSelectively(other.doc, config)
	if err != nil {
		return nil, err
	}
	return &File{doc: merged}, nil
}

// Migrate migrates the ATV file to the specified version (upwards only).
func (file *File) Migrate(targetVersion Version) (*File, error) {

	if file == nil {
		return nil, ErrNilReceiver
	}

	currentVersion, err := file.GetVersion()
	if err != nil {
		return nil, err
	}

	if currentVersion.Compare(targetVersion) > 0 {
		return nil, fmt.Errorf("Current atv document version (%s) must not be greater than the target version (%s)", currentVersion, targetVersion)
	}

	// migrations in ascending order
	migrations := []migrationProvider{
		migration_7_5_0_to_8_0_2{},
		migration_8_0_2_to_8_1_0{},
		migration_8_1_0_to_8_8_1{},
	}

	// run migrations
	current := file.Dupe()
	for _, migration := range migrations {

		// abort, if the migration would lead to a higher version number than specified
		if migration.ToVersion().Compare(targetVersion) > 0 {
			break
		}

		// get version of the current document
		currentVersion, err = current.GetVersion()
		if err != nil {
			return nil, err
		}

		// migrate
		current, err = migration.Migrate(current)
		if err != nil {
			return nil, err
		}
	}

	// check whether the target version was reached
	currentVersion, err = current.GetVersion()
	if err != nil {
		return nil, err
	}
	if currentVersion.Compare(targetVersion) != 0 {
		return nil, fmt.Errorf("Migration failed, could not reach target version (%s)", targetVersion)
	}

	return current, nil
}

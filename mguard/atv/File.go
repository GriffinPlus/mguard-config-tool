package atv

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/griffinplus/mguard-config-tool/mguard/atv/model"
)

// RowID is the id of a table row in an ATV document.
type RowID string

// RowRef represents a reference to a table row.
type RowRef string

// File represents a mGuard configuration file.
type File struct {
	doc *model.Document
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

	doc, err := model.FromReader(reader)
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
	content := file.String()
	_, err := writer.Write([]byte(content))
	return err
}

// GetVersion gets the version of the document.
func (file *File) GetVersion() (Version, error) {

	versionPragma := file.doc.GetPragma("version")
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

// GetRowReferences returns all row references recursively.
func (file *File) GetRowReferences() []RowRef {

	if file == nil {
		return []RowRef{}
	}

	return file.GetRowReferences()
}

// GetRowIDs returns all row ids recursively.
func (file *File) GetRowIDs() []RowID {

	if file == nil {
		return []RowID{}
	}

	return file.GetRowIDs()
}

// GetPragma returns the value of the pragma with the specified name.
func (file *File) GetPragma(name string) *string {

	if file == nil {
		return nil
	}

	pragma := file.doc.GetPragma(name)
	if pragma != nil {
		copy := string([]byte(pragma.Value)) // avoids referencing string in model
		return &copy
	}

	return nil
}

// SetPragma sets the value of the pragma with the specified name.
func (file *File) SetPragma(name string, value string) {

	if file == nil {
		return
	}

	file.doc.SetPragma(name, value)
}

// Merge merges the specified ATV document into the current one.
func (file *File) Merge(other *File) (*File, error) {
	merged, err := file.doc.Merge(other.doc)
	if err != nil {
		return nil, err
	}
	return &File{doc: merged}, nil
}

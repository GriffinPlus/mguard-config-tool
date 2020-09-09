package ecs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Path of the the openssl executable that is used to encrypt ECS containers
// (can be set to explicitly use it instead of an executable that can be found searching the PATH variable).
var opensslExecutablePath string

// GetOpensslExecutablePath returns the absolute path to the openssl executable.
// If SetOpensslExecutablePath was called, it simply returns the set path.
// If SetOpensslExecutablePath was not called, it tries to find the executable using the PATH variable.
func GetOpensslExecutablePath() (string, error) {

	if len(opensslExecutablePath) > 0 {
		return opensslExecutablePath, nil
	}

	// determine the name of the openssl binary
	var opensslBinaryFilename string
	if runtime.GOOS == "windows" {
		opensslBinaryFilename = "openssl.exe"
	} else {
		opensslBinaryFilename = "openssl"
	}

	// try to look up the openssl binary
	path, err := exec.LookPath(opensslBinaryFilename)
	if err != nil {
		return "", err
	}

	// found executable with that name
	// => make path to it absolute
	path, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return path, nil
}

// SetOpensslExecutablePath explicitly sets the path to the openssl executable.
// Specify an empty string to enable searching for it using the PATH variable.
func SetOpensslExecutablePath(path string) error {

	var err error
	if filepath.IsAbs(path) {
		path = filepath.Clean(path)
	} else {
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) || info.IsDir() {
		return fmt.Errorf("The openssl executable was not found at the specified location (%s)", path)
	}

	opensslExecutablePath = path
	return nil
}

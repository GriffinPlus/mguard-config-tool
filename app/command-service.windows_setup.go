// +build windows

package main

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// setupFilesystem creates the configured directories, if necessary and checks for the existence of configured files.
func (cmd *ServiceCommand) setupFilesystem() error {

	// create all configured directories recursively, if necessary
	dirs := []string{
		cmd.sdcardTemplateDirectory,
		cmd.hotFolderPath,
		cmd.mergedConfigurationDirectory,
		cmd.updatePackageDirectory,
		filepath.Dir(cmd.baseConfigurationPath),
	}
	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			log.Errorf("Creating directory (%s) failed: %v", dir, err)
			return err
		}
	}

	return nil
}

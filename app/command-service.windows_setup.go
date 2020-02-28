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
	}

	if len(cmd.baseConfigurationPath) > 0 {
		dirs = append(dirs, filepath.Dir(cmd.baseConfigurationPath))
	}

	if len(cmd.mergeConfigurationPath) > 0 {
		dirs = append(dirs, filepath.Dir(cmd.mergeConfigurationPath))
	}

	for _, dir := range dirs {
		if len(dir) > 0 {
			err := os.MkdirAll(dir, 0777)
			if err != nil {
				log.Errorf("Creating directory (%s) failed: %v", dir, err)
				return err
			}
		}
	}

	return nil
}

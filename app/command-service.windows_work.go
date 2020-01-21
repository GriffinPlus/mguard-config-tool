// +build windows

package main

import (
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// processFileInHotfolder is called when a new file arrives in the input directory for configuration files.
// It processes .atv/.ecs files, merges them with the common configuration and writes them to the output directory.
func (cmd *ServiceCommand) processFileInHotfolder(path string) error {

	filename := filepath.Base(path)
	filenameWithoutExtension := strings.TrimSuffix(filename, filepath.Ext(filename))

	// load the configuration file (in the hot folder)
	ecs, err := loadConfigurationFile(path)
	if err != nil {
		return err
	}

	// load the base configuration file
	baseEcs, err := loadConfigurationFile(cmd.baseConfigurationPath)
	if err != nil {
		return err
	}

	// merge the base configuration with the loaded configuration
	mergedAtv, err := baseEcs.Atv.Merge(ecs.Atv)
	if err != nil {
		return err
	}

	// keep the base ECS container, but update the configuration
	mergedEcs := baseEcs.Dupe()
	mergedEcs.Atv = mergedAtv
	if err != nil {
		return err
	}

	// write ATV/ECS files containing the merged result
	if len(cmd.mergedConfigurationDirectory) > 0 {

		// write ATV file, if requested
		if cmd.mergedConfigurationsWriteAtv {
			atvFileName := filenameWithoutExtension + ".atv"
			atvFilePath := filepath.Join(cmd.mergedConfigurationDirectory, atvFileName)
			log.Infof("Writing ATV file (%s)...", atvFilePath)
			err = mergedEcs.Atv.ToFile(atvFilePath)
			if err != nil {
				log.Errorf("Writing ATV file (%s) failed: %s", atvFilePath, err)
				return err
			}
		}

		// write ECS file, if requested
		if cmd.mergedConfigurationsWriteEcs {
			ecsFileName := filenameWithoutExtension + ".ecs"
			ecsFilePath := filepath.Join(cmd.mergedConfigurationDirectory, ecsFileName)
			log.Infof("Writing ECS file (%s)...", ecsFilePath)
			err = mergedEcs.ToFile(ecsFilePath)
			if err != nil {
				log.Errorf("Writing ECS file (%s) failed: %s", ecsFilePath, err)
				return err
			}
		}
	}

	return nil
}

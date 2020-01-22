// +build windows

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
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
	} else {
		log.Info("Output directory is not specified. Skipping generation of ATV/ECS files with the merged configuration.")
	}

	// build archive containing the contents of an sdcard that can be used to flash an mGuard with a
	// the defined firmware and load the merged configuration
	if len(cmd.updatePackageDirectory) > 0 {

		// create temporary directory to prepare the package in
		scratchDir, err := ioutil.TempDir("", "sdcard")
		if err != nil {
			log.Errorf("%v", err)
			return err
		}
		defer os.RemoveAll(scratchDir)

		// copy firmware files
		if len(cmd.firmwareDirectory) > 0 {
			src := cmd.firmwareDirectory + string(filepath.Separator)
			dest := filepath.Join(scratchDir, "Firmware") + string(filepath.Separator)
			err := copy.Copy(src, dest)
			if err != nil {
				log.Errorf("Copying firmware files into scratch directory failed: %s", err)
				return err
			}
		} else {
			log.Info("Firmware directory is not specified. Skipping adding firmware to update package.")
		}

		// write ECS container with the merged configuration
		ecsFilePath := filepath.Join(scratchDir, "ECS.tgz")
		err = mergedEcs.ToFile(ecsFilePath)
		if err != nil {
			log.Errorf("Writing ECS file (%s) failed: %s", ecsFilePath, err)
			return err
		}

		// create a package wrapping everything up using zip
		zipPath := filepath.Join(cmd.updatePackageDirectory, filenameWithoutExtension+".zip")
		err = zipFiles(scratchDir, zipPath)
		if err != nil {
			log.Errorf("Creating update package (%s) failed: %s", zipPath, err)
			return err
		}

	} else {
		log.Info("Output directory is not specified. Skipping generation of update package.")
	}

	return nil
}

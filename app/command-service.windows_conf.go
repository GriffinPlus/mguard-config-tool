// +build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type setting struct {
	path         string
	defaultValue interface{}
}

var settingServiceCycleTimeMs = setting{
	"service.cycle_time_ms",
	1000,
}

var settingInputFirmwareDirectory = setting{
	"input.firmware_directory",
	"./firmware",
}

var settingInputBaseConfigurationFile = setting{
	"input.base_configuration_path",
	"./configs/default.tgz",
}

var settingInputWatchedConfigurationDirectory = setting{
	"input.watched_configuration_directory",
	"./input",
}

var settingOutputMergedConfigurationDirectory = setting{
	"output.merged_configuration_directory",
	"./output-ecs",
}

var settingOutputUpdatePackageDirectory = setting{
	"output.update_package_directory",
	"./output-package",
}

var allSettings = []setting{
	settingServiceCycleTimeMs,
	settingInputFirmwareDirectory,
	settingInputBaseConfigurationFile,
	settingInputWatchedConfigurationDirectory,
	settingOutputMergedConfigurationDirectory,
	settingOutputUpdatePackageDirectory,
}

// loadServiceConfiguration loads the service configuration from the specified file.
func (cmd *ServiceCommand) loadServiceConfiguration(path string, createIfNotExist bool) error {

	// set up new viper configuration with default settings
	conf := viper.New()
	for _, setting := range allSettings {
		conf.SetDefault(setting.path, setting.defaultValue)
	}

	// read configuration from file
	log.Debugf("Loading configuration file '%s'...", path)
	basename := filepath.Base(path)
	configName := strings.TrimSuffix(basename, filepath.Ext(basename))
	configDir := filepath.Dir(path) + string(os.PathSeparator)
	conf.SetConfigName(configName)
	conf.SetConfigType("yaml")
	conf.AddConfigPath(configDir)
	err := conf.ReadInConfig()
	if err != nil {
		switch err := err.(type) {
		case viper.ConfigFileNotFoundError:
			log.Debugf("Configuration file '%s' does not exist.", path)
			if createIfNotExist {
				log.Debugf("Creating new configuration file '%s'...", path)
				err := conf.WriteConfigAs(path)
				if err != nil {
					log.Errorf("Saving configuration file '%s' failed: %v", path, err)
					// proceed with default settings...
				}
			}
		default:
			// some severe error, cannot proceed...
			return err
		}
	}

	// configuration is available now
	// => validate settings

	// service: cycle time
	log.Debugf("Setting '%s': '%s'", settingServiceCycleTimeMs.path, conf.GetString(settingServiceCycleTimeMs.path))
	cmd.cycleTime = time.Duration(conf.GetInt32(settingServiceCycleTimeMs.path)) * time.Millisecond
	if cmd.cycleTime <= time.Duration(0) {
		return fmt.Errorf("Invalid value '%s' for setting '%s'", conf.GetString(settingServiceCycleTimeMs.path), settingServiceCycleTimeMs.path)
	}

	// input: firmware directory
	log.Debugf("Setting '%s': '%s'", settingInputFirmwareDirectory.path, conf.GetString(settingInputFirmwareDirectory.path))
	cmd.firmwareDirectory = conf.GetString(settingInputFirmwareDirectory.path)
	path, err = filepath.Abs(cmd.firmwareDirectory)
	if err != nil {
		return err
	}
	cmd.firmwareDirectory = path

	// input: base configuration file
	log.Debugf("Setting '%s': '%s'", settingInputBaseConfigurationFile.path, conf.GetString(settingInputBaseConfigurationFile.path))
	cmd.baseConfigurationPath = conf.GetString(settingInputBaseConfigurationFile.path)
	path, err = filepath.Abs(cmd.baseConfigurationPath)
	if err != nil {
		return err
	}
	cmd.baseConfigurationPath = path

	// input: watched configuration directory
	log.Debugf("Setting '%s': '%s'", settingInputWatchedConfigurationDirectory.path, conf.GetString(settingInputWatchedConfigurationDirectory.path))
	cmd.watchedConfigurationDirectory = conf.GetString(settingInputWatchedConfigurationDirectory.path)
	path, err = filepath.Abs(cmd.watchedConfigurationDirectory)
	if err != nil {
		return err
	}
	cmd.watchedConfigurationDirectory = path

	// output: merged configuration directory
	log.Debugf("Setting '%s': '%s'", settingOutputMergedConfigurationDirectory.path, conf.GetString(settingOutputMergedConfigurationDirectory.path))
	cmd.mergedConfigurationDirectory = conf.GetString(settingOutputMergedConfigurationDirectory.path)
	path, err = filepath.Abs(cmd.mergedConfigurationDirectory)
	if err != nil {
		return err
	}
	cmd.mergedConfigurationDirectory = path

	// output: update package directory
	log.Debugf("Setting '%s': '%s'", settingOutputUpdatePackageDirectory.path, conf.GetString(settingOutputUpdatePackageDirectory.path))
	cmd.updatePackageDirectory = conf.GetString(settingOutputUpdatePackageDirectory.path)
	path, err = filepath.Abs(cmd.updatePackageDirectory)
	if err != nil {
		return err
	}
	cmd.updatePackageDirectory = path

	// log configuration
	log.Info("--- Configuration ---")
	log.Infof("Cycle time:                      %s", cmd.cycleTime)
	log.Infof("Firmware Directory:              %s", cmd.firmwareDirectory)
	log.Infof("Base Configuration File:         %s", cmd.baseConfigurationPath)
	log.Infof("Watched Configuration Directory: %s", cmd.watchedConfigurationDirectory)
	log.Infof("Merged Configuration Directory:  %s", cmd.mergedConfigurationDirectory)
	log.Infof("Update Package Directory:        %s", cmd.updatePackageDirectory)
	log.Info("--- Configuration End ---")

	return nil
}

// +build windows

package main

import (
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type setting struct {
	path         string
	defaultValue interface{}
}

var settingInputFirmwarePath = setting{
	"input.firmware.path",
	"./data/firmware",
}

var settingInputBaseConfigurationPath = setting{
	"input.base_configuration.path",
	"./data/configs/default.tgz",
}

var settingInputHotfolderPath = setting{
	"input.hotfolder.path",
	"./data/input",
}

var settingOutputMergedConfigurationsPath = setting{
	"output.merged_configurations.path",
	"./data/output-merged-configs",
}

var settingOutputMergedConfigurationsWriteAtv = setting{
	"output.merged_configurations.write_atv",
	true,
}

var settingOutputMergedConfigurationsWriteEcs = setting{
	"output.merged_configurations.write_ecs",
	true,
}

var settingOutputUpdatePackagesPath = setting{
	"output.update_packages.path",
	"./data/output-update-packages",
}

var allSettings = []setting{
	settingInputFirmwarePath,
	settingInputBaseConfigurationPath,
	settingInputHotfolderPath,
	settingOutputMergedConfigurationsPath,
	settingOutputMergedConfigurationsWriteAtv,
	settingOutputMergedConfigurationsWriteEcs,
	settingOutputUpdatePackagesPath,
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

	// input: firmware path (must be a directory)
	log.Debugf("Setting '%s': '%s'", settingInputFirmwarePath.path, conf.GetString(settingInputFirmwarePath.path))
	cmd.firmwareDirectory = conf.GetString(settingInputFirmwarePath.path)
	path, err = filepath.Abs(cmd.firmwareDirectory)
	if err != nil {
		return err
	}
	cmd.firmwareDirectory = path

	// input: base configuration file
	log.Debugf("Setting '%s': '%s'", settingInputBaseConfigurationPath.path, conf.GetString(settingInputBaseConfigurationPath.path))
	cmd.baseConfigurationPath = conf.GetString(settingInputBaseConfigurationPath.path)
	path, err = filepath.Abs(cmd.baseConfigurationPath)
	if err != nil {
		return err
	}
	cmd.baseConfigurationPath = path

	// input: hot folder path
	log.Debugf("Setting '%s': '%s'", settingInputHotfolderPath.path, conf.GetString(settingInputHotfolderPath.path))
	cmd.hotFolderPath = conf.GetString(settingInputHotfolderPath.path)
	path, err = filepath.Abs(cmd.hotFolderPath)
	if err != nil {
		return err
	}
	cmd.hotFolderPath = path

	// output: merged configuration directory
	log.Debugf("Setting '%s': '%s'", settingOutputMergedConfigurationsPath.path, conf.GetString(settingOutputMergedConfigurationsPath.path))
	cmd.mergedConfigurationDirectory = conf.GetString(settingOutputMergedConfigurationsPath.path)
	path, err = filepath.Abs(cmd.mergedConfigurationDirectory)
	if err != nil {
		return err
	}
	cmd.mergedConfigurationDirectory = path

	// output: merged configuration directory - write atv
	log.Debugf("Setting '%s': '%s'", settingOutputMergedConfigurationsWriteAtv.path, conf.GetString(settingOutputMergedConfigurationsWriteAtv.path))
	cmd.mergedConfigurationsWriteAtv = conf.GetBool(settingOutputMergedConfigurationsWriteAtv.path)

	// output: merged configuration directory - write ecs
	log.Debugf("Setting '%s': '%s'", settingOutputMergedConfigurationsWriteEcs.path, conf.GetString(settingOutputMergedConfigurationsWriteEcs.path))
	cmd.mergedConfigurationsWriteEcs = conf.GetBool(settingOutputMergedConfigurationsWriteEcs.path)

	// output: update package directory
	log.Debugf("Setting '%s': '%s'", settingOutputUpdatePackagesPath.path, conf.GetString(settingOutputUpdatePackagesPath.path))
	cmd.updatePackageDirectory = conf.GetString(settingOutputUpdatePackagesPath.path)
	path, err = filepath.Abs(cmd.updatePackageDirectory)
	if err != nil {
		return err
	}
	cmd.updatePackageDirectory = path

	// log configuration
	log.Info("--- Configuration ---")
	log.Infof("Firmware Directory:               %v", cmd.firmwareDirectory)
	log.Infof("Base Configuration File:          %v", cmd.baseConfigurationPath)
	log.Infof("Hot folder:                       %v", cmd.hotFolderPath)
	log.Infof("Merged Configuration Directory:   %v", cmd.mergedConfigurationDirectory)
	log.Infof("  - Write ATV:                    %v", cmd.mergedConfigurationsWriteAtv)
	log.Infof("  - Write ECS:                    %v", cmd.mergedConfigurationsWriteEcs)
	log.Infof("Update Package Directory:         %v", cmd.updatePackageDirectory)
	log.Info("--- Configuration End ---")

	return nil
}

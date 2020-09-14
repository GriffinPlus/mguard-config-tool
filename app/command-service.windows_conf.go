// +build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/griffinplus/mguard-config-tool/mguard/certmgr"
	"github.com/griffinplus/mguard-config-tool/mguard/ecs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type setting struct {
	path         string
	defaultValue interface{}
}

var settingCachePath = setting{
	"cache.path",
	"./data/cache",
}

var settingDeviceDatabaseUser = setting{
	"device_database.user",
	"",
}

var settingDeviceDatabasePassword = setting{
	"device_database.password",
	"",
}

var settingInputSdCardTemplatePath = setting{
	"input.sdcard_template.path",
	"./data/sdcard-template",
}

var settingInputBaseConfigurationPath = setting{
	"input.base_configuration.path",
	"./data/configs/default.atv",
}

var settingInputMergeConfigurationPath = setting{
	"input.merge_configuration.path",
	"",
}

var settingInputHotfolderPath = setting{
	"input.hotfolder.path",
	"./data/input",
}

var settingInputPasswordsRoot = setting{
	"input.passwords.root",
	"",
}

var settingInputPasswordsAdmin = setting{
	"input.passwords.admin",
	"",
}

var settingOutputMergedConfigurationsPath = setting{
	"output.merged_configurations.path",
	"./data/output-merged-configs",
}

var settingOutputMergedConfigurationsWriteAtv = setting{
	"output.merged_configurations.write_atv",
	true,
}

var settingOutputMergedConfigurationsWriteUnencryptedEcs = setting{
	"output.merged_configurations.write_unencrypted_ecs",
	true,
}

var settingOutputMergedConfigurationsWriteEncryptedEcs = setting{
	"output.merged_configurations.write_encrypted_ecs",
	false,
}

var settingOutputUpdatePackagesPath = setting{
	"output.update_packages.path",
	"./data/output-update-packages",
}

var settingOutputUpdatePackagesConfiguration = setting{
	"output.update_packages.configuration",
	"encrypted_ecs",
}

var settingOpenSslBinaryPath = setting{
	"tools.openssl.path",
	"",
}

var allSettings = []setting{
	settingCachePath,
	settingDeviceDatabaseUser,
	settingDeviceDatabasePassword,
	settingInputSdCardTemplatePath,
	settingInputBaseConfigurationPath,
	settingInputMergeConfigurationPath,
	settingInputHotfolderPath,
	settingInputPasswordsRoot,
	settingInputPasswordsAdmin,
	settingOutputMergedConfigurationsPath,
	settingOutputMergedConfigurationsWriteAtv,
	settingOutputMergedConfigurationsWriteUnencryptedEcs,
	settingOutputMergedConfigurationsWriteEncryptedEcs,
	settingOutputUpdatePackagesPath,
	settingOutputUpdatePackagesConfiguration,
	settingOpenSslBinaryPath,
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

	// cache: base path (must be a directory)
	log.Debugf("Setting '%s': '%s'", settingCachePath.path, conf.GetString(settingCachePath.path))
	cmd.cacheDirectory = conf.GetString(settingCachePath.path)
	if len(cmd.cacheDirectory) > 0 {
		if filepath.IsAbs(cmd.cacheDirectory) {
			cmd.cacheDirectory = filepath.Clean(cmd.cacheDirectory)
		} else {
			path, err := filepath.Abs(filepath.Join(configDir, cmd.cacheDirectory))
			if err != nil {
				return err
			}
			cmd.cacheDirectory = path
		}
	} else {
		return fmt.Errorf("setting '%s' is not set.", settingCachePath.path)
	}

	// device database: user
	log.Debugf("Setting '%s': '%s'", settingDeviceDatabaseUser.path, conf.GetString(settingDeviceDatabaseUser.path))
	deviceDatabaseUser := conf.GetString(settingDeviceDatabaseUser.path)

	// device database: password
	log.Debugf("Setting '%s': '%s'", settingDeviceDatabasePassword.path, conf.GetString(settingDeviceDatabasePassword.path))
	deviceDatabasePassword := conf.GetString(settingDeviceDatabasePassword.path)

	// input: sdcard template path (must be a directory)
	log.Debugf("Setting '%s': '%s'", settingInputSdCardTemplatePath.path, conf.GetString(settingInputSdCardTemplatePath.path))
	cmd.sdcardTemplateDirectory = conf.GetString(settingInputSdCardTemplatePath.path)
	if len(cmd.sdcardTemplateDirectory) > 0 {
		if filepath.IsAbs(cmd.sdcardTemplateDirectory) {
			cmd.sdcardTemplateDirectory = filepath.Clean(cmd.sdcardTemplateDirectory)
		} else {
			path, err := filepath.Abs(filepath.Join(configDir, cmd.sdcardTemplateDirectory))
			if err != nil {
				return err
			}
			cmd.sdcardTemplateDirectory = path
		}
	} else {
		return fmt.Errorf("setting '%s' is not set.", settingInputSdCardTemplatePath.path)
	}

	// input: base configuration file
	log.Debugf("Setting '%s': '%s'", settingInputBaseConfigurationPath.path, conf.GetString(settingInputBaseConfigurationPath.path))
	cmd.baseConfigurationPath = conf.GetString(settingInputBaseConfigurationPath.path)
	if len(cmd.baseConfigurationPath) > 0 {
		if filepath.IsAbs(cmd.baseConfigurationPath) {
			cmd.baseConfigurationPath = filepath.Clean(cmd.baseConfigurationPath)
		} else {
			path, err := filepath.Abs(filepath.Join(configDir, cmd.baseConfigurationPath))
			if err != nil {
				return err
			}
			cmd.baseConfigurationPath = path
		}
	} else {
		return fmt.Errorf("setting '%s' is not set.", settingInputBaseConfigurationPath.path)
	}

	// input: merge configuration file
	log.Debugf("Setting '%s': '%s'", settingInputMergeConfigurationPath.path, conf.GetString(settingInputMergeConfigurationPath.path))
	cmd.mergeConfigurationPath = conf.GetString(settingInputMergeConfigurationPath.path)
	if len(cmd.mergeConfigurationPath) > 0 { // setting is optional
		if filepath.IsAbs(cmd.mergeConfigurationPath) {
			cmd.mergeConfigurationPath = filepath.Clean(cmd.mergeConfigurationPath)
		} else {
			path, err := filepath.Abs(filepath.Join(configDir, cmd.mergeConfigurationPath))
			if err != nil {
				return err
			}
			cmd.mergeConfigurationPath = path
		}
	}

	// input: hot folder path
	log.Debugf("Setting '%s': '%s'", settingInputHotfolderPath.path, conf.GetString(settingInputHotfolderPath.path))
	cmd.hotFolderPath = conf.GetString(settingInputHotfolderPath.path)
	if len(cmd.hotFolderPath) > 0 {
		if filepath.IsAbs(cmd.hotFolderPath) {
			cmd.hotFolderPath = filepath.Clean(cmd.hotFolderPath)
		} else {
			path, err := filepath.Abs(filepath.Join(configDir, cmd.hotFolderPath))
			if err != nil {
				return err
			}
			cmd.hotFolderPath = path
		}
	} else {
		return fmt.Errorf("setting '%s' is not set", settingInputHotfolderPath.path)
	}

	// input: password for user 'root'
	log.Debugf("Setting '%s': '%s'", settingInputPasswordsRoot.path, conf.GetString(settingInputPasswordsRoot.path))
	cmd.passwordsRoot = conf.GetString(settingInputPasswordsRoot.path)

	// input: password for user 'admin'
	log.Debugf("Setting '%s': '%s'", settingInputPasswordsAdmin.path, conf.GetString(settingInputPasswordsAdmin.path))
	cmd.passwordsAdmin = conf.GetString(settingInputPasswordsAdmin.path)

	// output: merged configuration directory
	log.Debugf("Setting '%s': '%s'", settingOutputMergedConfigurationsPath.path, conf.GetString(settingOutputMergedConfigurationsPath.path))
	cmd.mergedConfigurationDirectory = conf.GetString(settingOutputMergedConfigurationsPath.path)
	if len(cmd.mergedConfigurationDirectory) > 0 { // setting is optional
		if filepath.IsAbs(cmd.mergedConfigurationDirectory) {
			cmd.mergedConfigurationDirectory = filepath.Clean(cmd.mergedConfigurationDirectory)
		} else {
			path, err := filepath.Abs(filepath.Join(configDir, cmd.mergedConfigurationDirectory))
			if err != nil {
				return err
			}
			cmd.mergedConfigurationDirectory = path
		}
	}

	// output: merged configuration directory - write atv
	// Valid: true, false
	log.Debugf("Setting '%s': '%s'", settingOutputMergedConfigurationsWriteAtv.path, conf.GetString(settingOutputMergedConfigurationsWriteAtv.path))
	cmd.mergedConfigurationsWriteAtv = conf.GetBool(settingOutputMergedConfigurationsWriteAtv.path)

	// output: merged configuration directory - write unencrypted ecs
	// Valid: true, false
	log.Debugf("Setting '%s': '%s'", settingOutputMergedConfigurationsWriteUnencryptedEcs.path, conf.GetString(settingOutputMergedConfigurationsWriteUnencryptedEcs.path))
	cmd.mergedConfigurationsWriteUnencryptedEcs = conf.GetBool(settingOutputMergedConfigurationsWriteUnencryptedEcs.path)

	// output: merged configuration directory - write encrypted ecs
	// Valid: true, false
	log.Debugf("Setting '%s': '%s'", settingOutputMergedConfigurationsWriteEncryptedEcs.path, conf.GetString(settingOutputMergedConfigurationsWriteEncryptedEcs.path))
	cmd.mergedConfigurationsWriteEncryptedEcs = conf.GetBool(settingOutputMergedConfigurationsWriteEncryptedEcs.path)

	// output: update package directory
	log.Debugf("Setting '%s': '%s'", settingOutputUpdatePackagesPath.path, conf.GetString(settingOutputUpdatePackagesPath.path))
	cmd.updatePackageDirectory = conf.GetString(settingOutputUpdatePackagesPath.path)
	if len(cmd.updatePackageDirectory) > 0 { // setting is optional
		if filepath.IsAbs(cmd.updatePackageDirectory) {
			cmd.updatePackageDirectory = filepath.Clean(cmd.updatePackageDirectory)
		} else {
			path, err := filepath.Abs(filepath.Join(configDir, cmd.updatePackageDirectory))
			if err != nil {
				return err
			}
			cmd.updatePackageDirectory = path
		}
	}

	// output: update package configuration
	log.Debugf("Setting '%s': '%s'", settingOutputUpdatePackagesConfiguration.path, conf.GetString(settingOutputUpdatePackagesConfiguration.path))
	updatePackageConfiguration := conf.GetString(settingOutputUpdatePackagesConfiguration.path)
	switch updatePackageConfiguration {
	case "atv":
		cmd.updatePackageConfiguration = config_atv
	case "unencrypted_ecs":
		cmd.updatePackageConfiguration = config_unencrypted_ecs
	case "encrypted_ecs":
		cmd.updatePackageConfiguration = config_encrypted_ecs
	default:
		return fmt.Errorf("setting '%s' is invalid (please choose one of the following: 'atv', 'unencrypted_ecs', 'encrypted_ecs')", settingOutputUpdatePackagesConfiguration.path)
	}

	// tools: openssl binary path
	log.Debugf("Setting '%s': '%s'", settingOpenSslBinaryPath.path, conf.GetString(settingOpenSslBinaryPath.path))
	opensslBinaryPath := conf.GetString(settingOpenSslBinaryPath.path)
	if len(opensslBinaryPath) > 0 {

		// the openssl binary path was specified explicitly
		// => tell the ecs container module to use it (the module checks the existence of the executable)

		if filepath.IsAbs(opensslBinaryPath) {
			opensslBinaryPath = filepath.Clean(opensslBinaryPath)
		} else {
			path, err := filepath.Abs(filepath.Join(configDir, opensslBinaryPath))
			if err != nil {
				return err
			}
			opensslBinaryPath = path
		}

		err := ecs.SetOpensslExecutablePath(opensslBinaryPath)
		if err != nil {
			return err
		}
	}

	// let the ecs module determine the openssl executable
	// (searches using the PATH variable, if the path was not configured explicitly)
	opensslBinaryPath, err = ecs.GetOpensslExecutablePath()
	if err != nil {
		return err
	}

	// log configuration
	logtext := strings.Builder{}
	logtext.WriteString(fmt.Sprintf("--- Configuration ---\n"))
	logtext.WriteString(fmt.Sprintf("Cache Directory:                  %s\n", cmd.cacheDirectory))
	logtext.WriteString(fmt.Sprintf("SD Card Template Directory:       %s\n", cmd.sdcardTemplateDirectory))
	logtext.WriteString(fmt.Sprintf("Base Configuration File:          %s\n", cmd.baseConfigurationPath))
	logtext.WriteString(fmt.Sprintf("Merge Configuration File:         %s\n", cmd.mergeConfigurationPath))
	logtext.WriteString(fmt.Sprintf("Hot folder:                       %s\n", cmd.hotFolderPath))
	logtext.WriteString(fmt.Sprintf("Passwords:\n"))
	logtext.WriteString(fmt.Sprintf("  - root:                         %s\n", cmd.passwordsRoot))
	logtext.WriteString(fmt.Sprintf("  - admin:                        %s\n", cmd.passwordsAdmin))
	logtext.WriteString(fmt.Sprintf("Merged Configuration Directory:   %s\n", cmd.mergedConfigurationDirectory))
	logtext.WriteString(fmt.Sprintf("  - Write ATV:                    %v\n", cmd.mergedConfigurationsWriteAtv))
	logtext.WriteString(fmt.Sprintf("  - Write ECS (unencrypted):      %v\n", cmd.mergedConfigurationsWriteUnencryptedEcs))
	logtext.WriteString(fmt.Sprintf("  - Write ECS (encrypted):        %v\n", cmd.mergedConfigurationsWriteEncryptedEcs))
	logtext.WriteString(fmt.Sprintf("Update Package Directory:         %s\n", cmd.updatePackageDirectory))
	logtext.WriteString(fmt.Sprintf("  - Configuration:                %s\n", cmd.updatePackageConfiguration))
	logtext.WriteString(fmt.Sprintf("External Tools:\n"))
	logtext.WriteString(fmt.Sprintf("  - OpenSSL:                      %s\n", opensslBinaryPath))
	logtext.WriteString(fmt.Sprintf("--- Configuration End ---"))
	log.Info(logtext.String())

	// abort, if writing encrypted ecs containers is enabled, but the openssl is not available
	if cmd.mergedConfigurationsWriteEncryptedEcs && len(opensslBinaryPath) == 0 {
		return fmt.Errorf("Writing encrypted ECS containers is enabled, but the OpenSSL executable was not found")
	}

	// initialize the certificate manager
	certificateCacheDirectory := filepath.Join(cmd.cacheDirectory, "certificates")
	cmd.certificateManager = certmgr.NewCertificateManager(certificateCacheDirectory, deviceDatabaseUser, deviceDatabasePassword)

	return nil
}

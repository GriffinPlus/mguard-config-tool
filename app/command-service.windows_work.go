// +build windows

package main

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/griffinplus/mguard-config-tool/mguard/atv"

	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
)

// processFileInHotfolder is called when a new file arrives in the input directory for configuration files.
// It processes .atv/.ecs files, merges them with the common configuration and writes them to the output directory.
func (cmd *ServiceCommand) processFileInHotfolder(path string) error {

	filename := filepath.Base(path)
	filenameWithoutExtension := strings.TrimSuffix(filename, filepath.Ext(filename))
	timestamp := time.Now().Format("20060102150405")
	var err error

	// extract the serial number of the mGuard, if ECS containers should be encrypted
	var deviceCertificate *x509.Certificate
	if cmd.mergedConfigurationsWriteEncryptedEcs {

		// encrypted ECS containers should be generated
		// => the files must bring along the serial number with the file name
		serial, err := getSerialNumberFrommGuardConfigurationFileName(path)
		if err != nil {
			return err
		}
		if serial == nil {
			return fmt.Errorf("The name of the configuration file (%s) does not match the required pattern (<serial>.(atv|ecs|tgz)", path)
		}

		// query the certificate manager for the appropriate device certificate
		deviceCertificate, err = cmd.certificateManager.GetCertificate(*serial)
		if err != nil {
			return err
		}
	}

	// load the merge configuration file
	var mergeConfig *atv.MergeConfiguration
	if len(cmd.mergeConfigurationPath) > 0 {
		mergeConfig, err = atv.LoadMergeConfiguration(cmd.mergeConfigurationPath)
		if err != nil {
			return err
		}
	}

	// load the configuration file (in the hot folder)
	ecs, err := loadConfigurationFile(path)
	if err != nil {
		return err
	}

	// determine the version of the configuration file in the hot folder
	ecsVersion, err := ecs.Atv.GetVersion()
	if err != nil {
		return err
	}

	// load the base configuration file
	baseEcs, err := loadConfigurationFile(cmd.baseConfigurationPath)
	if err != nil {
		return err
	}

	// determine the version of the base configuration file
	baseEcsVersion, err := baseEcs.Atv.GetVersion()
	if err != nil {
		return err
	}

	// ensure that the version of the base configuration file has the same or a higher version than the configuration file in the hot folder
	if baseEcsVersion.Compare(ecsVersion) < 0 {
		return fmt.Errorf(
			"The configuration file (%s, version: %s) must have the same or a higher version than the base configuration file (%s, version: %s)",
			path, ecsVersion,
			cmd.baseConfigurationPath, baseEcsVersion)
	}

	// migrate configuration file in the hot folder to the version of the base configuration file, if necessary
	ecs.Atv, err = ecs.Atv.Migrate(baseEcsVersion)
	if err != nil {
		return err
	}

	// merge the base configuration with the loaded configuration
	mergedAtv, err := baseEcs.Atv.MergeSelectively(ecs.Atv, mergeConfig)
	if err != nil {
		return err
	}

	// keep the base ECS container, but update the configuration
	mergedEcs := baseEcs.Dupe()
	mergedEcs.Atv = mergedAtv
	if err != nil {
		return err
	}

	// set the password for user 'root', if configured
	if len(cmd.passwordsRoot) > 0 {
		mergedEcs.Users.SetPassword("root", cmd.passwordsRoot)
	}

	// set the password for user 'admin', if configured
	if len(cmd.passwordsAdmin) > 0 {
		mergedEcs.Users.SetPassword("admin", cmd.passwordsAdmin)
	}

	// write ATV/ECS files containing the merged result
	if len(cmd.mergedConfigurationDirectory) > 0 {

		// write ATV file, if requested
		if cmd.mergedConfigurationsWriteAtv {

			atvFileName := filenameWithoutExtension + " (" + timestamp + ")" + ".atv"
			atvFilePath := filepath.Join(cmd.mergedConfigurationDirectory, atvFileName)
			log.Infof("Writing ATV file (%s)...", atvFilePath)
			err = mergedEcs.Atv.ToFile(atvFilePath)
			if err != nil {
				log.Errorf("Writing ATV file (%s) failed: %s", atvFilePath, err)
				return err
			}
		}

		// write unencrypted ECS file, if requested
		if cmd.mergedConfigurationsWriteUnencryptedEcs {

			ecsFileName := filenameWithoutExtension + " (" + timestamp + ")" + ".ecs"
			ecsFilePath := filepath.Join(cmd.mergedConfigurationDirectory, ecsFileName)

			log.Infof("Writing unencrypted ECS file (%s)...", ecsFilePath)
			err = mergedEcs.ToFile(ecsFilePath)
			if err != nil {
				log.Errorf("Writing unencrypted ECS file (%s) failed: %s", ecsFilePath, err)
				return err
			}
		}

		// write encrypted ECS file, if requested
		if cmd.mergedConfigurationsWriteEncryptedEcs {

			ecsFileName := filenameWithoutExtension + " (" + timestamp + ")" + ".ecs.p7e"
			ecsFilePath := filepath.Join(cmd.mergedConfigurationDirectory, ecsFileName)

			log.Infof("Writing encrypted ECS file (%s)...", ecsFilePath)
			err := mergedEcs.ToEncryptedFile(ecsFilePath, deviceCertificate)
			if err != nil {
				log.Errorf("Writing encrypted ECS file (%s) failed: %s", ecsFilePath, err)
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

		// copy files from sdcard template (always configured)
		src := cmd.sdcardTemplateDirectory + string(filepath.Separator)
		dest := scratchDir + string(filepath.Separator)
		err = copy.Copy(src, dest)
		if err != nil {
			log.Errorf("Copying sdcard template files into scratch directory failed: %s", err)
			return err
		}

		// process template 'preconfig.sh.tmpl'
		// ------------------------------------------------------------------------------------------------------------

		// parse the template
		templatePath := filepath.Join(scratchDir, "Rescue Config", "preconfig.sh.tmpl")
		renderedTemplatePath := filepath.Join(scratchDir, "Rescue Config", "preconfig.sh")
		template, err := template.ParseFiles(templatePath)
		if err != nil {
			log.Errorf("Parsing preconfig.sh.tmpl failed: %s", err)
			return err
		}

		// initialize dummy template data
		data := struct {
			Atv           bool
			RootPassword  string
			AdminPassword string
		}{
			false,
			"XXX",
			"XXX",
		}

		// replace passwords with real passwords, if an ATV file is used
		// (ECS containers encorporate hashed passwords and don't need them)
		if cmd.updatePackageConfiguration == config_atv {
			data.Atv = true
			data.RootPassword = cmd.passwordsRoot
			data.AdminPassword = cmd.passwordsAdmin
		}

		// render the template
		f, err := os.Create(renderedTemplatePath)
		if err != nil {
			log.Errorf("Opening preconfig.sh failed: %s", err)
			return err
		}
		defer f.Close()
		err = template.Execute(f, data)
		if err != nil {
			log.Errorf("Executing template preconfig.sh.tmpl and writing to file failed: %s", err)
			return err
		}
		f.Close()

		// remove template file from the temporary directory
		err = os.Remove(templatePath)
		if err != nil {
			log.Errorf("Removing preconfig.sh.tmpl failed: %s", err)
			return err
		}

		// write configuration
		// ------------------------------------------------------------------------------------------------------------
		switch cmd.updatePackageConfiguration {

		case config_atv:
			atvFilePath := filepath.Join(scratchDir, "Rescue Config", "preconfig.atv")
			err = mergedEcs.Atv.ToFile(atvFilePath)
			if err != nil {
				log.Errorf("Writing ATV file (%s) failed: %s", atvFilePath, err)
				return err
			}

		case config_unencrypted_ecs:
			ecsFileName := "ECS.tgz"
			ecsFilePath := filepath.Join(scratchDir, ecsFileName)
			err = mergedEcs.ToFile(ecsFilePath)
			if err != nil {
				log.Errorf("Writing unencrypted ECS file (%s) failed: %s", ecsFilePath, err)
				return err
			}

		case config_encrypted_ecs:
			ecsFileName := filenameWithoutExtension + ".ecs.p7e"
			ecsFilePath := filepath.Join(scratchDir, ecsFileName)
			err := mergedEcs.ToEncryptedFile(ecsFilePath, deviceCertificate)
			if err != nil {
				log.Errorf("Writing encrypted ECS file (%s) failed: %s", ecsFilePath, err)
				return err
			}
		}

		// create a package wrapping everything up using zip
		zipPath := filepath.Join(cmd.updatePackageDirectory, filenameWithoutExtension+" ("+timestamp+")"+".zip")
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

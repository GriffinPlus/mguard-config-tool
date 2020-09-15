package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/griffinplus/mguard-config-tool/mguard/certmgr"
	"github.com/integrii/flaggy"
	log "github.com/sirupsen/logrus"
)

// EncryptCommand represents the 'encrypt' subcommand.
type EncryptCommand struct {
	inFilePath     string             // the file to process
	outEcsFilePath string             // the file receiving the conditioned result (ECS container, encrypted)
	serial         string             // serial number of the mGuard to encrypt for
	cacheDirectory string             // path of the directory where certificates are cached
	subcommand     *flaggy.Subcommand // flaggy's subcommand representing the 'encrypt' subcommand
}

// NewConditionCommand creates a new command handling the 'condition' subcommand.
func NewEncryptCommand() *EncryptCommand {
	return &EncryptCommand{}
}

// AddFlaggySubcommand adds the 'encrypt' subcommand to flaggy.
func (cmd *EncryptCommand) AddFlaggySubcommand() *flaggy.Subcommand {

	cmd.subcommand = flaggy.NewSubcommand("encrypt")
	cmd.subcommand.Description = "Encrypt a mGuard configuration for a mGuard with a specific serial number (requires device database access)"
	cmd.subcommand.AddPositionalValue(&cmd.serial, "serial", 1, true, "Serial number of the mGuard")
	cmd.subcommand.String(&cmd.inFilePath, "", "in", "File containing the mGuard configuration to encrypt (ATV format or unencrypted ECS container)")
	cmd.subcommand.String(&cmd.outEcsFilePath, "", "ecs-out", "File receiving the encrypted configuration (ECS container, encrypted, instead of stdout)")
	cmd.subcommand.String(&cmd.cacheDirectory, "", "cache", "Directory where certificates are cached")

	flaggy.AttachSubcommand(cmd.subcommand, 1)

	return cmd.subcommand
}

// IsSubcommandUsed checks whether the 'encrypt' subcommand was used in the command line.
func (cmd *EncryptCommand) IsSubcommandUsed() bool {
	return cmd.subcommand.Used
}

// ValidateArguments checks whether the specified arguments for the 'encrypt' subcommand are valid.
func (cmd *EncryptCommand) ValidateArguments() error {

	// ensure that the specified files exist and are readable
	files := []string{cmd.inFilePath}
	for _, path := range files {
		if len(path) > 0 {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			file.Close()
		}
	}

	return nil
}

// ExecuteCommand performs the actual work of the 'encrypt' subcommand.
func (cmd *EncryptCommand) ExecuteCommand() error {

	fileWritten := false

	// initialize the certificate manager
	// (load credentials from mguard-device-database.yaml)
	certificateManager, err := certmgr.NewCertificateManager(cmd.cacheDirectory, "", "")
	if err != nil {
		return fmt.Errorf("Initializing the certificate manager failed: %v", err)
	}

	// fetch the required device certificate
	deviceCertificate, err := certificateManager.GetCertificate(cmd.serial)
	if err != nil {
		return err
	}

	// load configuration file (can be ATV or ECS)
	// (the configuration is always loaded into an ECS container, missing parts are filled with defaults)
	ecs, err := loadConfigurationFile(cmd.inFilePath)
	if err != nil {
		return err
	}

	// write encrypted ECS container, if requested
	if len(cmd.outEcsFilePath) > 0 {
		fileWritten = true
		log.Infof("Writing encrypted ECS file (%s)...", cmd.outEcsFilePath)
		err := ecs.ToEncryptedFile(cmd.outEcsFilePath, deviceCertificate)
		if err != nil {
			log.Errorf("Writing encrypted ECS file (%s) failed: %s", cmd.outEcsFilePath, err)
			return err
		}
	}

	// write the encrypted ECS container to stdout, if no output file was specified
	if !fileWritten {
		log.Info("Writing encrypted ECS file to stdout...")
		buffer := bytes.Buffer{}
		err := ecs.ToEncryptedWriter(&buffer, deviceCertificate)
		if err != nil {
			return err
		}
		os.Stdout.Write(buffer.Bytes())
	}

	return nil
}

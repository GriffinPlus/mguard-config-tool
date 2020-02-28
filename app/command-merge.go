package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/griffinplus/mguard-config-tool/mguard/atv"

	"github.com/integrii/flaggy"
	log "github.com/sirupsen/logrus"
)

// MergeCommand represents the 'merge' subcommand.
type MergeCommand struct {
	inFilePath1       string             // the first file to merge
	inFilePath2       string             // the second file to merge
	inMergeConfigPath string             // the configuration file controlling the merge process (optional)
	outAtvFilePath    string             // the file receiving the merged result (ATV format)
	outEcsFilePath    string             // the file receiving the merged result (ECS container)
	subcommand        *flaggy.Subcommand // flaggy's subcommand representing the 'merge' subcommand
}

// NewMergeCommand creates a new command handling the 'merge' subcommand.
func NewMergeCommand() *MergeCommand {
	return &MergeCommand{}
}

// AddFlaggySubcommand adds the 'merge' subcommand to flaggy.
func (cmd *MergeCommand) AddFlaggySubcommand() *flaggy.Subcommand {

	cmd.subcommand = flaggy.NewSubcommand("merge")
	cmd.subcommand.Description = "Merge two mGuard configuration files into one"
	cmd.subcommand.AddPositionalValue(&cmd.inFilePath1, "1st-file", 1, true, "First configuration file to merge")
	cmd.subcommand.AddPositionalValue(&cmd.inFilePath2, "2nd-file", 2, true, "Second configuration file to merge")
	cmd.subcommand.String(&cmd.inMergeConfigPath, "", "config", "Merge configuration file")
	cmd.subcommand.String(&cmd.outAtvFilePath, "", "atv-out", "File receiving the merged configuration (ATV format, instead of stdout)")
	cmd.subcommand.String(&cmd.outEcsFilePath, "", "ecs-out", "File receiving the merged configuration (ECS container, instead of stdout)")

	flaggy.AttachSubcommand(cmd.subcommand, 1)

	return cmd.subcommand
}

// IsSubcommandUsed checks whether the 'merge' subcommand was used in the command line.
func (cmd *MergeCommand) IsSubcommandUsed() bool {
	return cmd.subcommand.Used
}

// ValidateArguments checks whether the specified arguments for the 'merge' subcommand are valid.
func (cmd *MergeCommand) ValidateArguments() error {

	// ensure that the specified files exist and are readable
	files := []string{cmd.inFilePath1, cmd.inFilePath2, cmd.inMergeConfigPath}
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

// ExecuteCommand performs the actual work of the 'merge' subcommand.
func (cmd *MergeCommand) ExecuteCommand() error {

	// load the first file (can be ATV or ECS)
	// (the configuration is always loaded into an ECS container, missing parts are filled with defaults)
	ecs1, err := loadConfigurationFile(cmd.inFilePath1)
	if err != nil {
		return err
	}

	// determine the version of the first file
	version1, err := ecs1.Atv.GetVersion()
	if err != nil {
		return err
	}

	// load the second file (can be ATV or ECS)
	// (the configuration is always loaded into an ECS container, missing parts are filled with defaults)
	ecs2, err := loadConfigurationFile(cmd.inFilePath2)
	if err != nil {
		return err
	}

	// determine the version of the second file
	version2, err := ecs2.Atv.GetVersion()
	if err != nil {
		return err
	}

	// load the merge configuration file, if specified
	// (if no configuration file is specified, all settings are merged)
	var mergeConfig *atv.MergeConfiguration
	if len(cmd.inMergeConfigPath) > 0 {
		mergeConfig, err = atv.LoadMergeConfiguration(cmd.inMergeConfigPath)
		if err != nil {
			return err
		}
	}

	// ensure that the first file has the same or a higher version than the second file
	if version1.Compare(version2) < 0 {
		return fmt.Errorf(
			"The first file (%s, version: %s) must have the same or a higher version than the second file (%s, version: %s)",
			cmd.inFilePath1, version1,
			cmd.inFilePath2, version2)
	}

	// migrate second file to the version of the first file, if necessary
	atv2, err := ecs2.Atv.Migrate(version1)
	if err != nil {
		return err
	}

	// merge the configuration stored in both ECS containers
	mergedAtv, err := ecs1.Atv.MergeSelectively(atv2, mergeConfig)
	if err != nil {
		return err
	}

	// keep first ECS container, but update the configuration
	mergedEcs := ecs1.Dupe()
	mergedEcs.Atv = mergedAtv
	if err != nil {
		return err
	}

	// write ATV file, if requested
	fileWritten := false
	if len(cmd.outAtvFilePath) > 0 {
		fileWritten = true
		log.Infof("Writing ATV file (%s)...", cmd.outAtvFilePath)
		err := mergedEcs.Atv.ToFile(cmd.outAtvFilePath)
		if err != nil {
			log.Errorf("Writing ATV file (%s) failed: %s", cmd.outAtvFilePath, err)
			return err
		}
	}

	// write ECS file, if requested
	if len(cmd.outEcsFilePath) > 0 {
		fileWritten = true
		log.Infof("Writing ECS file (%s)...", cmd.outEcsFilePath)
		err := mergedEcs.ToFile(cmd.outEcsFilePath)
		if err != nil {
			log.Errorf("Writing ECS file (%s) failed: %s", cmd.outEcsFilePath, err)
			return err
		}
	}

	// write the ECS container to stdout, if no output file was specified
	if !fileWritten {
		log.Info("Writing ECS file to stdout...")
		buffer := bytes.Buffer{}
		err := mergedEcs.ToWriter(&buffer)
		if err != nil {
			return err
		}
		os.Stdout.Write(buffer.Bytes())
	}

	return nil
}

package main

import (
	"fmt"
	"os"

	"github.com/integrii/flaggy"
	log "github.com/sirupsen/logrus"
)

// ConditionCommand represents the 'show' subcommand.
type ConditionCommand struct {
	inFilePath     string             // the file to process
	outAtvFilePath string             // the file receiving the conditioned result (ATV format)
	outEcsFilePath string             // the file receiving the conditioned result (ECS container)
	subcommand     *flaggy.Subcommand // flaggy's subcommand representing the 'condition' subcommand
}

// NewConditionCommand creates a new command handling the 'condition' subcommand.
func NewConditionCommand() *ConditionCommand {
	return &ConditionCommand{}
}

// AddFlaggySubcommand adds the 'condition' subcommand to flaggy.
func (cmd *ConditionCommand) AddFlaggySubcommand() *flaggy.Subcommand {

	cmd.subcommand = flaggy.NewSubcommand("condition")
	cmd.subcommand.Description = "Condition and/or convert a mGuard configuration file"
	cmd.subcommand.AddPositionalValue(&cmd.inFilePath, "file", 1, true, "File containing the mGuard configuration to condition (ATV format or ECS container)")
	cmd.subcommand.String(&cmd.outAtvFilePath, "", "out-atv-file", "File receiving the conditioned configuration (ATV format)")
	cmd.subcommand.String(&cmd.outEcsFilePath, "", "out-ecs-file", "File receiving the conditioned configuration (ECS container)")

	flaggy.AttachSubcommand(cmd.subcommand, 1)

	return cmd.subcommand
}

// IsSubcommandUsed checks whether the 'condition' subcommand was used in the command line.
func (cmd *ConditionCommand) IsSubcommandUsed() bool {
	return cmd.subcommand.Used
}

// ValidateArguments checks whether the specified arguments for the 'merge' subcommand are valid.
func (cmd *ConditionCommand) ValidateArguments() error {

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

// Execute performs the actual work of the 'condition' subcommand.
func (cmd *ConditionCommand) Execute() error {

	fileWritten := false

	// load configuration file (can be ATV or ECS)
	// (the configuration is always loaded into an ECS container, missing parts are filled with defaults)
	ecs, err := loadConfigurationFile(cmd.inFilePath)
	if err != nil {
		return err
	}

	// write ATV file, if requested
	if len(cmd.outAtvFilePath) > 0 {
		fileWritten = true
		err := ecs.Atv.ToFile(cmd.outAtvFilePath)
		if err != nil {
			log.Errorf("Writing ATV file (%s) failed: %s", cmd.outAtvFilePath, err)
			return err
		}
	}

	// write ECS file, if requested
	if len(cmd.outEcsFilePath) > 0 {
		fileWritten = true
		err := ecs.ToFile(cmd.outEcsFilePath)
		if err != nil {
			log.Errorf("Writing ECS file (%s) failed: %s", cmd.outEcsFilePath, err)
			return err
		}
	}

	// print the configuration in ATV format to stdout,
	// if no output file was specified
	if !fileWritten {
		fmt.Print(ecs.Atv)
	}

	return nil
}

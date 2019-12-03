package main

import (
	"fmt"
	"mguard-config-tool/mguard/atv"
	"os"

	"github.com/integrii/flaggy"
)

// MergeCommand represents the 'merge' subcommand.
type MergeCommand struct {
	inFilePath1           string             // the first file to merge
	inFilePath2           string             // the second file to merge
	outAtvFilePath        string             // the file receiving the merged result (ATV format)
	outEcsFilePath        string             // the file receiving the merged result (ECS container)
	mergeInstructionsPath string             // a file containing instructions on how to merge the files
	subcommand            *flaggy.Subcommand // flaggy's subcommand representing the 'merge' subcommand
}

// NewMergeCommand creates a new command handling the 'merge' subcommand.
func NewMergeCommand() *MergeCommand {
	return &MergeCommand{}
}

// AddFlaggySubcommand adds the 'merge' subcommand to flaggy.
func (cmd *MergeCommand) AddFlaggySubcommand() *flaggy.Subcommand {

	cmd.subcommand = flaggy.NewSubcommand("merge")
	cmd.subcommand.Description = "Merge two mGuard configuration files into one"
	cmd.subcommand.AddPositionalValue(&cmd.inFilePath1, "1st-file", 1, true, "First mGuard configuration file to merge")
	cmd.subcommand.AddPositionalValue(&cmd.inFilePath2, "2nd-file", 2, true, "Second mGuard configuration file to merge")
	cmd.subcommand.String(&cmd.outAtvFilePath, "", "out-atv-file", "File receiving the merged configuration (ATV format)")
	cmd.subcommand.String(&cmd.outEcsFilePath, "", "out-ecs-file", "File receiving the merged configuration (ECS container)")
	cmd.subcommand.String(&cmd.mergeInstructionsPath, "", "merge-instructions-file", "A file defining how to merge the configurations (see manual).")

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
	files := []string{cmd.inFilePath1, cmd.inFilePath2, cmd.mergeInstructionsPath}
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

// Execute performs the actual work of the 'merge' subcommand.
func (cmd *MergeCommand) Execute() error {

	// open first file
	atv1, err := atv.DocumentFromFile(cmd.inFilePath1)
	if err != nil {
		return err
	}

	// open second file
	_, err = atv.DocumentFromFile(cmd.inFilePath2)
	if err != nil {
		return err
	}

	fmt.Printf("atv1: %s", atv1)
	//fmt.Printf("atv2: %s", atv2)

	return nil
}

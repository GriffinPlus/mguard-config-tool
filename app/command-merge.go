package main

import (
	"os"

	"github.com/integrii/flaggy"
)

type MergeCommand struct {
	inFilePath1           string             // the first file to merge
	inFilePath2           string             // the second file to merge
	outFilePath           string             // the file receiving the merged result
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
	cmd.subcommand.AddPositionalValue(&cmd.outFilePath, "out-file", 3, true, "File receiving the merged configuration")
	cmd.subcommand.String(&cmd.mergeInstructionsPath, "i", "merge-instructions-file", "A file defining how to merge the configurations (see manual).")

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
	return nil
}

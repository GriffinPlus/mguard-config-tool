// +build windows

package main

import (
	"github.com/integrii/flaggy"
)

// ServiceCommand represents the 'service' subcommand.
type ServiceCommand struct {
	subcommand          *flaggy.Subcommand // flaggy's subcommand representing the 'service' subcommand
	installSubcommand   *flaggy.Subcommand // flaggy's subcommand representing the 'service install' subcommand
	uninstallSubcommand *flaggy.Subcommand // flaggy's subcommand representing the 'service uninstall' subcommand
	startSubcommand     *flaggy.Subcommand // flaggy's subcommand representing the 'service start' subcommand
	stopSubcommand      *flaggy.Subcommand // flaggy's subcommand representing the 'service stop' subcommand
	debugSubcommand     *flaggy.Subcommand // flaggy's subcommand representing the 'service debug' subcommand
}

// NewServiceCommand creates a new command handling the 'service' subcommand.
func NewServiceCommand() *ServiceCommand {
	return &ServiceCommand{}
}

// AddFlaggySubcommand adds the 'service' subcommand to flaggy.
func (cmd *ServiceCommand) AddFlaggySubcommand() *flaggy.Subcommand {

	cmd.subcommand = flaggy.NewSubcommand("service")
	cmd.subcommand.Description = "Controls the mGuard configuration merging service."

	cmd.installSubcommand = flaggy.NewSubcommand("install")
	cmd.installSubcommand.Description = "Install the windows service."

	cmd.uninstallSubcommand = flaggy.NewSubcommand("uninstall")
	cmd.uninstallSubcommand.Description = "Uninstall the windows service."

	cmd.startSubcommand = flaggy.NewSubcommand("start")
	cmd.startSubcommand.Description = "Start the installed windows service."

	cmd.stopSubcommand = flaggy.NewSubcommand("stop")
	cmd.stopSubcommand.Description = "Stop the installed windows service."

	cmd.debugSubcommand = flaggy.NewSubcommand("debug")
	cmd.debugSubcommand.Description = "Run as a command line application for debugging purposes."

	cmd.subcommand.AttachSubcommand(cmd.installSubcommand, 1)
	cmd.subcommand.AttachSubcommand(cmd.uninstallSubcommand, 1)
	cmd.subcommand.AttachSubcommand(cmd.startSubcommand, 1)
	cmd.subcommand.AttachSubcommand(cmd.stopSubcommand, 1)
	cmd.subcommand.AttachSubcommand(cmd.debugSubcommand, 1)
	flaggy.AttachSubcommand(cmd.subcommand, 1)

	return cmd.subcommand
}

// IsSubcommandUsed checks whether the 'service' subcommand was used in the command line.
func (cmd *ServiceCommand) IsSubcommandUsed() bool {
	return cmd.subcommand.Used
}

// ValidateArguments checks whether the specified arguments for the 'service' subcommand are valid.
func (cmd *ServiceCommand) ValidateArguments() error {
	return nil
}

// Execute performs the actual work of the 'service' subcommand.
func (cmd *ServiceCommand) Execute() error {
	return nil
}

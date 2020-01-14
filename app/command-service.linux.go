// +build linux

package main

import (
	"github.com/integrii/flaggy"
)

// ServiceCommand represents the 'service' subcommand.
type ServiceCommand struct {
}

// NewServiceCommand creates a new command handling the 'service' subcommand.
func NewServiceCommand() *ServiceCommand {
	return &ServiceCommand{}
}

// AddFlaggySubcommand adds the 'service' subcommand to flaggy.
func (cmd *ServiceCommand) AddFlaggySubcommand() *flaggy.Subcommand {
	return nil // do not add subcommand
}

// IsSubcommandUsed checks whether the 'service' subcommand was used in the command line.
func (cmd *ServiceCommand) IsSubcommandUsed() bool {
	return false
}

// ValidateArguments checks whether the specified arguments for the 'service' subcommand are valid.
func (cmd *ServiceCommand) ValidateArguments() error {
	return nil
}

// ExecuteCommand performs the actual work of the 'service' subcommand.
func (cmd *ServiceCommand) ExecuteCommand() error {
	return nil
}

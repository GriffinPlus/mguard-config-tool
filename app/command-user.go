package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/griffinplus/mguard-config-tool/mguard/ecs"
	log "github.com/sirupsen/logrus"

	"github.com/integrii/flaggy"
)

// UserCommand represents the 'user' subcommand.
type UserCommand struct {
	inFilePath               string             // the file to process
	outFilePath              string             // the file receiving the updated ECS container
	username                 string             // login name of the user the operation applys to
	password                 string             // the password to set/verify
	subcommand               *flaggy.Subcommand // flaggy's subcommand representing the 'user' subcommand
	addSubcommand            *flaggy.Subcommand // flaggy's subcommand representing the 'user add' subcommand
	passwordSubcommand       *flaggy.Subcommand // flaggy's subcommand representing the 'user password' subcommand
	passwordSetSubcommand    *flaggy.Subcommand // flaggy's subcommand representing the 'user password set' subcommand
	passwordVerifySubcommand *flaggy.Subcommand // flaggy's subcommand representing the 'user password verify' subcommand
}

// NewUserCommand creates a new command handling the 'user' subcommand.
func NewUserCommand() *UserCommand {
	return &UserCommand{}
}

// AddFlaggySubcommand adds the 'user' subcommand to flaggy.
func (cmd *UserCommand) AddFlaggySubcommand() *flaggy.Subcommand {

	cmd.subcommand = flaggy.NewSubcommand("user")
	cmd.subcommand.Description = "Add user and set/verify user passwords (ECS containers only)"

	cmd.addSubcommand = flaggy.NewSubcommand("add")
	cmd.addSubcommand.Description = "Add a user."
	cmd.addSubcommand.AddPositionalValue(&cmd.username, "username", 1, true, "Login name of the user")
	cmd.addSubcommand.AddPositionalValue(&cmd.password, "password", 2, true, "Password of the user")
	cmd.addSubcommand.String(&cmd.inFilePath, "", "ecs-in", "The ECS container (unencrypted, instead of stdin)")
	cmd.addSubcommand.String(&cmd.outFilePath, "", "ecs-out", "File receiving the updated ECS container (unencrypted, instead of stdout)")

	cmd.passwordSubcommand = flaggy.NewSubcommand("password")
	cmd.passwordSubcommand.Description = "Set or verify the password of a user"

	cmd.passwordSetSubcommand = flaggy.NewSubcommand("set")
	cmd.passwordSetSubcommand.Description = "Set the password a user (ECS containers only)"
	cmd.passwordSetSubcommand.AddPositionalValue(&cmd.username, "username", 1, true, "Login name of the user")
	cmd.passwordSetSubcommand.AddPositionalValue(&cmd.password, "password", 2, true, "Password of the user")
	cmd.passwordSetSubcommand.String(&cmd.inFilePath, "", "ecs-in", "The ECS container (unencrypted, instead of stdin)")
	cmd.passwordSetSubcommand.String(&cmd.outFilePath, "", "ecs-out", "File receiving the updated ECS container (unencrypted, instead of stdout)")

	cmd.passwordVerifySubcommand = flaggy.NewSubcommand("verify")
	cmd.passwordVerifySubcommand.Description = "Verify the password of a user (ECS containers only)"
	cmd.passwordVerifySubcommand.AddPositionalValue(&cmd.username, "username", 1, true, "Login name of the user")
	cmd.passwordVerifySubcommand.AddPositionalValue(&cmd.password, "password", 2, true, "Password of the user")
	cmd.passwordVerifySubcommand.String(&cmd.inFilePath, "", "ecs-in", "The ECS container (unencrypted, instead of stdin)")

	// attach subcommands to flaggy
	cmd.subcommand.AttachSubcommand(cmd.addSubcommand, 1)
	cmd.subcommand.AttachSubcommand(cmd.passwordSubcommand, 1)
	cmd.passwordSubcommand.AttachSubcommand(cmd.passwordSetSubcommand, 1)
	cmd.passwordSubcommand.AttachSubcommand(cmd.passwordVerifySubcommand, 1)
	flaggy.AttachSubcommand(cmd.subcommand, 1)

	return cmd.subcommand
}

// IsSubcommandUsed checks whether the 'user' subcommand was used in the command line.
func (cmd *UserCommand) IsSubcommandUsed() bool {
	return cmd.subcommand.Used
}

// ValidateArguments checks whether the specified arguments for the 'user' subcommand are valid.
func (cmd *UserCommand) ValidateArguments() error {

	// ensure that one of the subcommands is specified
	if !cmd.addSubcommand.Used && !cmd.passwordSubcommand.Used {
		flaggy.ShowHelpAndExit("")
	}

	// ensure that the username is specified (needed for all operations)
	if len(cmd.username) == 0 {
		return fmt.Errorf("The username was not specified, please add '--username <user>' to the command line")
	}

	// ensure that the password is specified (needed for all operations)
	if len(cmd.password) == 0 {
		return fmt.Errorf("The password was not specified, please add '--pass <new-password>' to the command line")
	}

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

// ExecuteCommand performs the actual work of the 'user' subcommand.
func (cmd *UserCommand) ExecuteCommand() error {

	// load configuration file (can be ATV or ECS)
	// (the configuration is always loaded into an ECS container, missing parts are filled with defaults)
	ecs, err := loadConfigurationFile(cmd.inFilePath)
	if err != nil {
		return err
	}

	if cmd.addSubcommand.Used {
		err = cmd.executeAdd(ecs)
	} else if cmd.passwordSetSubcommand.Used {
		err = cmd.executeSet(ecs)
	} else if cmd.passwordVerifySubcommand.Used {
		return cmd.executeVerify(ecs)
	} else {
		panic("Unhandled subcommand")
	}

	// abort, if the operation failed
	if err != nil {
		return err
	}

	// check whether a different ECS file was specified as output and fall back to the
	// input file, if it was not specified
	effectiveOutFilePath := cmd.outFilePath
	if len(effectiveOutFilePath) == 0 {
		effectiveOutFilePath = cmd.inFilePath
	}

	// operation succeeded, write ECS container
	if len(effectiveOutFilePath) > 0 {
		log.Infof("Writing ECS file (%s)...", effectiveOutFilePath)
		err = ecs.ToFile(effectiveOutFilePath)
		if err != nil {
			log.Errorf("Writing ECS file (%s) failed: %s", effectiveOutFilePath, err)
			return err
		}
	} else {
		log.Info("Writing ECS file to stdout...")
		buffer := bytes.Buffer{}
		err := ecs.ToWriter(&buffer)
		if err != nil {
			return err
		}
		os.Stdout.Write(buffer.Bytes())
	}

	return nil
}

// executeAdd performs the actual work of the 'user add' subcommand.
func (cmd *UserCommand) executeAdd(ecs *ecs.Container) error {
	return ecs.Users.AddUser(cmd.username, cmd.password)
}

// executeSet performs the actual work of the 'user password set' subcommand.
func (cmd *UserCommand) executeSet(ecs *ecs.Container) error {
	return ecs.Users.SetPassword(cmd.username, cmd.password)
}

// executeVerify performs the actual work of the 'user password verify' subcommand.
func (cmd *UserCommand) executeVerify(ecs *ecs.Container) error {

	// verify the password
	ok, err := ecs.Users.VerifyPassword(cmd.username, cmd.password)
	if err != nil {
		return err
	}

	if ok {
		log.Infof("The specified password was verified successfully.")
		ExitCode = 0
	} else {
		log.Infof("The specified password does not match the stored password.")
		ExitCode = 1
	}

	return nil
}

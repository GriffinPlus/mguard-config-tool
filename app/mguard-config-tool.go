package main

import (
	"fmt"
	"os"

	"github.com/integrii/flaggy"
	log "github.com/sirupsen/logrus"
)

var (
	buildTime   string = "<not set>"
	version     string = "<not set>"
	fullVersion string = "<not set>"
)

type command interface {
	AddFlaggySubcommand() *flaggy.Subcommand // Adds the subcommand specific settings to flaggy.
	IsSubcommandUsed() bool                  // Determines whether the subcommand was used in the command line.
	ValidateArguments() error                // Validates parsed arguments after parsing has completed.
	ExecuteCommand() error                   // Executes the command.
}

type arguments struct {
	verbose    bool    // true to write additional messages, otherwise false
	subcommand command // subcommand specific arguments
}

// ExitCode represents the exit code the process returns when finishing.
var ExitCode int = 0

func main() {

	// flaggy.DebugMode = true

	// configure logging
	log.SetOutput(os.Stderr)
	log.SetLevel(log.WarnLevel)
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true, DisableLevelTruncation: true, ForceColors: true})

	// parse arguments
	args := parseArgs()

	// log version information
	log.Infof("mguard-config-tool (version: %s)", version)
	log.Debugf("mguard-config-tool (full version: %s)", fullVersion)
	log.Debugf("mguard-config-tool (build time: %s)", buildTime)

	// execute the appropriate handler
	err := args.subcommand.ExecuteCommand()
	if err != nil {
		log.Errorf("Processing command failed: %s", err)
		ExitCode = 1
	}

	// exit with the specified code
	log.Infof("Exiting with code %d.", ExitCode)
	os.Exit(ExitCode)
}

func parseArgs() arguments {

	args := arguments{}

	flaggy.ResetParser()
	flaggy.SetName("mguard-config-tool")

	// add global flags
	flaggy.Bool(&args.verbose, "", "verbose", "Include additional messages that might help when problems occur.")

	// set version shown when explicitly requesting the version with --version
	flaggy.SetVersion(version)

	// add additional text shown before the flags help
	flaggy.DefaultParser.AdditionalHelpPrepend = "A tool for handling ATV/ECS configuration files for the mGuard security router family of the PHOENIX CONTACT Cyber Security AG."

	// add additional text shown after the flags help
	flaggy.DefaultParser.AdditionalHelpAppend = "\n" +
		"------------------------------------------------------------------------------------------------------------------------\n" +
		"Project:      https://github.com/griffinplus/mguard-config-tool\n" +
		"Full Version: " + fullVersion + "\n" +
		"Build Time:   " + buildTime + "\n" +
		"------------------------------------------------------------------------------------------------------------------------"

	flaggy.DefaultParser.ShowHelpWithHFlag = true
	flaggy.DefaultParser.ShowHelpOnUnexpected = true
	flaggy.DefaultParser.ShowVersionWithVersionFlag = true
	flaggy.DefaultParser.SetHelpTemplate(helpTemplate)

	// add subcommands
	subcommands := []command{
		NewUserCommand(),
		NewConditionCommand(),
		NewMergeCommand(),
		NewEncryptCommand(),
		NewServiceCommand(),
	}
	for _, cmd := range subcommands {
		cmd.AddFlaggySubcommand()
	}

	// let flaggy parse the arguments
	flaggy.Parse()

	// let subcommands validate their arguments
	for _, cmd := range subcommands {
		if cmd.IsSubcommandUsed() {
			err := cmd.ValidateArguments()
			if err != nil {
				flaggy.ShowHelpAndExit(fmt.Sprintf("ERROR: %s", err))
			}
		}
	}

	// configure logging
	if args.verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	// assign the appropriate subcommand arguments
	for _, cmd := range subcommands {
		if cmd.IsSubcommandUsed() {
			args.subcommand = cmd
			break
		}
	}

	// if no valid subcommand was specified, show help
	if args.subcommand == nil {
		flaggy.ShowHelpAndExit("No valid subcommand was specified.")
	}

	return args
}

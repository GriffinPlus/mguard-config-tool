// +build windows

package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/integrii/flaggy"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

// ServiceCommand represents the 'service' subcommand.
type ServiceCommand struct {
	serviceName                   string             // name of the service
	serviceConfig                 mgr.Config         // configuration of the service
	elog                          debug.Log          // interface to the windows event log
	configPath                    string             // path of the configuration file
	cycleTime                     time.Duration      // interval between two checking cycles
	firmwareDirectory             string             // path of the directory containing firmware files to put onto the sdcard
	baseConfigurationPath         string             // path of the mguard configuration file to use as base configuration
	watchedConfigurationDirectory string             // path of the directory to watch for atv/ecs files with configurations to merge with the base configuration
	mergedConfigurationDirectory  string             // path of the directory where to store merged mguard configurations
	updatePackageDirectory        string             // path of the directory where to store update packages (for use on an sdcard)
	subcommand                    *flaggy.Subcommand // flaggy's subcommand representing the 'service' subcommand
	installSubcommand             *flaggy.Subcommand // flaggy's subcommand representing the 'service install' subcommand
	uninstallSubcommand           *flaggy.Subcommand // flaggy's subcommand representing the 'service uninstall' subcommand
	startSubcommand               *flaggy.Subcommand // flaggy's subcommand representing the 'service start' subcommand
	stopSubcommand                *flaggy.Subcommand // flaggy's subcommand representing the 'service stop' subcommand
	debugSubcommand               *flaggy.Subcommand // flaggy's subcommand representing the 'service debug' subcommand
}

// NewServiceCommand creates a new command handling the 'service' subcommand.
func NewServiceCommand() *ServiceCommand {
	exePath, _ := exePath()
	defaultConfigPath := filepath.Join(filepath.Dir(exePath), "mguard-config-tool.yaml")
	return &ServiceCommand{
		configPath:  defaultConfigPath,
		cycleTime:   1000 * time.Millisecond,
		serviceName: "mg_cfg_svc",
		serviceConfig: mgr.Config{
			DisplayName:      "Griffin+ mGuard Configuration Merging Service",
			Description:      "Monitors hot-folder for mGuard configuration files (.atv/.ecs) and merges these files with a common parameter set.",
			ErrorControl:     mgr.ErrorNormal,
			StartType:        mgr.StartAutomatic,
			Dependencies:     []string{},
			DelayedAutoStart: false,
		},
	}
}

// AddFlaggySubcommand adds the 'service' subcommand to flaggy.
func (cmd *ServiceCommand) AddFlaggySubcommand() *flaggy.Subcommand {

	cmd.subcommand = flaggy.NewSubcommand("service")
	cmd.subcommand.Description = "Controls the mGuard configuration merging service"

	cmd.installSubcommand = flaggy.NewSubcommand("install")
	cmd.installSubcommand.Description = "Install the windows service"

	cmd.uninstallSubcommand = flaggy.NewSubcommand("uninstall")
	cmd.uninstallSubcommand.Description = "Uninstall the windows service"

	cmd.startSubcommand = flaggy.NewSubcommand("start")
	cmd.startSubcommand.Description = "Start the installed windows service"

	cmd.stopSubcommand = flaggy.NewSubcommand("stop")
	cmd.stopSubcommand.Description = "Stop the installed windows service"

	cmd.debugSubcommand = flaggy.NewSubcommand("debug")
	cmd.debugSubcommand.Description = "Run as a command line application for debugging purposes"

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

// ExecuteCommand performs the actual work of the 'service' subcommand.
func (cmd *ServiceCommand) ExecuteCommand() error {

	// check if running in an interactive session
	isInteractiveSession, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("Failed to determine if we are running in an interactive session: %v", err)
	}

	// run as a service, if no interactive session
	if !isInteractiveSession {
		return cmd.runService(false)
	}

	if cmd.debugSubcommand.Used {
		return cmd.runService(true)
	} else if cmd.installSubcommand.Used {
		err = cmd.installService()
		if err == nil {
			log.Infof("Service %s was installed successfully.", cmd.serviceName)
		}
	} else if cmd.uninstallSubcommand.Used {
		err = cmd.uninstallService()
		if err == nil {
			log.Infof("Service %s was uninstalled successfully.", cmd.serviceName)
		}
	} else if cmd.startSubcommand.Used {
		err = cmd.startService()
		if err == nil {
			log.Infof("Service %s was started successfully.", cmd.serviceName)
		}
	} else if cmd.stopSubcommand.Used {
		err = cmd.controlService(svc.Stop, svc.Stopped)
		if err == nil {
			log.Infof("Service %s was stopped successfully.", cmd.serviceName)
		}
	} else {
		flaggy.ShowHelpAndExit("No command specified.")
	}

	return err
}

// runService runs the service and blocks until it completes.
func (cmd *ServiceCommand) runService(isDebug bool) error {

	// initialize the event log
	var err error
	if isDebug {
		cmd.elog = debug.New(cmd.serviceName)
	} else {
		cmd.elog, err = eventlog.Open(cmd.serviceName)
		if err != nil {
			return err
		}
	}
	defer cmd.elog.Close()

	// load the service configuration file
	err = cmd.loadServiceConfiguration(cmd.configPath, true)
	if err != nil {
		return err
	}

	// create involved directories as configured in the service configuration
	err = cmd.setupFilesystem()
	if err != nil {
		return err
	}

	// ensure that the specified base configuration file is valid
	_, err = loadConfigurationFile(cmd.baseConfigurationPath)
	if err != nil {
		log.Errorf("Loading base configuration file failed: %v", err)
		return err
	}

	// start the service
	cmd.elog.Info(1, fmt.Sprintf("Starting %s service", cmd.serviceName))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(cmd.serviceName, cmd)
	if err != nil {
		cmd.elog.Error(1, fmt.Sprintf("%s service failed: %v", cmd.serviceName, err))
		return err
	}

	// the service has stopped
	cmd.elog.Info(1, fmt.Sprintf("%s service stopped", cmd.serviceName))
	return nil
}

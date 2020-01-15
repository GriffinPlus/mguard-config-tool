// +build windows

package main

import (
	"fmt"
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
	serviceName         string             // name of the service
	serviceConfig       mgr.Config         // configuration of the service
	elog                debug.Log          // interface to the windows event log
	cycleTime           time.Duration      // interval between two checking cycles
	subcommand          *flaggy.Subcommand // flaggy's subcommand representing the 'service' subcommand
	installSubcommand   *flaggy.Subcommand // flaggy's subcommand representing the 'service install' subcommand
	uninstallSubcommand *flaggy.Subcommand // flaggy's subcommand representing the 'service uninstall' subcommand
	startSubcommand     *flaggy.Subcommand // flaggy's subcommand representing the 'service start' subcommand
	stopSubcommand      *flaggy.Subcommand // flaggy's subcommand representing the 'service stop' subcommand
	debugSubcommand     *flaggy.Subcommand // flaggy's subcommand representing the 'service debug' subcommand
}

// NewServiceCommand creates a new command handling the 'service' subcommand.
func NewServiceCommand() *ServiceCommand {
	return &ServiceCommand{
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

// Execute is the handler for commands coming in from the service control manager.
func (cmd *ServiceCommand) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	tick := time.Tick(cmd.cycleTime)
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	cmd.doWork()
loop:
	for {
		select {
		case <-tick:
			cmd.doWork()
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				break loop
			default:
				cmd.elog.Error(1, fmt.Sprintf("Unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

// doWork checks whether there are .atv/.ecs files in the input directory, merges them with the common configuration
// and writes them to the output directory.
func (cmd *ServiceCommand) doWork() error {
	return nil
}

// installService registers the service with the service control manager.
func (cmd *ServiceCommand) installService() error {

	exepath, err := exePath()
	if err != nil {
		return err
	}

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(cmd.serviceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("Service %s exists already", cmd.serviceName)
	}

	s, err = m.CreateService(cmd.serviceName, exepath, cmd.serviceConfig, "service")
	if err != nil {
		return err
	}
	defer s.Close()

	err = eventlog.InstallAsEventCreate(cmd.serviceName, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %v", err)
	}

	return nil
}

// uninstallService deregisters the service with the service control manager.
func (cmd *ServiceCommand) uninstallService() error {

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(cmd.serviceName)
	if err != nil {
		return fmt.Errorf("Service %s is not installed", cmd.serviceName)
	}
	defer s.Close()

	err = s.Delete()
	if err != nil {
		return err
	}

	err = eventlog.Remove(cmd.serviceName)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %v", err)
	}

	return nil
}

// startService starts the registered windows service.
func (cmd *ServiceCommand) startService() error {

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(cmd.serviceName)
	if err != nil {
		return fmt.Errorf("Could not access service: %v", err)
	}
	defer s.Close()

	err = s.Start()
	if err != nil {
		return fmt.Errorf("Could not start service: %v", err)
	}

	return nil
}

// controlService send a control command to the running windows service.
func (cmd *ServiceCommand) controlService(c svc.Cmd, to svc.State) error {

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(cmd.serviceName)
	if err != nil {
		return fmt.Errorf("Could not access service: %v", err)
	}
	defer s.Close()

	status, err := s.Control(c)
	if err != nil {
		return fmt.Errorf("Could not send control=%d: %v", c, err)
	}

	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("Timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("Could not retrieve service status: %v", err)
		}
	}

	return nil
}

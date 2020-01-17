// +build windows

package main

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

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

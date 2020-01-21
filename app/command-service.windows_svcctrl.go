// +build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

// Execute is the handler for commands coming in from the service control manager.
func (cmd *ServiceCommand) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {

	// signal that the service is starting
	changes <- svc.Status{State: svc.StartPending}

	// initialize file system watcher for input directory (hot folder)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Errorf("%v", err)
		changes <- svc.Status{State: svc.StopPending}
		return
	}
	defer watcher.Close()

	// start watching hot folder
	err = watcher.Add(cmd.hotFolderPath)
	if err != nil {
		log.Errorf("%v", err)
		changes <- svc.Status{State: svc.StopPending}
		return
	}

	// signal that the service is running now
	changes <- svc.Status{
		State:   svc.Running,
		Accepts: svc.AcceptStop | svc.AcceptShutdown,
	}

	// populate list of files in hot folder to start with
	filesInHotFolder := make(map[string]time.Time)
	err = filepath.Walk(cmd.hotFolderPath, func(path string, info os.FileInfo, err error) error {
		path, err = filepath.Abs(path)
		if err == nil {
			isConfFile, err := isPossiblemGuardConfigurationFile(path)
			if err == nil && isConfFile {
				log.Debugf("Found possible configuration file: %s", path)
				filesInHotFolder[path] = time.Now()
			}
		}
		return nil
	})
	if err != nil {
		log.Errorf("%v", err)
		changes <- svc.Status{State: svc.StopPending}
		return
	}

	// set up timer that triggers processing files in the hot folder
	hotFolderDetectedFileCooldown := 5 * time.Second          // time after which a file is considered "stable"
	hotFolderProcessingCycle := 1 * time.Second               // time after which the files in the hot folder are checked
	hotFolderTimer := time.NewTimer(hotFolderProcessingCycle) // timer triggering processing files in the hot folder
	hotFolderTimerRunning := true

loop:
	for {
		select {

		// handle service control commands
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

		// handle file system watcher events
		case event, ok := <-watcher.Events:

			if !ok {
				continue
			}

			log.Debugf("fs event: %v", event)

			if (event.Op & fsnotify.Create) == fsnotify.Create {
				log.Debugf("Created file: %s", event.Name)
				path, err := filepath.Abs(event.Name)
				if err == nil {
					isConfFile, err := isPossiblemGuardConfigurationFile(path)
					if err == nil && isConfFile {
						filesInHotFolder[path] = time.Now()
						if !hotFolderTimerRunning {
							hotFolderTimer.Reset(hotFolderProcessingCycle)
							hotFolderTimerRunning = true
						}
					}
				}
			}

			if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {

				log.Debugf("Removed/Renamed file: %s", event.Name)
				path, err := filepath.Abs(event.Name)
				if err == nil {
					delete(filesInHotFolder, path)
				}
				if len(filesInHotFolder) == 0 {
					if hotFolderTimerRunning {
						if !hotFolderTimer.Stop() {
							<-hotFolderTimer.C
						}
						hotFolderTimerRunning = false
					}
				}
			}

			if (event.Op & fsnotify.Write) == fsnotify.Write {
				log.Debugf("Modified file: %s", event.Name)
				path, err := filepath.Abs(event.Name)
				if err == nil {
					isConfFile, err := isPossiblemGuardConfigurationFile(path)
					if err == nil && isConfFile {
						filesInHotFolder[path] = time.Now()
						if !hotFolderTimerRunning {
							hotFolderTimer.Reset(hotFolderProcessingCycle)
							hotFolderTimerRunning = true
						}
					}
				}
			}

		// handle file system watcher errors
		case err, ok := <-watcher.Errors:
			if !ok {
				continue
			}
			log.Errorf("watcher error: %v", err)

		// process files in the hot folder that did not change within the specified time
		case <-hotFolderTimer.C:
			for file, lastwritten := range filesInHotFolder {
				if time.Since(lastwritten) > hotFolderDetectedFileCooldown {
					path, _ := filepath.Abs(file)
					err = cmd.processFileInHotfolder(path)
					if err == nil {
						log.Infof("Processing file '%s' succeeded.", path)
						err = os.Remove(path)
						if err != nil {
							log.Errorf("%v", err)
						} else {
							log.Debugf("Removing file '%s' succeeded.", path)
						}
					} else {
						newPath := path + ".err"
						log.Errorf("Processing file '%s' failed, renaming it to '%s'.", path, newPath)
						err = os.Rename(path, newPath)
						if err != nil {
							log.Errorf("Renaming file '%s' to '%s' failed: %v", path, newPath, err)
						}
					}
					delete(filesInHotFolder, path)
				}
			}
			if len(filesInHotFolder) > 0 {
				hotFolderTimer.Reset(hotFolderProcessingCycle)
				hotFolderTimerRunning = true
			} else {
				hotFolderTimerRunning = false
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

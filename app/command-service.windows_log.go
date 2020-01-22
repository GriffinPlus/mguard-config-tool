// +build windows

package main

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

// WindowsEventLogAdapter is a log adapter forwarding writting log messages to the windows event log.
type WindowsEventLogAdapter struct {
	serviceName string
	isDebug     bool
	elog        debug.Log
}

// NewWindowsEventLogAdapter creates a new windows log adapter.
func NewWindowsEventLogAdapter(serviceName string, isDebug bool) *WindowsEventLogAdapter {
	return &WindowsEventLogAdapter{
		serviceName: serviceName,
		isDebug:     isDebug,
	}
}

// Open initializes the windows log adapter.
func (adapter *WindowsEventLogAdapter) Open() error {
	if adapter.elog == nil {
		if adapter.isDebug {
			adapter.elog = debug.New(adapter.serviceName)
		} else {
			log, err := eventlog.Open(adapter.serviceName)
			if err != nil {
				return err
			}
			adapter.elog = log
		}
		log.AddHook(adapter)
	}

	return nil
}

// Close shuts the windows log adapter down.
func (adapter *WindowsEventLogAdapter) Close() {
	if adapter != nil {
		if adapter.elog != nil {
			adapter.elog.Close()
			adapter.elog = nil
		}
	}
}

// Levels returns the log levels the log hook is called for.
func (adapter *WindowsEventLogAdapter) Levels() []log.Level {
	return log.AllLevels
}

// Fire processes the written log entry.
func (adapter *WindowsEventLogAdapter) Fire(entry *log.Entry) error {
	switch entry.Level {
	case log.DebugLevel:
		return adapter.elog.Info(1, entry.Message)
	case log.InfoLevel:
		return adapter.elog.Info(1, entry.Message)
	case log.WarnLevel:
		return adapter.elog.Warning(1, entry.Message)
	case log.ErrorLevel:
		return adapter.elog.Error(1, entry.Message)
	case log.FatalLevel:
		return adapter.elog.Error(1, entry.Message)
	default:
		return adapter.elog.Info(1, entry.Message)
	}
}

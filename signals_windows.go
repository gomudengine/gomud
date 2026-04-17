//go:build windows

package main

import (
	"os"
	"os/signal"
	"syscall"
)

func registerShutdownSignals(sigCh chan os.Signal) {
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
}

func startCopyoverSignalHandler() {}

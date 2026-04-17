//go:build !windows

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

func registerShutdownSignals(sigCh chan os.Signal) {
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
}

func startCopyoverSignalHandler() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGUSR1)
	go func() {
		for range sigCh {
			mudlog.Info("SIGUSR1 received, initiating copyover")
			if err := triggerCopyover(); err != nil {
				mudlog.Error("copyover failed", "error", err)
			}
		}
	}()
}

//go:build !windows

package main

import (
	"os/signal"
	"syscall"

	"golang.org/x/crypto/ssh"
)

// watchTerminalResize monitors terminal resize on Unix using SIGWINCH
func watchTerminalResize(session *ssh.Session) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)
	go func() {
		for range sigChan {
			sendWindowChange(session)
		}
	}()
}

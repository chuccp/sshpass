package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHClient creates an SSH client connection
func SSHClient(config *Config) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	// use private key authentication first
	if config.KeyPath != "" {
		key, err := os.ReadFile(config.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// add password authentication if available (as fallback or primary)
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method provided (password or key required)")
	}

	// set host key verification callback
	var hostKeyCallback ssh.HostKeyCallback
	if config.StrictHostKey {
		// use known_hosts file for verification
		knownHostsPath := getKnownHostsPath()
		callback, err := knownhosts.New(knownHostsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read known_hosts file (%s): %w\nhint: connect to the server manually first to add the host key", knownHostsPath, err)
		}
		hostKeyCallback = callback
	} else {
		// ignore host key verification (default)
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
	}

	address := fmt.Sprintf("%s:%s", config.Host, config.Port)
	client, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	return client, nil
}

// getKnownHostsPath returns the known_hosts file path
func getKnownHostsPath() string {
	// use environment variable if set
	if path := os.Getenv("SSH_KNOWN_HOSTS"); path != "" {
		return path
	}

	// Windows default path: %USERPROFILE%\.ssh\known_hosts
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.Getenv("USERPROFILE")
	}
	return filepath.Join(homeDir, ".ssh", "known_hosts")
}

// runShell starts an interactive shell
func runShell(client *ssh.Client) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// set terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// request pseudo-terminal
	if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
		return fmt.Errorf("failed to request terminal: %w", err)
	}

	// set standard I/O
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	// start remote shell
	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	return session.Wait()
}

// executeCommand executes a single command
func executeCommand(client *ssh.Client, command string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	return session.Run(command)
}
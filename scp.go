package main

import (
	"fmt"
	"strings"
)

// runSCP executes the scp command (file transfer over SSH)
func runSCP(config *Config, args []string) error {
	// establish SSH connection
	client, err := SSHClient(config)
	if err != nil {
		return err
	}
	defer client.Close()

	// parse scp arguments to determine source and target
	var localFile, remotePath string
	var isUpload bool
	var nonFlagArgs []string

	// collect non-flag arguments
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			nonFlagArgs = append(nonFlagArgs, arg)
		}
	}

	// parse source and target
	for _, arg := range nonFlagArgs {
		if strings.Contains(arg, "@") && strings.Contains(arg, ":") {
			// remote path user@host:path (supports IPv6)
			_, _, remotePath = parseUserHostPath(arg)
		} else if arg != "scp" {
			// local file
			if localFile == "" {
				localFile = arg
			}
		}
	}

	// determine upload or download: remote path in last argument means upload
	for i, arg := range nonFlagArgs {
		if strings.Contains(arg, "@") && strings.Contains(arg, ":") {
			isUpload = (i == len(nonFlagArgs)-1)
			break
		}
	}

	if localFile == "" || remotePath == "" {
		return fmt.Errorf("failed to parse scp arguments: %v", args)
	}

	// clean remote path (handle Git Bash // prefix and path conversion)
	remotePath = cleanRemotePath(remotePath)

	if isUpload {
		return uploadFile(client, localFile, remotePath)
	}
	return downloadFile(client, remotePath, localFile)
}

// runRsync executes rsync command (file sync over SSH)
func runRsync(config *Config, args []string) error {
	// simple implementation: parse source and target
	var src, dst string
	for _, arg := range args {
		if strings.Contains(arg, "@") && strings.Contains(arg, ":") {
			if src == "" {
				src = arg
			} else {
				dst = arg
			}
		} else if !strings.HasPrefix(arg, "-") {
			if src == "" {
				src = arg
			} else {
				dst = arg
			}
		}
	}

	if src == "" || dst == "" {
		return fmt.Errorf("failed to parse rsync arguments")
	}

	// establish SSH connection
	client, err := SSHClient(config)
	if err != nil {
		return err
	}
	defer client.Close()

	// determine direction
	if strings.Contains(src, "@") {
		// remote to local (download)
		_, _, remotePath := parseUserHostPath(src)
		return downloadFile(client, cleanRemotePath(remotePath), dst)
	}
	// local to remote (upload)
	_, _, remotePath := parseUserHostPath(dst)
	return uploadFile(client, src, cleanRemotePath(remotePath))
}

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	// command line arguments
	configFile := flag.String("f", "", "password file or config file path")
	host := flag.String("h", "", "host address")
	user := flag.String("u", "root", "username")
	password := flag.String("p", "", "password")
	port := flag.String("P", "22", "port")
	keyPath := flag.String("i", "", "private key file path")
	command := flag.String("c", "", "command to execute")
	localPath := flag.String("local", "", "local file path (for upload/download)")
	remotePath := flag.String("remote", "", "remote file path (for upload/download)")
	download := flag.Bool("d", false, "download mode (remote to local)")
	useEnv := flag.Bool("e", false, "read password from environment variable SSHPASS")
	strictHostKey := flag.Bool("k", false, "enable strict host key verification")
	showVersion := flag.Bool("v", false, "show version")
	flag.Parse()

	// display version
	if *showVersion {
		printVersion()
		return
	}

	var config *Config
	var err error
	var cmdToRun string

	// get remaining arguments (for sshpass-style commands)
	remainingArgs := flag.Args()

	// get password: priority -p > config file > password file > -e > SSHPASS
	pass := *password
	if pass == "" && *configFile != "" {
		// try parsing as config file first
		config, err = parseConfigFile(*configFile)
		if err != nil {
			// not a config file, try as single-line password file
			pass, err = readPasswordFile(*configFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			// config file parsed successfully, set StrictHostKey from command line
			config.StrictHostKey = *strictHostKey
		}
	}
	if pass == "" && *useEnv {
		pass = getEnvPassword()
	}

	// detect command type
	cmdType := detectCommandType(remainingArgs)

	// handle based on command type
	switch cmdType {
	case CommandSCP:
		config, scpArgs := parseSCPArgs(remainingArgs)
		if pass != "" {
			config.Password = pass
		}
		if *keyPath != "" {
			config.KeyPath = *keyPath
		}
		if *host != "" {
			config.Host = *host
		}
		if *user != "" {
			config.User = *user
		}
		if *port != "" && *port != "22" {
			config.Port = *port
		}
		config.StrictHostKey = *strictHostKey
		err = runSCP(config, scpArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SCP failed: %v\n", err)
			os.Exit(1)
		}
		return

	case CommandRsync:
		config, rsyncArgs := parseRsyncArgs(remainingArgs)
		if pass != "" {
			config.Password = pass
		}
		if *keyPath != "" {
			config.KeyPath = *keyPath
		}
		if *host != "" {
			config.Host = *host
		}
		if *user != "" {
			config.User = *user
		}
		if *port != "" && *port != "22" {
			config.Port = *port
		}
		config.StrictHostKey = *strictHostKey
		err = runRsync(config, rsyncArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Rsync failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// SSH command handling
	if config == nil {
		if len(remainingArgs) > 0 && (pass != "" || *keyPath != "") {
			// sshpass style: -p password or -i keyfile ssh user@host [command]
			config, cmdToRun = parseSSHArgs(remainingArgs)
			if pass != "" {
				config.Password = pass
			}
			if *keyPath != "" {
				config.KeyPath = *keyPath
			}
			if *host != "" {
				config.Host = *host
			}
			if *user != "" {
				config.User = *user
			}
			if *port != "" && *port != "22" {
				config.Port = *port
			}
			config.StrictHostKey = *strictHostKey
		} else if *host != "" && (pass != "" || *keyPath != "") {
			// read from command line arguments (including file transfer mode)
			config = &Config{
				Host:          *host,
				User:          *user,
				Password:      pass,
				Port:          *port,
				KeyPath:       *keyPath,
				StrictHostKey: *strictHostKey,
			}
		} else {
			printUsage()
			os.Exit(1)
		}
	}

	// validate config
	if config.Host == "" {
		fmt.Fprintf(os.Stderr, "Error: host address not specified\n")
		os.Exit(1)
	}
	if config.Password == "" && config.KeyPath == "" {
		fmt.Fprintf(os.Stderr, "Error: no authentication method provided (password or key required)\n")
		os.Exit(1)
	}

	// establish SSH connection
	client, err := SSHClient(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "SSH connection failed: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// file transfer
	if *localPath != "" && *remotePath != "" {
		if *download {
			fmt.Printf("Downloading %s -> %s...\n", *remotePath, *localPath)
			err = downloadFile(client, *remotePath, *localPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Download successful!")
		} else {
			fmt.Printf("Uploading %s -> %s...\n", *localPath, *remotePath)
			err = uploadFile(client, *localPath, *remotePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Upload failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Upload successful!")
		}
		return
	} else if *localPath != "" || *remotePath != "" {
		fmt.Fprintf(os.Stderr, "Error: file transfer requires both -local and -remote arguments\n")
		os.Exit(1)
	}

	// execute command
	cmd := *command
	if cmd == "" {
		cmd = cmdToRun
	}

	if cmd != "" {
		err = executeCommand(client, cmd)
	} else {
		err = runShell(client)
	}

	if err != nil {
		if !strings.Contains(err.Error(), "closed network connection") &&
			!strings.Contains(err.Error(), "use of closed network connection") {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}
}

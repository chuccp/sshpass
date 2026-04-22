package main

import (
	"fmt"
	"strings"
)

// CommandType represents the command type
type CommandType int

const (
	CommandSSH CommandType = iota
	CommandSCP
	CommandRsync
)

// ParsedCommand represents the parsed command
type ParsedCommand struct {
	Config  *Config
	Type    CommandType
	Command string   // command to execute
	SCPArgs []string // scp arguments
}

// parseUserHostPath parses user@host:path format, supporting IPv6
// returns user, host, path
func parseUserHostPath(arg string) (user, host, remotePath string) {
	atIdx := strings.Index(arg, "@")
	if atIdx <= 0 {
		return "", "", ""
	}
	user = arg[:atIdx]
	remainder := arg[atIdx+1:]

	// check if IPv6 address (starts with [)
	if strings.HasPrefix(remainder, "[") {
		// IPv6 format: [::1]:path or [2001:db8::1]:path
		closeBracket := strings.Index(remainder, "]")
		if closeBracket > 0 {
			host = remainder[:closeBracket+1] // including square brackets
			// check if there is a path after ]:
			if closeBracket+1 < len(remainder) && remainder[closeBracket+1] == ':' {
				remotePath = remainder[closeBracket+2:]
			}
		}
	} else {
		// IPv4 or hostname: host:path
		colonIdx := strings.Index(remainder, ":")
		if colonIdx > 0 {
			host = remainder[:colonIdx]
			remotePath = remainder[colonIdx+1:]
		} else {
			host = remainder
		}
	}
	return user, host, remotePath
}

// parseSSHArgs parses ssh-style arguments (user@host or -p port user@host)
func parseSSHArgs(args []string) (*Config, string) {
	config := &Config{
		User: "root",
		Port: "22",
	}
	var command string

	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "ssh" {
			// skip the ssh command itself
			i++
			continue
		}
		if arg == "-p" && i+1 < len(args) {
			config.Port = args[i+1]
			i += 2
			continue
		}
		if arg == "-i" && i+1 < len(args) {
			config.KeyPath = args[i+1]
			i += 2
			continue
		}
		if arg == "-o" && i+1 < len(args) {
			// skip ssh options like StrictHostKeyChecking=no
			i += 2
			continue
		}
		if strings.Contains(arg, "@") {
			// user@host format (supports IPv6)
			parts := strings.SplitN(arg, "@", 2)
			if len(parts) == 2 {
				config.User = parts[0]
				config.Host = parts[1]
			}
			i++
			continue
		}
		// remaining args as command
		if config.Host != "" {
			command = strings.Join(args[i:], " ")
			break
		}
		i++
	}

	return config, command
}

// parseSCPArgs parses scp command arguments
func parseSCPArgs(args []string) (*Config, []string) {
	config := &Config{
		User: "root",
		Port: "22",
	}
	var scpArgs []string

	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "scp" {
			i++
			continue
		}
		if arg == "-P" && i+1 < len(args) {
			// scp uses uppercase -P for port
			config.Port = args[i+1]
			scpArgs = append(scpArgs, "-P", args[i+1])
			i += 2
			continue
		}
		if arg == "-i" && i+1 < len(args) {
			config.KeyPath = args[i+1]
			i += 2
			continue
		}
		if arg == "-o" && i+1 < len(args) {
			i += 2
			continue
		}
		if strings.Contains(arg, "@") && strings.Contains(arg, ":") {
			// user@host:path format (supports IPv6)
			user, host, _ := parseUserHostPath(arg)
			if user != "" && host != "" {
				config.User = user
				config.Host = host
			}
		}
		scpArgs = append(scpArgs, arg)
		i++
	}

	return config, scpArgs
}

// parseRsyncArgs parses rsync command arguments
func parseRsyncArgs(args []string) (*Config, []string) {
	config := &Config{
		User: "root",
		Port: "22",
	}
	var rsyncArgs []string

	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "rsync" {
			i++
			continue
		}
		if arg == "-e" && i+1 < len(args) {
			// skip -e ssh option
			i += 2
			continue
		}
		if strings.HasPrefix(arg, "--rsh=") {
			// skip --rsh=ssh option
			i++
			continue
		}
		if strings.HasPrefix(arg, "-p") && len(arg) > 2 {
			// -p22 format port
			config.Port = arg[2:]
			i++
			continue
		}
		if strings.Contains(arg, "@") && strings.Contains(arg, ":") {
			// user@host:path format (supports IPv6)
			user, host, _ := parseUserHostPath(arg)
			if user != "" && host != "" {
				config.User = user
				config.Host = host
			}
		}
		rsyncArgs = append(rsyncArgs, arg)
		i++
	}

	return config, rsyncArgs
}

// detectCommandType detects the command type
func detectCommandType(args []string) CommandType {
	for _, arg := range args {
		if arg == "scp" {
			return CommandSCP
		}
		if arg == "rsync" {
			return CommandRsync
		}
	}
	return CommandSSH
}

// printUsage prints the usage
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  sshpass [-p <password> | -f <passfile>] ssh [user@host] [command]")
	fmt.Println("  sshpass [-p <password> | -f <passfile>] scp [options] [user@host:]path")
	fmt.Println("  sshpass [-p <password> | -f <passfile>] rsync [options] [user@host:]path")
	fmt.Println("  sshpass -i <keypath> ssh [user@host] [command]")
	fmt.Println("  sshpass -f <configfile> [-c <command>]")
	fmt.Println("  sshpass -h <host> -p <password> [-u <user>] [-P <port>]")
	fmt.Println("\nOptions:")
	fmt.Println("  -p <password>      specify password directly")
	fmt.Println("  -f <file>      read password from file (single line) or config file")
	fmt.Println("  -i <key>      use private key authentication")
	fmt.Println("  -e             read password from environment variable SSHPASS")
	fmt.Println("  -k             enable strict host key verification (use known_hosts file)")
	fmt.Println("  -v             show version")
	fmt.Println("\nExamples:")
	fmt.Println("  sshpass -p 'pass' ssh user@example.com 'whoami'")
	fmt.Println("  sshpass -f pass.txt ssh user@example.com")
	fmt.Println("  SSHPASS='pass' sshpass -e ssh user@example.com")
	fmt.Println("  sshpass -i ~/.ssh/id_ed25519 ssh user@example.com")
	fmt.Println("  sshpass -p 'pass' scp file.txt user@example.com:/tmp/")
	fmt.Println("  sshpass -p 'pass' rsync -avz ./ user@example.com:/backup/")
}

// printVersion prints version info
func printVersion() {
	fmt.Printf("sshpass version %s (Windows)\n", version)
}

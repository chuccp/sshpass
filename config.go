package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config represents SSH connection configuration
type Config struct {
	Host          string
	User          string
	Password      string
	Port          string
	KeyPath       string // private key file path
	StrictHostKey bool   // whether to verify host key
}

// parseConfigFile parses a config file (format: key: value)
func parseConfigFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	config := &Config{
		Port: "22",
		User: "root",
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "host":
			config.Host = value
		case "user", "username":
			config.User = value
		case "password":
			config.Password = value
		case "port":
			config.Port = value
		case "key", "keypath":
			config.KeyPath = value
		}
	}

	if config.Host == "" {
		return nil, fmt.Errorf("config file missing host")
	}
	if config.Password == "" && config.KeyPath == "" {
		return nil, fmt.Errorf("config file missing password or key")
	}

	return config, nil
}

// readPasswordFile reads password from a single-line password file
func readPasswordFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read password file: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// getEnvPassword gets password from environment variable
func getEnvPassword() string {
	return os.Getenv("SSHPASS")
}
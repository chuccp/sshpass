package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config 表示SSH连接配置
type Config struct {
	Host     string
	User     string
	Password string
	Port     string
	KeyPath  string // 私钥文件路径
}

// parseConfigFile 解析配置文件 (格式: key: value)
func parseConfigFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("无法打开配置文件: %w", err)
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
		return nil, fmt.Errorf("配置文件缺少 host")
	}
	if config.Password == "" && config.KeyPath == "" {
		return nil, fmt.Errorf("配置文件缺少 password 或 key")
	}

	return config, nil
}

// readPasswordFile 从单行密码文件读取密码
func readPasswordFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("无法读取密码文件: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// getEnvPassword 从环境变量获取密码
func getEnvPassword() string {
	return os.Getenv("SSHPASS")
}
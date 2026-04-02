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

// parseConfigFile 解析配置文件
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
		if line == "" {
			continue
		}

		// 解析各种格式的配置行
		if strings.Contains(line, "IP Address:") {
			config.Host = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.Contains(line, "Username:") {
			config.User = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.Contains(line, "Root Password:") || strings.Contains(line, "Password:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				config.Password = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "SSH Port") {
			parts := strings.Split(line, "Port")
			if len(parts) == 2 {
				config.Port = strings.TrimSpace(parts[1])
			}
		}
	}

	if config.Host == "" || config.Password == "" {
		return nil, fmt.Errorf("配置文件缺少必要信息(主机或密码)")
	}

	return config, nil
}
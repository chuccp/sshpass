package main

import (
	"fmt"
	"strings"
)

// parseSSHArgs 解析ssh风格的参数 (user@host 或 -p port user@host)
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
			// 跳过 ssh 命令本身
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
			// 跳过 ssh 选项如 StrictHostKeyChecking=no
			i += 2
			continue
		}
		if strings.Contains(arg, "@") {
			// user@host 格式
			parts := strings.Split(arg, "@")
			if len(parts) == 2 {
				config.User = parts[0]
				config.Host = parts[1]
			}
			i++
			continue
		}
		// 剩余的作为命令
		if config.Host != "" {
			command = strings.Join(args[i:], " ")
			break
		}
		i++
	}

	return config, command
}

// printUsage 打印使用说明
func printUsage() {
	fmt.Println("用法:")
	fmt.Println("  sshpass -f <配置文件> [-c <命令>]")
	fmt.Println("  sshpass -h <主机> -p <密码> [-u <用户>] [-P <端口>] [-c <命令>]")
	fmt.Println("  sshpass -h <主机> -i <私钥路径> [-u <用户>] [-P <端口>] [-c <命令>]")
	fmt.Println("  sshpass -p <密码> ssh [user@host] [命令]  # sshpass风格")
	fmt.Println("  sshpass -i <私钥路径> ssh [user@host] [命令]  # 私钥登录")
	fmt.Println("  sshpass -f <配置文件> -local <本地路径> -remote <远程路径>  # 上传")
	fmt.Println("  sshpass -f <配置文件> -d -remote <远程路径> -local <本地路径>  # 下载")
	fmt.Println("\n示例:")
	fmt.Println("  sshpass -f config.txt")
	fmt.Println("  sshpass -h example.com -p password123 -c 'ls -la'")
	fmt.Println("  sshpass -p 'password' ssh ubuntu@example.com 'pkill -f antilost'")
	fmt.Println("  sshpass -p 'password' ssh -o StrictHostKeyChecking=no -p 22 user@example.com")
	fmt.Println("  sshpass -i ~/.ssh/id_ed25519 ssh ubuntu@example.com 'whoami'")
	fmt.Println("  sshpass -i ~/.ssh/id_ed25519 ssh -o StrictHostKeyChecking=no ubuntu@example.com 'sudo journalctl -u caddy -n 30'")
	fmt.Println("  sshpass -f config.txt -local README.md -remote /root/README.md")
	fmt.Println("  sshpass -f config.txt -d -remote /root/v2ray.log -local v2ray.log")
}
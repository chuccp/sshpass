package main

import (
	"fmt"
	"strings"
)

// Command_type 命令类型
type CommandType int

const (
	CommandSSH CommandType = iota
	CommandSCP
	CommandRsync
)

// ParsedCommand 解析后的命令
type ParsedCommand struct {
	Config  *Config
	Type    CommandType
	Command string   // 要执行的命令
	SCPArgs []string // scp 参数
}

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

// parseSCPArgs 解析 scp 命令参数
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
			// scp 使用大写 -P 指定端口
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
			// user@host:path 格式
			atIdx := strings.Index(arg, "@")
			colonIdx := strings.Index(arg, ":")
			if atIdx > 0 && colonIdx > atIdx {
				config.User = arg[:atIdx]
				config.Host = arg[atIdx+1 : colonIdx]
			}
		}
		scpArgs = append(scpArgs, arg)
		i++
	}

	return config, scpArgs
}

// parseRsyncArgs 解析 rsync 命令参数
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
			// 跳过 -e ssh 选项
			i += 2
			continue
		}
		if strings.HasPrefix(arg, "--rsh=") {
			// 跳过 --rsh=ssh 选项
			i++
			continue
		}
		if strings.HasPrefix(arg, "-p") && len(arg) > 2 {
			// -p22 格式的端口
			config.Port = arg[2:]
			i++
			continue
		}
		if strings.Contains(arg, "@") && strings.Contains(arg, ":") {
			// user@host:path 格式
			atIdx := strings.Index(arg, "@")
			colonIdx := strings.Index(arg, ":")
			if atIdx > 0 && colonIdx > atIdx {
				config.User = arg[:atIdx]
				config.Host = arg[atIdx+1 : colonIdx]
			}
		}
		rsyncArgs = append(rsyncArgs, arg)
		i++
	}

	return config, rsyncArgs
}

// detectCommandType 检测命令类型
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

// printUsage 打印使用说明
func printUsage() {
	fmt.Println("用法:")
	fmt.Println("  sshpass [-p <密码> | -f <密码文件>] ssh [user@host] [命令]")
	fmt.Println("  sshpass [-p <密码> | -f <密码文件>] scp [选项] [user@host:]path")
	fmt.Println("  sshpass [-p <密码> | -f <密码文件>] rsync [选项] [user@host:]path")
	fmt.Println("  sshpass -i <私钥路径> ssh [user@host] [命令]")
	fmt.Println("  sshpass -f <配置文件> [-c <命令>]")
	fmt.Println("  sshpass -h <主机> -p <密码> [-u <用户>] [-P <端口>]")
	fmt.Println("\n选项:")
	fmt.Println("  -p <密码>      直接指定密码")
	fmt.Println("  -f <文件>      从文件读取密码（单行）或配置文件")
	fmt.Println("  -i <私钥>      使用私钥认证")
	fmt.Println("  -e             从环境变量 SSHPASS 读取密码")
	fmt.Println("  -v             显示版本")
	fmt.Println("\n示例:")
	fmt.Println("  sshpass -p 'pass' ssh user@example.com 'whoami'")
	fmt.Println("  sshpass -f pass.txt ssh user@example.com")
	fmt.Println("  SSHPASS='pass' sshpass -e ssh user@example.com")
	fmt.Println("  sshpass -i ~/.ssh/id_ed25519 ssh user@example.com")
	fmt.Println("  sshpass -p 'pass' scp file.txt user@example.com:/tmp/")
	fmt.Println("  sshpass -p 'pass' rsync -avz ./ user@example.com:/backup/")
}

// printVersion 打印版本信息
func printVersion() {
	fmt.Println("sshpass version 0.2.3 (Windows)")
}

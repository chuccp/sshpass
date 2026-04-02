package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	version = "v0.3.1"
)

func main() {
	// 命令行参数
	configFile := flag.String("f", "", "密码文件或配置文件路径")
	host := flag.String("h", "", "主机地址")
	user := flag.String("u", "root", "用户名")
	password := flag.String("p", "", "密码")
	port := flag.String("P", "22", "端口")
	keyPath := flag.String("i", "", "私钥文件路径")
	command := flag.String("c", "", "要执行的命令")
	localPath := flag.String("local", "", "本地文件路径(用于上传/下载)")
	remotePath := flag.String("remote", "", "远程文件路径(用于上传/下载)")
	download := flag.Bool("d", false, "下载模式（从远程下载到本地）")
	useEnv := flag.Bool("e", false, "从环境变量 SSHPASS 读取密码")
	showVersion := flag.Bool("v", false, "显示版本")
	flag.Parse()

	// 显示版本
	if *showVersion {
		printVersion()
		return
	}

	var config *Config
	var err error
	var cmdToRun string

	// 获取剩余参数（用于sshpass风格命令）
	remainingArgs := flag.Args()

	// 获取密码：优先级 -p > 配置文件 > 密码文件 > -e > SSHPASS
	pass := *password
	if pass == "" && *configFile != "" {
		// 先尝试作为配置文件解析
		config, err = parseConfigFile(*configFile)
		if err != nil {
			// 不是配置文件，尝试作为单行密码文件
			pass, err = readPasswordFile(*configFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				os.Exit(1)
			}
		}
	}
	if pass == "" && *useEnv {
		pass = getEnvPassword()
	}

	// 检测命令类型
	cmdType := detectCommandType(remainingArgs)

	// 根据命令类型处理
	switch cmdType {
	case CommandSCP:
		config, scpArgs := parseSCPArgs(remainingArgs)
		config.Password = pass
		config.KeyPath = *keyPath
		if *host != "" {
			config.Host = *host
		}
		if *user != "" && *user != "root" {
			config.User = *user
		}
		if *port != "" && *port != "22" {
			config.Port = *port
		}
		err = runSCP(config, scpArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SCP执行失败: %v\n", err)
			os.Exit(1)
		}
		return

	case CommandRsync:
		config, rsyncArgs := parseRsyncArgs(remainingArgs)
		config.Password = pass
		config.KeyPath = *keyPath
		if *host != "" {
			config.Host = *host
		}
		if *user != "" && *user != "root" {
			config.User = *user
		}
		if *port != "" && *port != "22" {
			config.Port = *port
		}
		err = runRsync(config, rsyncArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Rsync执行失败: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// SSH 命令处理
	if config == nil {
		if len(remainingArgs) > 0 && (pass != "" || *keyPath != "") {
			// sshpass风格: -p password 或 -i keyfile ssh user@host [command]
			config, cmdToRun = parseSSHArgs(remainingArgs)
			config.Password = pass
			config.KeyPath = *keyPath
			if *host != "" {
				config.Host = *host
			}
			if *user != "" && *user != "root" {
				config.User = *user
			}
			if *port != "" && *port != "22" {
				config.Port = *port
			}
		} else if *host != "" && (pass != "" || *keyPath != "") {
			// 从命令行参数读取（包括文件传输模式）
			config = &Config{
				Host:     *host,
				User:     *user,
				Password: pass,
				Port:     *port,
				KeyPath:  *keyPath,
			}
		} else {
			printUsage()
			os.Exit(1)
		}
	}

	// 验证配置
	if config.Host == "" {
		fmt.Fprintf(os.Stderr, "错误: 未指定主机地址\n")
		os.Exit(1)
	}
	if config.Password == "" && config.KeyPath == "" {
		fmt.Fprintf(os.Stderr, "错误: 未提供认证方式（需要密码或私钥）\n")
		os.Exit(1)
	}

	// 建立SSH连接
	client, err := SSHClient(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "SSH连接失败: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// 文件传输
	if *localPath != "" && *remotePath != "" {
		if *download {
			fmt.Printf("正在下载 %s -> %s...\n", *remotePath, *localPath)
			err = downloadFile(client, *remotePath, *localPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "下载失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("下载成功!")
		} else {
			fmt.Printf("正在上传 %s -> %s...\n", *localPath, *remotePath)
			err = uploadFile(client, *localPath, *remotePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "上传失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("上传成功!")
		}
		return
	} else if *localPath != "" || *remotePath != "" {
		fmt.Fprintf(os.Stderr, "错误: 文件传输需要同时指定 -local 和 -remote 参数\n")
		os.Exit(1)
	}

	// 执行命令
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
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		}
	}
}

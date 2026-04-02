package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHClient 创建SSH客户端连接
func SSHClient(config *Config) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	// 优先使用私钥认证
	if config.KeyPath != "" {
		key, err := os.ReadFile(config.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("读取私钥失败: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %w", err)
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// 如果有密码，添加密码认证（作为备选或主要方式）
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("未提供认证方式（密码或私钥）")
	}

	// 设置主机密钥验证回调
	var hostKeyCallback ssh.HostKeyCallback
	if config.StrictHostKey {
		// 使用 known_hosts 文件验证
		knownHostsPath := getKnownHostsPath()
		callback, err := knownhosts.New(knownHostsPath)
		if err != nil {
			return nil, fmt.Errorf("无法读取 known_hosts 文件 (%s): %w\n提示: 请先手动连接一次服务器以添加主机密钥", knownHostsPath, err)
		}
		hostKeyCallback = callback
	} else {
		// 忽略主机密钥验证（默认）
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
		return nil, fmt.Errorf("连接失败: %w", err)
	}

	return client, nil
}

// getKnownHostsPath 获取 known_hosts 文件路径
func getKnownHostsPath() string {
	// 优先使用环境变量
	if path := os.Getenv("SSH_KNOWN_HOSTS"); path != "" {
		return path
	}

	// Windows 默认路径: %USERPROFILE%\.ssh\known_hosts
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.Getenv("USERPROFILE")
	}
	return filepath.Join(homeDir, ".ssh", "known_hosts")
}

// runShell 启动交互式Shell
func runShell(client *ssh.Client) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	// 设置终端模式
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// 请求伪终端
	if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
		return fmt.Errorf("请求终端失败: %w", err)
	}

	// 设置标准输入输出
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	// 启动远程shell
	if err := session.Shell(); err != nil {
		return fmt.Errorf("启动Shell失败: %w", err)
	}

	return session.Wait()
}

// executeCommand 执行单个命令
func executeCommand(client *ssh.Client, command string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	return session.Run(command)
}
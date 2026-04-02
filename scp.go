package main

import (
	"fmt"
	"strings"
)

// runSCP 执行 scp 命令（通过 SSH 连接传输文件）
func runSCP(config *Config, args []string) error {
	// 建立SSH连接
	client, err := SSHClient(config)
	if err != nil {
		return err
	}
	defer client.Close()

	// 解析 scp 参数，确定源和目标
	var localFile, remotePath string
	var isUpload bool
	var nonFlagArgs []string

	// 收集非标志参数
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			nonFlagArgs = append(nonFlagArgs, arg)
		}
	}

	// 解析源和目标
	for _, arg := range nonFlagArgs {
		if strings.Contains(arg, "@") && strings.Contains(arg, ":") {
			// 远程路径 user@host:path
			_, remotePath, _ = strings.Cut(arg, ":")
		} else if arg != "scp" {
			// 本地文件
			if localFile == "" {
				localFile = arg
			}
		}
	}

	// 判断上传还是下载：如果远程路径在最后一个参数，则是上传
	for i, arg := range nonFlagArgs {
		if strings.Contains(arg, "@") && strings.Contains(arg, ":") {
			isUpload = (i == len(nonFlagArgs)-1)
			break
		}
	}

	if localFile == "" || remotePath == "" {
		return fmt.Errorf("无法解析 scp 参数: %v", args)
	}

	if isUpload {
		return uploadFile(client, localFile, remotePath)
	}
	return downloadFile(client, remotePath, localFile)
}

// runRsync 执行 rsync 命令（通过 SSH 连接同步文件）
func runRsync(config *Config, args []string) error {
	// 简单实现：解析源和目标
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
		return fmt.Errorf("无法解析 rsync 参数")
	}

	// 建立SSH连接
	client, err := SSHClient(config)
	if err != nil {
		return err
	}
	defer client.Close()

	// 判断方向
	if strings.Contains(src, "@") {
		// 远程到本地（下载）
		_, remotePath, _ := strings.Cut(src, ":")
		return downloadFile(client, remotePath, dst)
	}
	// 本地到远程（上传）
	_, remotePath, _ := strings.Cut(dst, ":")
	return uploadFile(client, src, remotePath)
}
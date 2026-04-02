package main

import (
	"fmt"
	"io"
	"os"
	"path" // Unix 风格路径（始终使用 /）
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/crypto/ssh"
)

// uploadFile 上传文件或目录到远程服务器
func uploadFile(client *ssh.Client, localPath, remotePath string) error {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("创建SFTP客户端失败: %w", err)
	}
	defer sftpClient.Close()

	// 检查本地是文件还是目录
	localInfo, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("获取本地文件信息失败: %w", err)
	}

	if localInfo.IsDir() {
		return uploadDirectory(sftpClient, localPath, remotePath)
	}
	return uploadSingleFile(sftpClient, localPath, remotePath)
}

// uploadSingleFile 上传单个文件
func uploadSingleFile(sftpClient *sftp.Client, localPath, remotePath string) error {
	// 检查远程路径是否是目录
	remoteFileInfo, err := sftpClient.Stat(remotePath)
	if err == nil && remoteFileInfo.IsDir() {
		remotePath = path.Join(remotePath, filepath.Base(localPath))
	}

	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %w", err)
	}
	defer localFile.Close()

	// 获取文件大小
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}
	fileSize := fileInfo.Size()

	// 确保远程目录存在
	remoteDir := path.Dir(remotePath)
	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("创建远程目录失败: %w", err)
	}

	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("创建远程文件失败: %w", err)
	}
	defer remoteFile.Close()

	// 创建进度条
	bar := progressbar.NewOptions64(
		fileSize,
		progressbar.OptionSetDescription(fmt.Sprintf("上传 %s", filepath.Base(localPath))),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionFullWidth(),
		progressbar.OptionUseANSICodes(true),
	)

	_, err = io.Copy(remoteFile, io.TeeReader(localFile, bar))
	if err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}

	return nil
}

// uploadDirectory 上传整个目录
func uploadDirectory(sftpClient *sftp.Client, localPath, remotePath string) error {
	// 获取本地目录名
	localBase := filepath.Base(localPath)

	// 确保远程目录存在（使用 path.Join 保证 Unix 风格路径）
	remoteDir := path.Join(remotePath, localBase)
	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("创建远程目录失败: %w", err)
	}

	// 遍历本地目录
	return filepath.Walk(localPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(localPath, filePath)
		if err != nil {
			return err
		}

		// 将 Windows 相对路径转换为 Unix 风格
		relPath = filepath.ToSlash(relPath)

		// 远程完整路径（使用 path.Join 保证 Unix 风格路径）
		remoteFullPath := path.Join(remoteDir, relPath)

		if info.IsDir() {
			// 创建远程目录
			return sftpClient.MkdirAll(remoteFullPath)
		}

		// 上传文件
		return uploadSingleFile(sftpClient, filePath, remoteFullPath)
	})
}

// downloadFile 从远程服务器下载文件或目录
func downloadFile(client *ssh.Client, remotePath, localPath string) error {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("创建SFTP客户端失败: %w", err)
	}
	defer sftpClient.Close()

	// 检查远程是文件还是目录
	remoteInfo, err := sftpClient.Stat(remotePath)
	if err != nil {
		return fmt.Errorf("获取远程文件信息失败: %w", err)
	}

	if remoteInfo.IsDir() {
		return downloadDirectory(sftpClient, remotePath, localPath)
	}
	return downloadSingleFile(sftpClient, remotePath, localPath)
}

// downloadSingleFile 下载单个文件
func downloadSingleFile(sftpClient *sftp.Client, remotePath, localPath string) error {
	// 检查本地路径是否是目录
	localFileInfo, err := os.Stat(localPath)
	if err == nil && localFileInfo.IsDir() {
		localPath = filepath.Join(localPath, path.Base(remotePath))
	}

	// 确保本地目录存在
	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}

	remoteFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("打开远程文件失败: %w", err)
	}
	defer remoteFile.Close()

	// 获取文件大小
	fileInfo, err := remoteFile.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}
	fileSize := fileInfo.Size()

	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("创建本地文件失败: %w", err)
	}
	defer localFile.Close()

	// 创建进度条
	bar := progressbar.NewOptions64(
		fileSize,
		progressbar.OptionSetDescription(fmt.Sprintf("下载 %s", path.Base(remotePath))),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionFullWidth(),
		progressbar.OptionUseANSICodes(true),
	)

	_, err = io.Copy(localFile, io.TeeReader(remoteFile, bar))
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}

	return nil
}

// downloadDirectory 下载整个目录
func downloadDirectory(sftpClient *sftp.Client, remotePath, localPath string) error {
	// 获取远程目录名
	remoteBase := path.Base(remotePath)

	// 确保本地目录存在
	localDir := filepath.Join(localPath, remoteBase)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}

	// 确保远程路径以 / 结尾，便于计算相对路径
	remotePath = strings.TrimSuffix(remotePath, "/") + "/"

	// 遍历远程目录
	walker := sftpClient.Walk(remotePath)
	for walker.Step() {
		if err := walker.Err(); err != nil {
			return err
		}

		remoteFilePath := walker.Path()

		// 计算相对路径（去掉远程基础路径）
		relPath := strings.TrimPrefix(remoteFilePath, remotePath)
		if relPath == "" {
			continue
		}

		// 本地完整路径
		localFullPath := filepath.Join(localDir, relPath)

		info := walker.Stat()
		if info.IsDir() {
			// 创建本地目录
			if err := os.MkdirAll(localFullPath, 0755); err != nil {
				return err
			}
		} else {
			// 下载文件
			if err := downloadSingleFile(sftpClient, remoteFilePath, localFullPath); err != nil {
				return err
			}
		}
	}

	return nil
}
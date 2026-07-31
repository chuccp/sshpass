# Go SDK

win-sshpass 不仅是一个命令行工具，也是一个可复用的 Go 库（`package sshpass`）。你可以将 SSH/SFTP/Shell 功能嵌入到自己的应用中。

## 安装

```bash
go get github.com/chuccp/win-sshpass
```

## 快速开始

```go
package main

import (
    "log"

    sshpass "github.com/chuccp/win-sshpass"
)

func main() {
    cfg := sshpass.NewConfig()
    cfg.Host = "example.com"
    cfg.User = "root"
    cfg.Password = "secret"

    // NewClient 建立 SSH 连接并返回可用的客户端
    client, err := sshpass.NewClient(cfg, sshpass.WithSignalHandler())
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 执行命令（输出流默认为 os.Stdout/os.Stderr）
    if err := client.Exec("uname -a"); err != nil {
        log.Fatal(err)
    }
}
```

## Client 核心方法

### NewClient

建立 SSH 连接并返回客户端：

```go
client, err := sshpass.NewClient(cfg, opts...)
```

- `cfg`：连接配置（`*sshpass.Config`）
- `opts`：可选的函数选项（`...sshpass.Option`）
- 如果 `cfg.Timeout > 0`，会启动操作计时器，超时后自动关闭连接

### Exec

执行单条远程命令：

```go
err := client.Exec("ls -la")
```

- 输出流通过 `WithStdout`/`WithStderr` 配置（默认 `os.Stdout`/`os.Stderr`）
- 输入流通过 `WithStdin` 配置（默认 `os.Stdin`）

### ExecCapture

执行命令并捕获输出（不流式输出），适用于程序化处理：

```go
stdout, stderr, exitCode, err := client.ExecCapture("whoami")
if exitCode == 0 && err == nil {
    fmt.Println("Output:", stdout)
} else {
    fmt.Println("Error:", stderr)
}
```

- `exitCode`: 0 表示成功，非零为远程退出码，-1 表示会话/连接错误
- `err`: 仅在会话创建或连接级别失败时非 nil；正常退出（包括非零退出码）时为 nil

### Shell

启动交互式 Shell：

```go
err := client.Shell()
```

- 自动检测终端，支持 PTY 和 raw 模式
- 支持动态终端调整大小
- 支持 rz/sz 文件传输（需要配置 `WithFileSelector`）

### SFTP

打开 SFTP 子通道：

```go
sftp, err := client.SFTP()
if err != nil {
    log.Fatal(err)
}
defer sftp.Close()

// 上传文件
err = sftp.Upload("./local.txt", "/tmp/remote.txt")

// 下载文件
err = sftp.Download("/tmp/remote.txt", "./local.txt")

// 访问底层 *sftp.Client（高级用法）
rawClient := sftp.SFTP()
```

### Close

关闭 SSH 连接：

```go
err := client.Close()
```

- 幂等操作，多次调用返回相同的错误
- 自动停止计时器和信号处理器

### TimedOut

检查是否因超时而失败：

```go
if client.TimedOut() {
    fmt.Println("操作超时")
}
```

### 端口转发

通过 SSH 连接创建 TCP 隧道。本地转发和远程转发均返回一个在后台运行的 `*Forwarder`。

#### LocalForward

在本地地址上监听，将每个连接通过 SSH 转发到远程目标：

```go
// 将 localhost:8080 → db.internal:3306 通过 SSH 服务器转发
fwd, err := client.LocalForward("127.0.0.1:8080", "db.internal:3306")
if err != nil {
    log.Fatal(err)
}
defer fwd.Close()
// Forwarder 在后台运行，直到调用 Close() 停止
```

#### RemoteForward

要求 SSH 服务器在远程地址上监听，并将连接转发回本地目标：

```go
// 将 localhost:8080 暴露在远程服务器的 9090 端口
fwd, err := client.RemoteForward("0.0.0.0:9090", "127.0.0.1:8080")
if err != nil {
    log.Fatal(err)
}
defer fwd.Close()
```

#### Forwarder 生命周期

- `Close()` — 停止隧道并释放监听器（可安全多次调用）
- `Wait()` — 阻塞直到转发器停止（通过 `Close()` 或监听器错误）
- 两个方法都是 goroutine 安全的

!!! info "端口转发与连接共享"
    `LocalForward` 和 `RemoteForward` 共享 Client 的 SSH 连接。关闭 `Client` 也会终止所有转发器。

## 配置（Config）

```go
cfg := sshpass.NewConfig()
cfg.Host = "example.com"
cfg.User = "root"
cfg.Password = "secret"
cfg.Port = "22"
cfg.KeyPath = "~/.ssh/id_ed25519"
cfg.StrictHostKey = true
cfg.Timeout = 30          // 操作超时（秒），0 = 无限制
cfg.ConnectTimeout = 10   // TCP 连接超时（秒）
cfg.Retries = 3           // 连接重试次数
```

### 从配置文件加载

```go
cfg, err := sshpass.LoadConfig("server.config")
```

### 加载配置文件或密码文件

```go
cfg, pass, err := sshpass.LoadConfigOrPasswordFile("file.txt", "", false)
if cfg != nil {
    // 是配置文件
} else {
    // 是密码文件，pass 包含密码
}
```

## 密钥生成

以编程方式生成 SSH 密钥对：

```go
// 生成 Ed25519 密钥对（推荐）
pair, err := sshpass.GenerateKeyPair(sshpass.KeyEd25519, "user@host")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("私钥：\n%s\n", pair.PrivateKey)
fmt.Printf("公钥：\n%s\n", pair.PublicKey)

// 生成 RSA 密钥对（至少 2048 位）
pair, err = sshpass.GenerateRSAKeyPair(4096, "user@host")

// 将密钥对保存到文件
err = sshpass.SaveKeyPair(pair, "~/.ssh/mykey")
// 创建：~/.ssh/mykey（私钥，0600）和 ~/.ssh/mykey.pub（公钥）

// 获取默认密钥路径
path := sshpass.DefaultKeyPath(sshpass.KeyEd25519) // ~/.ssh/id_ed25519

// 将公钥部署到远程服务器（需要已有的客户端连接）
err = sshpass.DeployPublicKey(client, pair.PublicKey)
```

## 函数选项（Options）

通过函数选项配置 Client 的行为：

### I/O 流

```go
// 重定向输入/输出
client, err := sshpass.NewClient(cfg,
    sshpass.WithStdin(myReader),
    sshpass.WithStdout(myWriter),
    sshpass.WithStderr(myWriter),
)
```

### 进度回调

```go
client, err := sshpass.NewClient(cfg,
    sshpass.WithProgress(func(desc string, sent, total int64) {
        fmt.Printf("\r%s %d/%d bytes", desc, sent, total)
    }),
)
```

- `desc`：传输描述（如 "Uploading file.txt"）
- `sent`：已传输字节数
- `total`：文件总大小
- SDK 不做任何渲染，调用者自行显示进度

### 文件选择器

```go
type myFileSelector struct{}

func (s myFileSelector) OpenFile() (string, error) {
    // 实现文件打开对话框
    return "/path/to/file", nil
}

func (s myFileSelector) SaveFile(defaultName string) (string, error) {
    // 实现文件保存对话框
    return "/path/to/save", nil
}

client, err := sshpass.NewClient(cfg,
    sshpass.WithFileSelector(myFileSelector{}),
)
```

用于 rz/sz Shell 传输的文件选择。SDK 不提供默认实现。

### 信号处理

```go
client, err := sshpass.NewClient(cfg,
    sshpass.WithSignalHandler(),
)
```

注册 Ctrl+C 处理器，按下时关闭连接。默认不注册，以免干扰宿主进程的信号处理。

### 断点续传

```go
client, err := sshpass.NewClient(cfg,
    sshpass.WithResume(),
)
```

启用 SFTP 传输的断点续传功能 —— 中断的上传/下载将从断点处继续。

### SSH Agent 转发

```go
client, err := sshpass.NewClient(cfg,
    sshpass.WithAgentForward(),
)
```

启用 ssh-agent 转发 —— 远程服务器可以使用本地 ssh-agent 进行进一步的 SSH 连接（例如从跳板机执行 `git clone`）。需要运行中的本地 ssh-agent 并已加载密钥。

### SSH Agent 认证

在配置上设置 `UseAgent` 以使用本地 ssh-agent 进行认证：

```go
cfg := sshpass.NewConfig()
cfg.Host = "example.com"
cfg.User = "root"
cfg.UseAgent = true  // 自动检测并使用本地 ssh-agent
```

当 `UseAgent` 为 true 且未配置密码或密钥路径时，win-sshpass 会自动连接到本地 ssh-agent 进行认证。这是 CLI 中未指定 `-p` 或 `-i` 时的默认行为。

### 代理配置

```go
cfg := sshpass.NewConfig()
cfg.Host = "example.com"
cfg.User = "root"
cfg.Password = "secret"
cfg.ProxyURL = "socks5://user:pass@127.0.0.1:1080" // 或 http://、https://、socks4://
```

设置 `ProxyURL` 后，SSH 连接将通过指定的代理服务器进行隧道传输。支持的协议：SOCKS5（可选认证）、SOCKS4、SOCKS4A、HTTP CONNECT、HTTPS CONNECT。

## 底层 API

### Dial

直接创建 SSH 客户端连接：

```go
sshClient, err := sshpass.Dial(cfg)
```

返回 `*ssh.Client`，适用于需要更底层控制的场景。

### 参数解析

```go
// 解析 SSH 参数
config, cmd := sshpass.ParseSSHArgs([]string{"ssh", "user@host", "ls"})

// 解析 SCP 参数
config, args := sshpass.ParseSCPArgs([]string{"scp", "file.txt", "user@host:/tmp/"})

// 解析 Rsync 参数
config, args := sshpass.ParseRsyncArgs([]string{"rsync", "-avz", "./", "user@host:/backup/"})

// 检测命令类型
cmdType := sshpass.DetectCommandType(args)
```

### 工具函数

```go
// 运行 SCP 传输
err := sshpass.RunSCP(client, args)

// 运行 Rsync 传输
err := sshpass.RunRsync(client, args)

// 清理远程路径（处理 Git Bash 路径转换）
path, err := sshpass.CleanRemotePath("//tmp/file.txt")

// 解析 user@host:path 格式
user, host, path := sshpass.ParseUserHostPath("user@host:/tmp/file.txt")

// 分割路径（逗号或空格分隔）
paths, err := sshpass.SplitPaths("a.txt,b.txt,c.txt", "local")

// 从错误中提取退出码
code, ok := sshpass.ExitCodeFromError(err)
```

## 完整示例

### 批量执行命令

```go
package main

import (
    "fmt"
    "log"

    sshpass "github.com/chuccp/win-sshpass"
)

func main() {
    hosts := []string{"192.168.1.101", "192.168.1.102", "192.168.1.103"}

    for _, host := range hosts {
        cfg := sshpass.NewConfig()
        cfg.Host = host
        cfg.User = "root"
        cfg.Password = "secret"

        client, err := sshpass.NewClient(cfg)
        if err != nil {
            log.Printf("[%s] 连接失败: %v", host, err)
            continue
        }

        fmt.Printf("[%s] 执行命令...\n", host)
        if err := client.Exec("uptime"); err != nil {
            log.Printf("[%s] 执行失败: %v", host, err)
        }
        client.Close()
    }
}
```

### 带进度的文件上传

```go
package main

import (
    "fmt"
    "log"

    sshpass "github.com/chuccp/win-sshpass"
)

func main() {
    cfg := sshpass.NewConfig()
    cfg.Host = "example.com"
    cfg.User = "root"
    cfg.Password = "secret"

    client, err := sshpass.NewClient(cfg,
        sshpass.WithProgress(func(desc string, sent, total int64) {
            pct := sent * 100 / total
            fmt.Printf("\r%s %d%%", desc, pct)
        }),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    sftp, err := client.SFTP()
    if err != nil {
        log.Fatal(err)
    }
    defer sftp.Close()

    if err := sftp.Upload("./large-file.zip", "/tmp/large-file.zip"); err != nil {
        log.Fatal(err)
    }
    fmt.Println("\n上传完成!")
}
```

## 下一步

- [最佳实践](best-practices.md) - 安全与效率建议

# sshpass

Windows 版 sshpass 工具，让你在 Windows 上体验类似 Linux sshpass 的功能。

## 功能

- SSH 密码认证登录（无需手动输入密码）
- 执行远程命令
- 文件上传/下载（SFTP）
- 交互式 Shell
- 支持原版 sshpass 命令格式

## 安装

```bash
go build -o sshpass.exe .
```

## 使用方法

### 1. sshpass 风格（推荐）

```bash
# 基本格式
sshpass -p '密码' ssh user@host '命令'
sshpass -p '密码' ssh -o StrictHostKeyChecking=no user@host '命令'

# 示例
sshpass -p 'your_password' ssh ubuntu@example.com 'pkill -f app'
sshpass -p 'your_password' ssh -p 2222 user@example.com 'ls -la'
```

### 2. 配置文件模式

创建配置文件（如 `server.config`）：

```
IP Address:your.server.com
Username:root
Root Password: your_password
SSH Port 22
```

```bash
sshpass -f server.config              # 登录
sshpass -f server.config -c 'ls -la'  # 执行命令
```

### 3. 命令行参数模式

```bash
sshpass -h <主机> -p <密码> [-u <用户>] [-P <端口>] [-c <命令>]
```

### 4. 文件传输

```bash
# 上传
sshpass -f config.txt -local <本地路径> -remote <远程路径>

# 下载
sshpass -f config.txt -d -remote <远程路径> -local <本地路径>
```

## 示例

```bash
# sshpass风格
sshpass -p 'password' ssh root@example.com 'whoami'

# 带ssh选项
sshpass -p 'password' ssh -o StrictHostKeyChecking=no ubuntu@example.com 'pkill -f app || true'

# 指定端口
sshpass -p 'password' ssh -p 2222 user@example.com 'hostname'

# 配置文件
sshpass -f server.config
sshpass -f server.config -c 'tail -f /var/log/app.log'

# 上传文件
sshpass -f server.config -local file.txt -remote /tmp/file.txt

# 下载文件
sshpass -f server.config -d -remote /var/log/app.log -local app.log
```

## 注意事项

- 在 Git Bash 中使用时，远程路径要用 `//` 开头（如 `//root/file.txt`）避免路径被自动转换
- 配置文件中密码字段支持 `Password:` 或 `Root Password:` 格式

## 依赖

- Go 1.23+
- golang.org/x/crypto/ssh
- github.com/pkg/sftp
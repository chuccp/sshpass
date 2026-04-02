# sshpass

Windows 版 sshpass 工具，让你在 Windows 上体验类似 Linux sshpass 的功能。

## 功能

- SSH 密码认证登录
- SSH 私钥认证登录
- 执行远程命令
- 文件/目录上传下载（带进度条）
- 交互式 Shell
- 支持原版 sshpass 命令格式
- 支持 scp 风格文件传输
- 支持 rsync 风格文件同步

## 安装

```bash
go build -o sshpass.exe .
```

## 使用方法

### 1. 密码传递方式

```bash
# 直接指定密码
sshpass -p 'password' ssh user@example.com 'whoami'

# 从文件读取密码（单行密码文件）
echo 'password' > pass.txt
sshpass -f pass.txt ssh user@example.com

# 从环境变量读取密码
export SSHPASS='password'
sshpass -e ssh user@example.com
```

### 2. 私钥登录

```bash
sshpass -i ~/.ssh/id_ed25519 ssh user@example.com 'whoami'
sshpass -i ~/.ssh/id_ed25519 ssh -o StrictHostKeyChecking=no ubuntu@example.com 'sudo journalctl -u caddy -n 30'
```

### 3. 文件/目录传输

```bash
# 上传文件
sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# 上传目录
sshpass -h example.com -p 'password' -local mydir -remote /tmp/mydir

# 下载文件
sshpass -h example.com -p 'password' -d -remote /tmp/file.txt -local ./file.txt

# 下载目录
sshpass -h example.com -p 'password' -d -remote /tmp/mydir -local ./mydir
```

### 4. SCP 风格传输

```bash
# 上传文件
sshpass -p 'password' scp local.txt user@example.com:/remote/path/

# 下载文件
sshpass -p 'password' scp user@example.com:/remote/file.txt ./local/
```

### 5. Rsync 风格同步

```bash
sshpass -p 'password' rsync -avz ./local/ user@example.com:/remote/backup/
```

### 6. 配置文件模式

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

## 选项

| 选项 | 说明 |
|------|------|
| `-p <密码>` | 直接指定密码 |
| `-f <文件>` | 从文件读取密码（单行）或配置文件 |
| `-i <私钥>` | 使用私钥认证 |
| `-e` | 从环境变量 `SSHPASS` 读取密码 |
| `-v` | 显示版本 |
| `-h <主机>` | 指定主机地址 |
| `-u <用户>` | 指定用户名（默认 root） |
| `-P <端口>` | 指定端口（默认 22） |
| `-c <命令>` | 要执行的命令 |
| `-local <路径>` | 本地文件/目录路径 |
| `-remote <路径>` | 远程文件/目录路径 |
| `-d` | 下载模式 |

## 示例

```bash
# 密码登录
sshpass -p 'password' ssh root@example.com 'whoami'

# 私钥登录
sshpass -i ~/.ssh/id_ed25519 ssh ubuntu@example.com 'hostname'

# 环境变量密码
export SSHPASS='password'
sshpass -e ssh user@example.com

# 密码文件
echo 'password' > ~/.ssh/pass
sshpass -f ~/.ssh/pass ssh user@example.com

# 上传目录（带进度条）
sshpass -h example.com -p 'password' -local ./project -remote /var/www/

# 下载目录
sshpass -h example.com -p 'password' -d -remote /var/log/app -local ./logs
```

## 注意事项

- 在 Git Bash 中使用时，远程路径要用 `//` 开头（如 `//tmp/file.txt`）避免路径被自动转换
- 配置文件中密码字段支持 `Password:` 或 `Root Password:` 格式
- 私钥和密码可同时指定，私钥认证优先
- 文件传输支持目录递归上传/下载
- 传输过程显示进度条

## 依赖

- Go 1.23+
- golang.org/x/crypto/ssh
- github.com/pkg/sftp
- github.com/schollz/progressbar/v3
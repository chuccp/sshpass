# win-win-sshpass

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md)

Windows 版 win-sshpass 工具，实现类似 Linux sshpass 的功能。

## 下载

从 [GitHub Releases](https://github.com/chuccp/win-sshpass/releases) 下载最新版本：

1. 打开 [Releases](https://github.com/chuccp/win-sshpass/releases) 页面
2. 下载最新版本的 `win-sshpass-*.zip`
3. 解压 `win-sshpass.exe` 到你想要的位置

## 快速开始

```bash
# 密码登录执行命令
win-sshpass -p 'password' ssh user@example.com 'whoami'

# 私钥登录执行命令
win-sshpass -i ~/.ssh/id_ed25519 ssh user@example.com 'hostname'

# 上传文件
win-sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# 下载文件
win-sshpass -h example.com -p 'password' -d -remote /tmp/file.txt -local ./file.txt
```

## 命令格式

### SSH 登录

```bash
# 密码认证
win-sshpass -p <密码> ssh [user@host] [命令]
win-sshpass -p <密码> ssh -p <端口> user@host '命令'
win-sshpass -p <密码> ssh -o StrictHostKeyChecking=no user@host

# 私钥认证
win-sshpass -i <私钥路径> ssh [user@host] [命令]

# 环境变量密码
SSHPASS=<密码> win-sshpass -e ssh user@host

# 密码文件
echo 'password' > pass.txt
win-sshpass -f pass.txt ssh user@host

# 配置文件（多行格式）
win-sshpass -f server.config
```

### 文件传输

```bash
# 上传文件
win-sshpass -h <主机> -p <密码> -local <本地路径> -remote <远程路径>

# 上传目录（自动递归）
win-sshpass -h <主机> -p <密码> -local <本地目录> -remote <远程目录>

# 下载文件/目录
win-sshpass -h <主机> -p <密码> -d -remote <远程路径> -local <本地路径>
```

### SCP 风格

```bash
win-sshpass -p <密码> scp <本地文件> user@host:<远程路径>
win-sshpass -p <密码> scp user@host:<远程文件> <本地路径>
```

### Rsync 风格

```bash
win-sshpass -p <密码> rsync -avz <本地路径> user@host:<远程路径>
```

## 参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| `-p` | 密码 | `-p 'secret123'` |
| `-i` | 私钥路径 | `-i ~/.ssh/id_ed25519` |
| `-f` | 密码文件/配置文件 | `-f pass.txt` |
| `-e` | 从环境变量 SSHPASS 读密码 | `SSHPASS='pass' win-sshpass -e ssh ...` |
| `-h` | 主机地址 | `-h example.com` |
| `-u` | 用户名，默认 root | `-u ubuntu` |
| `-P` | 端口，默认 22 | `-P 2222` |
| `-c` | 执行的命令 | `-c 'ls -la'` |
| `-local` | 本地路径（上传/下载） | `-local ./file.txt` |
| `-remote` | 远程路径（上传/下载） | `-remote /tmp/file.txt` |
| `-d` | 下载模式 | `-d` |
| `-v` | 显示版本 | `-v` |

## 配置文件格式

```yaml
host: example.com
username: root
password: your_password
port: 22
# key: ~/.ssh/id_ed25519  # 可选，使用私钥代替密码
```

使用方式：
```bash
win-sshpass -f server.config -c 'ls -la'
```

## 完整示例

```bash
# 1. 密码登录执行命令
win-sshpass -p 'mypass' ssh root@192.168.1.100 'docker ps'

# 2. 私钥登录执行 sudo 命令
win-sshpass -i ~/.ssh/id_ed25519 ssh ubuntu@server.com 'sudo systemctl restart nginx'

# 3. 上传整个目录到服务器
win-sshpass -h server.com -p 'mypass' -local ./dist -remote /var/www/html

# 4. 下载服务器日志目录
win-sshpass -h server.com -p 'mypass' -d -remote /var/log/nginx -local ./logs

# 5. SCP 上传文件
win-sshpass -p 'mypass' scp ./app.jar user@server.com:/opt/app/

# 6. 环境变量传递密码（更安全）
export SSHPASS='mypass'
win-sshpass -e ssh user@server.com 'whoami'
```

## Git Bash 注意事项

远程路径用 `//` 开头避免路径转换：
```bash
# 错误：/tmp 会被转换为 Windows 路径
win-sshpass ... -remote /tmp/file.txt

# 正确：使用双斜杠
win-sshpass ... -remote //tmp/file.txt
```

## 编译

```bash
go build -o win-sshpass.exe .
```

## 依赖

- Go 1.23+
- golang.org/x/crypto/ssh
- github.com/pkg/sftp
- github.com/schollz/progressbar/v3
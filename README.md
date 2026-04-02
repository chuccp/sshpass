# sshpass

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md)

A Windows implementation of sshpass, providing similar functionality to the Linux sshpass tool.

## Quick Start

```bash
# Password login and execute command
sshpass -p 'password' ssh user@example.com 'whoami'

# Private key login and execute command
sshpass -i ~/.ssh/id_ed25519 ssh user@example.com 'hostname'

# Upload file
sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# Download file
sshpass -h example.com -p 'password' -d -remote /tmp/file.txt -local ./file.txt
```

## Command Format

### SSH Login

```bash
# Password authentication
sshpass -p <password> ssh [user@host] [command]
sshpass -p <password> ssh -p <port> user@host 'command'
sshpass -p <password> ssh -o StrictHostKeyChecking=no user@host

# Private key authentication
sshpass -i <private_key_path> ssh [user@host] [command]

# Password from environment variable
SSHPASS=<password> sshpass -e ssh user@host

# Password from file
echo 'password' > pass.txt
sshpass -f pass.txt ssh user@host

# Configuration file (multi-line format)
sshpass -f server.config
```

### File Transfer

```bash
# Upload file
sshpass -h <host> -p <password> -local <local_path> -remote <remote_path>

# Upload directory (auto-recursive)
sshpass -h <host> -p <password> -local <local_dir> -remote <remote_dir>

# Download file/directory
sshpass -h <host> -p <password> -d -remote <remote_path> -local <local_path>
```

### SCP Style

```bash
sshpass -p <password> scp <local_file> user@host:<remote_path>
sshpass -p <password> scp user@host:<remote_file> <local_path>
```

### Rsync Style

```bash
sshpass -p <password> rsync -avz <local_path> user@host:<remote_path>
```

## Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `-p` | Password | `-p 'secret123'` |
| `-i` | Private key path | `-i ~/.ssh/id_ed25519` |
| `-f` | Password file / config file | `-f pass.txt` |
| `-e` | Read password from SSHPASS env var | `SSHPASS='pass' sshpass -e ssh ...` |
| `-h` | Host address | `-h example.com` |
| `-u` | Username, default: root | `-u ubuntu` |
| `-P` | Port, default: 22 | `-P 2222` |
| `-c` | Command to execute | `-c 'ls -la'` |
| `-local` | Local path (upload/download) | `-local ./file.txt` |
| `-remote` | Remote path (upload/download) | `-remote /tmp/file.txt` |
| `-d` | Download mode | `-d` |
| `-v` | Show version | `-v` |

## Configuration File Format

```yaml
host: example.com
username: root
password: your_password
port: 22
# key: ~/.ssh/id_ed25519  # optional, use private key instead of password
```

Usage:
```bash
sshpass -f server.config -c 'ls -la'
```

## Complete Examples

```bash
# 1. Password login and execute command
sshpass -p 'mypass' ssh root@192.168.1.100 'docker ps'

# 2. Private key login and execute sudo command
sshpass -i ~/.ssh/id_ed25519 ssh ubuntu@server.com 'sudo systemctl restart nginx'

# 3. Upload entire directory to server
sshpass -h server.com -p 'mypass' -local ./dist -remote /var/www/html

# 4. Download server log directory
sshpass -h server.com -p 'mypass' -d -remote /var/log/nginx -local ./logs

# 5. SCP upload file
sshpass -p 'mypass' scp ./app.jar user@server.com:/opt/app/

# 6. Password via environment variable (more secure)
export SSHPASS='mypass'
sshpass -e ssh user@server.com 'whoami'
```

## Git Bash Notes

Use `//` prefix for remote paths to avoid path conversion:
```bash
# Wrong: /tmp will be converted to Windows path
sshpass ... -remote /tmp/file.txt

# Correct: use double slashes
sshpass ... -remote //tmp/file.txt
```

## Build

```bash
go build -o sshpass.exe .
```

## Dependencies

- Go 1.23+
- golang.org/x/crypto/ssh
- github.com/pkg/sftp
- github.com/schollz/progressbar/v3
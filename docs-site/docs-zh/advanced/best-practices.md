# 最佳实践

## 安全建议

### 1. 避免在命令行中直接传递密码

```bash
# 不推荐：密码会出现在命令历史中
win-sshpass -p 'mypassword' ssh user@host

# 推荐：使用环境变量
export SSHPASS='mypassword'
win-sshpass -e ssh user@host

# 推荐：使用密码文件
win-sshpass -f pass.txt ssh user@host

# 推荐：使用配置文件
win-sshpass -f server.config ssh user@host
```

### 2. 使用私钥认证

私钥认证比密码认证更安全：

```bash
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host
```

### 3. 保护配置文件权限

```bash
# Linux/macOS
chmod 600 server.config

# Windows（PowerShell）
$acl = Get-Acl server.config
$acl.SetAccessRuleProtection($true, $false)
$rule = New-Object System.Security.AccessControl.FileSystemAccessRule($env:USERNAME, "FullControl", "Allow")
$acl.AddAccessRule($rule)
Set-Acl server.config $acl
```

### 4. 启用主机密钥验证

在生产环境中，建议启用严格主机密钥验证：

```bash
win-sshpass -k -f server.config ssh user@host
```

或在配置文件中：

```yaml
strict_host_key: true
```

## 效率建议

### 1. 使用配置文件管理多台服务器

为常用服务器创建配置文件，避免重复输入参数：

```bash
# ~/.ssh/prod-web.config
host: web.example.com
username: deploy
key: ~/.ssh/id_ed25519

# ~/.ssh/prod-db.config
host: db.example.com
username: admin
key: ~/.ssh/id_ed25519
```

### 2. 批量操作

结合 Shell 脚本进行批量操作：

```bash
#!/bin/bash
for host in web1 web2 web3; do
    win-sshpass -f ~/.ssh/$host.config 'sudo systemctl restart nginx' &
done
wait
```

### 3. 使用 SSH 风格语法

对于熟悉 SSH 的用户，可以使用更自然的语法：

```bash
# 标准 SSH 语法
win-sshpass -p 'pass' ssh user@host 'command'

# SCP 语法
win-sshpass -p 'pass' scp file.txt user@host:/tmp/

# Rsync 语法
win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/
```

### 4. 设置合理的超时

```bash
# 快速命令：短超时
win-sshpass -p 'pass' -ct 5 -t 10 ssh user@host 'echo ok'

# 长时间操作：长超时或无超时
win-sshpass -p 'pass' -t 300 ssh user@host 'backup.sh'
```

### 5. 使用 SSH Agent 实现无缝认证

当本地 ssh-agent 中已加载密钥时，win-sshpass 会自动检测 —— 无需 `-p` 或 `-i`：

```bash
# 将密钥加载到 agent 中
ssh-add ~/.ssh/id_ed25519

# 无需任何凭据参数即可连接
win-sshpass ssh user@host 'whoami'

# 适用于所有操作类型
win-sshpass scp file.txt user@host:/tmp/
win-sshpass rsync -avz ./ user@host:/backup/
win-sshpass -h host -local file.txt -remote /tmp/file.txt
```

这是最安全的方法 —— 凭据永远不会出现在命令历史或配置文件中。

### 6. 使用端口转发实现安全访问

通过跳板机访问内部服务，无需将其暴露在公网：

```bash
# 通过跳板机访问内部数据库
win-sshpass -i ~/.ssh/id_ed25519 -L 3306:db.internal:3306 ssh user@jumphost

# 访问多个服务
win-sshpass -A -L 8080:app.internal:80 -L 6379:redis.internal:6379 ssh user@jumphost

# 将本地开发服务器暴露给远程服务器进行测试
win-sshpass -i ~/.ssh/id_ed25519 -R 9090:localhost:3000 ssh user@dev-server
```

!!! tip "结合 agent 转发使用"
    使用 `-A` 配合 `-L`/`-R`，让跳板机也能使用你的本地密钥进行进一步的 SSH 跳转。

## 故障排查

### 连接失败

```bash
# 增加重试次数
win-sshpass -p 'pass' -retry 5 ssh user@host

# 增加连接超时
win-sshpass -p 'pass' -ct 30 ssh user@host
```

### 认证失败

- 检查密码是否正确
- 检查私钥路径是否正确
- 检查远程服务器是否允许密码/密钥认证
- 注意：不支持加密的私钥
- 对于需要键盘交互式认证的 PAM/Cisco 服务器：win-sshpass 会自动回退 — 无需额外参数
- 对于 ssh-agent 问题：确保 agent 正在运行（`ssh-add -l`）并已加载密钥（`ssh-add ~/.ssh/your_key`）

### Git Bash 路径问题

```bash
# 错误：/tmp 会被 Git Bash 转换
win-sshpass ... -remote /tmp/file.txt

# 正确：使用双斜杠
win-sshpass ... -remote //tmp/file.txt
```

## JSON 输出模式

win-sshpass 支持 `-json` 标志，将命令执行结果以结构化 JSON 格式输出到 stdout。这非常适合 AI Agent 和自动化脚本解析。

### 基本用法

```bash
win-sshpass -json -p 'password' ssh user@host 'whoami && uptime'
```

### 成功输出示例

```json
{
  "success": true,
  "host": "user@host",
  "command": "whoami && uptime",
  "exit_code": 0,
  "stdout": "root\n14:32:45 up 3 days,  2:15,  1 user,  load average: 0.00, 0.01, 0.05",
  "duration_ms": 1245
}
```

### 失败输出示例

```json
{
  "success": false,
  "host": "user@host",
  "command": "ls /nonexistent",
  "exit_code": 2,
  "stdout": "",
  "stderr": "ls: cannot access '/nonexistent': No such file or directory",
  "error": "command exited with code 2",
  "duration_ms": 892
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | bool | 操作是否成功 |
| `host` | string | 目标主机（user@host） |
| `command` | string | 执行的命令 |
| `exit_code` | int | 远程命令退出码（0=成功，-1=连接失败） |
| `stdout` | string | 命令的标准输出 |
| `stderr` | string | 命令的标准错误输出（为空时省略） |
| `error` | string | 错误摘要（成功时省略） |
| `duration_ms` | int64 | 执行耗时（毫秒） |

### 支持的命令

JSON 模式支持所有非交互式命令：

```bash
# SSH 命令执行
win-sshpass -json -p 'pass' ssh user@host 'ls -la'

# 文件传输
win-sshpass -json -p 'pass' -h host -local file.txt -remote /tmp/file.txt
win-sshpass -json -p 'pass' -h host -local /tmp/file.txt -remote file.txt -d

# SCP / Rsync
win-sshpass -json -p 'pass' scp file.txt user@host:/tmp/
win-sshpass -json -p 'pass' rsync -avz ./ user@host:/backup/

# 文件哈希校验
win-sshpass -json hash sha256 file.txt
win-sshpass -json verify sha256 <hash> file.txt

# 密钥生成
win-sshpass -json keygen -out ~/.ssh/mykey
```

!!! warning "交互式 Shell 不支持 JSON 模式"
    JSON 模式需要捕获完整输出后再返回，因此不适用于交互式 Shell 会话。如果在不提供命令的情况下使用 `-json`，将返回错误。

## 下一步

- [Go SDK](sdk.md) - 以编程方式使用
- [更新日志](../changelog.md) - 版本更新记录

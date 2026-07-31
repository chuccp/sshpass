# Best Practices

## Security Tips

### 1. Avoid Passing Passwords on the Command Line

```bash
# Not recommended: password appears in command history
win-sshpass -p 'mypassword' ssh user@host

# Recommended: use environment variable
export SSHPASS='mypassword'
win-sshpass -e ssh user@host

# Recommended: use password file
win-sshpass -f pass.txt ssh user@host

# Recommended: use config file
win-sshpass -f server.config ssh user@host
```

### 2. Use Private Key Authentication

Private key authentication is more secure than password authentication:

```bash
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host
```

### 3. Protect Config File Permissions

```bash
# Linux/macOS
chmod 600 server.config

# Windows (PowerShell)
$acl = Get-Acl server.config
$acl.SetAccessRuleProtection($true, $false)
$rule = New-Object System.Security.AccessControl.FileSystemAccessRule($env:USERNAME, "FullControl", "Allow")
$acl.AddAccessRule($rule)
Set-Acl server.config $acl
```

### 4. Enable Host Key Verification

In production environments, it's recommended to enable strict host key verification:

```bash
win-sshpass -k -f server.config ssh user@host
```

Or in config file:

```yaml
strict_host_key: true
```

## Efficiency Tips

### 1. Use Config Files for Server Management

Create config files for frequently used servers to avoid repeating parameters:

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

### 2. Batch Operations

Combine with shell scripts for batch operations:

```bash
#!/bin/bash
for host in web1 web2 web3; do
    win-sshpass -f ~/.ssh/$host.config 'sudo systemctl restart nginx' &
done
wait
```

### 3. Use SSH-Style Syntax

For users familiar with SSH, a more natural syntax is available:

```bash
# Standard SSH syntax
win-sshpass -p 'pass' ssh user@host 'command'

# SCP syntax
win-sshpass -p 'pass' scp file.txt user@host:/tmp/

# Rsync syntax
win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/
```

### 4. Set Reasonable Timeouts

```bash
# Quick commands: short timeout
win-sshpass -p 'pass' -ct 5 -t 10 ssh user@host 'echo ok'

# Long operations: long timeout or no timeout
win-sshpass -p 'pass' -t 300 ssh user@host 'backup.sh'
```

### 5. Use SSH Agent for Seamless Authentication

When you have keys loaded in your local ssh-agent, win-sshpass auto-detects them — no `-p` or `-i` needed:

```bash
# Load your key into the agent
ssh-add ~/.ssh/id_ed25519

# Connect without any credential flags
win-sshpass ssh user@host 'whoami'

# Works for all operation types
win-sshpass scp file.txt user@host:/tmp/
win-sshpass rsync -avz ./ user@host:/backup/
win-sshpass -h host -local file.txt -remote /tmp/file.txt
```

This is the most secure method — credentials never appear in command history or config files.

### 6. Use Port Forwarding for Secure Access

Access internal services through a jumphost without exposing them to the public internet:

```bash
# Access an internal database via a jumphost
win-sshpass -i ~/.ssh/id_ed25519 -L 3306:db.internal:3306 ssh user@jumphost

# Access multiple services
win-sshpass -A -L 8080:app.internal:80 -L 6379:redis.internal:6379 ssh user@jumphost

# Expose a local dev server to a remote server for testing
win-sshpass -i ~/.ssh/id_ed25519 -R 9090:localhost:3000 ssh user@dev-server
```

!!! tip "Combine with agent forwarding"
    Use `-A` with `-L`/`-R` to enable the jumphost itself to use your local keys for further SSH hops.

## Troubleshooting

### Connection Failures

```bash
# Increase retry count
win-sshpass -p 'pass' -retry 5 ssh user@host

# Increase connection timeout
win-sshpass -p 'pass' -ct 30 ssh user@host
```

### Authentication Failures

- Verify the password is correct
- Verify the private key path is correct
- Check if the remote server allows password/key authentication
- Note: encrypted private keys are not supported
- For PAM/Cisco servers that require keyboard-interactive auth: win-sshpass automatically falls back — no extra flags needed
- For ssh-agent issues: ensure the agent is running (`ssh-add -l`) and has keys loaded (`ssh-add ~/.ssh/your_key`)

### Git Bash Path Issues

```bash
# Wrong: /tmp will be converted by Git Bash
win-sshpass ... -remote /tmp/file.txt

# Correct: use double slashes
win-sshpass ... -remote //tmp/file.txt
```

## JSON Output Mode

win-sshpass supports the `-json` flag, which outputs command execution results as structured JSON to stdout. This is ideal for AI agents and automation scripts to parse.

### Basic Usage

```bash
win-sshpass -json -p 'password' ssh user@host 'whoami && uptime'
```

### Success Output Example

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

### Failure Output Example

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

### Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `success` | bool | Whether the operation succeeded |
| `host` | string | Target host (user@host) |
| `command` | string | The command executed |
| `exit_code` | int | Remote command exit code (0=success, -1=connection failure) |
| `stdout` | string | Command stdout |
| `stderr` | string | Command stderr (omitted when empty) |
| `error` | string | Error summary (omitted on success) |
| `duration_ms` | int64 | Execution duration in milliseconds |

### Supported Commands

JSON mode supports all non-interactive commands:

```bash
# SSH command execution
win-sshpass -json -p 'pass' ssh user@host 'ls -la'

# File transfer
win-sshpass -json -p 'pass' -h host -local file.txt -remote /tmp/file.txt
win-sshpass -json -p 'pass' -h host -local /tmp/file.txt -remote file.txt -d

# SCP / Rsync
win-sshpass -json -p 'pass' scp file.txt user@host:/tmp/
win-sshpass -json -p 'pass' rsync -avz ./ user@host:/backup/

# File hash verification
win-sshpass -json hash sha256 file.txt
win-sshpass -json verify sha256 <hash> file.txt

# Key generation
win-sshpass -json keygen -out ~/.ssh/mykey
```

!!! warning "Interactive Shell Not Supported in JSON Mode"
    JSON mode requires capturing the full output before returning, so it is not suitable for interactive shell sessions. Using `-json` without a command will return an error.

## Next Steps

- [Go SDK](sdk.md) - Use programmatically
- [Changelog](../changelog.md) - Version history

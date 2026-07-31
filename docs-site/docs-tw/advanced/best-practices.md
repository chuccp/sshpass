# 最佳實踐

## 安全建議

### 1. 避免在命令列中直接傳遞密碼

```bash
# 不推薦：密碼會出現在命令歷史中
win-sshpass -p 'mypassword' ssh user@host

# 推薦：使用環境變數
export SSHPASS='mypassword'
win-sshpass -e ssh user@host

# 推薦：使用密碼檔案
win-sshpass -f pass.txt ssh user@host

# 推薦：使用設定檔
win-sshpass -f server.config ssh user@host
```

### 2. 使用私鑰認證

私鑰認證比密碼認證更安全：

```bash
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host
```

### 3. 保護設定檔權限

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

### 4. 啟用主機金鑰驗證

在正式環境中，建議啟用嚴格主機金鑰驗證：

```bash
win-sshpass -k -f server.config ssh user@host
```

或在設定檔中：

```yaml
strict_host_key: true
```

## 效率建議

### 1. 使用設定檔管理多台伺服器

為常用伺服器建立設定檔，避免重複輸入參數：

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

### 2. 批次操作

結合 Shell 指令碼進行批次操作：

```bash
#!/bin/bash
for host in web1 web2 web3; do
    win-sshpass -f ~/.ssh/$host.config 'sudo systemctl restart nginx' &
done
wait
```

### 3. 使用 SSH 風格語法

對於熟悉 SSH 的使用者，可以使用更自然的語法：

```bash
# 標準 SSH 語法
win-sshpass -p 'pass' ssh user@host 'command'

# SCP 語法
win-sshpass -p 'pass' scp file.txt user@host:/tmp/

# Rsync 語法
win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/
```

### 4. 設定合理的逾時

```bash
# 快速命令：短逾時
win-sshpass -p 'pass' -ct 5 -t 10 ssh user@host 'echo ok'

# 長時間操作：長逾時或無逾時
win-sshpass -p 'pass' -t 300 ssh user@host 'backup.sh'
```

### 5. 使用 SSH Agent 實現無縫認證

當本機 ssh-agent 中已載入金鑰時，win-sshpass 會自動檢測 —— 無需 `-p` 或 `-i`：

```bash
# 將金鑰載入到 agent 中
ssh-add ~/.ssh/id_ed25519

# 無需任何憑據參數即可連線
win-sshpass ssh user@host 'whoami'

# 適用於所有操作類型
win-sshpass scp file.txt user@host:/tmp/
win-sshpass rsync -avz ./ user@host:/backup/
win-sshpass -h host -local file.txt -remote /tmp/file.txt
```

這是最安全的方法 —— 憑據永遠不會出現在命令歷史或設定檔中。

### 6. 使用連接埠轉送實現安全存取

透過跳板機存取內部服務，無需將其暴露在公網：

```bash
# 透過跳板機存取內部資料庫
win-sshpass -i ~/.ssh/id_ed25519 -L 3306:db.internal:3306 ssh user@jumphost

# 存取多個服務
win-sshpass -A -L 8080:app.internal:80 -L 6379:redis.internal:6379 ssh user@jumphost

# 將本機開發伺服器暴露給遠端伺服器進行測試
win-sshpass -i ~/.ssh/id_ed25519 -R 9090:localhost:3000 ssh user@dev-server
```

!!! tip "結合 agent 轉送使用"
    使用 `-A` 配合 `-L`/`-R`，讓跳板機也能使用你的本機金鑰進行進一步的 SSH 跳轉。

## 故障排除

### 連線失敗

```bash
# 增加重試次數
win-sshpass -p 'pass' -retry 5 ssh user@host

# 增加連線逾時
win-sshpass -p 'pass' -ct 30 ssh user@host
```

### 認證失敗

- 檢查密碼是否正確
- 檢查私鑰路徑是否正確
- 檢查遠端伺服器是否允許密碼/金鑰認證
- 注意：不支援加密的私鑰
- 對於需要鍵盤互動式認證的 PAM/Cisco 伺服器：win-sshpass 會自動回退 — 無需額外參數
- 對於 ssh-agent 問題：確保 agent 正在執行（`ssh-add -l`）並已載入金鑰（`ssh-add ~/.ssh/your_key`）

### Git Bash 路徑問題

```bash
# 錯誤：/tmp 會被 Git Bash 轉換
win-sshpass ... -remote /tmp/file.txt

# 正確：使用雙斜線
win-sshpass ... -remote //tmp/file.txt
```

## JSON 輸出模式

win-sshpass 支援 `-json` 標誌，將命令執行結果以結構化 JSON 格式輸出到 stdout。這非常適合 AI Agent 和自動化腳本解析。

### 基本用法

```bash
win-sshpass -json -p 'password' ssh user@host 'whoami && uptime'
```

### 成功輸出範例

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

### 失敗輸出範例

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

### 欄位說明

| 欄位 | 類型 | 說明 |
|------|------|------|
| `success` | bool | 操作是否成功 |
| `host` | string | 目標主機（user@host） |
| `command` | string | 執行的命令 |
| `exit_code` | int | 遠端命令退出碼（0=成功，-1=連接失敗） |
| `stdout` | string | 命令的標準輸出 |
| `stderr` | string | 命令的標準錯誤輸出（為空時省略） |
| `error` | string | 錯誤摘要（成功時省略） |
| `duration_ms` | int64 | 執行耗時（毫秒） |

### 支援的命令

JSON 模式支援所有非互動式命令：

```bash
# SSH 命令執行
win-sshpass -json -p 'pass' ssh user@host 'ls -la'

# 檔案傳輸
win-sshpass -json -p 'pass' -h host -local file.txt -remote /tmp/file.txt
win-sshpass -json -p 'pass' -h host -local /tmp/file.txt -remote file.txt -d

# SCP / Rsync
win-sshpass -json -p 'pass' scp file.txt user@host:/tmp/
win-sshpass -json -p 'pass' rsync -avz ./ user@host:/backup/

# 檔案雜湊校驗
win-sshpass -json hash sha256 file.txt
win-sshpass -json verify sha256 <hash> file.txt

# 金鑰生成
win-sshpass -json keygen -out ~/.ssh/mykey
```

!!! warning "互動式 Shell 不支援 JSON 模式"
    JSON 模式需要捕獲完整輸出後再返回，因此不適用於互動式 Shell 工作階段。如果在不提供命令的情況下使用 `-json`，將返回錯誤。

## 下一步

- [Go SDK](sdk.md) - 以程式設計方式使用
- [更新日誌](../changelog.md) - 版本更新記錄

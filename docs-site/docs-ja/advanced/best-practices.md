# ベストプラクティス

## セキュリティのヒント

### 1. コマンドラインでパスワードを渡すのを避ける

```bash
# 非推奨：パスワードがコマンド履歴に残る
win-sshpass -p 'mypassword' ssh user@host

# 推奨：環境変数を使用
export SSHPASS='mypassword'
win-sshpass -e ssh user@host

# 推奨：パスワードファイルを使用
win-sshpass -f pass.txt ssh user@host

# 推奨：設定ファイルを使用
win-sshpass -f server.config ssh user@host
```

### 2. 秘密鍵認証を使用する

秘密鍵認証はパスワード認証よりも安全です：

```bash
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host
```

### 3. 設定ファイルの権限を保護する

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

### 4. ホストキー検証を有効にする

本番環境では、厳密なホストキー検証を有効にすることをお勧めします：

```bash
win-sshpass -k -f server.config ssh user@host
```

または設定ファイルで：

```yaml
strict_host_key: true
```

## 効率化のヒント

### 1. 設定ファイルでサーバーを管理する

よく使うサーバーの設定ファイルを作成し、パラメータの繰り返し入力を避ける：

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

### 2. バッチ操作

シェルスクリプトと組み合わせてバッチ操作を実行：

```bash
#!/bin/bash
for host in web1 web2 web3; do
    win-sshpass -f ~/.ssh/$host.config 'sudo systemctl restart nginx' &
done
wait
```

### 3. SSH スタイル構文を使用する

SSH に慣れたユーザーには、より自然な構文が利用可能です：

```bash
# 標準 SSH 構文
win-sshpass -p 'pass' ssh user@host 'command'

# SCP 構文
win-sshpass -p 'pass' scp file.txt user@host:/tmp/

# Rsync 構文
win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/
```

### 4. 適切なタイムアウトを設定する

```bash
# 短時間のコマンド：短いタイムアウト
win-sshpass -p 'pass' -ct 5 -t 10 ssh user@host 'echo ok'

# 長時間の操作：長いタイムアウトまたはタイムアウトなし
win-sshpass -p 'pass' -t 300 ssh user@host 'backup.sh'
```

### 5. SSH Agent でシームレスな認証

ローカルの ssh-agent に鍵がロードされている場合、win-sshpass が自動検出します — `-p` も `-i` も不要：

```bash
# 鍵をエージェントにロード
ssh-add ~/.ssh/id_ed25519

# 認証情報フラグなしで接続
win-sshpass ssh user@host 'whoami'

# すべての操作タイプで動作
win-sshpass scp file.txt user@host:/tmp/
win-sshpass rsync -avz ./ user@host:/backup/
win-sshpass -h host -local file.txt -remote /tmp/file.txt
```

これは最も安全な方法です — 認証情報がコマンド履歴や設定ファイルに残りません。

### 6. ポート転送で安全なアクセス

内部サービスをインターネットに公開せずにジャンプホスト経由でアクセス：

```bash
# 内部データベースにジャンプホスト経由でアクセス
win-sshpass -i ~/.ssh/id_ed25519 -L 3306:db.internal:3306 ssh user@jumphost

# 複数サービスにアクセス
win-sshpass -A -L 8080:app.internal:80 -L 6379:redis.internal:6379 ssh user@jumphost

# ローカル開発サーバーをリモートサーバーに公開してテスト
win-sshpass -i ~/.ssh/id_ed25519 -R 9090:localhost:3000 ssh user@dev-server
```

!!! tip "エージェント転送と組み合わせる"
    `-A` を `-L`/`-R` と組み合わせて、ジャンプホスト自体があなたのローカルキーを使ってさらに SSH ホップできるようにします。

## トラブルシューティング

### 接続失敗

```bash
# リトライ回数を増やす
win-sshpass -p 'pass' -retry 5 ssh user@host

# 接続タイムアウトを増やす
win-sshpass -p 'pass' -ct 30 ssh user@host
```

### 認証失敗

- パスワードが正しいか確認
- 秘密鍵のパスが正しいか確認
- リモートサーバーがパスワード/鍵認証を許可しているか確認
- 注意：暗号化された秘密鍵はサポートされていません
- PAM/Cisco サーバーでキーボードインタラクティブ認証が必要な場合：win-sshpass が自動的にフォールバックします — 追加フラグ不要
- ssh-agent の問題：エージェントが実行中（`ssh-add -l`）かつ鍵がロードされていること（`ssh-add ~/.ssh/your_key`）を確認

### Git Bash のパス問題

```bash
# 間違い：/tmp は Git Bash によって変換されます
win-sshpass ... -remote /tmp/file.txt

# 正しい：二重スラッシュを使用
win-sshpass ... -remote //tmp/file.txt
```

## JSON 出力モード

win-sshpass は `-json` フラグをサポートしており、コマンド実行結果を構造化 JSON として stdout に出力します。AI エージェントや自動化スクリプトの解析に最適です。

### 基本的な使い方

```bash
win-sshpass -json -p 'password' ssh user@host 'whoami && uptime'
```

### 成功時の出力例

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

### 失敗時の出力例

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

### フィールド説明

| フィールド | 型 | 説明 |
|-----------|------|------|
| `success` | bool | 操作が成功したかどうか |
| `host` | string | ターゲットホスト（user@host） |
| `command` | string | 実行されたコマンド |
| `exit_code` | int | リモートコマンドの終了コード（0=成功、-1=接続失敗） |
| `stdout` | string | コマンドの標準出力 |
| `stderr` | string | コマンドの標準エラー出力（空の場合は省略） |
| `error` | string | エラーサマリー（成功時は省略） |
| `duration_ms` | int64 | 実行時間（ミリ秒） |

### 対応コマンド

JSON モードはすべての非対話型コマンドに対応しています：

```bash
# SSH コマンド実行
win-sshpass -json -p 'pass' ssh user@host 'ls -la'

# ファイル転送
win-sshpass -json -p 'pass' -h host -local file.txt -remote /tmp/file.txt
win-sshpass -json -p 'pass' -h host -local /tmp/file.txt -remote file.txt -d

# SCP / Rsync
win-sshpass -json -p 'pass' scp file.txt user@host:/tmp/
win-sshpass -json -p 'pass' rsync -avz ./ user@host:/backup/

# ファイルハッシュ検証
win-sshpass -json hash sha256 file.txt
win-sshpass -json verify sha256 <hash> file.txt

# 鍵生成
win-sshpass -json keygen -out ~/.ssh/mykey
```

!!! warning "対話型 Shell は JSON モード非対応"
    JSON モードは完全な出力をキャプチャしてから返す必要があるため、対話型 Shell セッションには適していません。コマンドを指定せずに `-json` を使用するとエラーが返されます。

## 次のステップ

- [Go SDK](sdk.md) - プログラムから使用
- [変更履歴](../changelog.md) - バージョン履歴

# sshpass

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md)

Windows 版 sshpass ツール。Linux の sshpass と同様の機能を提供します。

## クイックスタート

```bash
# パスワード認証でコマンド実行
sshpass -p 'password' ssh user@example.com 'whoami'

# 秘密鍵認証でコマンド実行
sshpass -i ~/.ssh/id_ed25519 ssh user@example.com 'hostname'

# ファイルをアップロード
sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# ファイルをダウンロード
sshpass -h example.com -p 'password' -d -remote /tmp/file.txt -local ./file.txt
```

## コマンド形式

### SSH ログイン

```bash
# パスワード認証
sshpass -p <パスワード> ssh [user@host] [コマンド]
sshpass -p <パスワード> ssh -p <ポート> user@host 'コマンド'
sshpass -p <パスワード> ssh -o StrictHostKeyChecking=no user@host

# 秘密鍵認証
sshpass -i <秘密鍵パス> ssh [user@host] [コマンド]

# 環境変数からパスワード読み込み
SSHPASS=<パスワード> sshpass -e ssh user@host

# パスワードファイル
echo 'password' > pass.txt
sshpass -f pass.txt ssh user@host

# 設定ファイル（複数行形式）
sshpass -f server.config
```

### ファイル転送

```bash
# ファイルをアップロード
sshpass -h <ホスト> -p <パスワード> -local <ローカルパス> -remote <リモートパス>

# ディレクトリをアップロード（自動再帰）
sshpass -h <ホスト> -p <パスワード> -local <ローカルディレクトリ> -remote <リモートディレクトリ>

# ファイル/ディレクトリをダウンロード
sshpass -h <ホスト> -p <パスワード> -d -remote <リモートパス> -local <ローカルパス>
```

### SCP スタイル

```bash
sshpass -p <パスワード> scp <ローカルファイル> user@host:<リモートパス>
sshpass -p <パスワード> scp user@host:<リモートファイル> <ローカルパス>
```

### Rsync スタイル

```bash
sshpass -p <パスワード> rsync -avz <ローカルパス> user@host:<リモートパス>
```

## パラメータ

| パラメータ | 説明 | 例 |
|-----------|------|-----|
| `-p` | パスワード | `-p 'secret123'` |
| `-i` | 秘密鍵パス | `-i ~/.ssh/id_ed25519` |
| `-f` | パスワードファイル/設定ファイル | `-f pass.txt` |
| `-e` | 環境変数 SSHPASS からパスワード読み込み | `SSHPASS='pass' sshpass -e ssh ...` |
| `-h` | ホストアドレス | `-h example.com` |
| `-u` | ユーザー名、デフォルト: root | `-u ubuntu` |
| `-P` | ポート、デフォルト: 22 | `-P 2222` |
| `-c` | 実行するコマンド | `-c 'ls -la'` |
| `-local` | ローカルパス（アップロード/ダウンロード） | `-local ./file.txt` |
| `-remote` | リモートパス（アップロード/ダウンロード） | `-remote /tmp/file.txt` |
| `-d` | ダウンロードモード | `-d` |
| `-v` | バージョン表示 | `-v` |

## 設定ファイル形式

```yaml
host: example.com
username: root
password: your_password
port: 22
# key: ~/.ssh/id_ed25519  # オプション、パスワードの代わりに秘密鍵を使用
```

使用方法：
```bash
sshpass -f server.config -c 'ls -la'
```

## 完全な例

```bash
# 1. パスワード認証でコマンド実行
sshpass -p 'mypass' ssh root@192.168.1.100 'docker ps'

# 2. 秘密鍵認証で sudo コマンド実行
sshpass -i ~/.ssh/id_ed25519 ssh ubuntu@server.com 'sudo systemctl restart nginx'

# 3. ディレクトリ全体をサーバーにアップロード
sshpass -h server.com -p 'mypass' -local ./dist -remote /var/www/html

# 4. サーバーのログディレクトリをダウンロード
sshpass -h server.com -p 'mypass' -d -remote /var/log/nginx -local ./logs

# 5. SCP でファイルをアップロード
sshpass -p 'mypass' scp ./app.jar user@server.com:/opt/app/

# 6. 環境変数でパスワードを渡す（より安全）
export SSHPASS='mypass'
sshpass -e ssh user@server.com 'whoami'
```

## Git Bash の注意事項

リモートパスは `//` で始めてパス変換を回避してください：
```bash
# 誤り: /tmp が Windows パスに変換される
sshpass ... -remote /tmp/file.txt

# 正しい: ダブルスラッシュを使用
sshpass ... -remote //tmp/file.txt
```

## ビルド

```bash
go build -o sshpass.exe .
```

## 依存関係

- Go 1.23+
- golang.org/x/crypto/ssh
- github.com/pkg/sftp
- github.com/schollz/progressbar/v3
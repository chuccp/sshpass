# win-win-sshpass

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md)

Windows 版 win-sshpass ツール。Linux の sshpass と同様の機能を提供します。

## ダウンロード

[GitHub Releases](https://github.com/chuccp/win-sshpass/releases) から最新版をダウンロード：

1. [Releases](https://github.com/chuccp/win-sshpass/releases) ページを開く
2. 最新版の `win-sshpass-*.zip` をダウンロード
3. `win-sshpass.exe` を任意の場所に展開

## クイックスタート

```bash
# パスワード認証でコマンド実行
win-sshpass -p 'password' ssh user@example.com 'whoami'

# 秘密鍵認証でコマンド実行
win-sshpass -i ~/.ssh/id_ed25519 ssh user@example.com 'hostname'

# ファイルをアップロード
win-sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# ファイルをダウンロード
win-sshpass -h example.com -p 'password' -d -remote /tmp/file.txt -local ./file.txt
```

## コマンド形式

### SSH ログイン

```bash
# パスワード認証
win-sshpass -p <パスワード> ssh [user@host] [コマンド]
win-sshpass -p <パスワード> ssh -p <ポート> user@host 'コマンド'
win-sshpass -p <パスワード> ssh -o StrictHostKeyChecking=no user@host

# 秘密鍵認証
win-sshpass -i <秘密鍵パス> ssh [user@host] [コマンド]

# 環境変数からパスワード読み込み
SSHPASS=<パスワード> win-sshpass -e ssh user@host

# パスワードファイル
echo 'password' > pass.txt
win-sshpass -f pass.txt ssh user@host

# 設定ファイル（複数行形式）
win-sshpass -f server.config
```

### ファイル転送

```bash
# ファイルをアップロード
win-sshpass -h <ホスト> -p <パスワード> -local <ローカルパス> -remote <リモートパス>

# ディレクトリをアップロード（自動再帰）
win-sshpass -h <ホスト> -p <パスワード> -local <ローカルディレクトリ> -remote <リモートディレクトリ>

# ファイル/ディレクトリをダウンロード
win-sshpass -h <ホスト> -p <パスワード> -d -remote <リモートパス> -local <ローカルパス>
```

### SCP スタイル

```bash
win-sshpass -p <パスワード> scp <ローカルファイル> user@host:<リモートパス>
win-sshpass -p <パスワード> scp user@host:<リモートファイル> <ローカルパス>
```

### Rsync スタイル

```bash
win-sshpass -p <パスワード> rsync -avz <ローカルパス> user@host:<リモートパス>
```

## パラメータ

| パラメータ | 説明 | 例 |
|-----------|------|-----|
| `-p` | パスワード | `-p 'secret123'` |
| `-i` | 秘密鍵パス | `-i ~/.ssh/id_ed25519` |
| `-f` | パスワードファイル/設定ファイル | `-f pass.txt` |
| `-e` | 環境変数 SSHPASS からパスワード読み込み | `SSHPASS='pass' win-sshpass -e ssh ...` |
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
win-sshpass -f server.config -c 'ls -la'
```

## 完全な例

```bash
# 1. パスワード認証でコマンド実行
win-sshpass -p 'mypass' ssh root@192.168.1.100 'docker ps'

# 2. 秘密鍵認証で sudo コマンド実行
win-sshpass -i ~/.ssh/id_ed25519 ssh ubuntu@server.com 'sudo systemctl restart nginx'

# 3. ディレクトリ全体をサーバーにアップロード
win-sshpass -h server.com -p 'mypass' -local ./dist -remote /var/www/html

# 4. サーバーのログディレクトリをダウンロード
win-sshpass -h server.com -p 'mypass' -d -remote /var/log/nginx -local ./logs

# 5. SCP でファイルをアップロード
win-sshpass -p 'mypass' scp ./app.jar user@server.com:/opt/app/

# 6. 環境変数でパスワードを渡す（より安全）
export SSHPASS='mypass'
win-sshpass -e ssh user@server.com 'whoami'
```

## Git Bash の注意事項

リモートパスは `//` で始めてパス変換を回避してください：
```bash
# 誤り: /tmp が Windows パスに変換される
win-sshpass ... -remote /tmp/file.txt

# 正しい: ダブルスラッシュを使用
win-sshpass ... -remote //tmp/file.txt
```

## ビルド

```bash
go build -o win-sshpass.exe .
```

## 依存関係

- Go 1.23+
- golang.org/x/crypto/ssh
- github.com/pkg/sftp
- github.com/schollz/progressbar/v3
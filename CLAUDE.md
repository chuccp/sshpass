# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build

```bash
go build -o win-sshpass.exe .
```

## Architecture

Single-package Go application split by functionality:

- `main.go` - Entry point, flag parsing, connection flow
- `config.go` - `Config` struct and config file parsing
- `ssh.go` - SSH client connection, shell session, command execution
- `sftp.go` - File upload/download via SFTP
- `args.go` - sshpass-style argument parsing (`user@host` format)

## Dependencies

- `golang.org/x/crypto/ssh` - SSH protocol
- `github.com/pkg/sftp` - SFTP file transfer

## Release

Push a `v*` tag to trigger GitHub Actions workflow that builds `sshpass.exe` and creates a release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Git Bash Path Note

Remote paths starting with `/` get converted by Git Bash. Use `//` prefix (e.g., `//root/file.txt`) to avoid this.
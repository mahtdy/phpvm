# phpvm — PHP Version Manager

[![CI](https://github.com/mahtdy/phpvm/actions/workflows/ci.yml/badge.svg)](https://github.com/mahtdy/phpvm/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mahtdy/phpvm)](https://goreportcard.com/report/github.com/mahtdy/phpvm)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/mahtdy/phpvm)](https://github.com/mahtdy/phpvm/releases)

> A cross-platform PHP version manager written in Go. Inspired by nvm, Volta, and Rustup.

---

## Features

- Install any PHP version with a single command
- Switch between versions instantly
- Per-project version pinning via `.phpvmrc`
- Manage `php.ini` settings and extensions
- Built-in Composer management
- Xdebug integration
- Shell auto-completion (bash, zsh, fish, PowerShell)
- Works on Windows, Linux, and macOS

---

## Installation

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/mahtdy/phpvm/main/scripts/install.sh | sh
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/mahtdy/phpvm/main/scripts/install.ps1 | iex
```

### From source

```bash
git clone https://github.com/mahtdy/phpvm.git
cd phpvm
make build
```

---

## Quick Start

```bash
phpvm install 8.4       # install PHP 8.4
phpvm use 8.4           # switch to PHP 8.4
phpvm list              # list installed versions
phpvm current           # show active version
phpvm doctor            # check environment health
```

---

## Command Reference

| Command | Description |
|---------|-------------|
| `phpvm install <version>` | Download and install a PHP version |
| `phpvm use <version>` | Switch active PHP version |
| `phpvm list` | List installed PHP versions |
| `phpvm current` | Show current PHP version |
| `phpvm remove <version>` | Uninstall a PHP version |
| `phpvm update [version]` | Update to latest patch |
| `phpvm local <version>` | Pin version for current project |
| `phpvm doctor` | Environment health check |
| `phpvm composer <cmd>` | Manage Composer |
| `phpvm env <get\|set\|list>` | Manage php.ini settings |
| `phpvm ext <list\|enable\|disable>` | Manage extensions |
| `phpvm xdebug <cmd>` | Manage Xdebug |

---

## Configuration

phpvm stores its data in `~/.phpvm/` by default.

Override with `PHPVM_ROOT` environment variable or `--config` flag.

---

## License

MIT — see [LICENSE](LICENSE)

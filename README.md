# ClawSec 🛡️⚔️

> **AI-Native Unified Offensive Security CLI Platform**

[![CI](https://github.com/jinyimeng01/clawsec/actions/workflows/ci.yml/badge.svg)](https://github.com/jinyimeng01/clawsec/actions)
[![Release](https://github.com/jinyimeng01/clawsec/actions/workflows/release.yml/badge.svg)](https://github.com/jinyimeng01/clawsec/releases)
[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-GPL--3.0-blue.svg)](LICENSE)

**ClawSec** is a unified AI-Native network offensive security testing platform, combining the power of high-performance Go networking with intelligent AI decision-making. It integrates port scanning, vulnerability verification (PoC), password brute-forcing, web crawling, AI-assisted analysis, and security product management into a single cohesive CLI tool.

---

## 📑 Table of Contents

- [Features](#-features)
- [Architecture](#-architecture)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Command Reference](#-command-reference)
  - [`scan` - Network Scanning](#scan---network-scanning)
  - [`poc` - Vulnerability Verification](#poc---vulnerability-verification)
  - [`brute` - Password Brute-forcing](#brute---password-brute-forcing)
  - [`crawl` - Web Crawling](#crawl---web-crawling)
  - [`ai` - AI Security Assistant](#ai---ai-security-assistant)
  - [`product` - Security Product Console](#product---security-product-console)
  - [`mcp` - MCP Server](#mcp---model-context-protocol-server)
  - [`workflow` - Automated Workflows](#workflow---automated-penetration-testing-workflows)
- [Configuration](#-configuration)
- [Output Formats](#-output-formats)
- [Development](#-development)
- [Legal Disclaimer](#-legal-disclaimer)
- [License](#-license)

---

## ✨ Features

| Module | Description | Protocols / Formats |
|--------|-------------|---------------------|
| **Port Scanner** | SYN / Connect / UDP scanning with adaptive rate control and banner grabbing | TCP, UDP, SYN (raw) |
| **PoC Engine** | Nuclei YAML-compatible vulnerability verification engine | HTTP, TCP, UDP, DNS, SSL, WebSocket |
| **Brute Force** | High-performance password brute-forcing | SSH, FTP, RDP, MySQL, Redis, MongoDB, PostgreSQL, MSSQL, SMB, LDAP, HTTP |
| **Web Crawler** | Directory enumeration, JS analysis, parameter fuzzing | HTTP/HTTPS |
| **AI Agent** | Claude-powered intelligent target analysis and exploit chain building | Anthropic MCP |
| **Product Console** | Unified WAF / Scanner / EDR management | SafeLine, X-Ray, CloudWalker, T-Answer, DDR |
| **MCP Server** | Expose security tools to external AI agents via HTTP API | Model Context Protocol |
| **Workflows** | AI-driven automated penetration testing chains | Multi-step orchestration |

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        ClawSec CLI                          │
├──────────┬──────────┬──────────┬──────────┬─────────────────┤
│   scan   │   poc    │  brute   │  crawl   │   ai / product  │
├──────────┴──────────┴──────────┴──────────┴─────────────────┤
│                    Engine Layer (Go)                        │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────────┐  │
│  │ Scanner │ │  PoC    │ │ Brute   │ │   Crawler       │  │
│  │ Engine  │ │ Engine  │ │ Engine  │ │   Engine        │  │
│  └─────────┘ └─────────┘ └─────────┘ └─────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│              Output / Config / AI / MCP Layer               │
│       JSON / CSV / HTML / Markdown / MCP Server             │
└─────────────────────────────────────────────────────────────┘
```

---

## 📦 Installation

### One-Line Installer

**Linux / macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/jinyimeng01/clawsec/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/jinyimeng01/clawsec/main/install.ps1 | iex
```

### Pre-built Binaries

Download the latest release for your platform from the [Releases](https://github.com/jinyimeng01/clawsec/releases) page.

Supported platforms:
- **Linux**: `amd64`, `arm64`
- **macOS**: `amd64`, `arm64` (Apple Silicon)
- **Windows**: `amd64`

```bash
# Example: Linux AMD64
curl -LO https://github.com/jinyimeng01/clawsec/releases/latest/download/clawsec-linux-amd64.tar.gz
tar -xzf clawsec-linux-amd64.tar.gz
chmod +x clawsec
sudo mv clawsec /usr/local/bin/
```

### Build from Source

Requirements: **Go 1.22+**

```bash
git clone https://github.com/jinyimeng01/clawsec.git
cd clawsec
go build -o clawsec ./cmd/clawsec
./clawsec version
```

Cross-compile for all platforms:
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o clawsec-linux-amd64 ./cmd/clawsec

# macOS
GOOS=darwin GOARCH=arm64 go build -o clawsec-darwin-arm64 ./cmd/clawsec

# Windows
GOOS=windows GOARCH=amd64 go build -o clawsec-windows-amd64.exe ./cmd/clawsec
```

---

## 🚀 Quick Start

```bash
# Show help
clawsec --help

# Port scan top 100 ports on a subnet
clawsec scan port -t 10.0.0.0/24 -p top100

# Full port SYN scan with banner grabbing (requires root)
sudo clawsec scan port -t 10.0.0.1 -p 1-65535 --syn --banner

# Run a PoC template against a target
clawsec poc run -t CVE-2021-41773.yaml -u http://target.com

# Run all critical/high PoCs from a directory
clawsec poc run -d ./nuclei-templates/ -u targets.txt -s critical,high

# SSH brute force
clawsec brute ssh -t 10.0.0.1 -u root -P passwords.txt --threads 50

# Directory enumeration
clawsec crawl dir -t http://target.com -w wordlist.txt --ext

# AI-assisted target analysis
clawsec ai analyze -t 10.0.0.1 --context "Apache 2.4.41, PHP 7.4"

# Interactive AI security assistant
clawsec ai chat

# Run automated penetration testing workflow
clawsec workflow run -t 10.0.0.0/24 --objective "find all vulnerabilities"
```

---

## 📖 Command Reference

### `scan` - Network Scanning

High-performance network scanning engine supporting multiple scan modes.

**Subcommands:**
- `port` — TCP/UDP port scanning (SYN/Connect/UDP)
- `service` — Service fingerprinting and version detection
- `web` — Web asset discovery and technology fingerprinting

```bash
# TCP Connect scan of top 100 ports
clawsec scan port -t 10.0.0.0/24

# Full port SYN scan with banner grabbing (root required)
clawsec scan port -t 10.0.0.1 -p 1-65535 --syn --banner

# UDP scan
clawsec scan port -t 10.0.0.1 -p 53,161 --udp

# Service version detection
clawsec scan service -t 10.0.0.1 -p 22,80,443,3306

# Web fingerprinting
clawsec scan web -t urls.txt
```

**Flags:**
| Flag | Description | Default |
|------|-------------|---------|
| `-t, --target` | Target hosts/CIDR/URLs (required) | — |
| `-p, --ports` | Port range (`80,443`, `1-65535`, `top100`, `top1000`) | `top100` |
| `--syn` | Use SYN stealth scan (requires root) | false |
| `--udp` | Use UDP scan | false |
| `--banner` | Grab service banners | false |
| `--rate` | Packets per second rate limit | — |
| `--threads` | Concurrent threads | 50 |
| `--timeout` | Connection timeout (seconds) | 3 |

---

### `poc` - Vulnerability Verification

Execute vulnerability proof-of-concept templates compatible with the Nuclei YAML format.

**Features:**
- Full Nuclei YAML template syntax support
- HTTP / TCP / UDP / DNS / SSL / WebSocket / Headless / Code protocols
- DSL expression engine with 50+ built-in functions
- Multi-step workflow chains with variable passing
- Automatic template updates from community repository

**Subcommands:**
- `run` — Run PoC templates against targets
- `list` — List available PoC templates
- `update` — Update PoC templates from remote repository

```bash
# Run a single template against a target
clawsec poc run -t CVE-2021-41773.yaml -u http://target.com

# Run all templates in a directory against multiple targets
clawsec poc run -d ./poc/ -u targets.txt

# Filter by severity and tags
clawsec poc run -d ./nuclei-templates/ -u targets.txt -s critical,high -t cve,rce

# Update templates from remote repository
clawsec poc update

# List available templates
clawsec poc list
```

---

### `brute` - Password Brute-forcing

High-performance password brute-forcing engine supporting 10+ protocols.

**Supported Protocols:**
`ssh`, `ftp`, `rdp`, `mysql`, `redis`, `mongodb`, `postgres`, `mssql`, `smb`, `ldap`, `http`

```bash
# SSH brute force with password list
clawsec brute ssh -t 10.0.0.1 -u root -P passwords.txt

# Multiple targets with multiple users
clawsec brute ssh -t targets.txt -U users.txt -P passwords.txt --threads 100

# Redis brute force (no username)
clawsec brute redis -t 10.0.0.1 -P passwords.txt

# HTTP Basic auth brute force
clawsec brute http -t http://target.com -u admin -P passwords.txt
```

---

### `crawl` - Web Crawling

Web crawling and directory enumeration engine.

**Subcommands:**
- `dir` — Directory and file enumeration (dirbuster-style)
- `js` — JavaScript file discovery and endpoint extraction
- `params` — Parameter enumeration and fuzzing

```bash
# Directory enumeration with default wordlist
clawsec crawl dir -t http://target.com

# Custom wordlist with smart extensions
clawsec crawl dir -t http://target.com -w /path/to/wordlist.txt --ext -T 50

# JavaScript endpoint extraction
clawsec crawl js -t http://target.com

# Parameter fuzzing
clawsec crawl params -t http://target.com/api
```

---

### `ai` - AI Security Assistant

Interact with the AI security brain for intelligent offensive security analysis. Powered by **Anthropic Claude** via Model Context Protocol (MCP).

**Requirements:**
- [Bun](https://bun.sh) runtime installed
- `ANTHROPIC_API_KEY` environment variable set
- `ai-brain/` TypeScript agent built (`cd ai-brain && bun install`)

**Subcommands:**
- `analyze` — Analyze target and suggest attack paths
- `suggest` — Suggest PoC templates based on fingerprints
- `chain` — Build exploit chains from discovered vulnerabilities
- `report` — Generate professional penetration test reports
- `chat` — Interactive AI security assistant

```bash
# Analyze a target and get attack recommendations
clawsec ai analyze -t 10.0.0.1 --context "Apache 2.4.41, PHP 7.4, MySQL 5.7"

# Suggest PoCs based on service fingerprint
clawsec ai suggest -t http://target.com --fingerprint "Apache/2.4.41, PHP/7.4"

# Generate report from scan results
clawsec ai report -i results.json -o report.md

# Interactive AI assistant
clawsec ai chat
```

---

### `product` - Security Product Console

Manage and interact with various security products from a unified CLI interface.

**Supported Products:**
| Product | Description |
|---------|-------------|
| `safeline` | Chaitin SafeLine WAF |
| `xray` | Chaitin X-Ray vulnerability scanner |
| `cloudwalker` | Chaitin CloudWalker CWPP |
| `tanswer` | Chaitin T-Answer traffic threat detection |
| `ddr` | Chaitin DDR data security |

**Subcommands:**
- `list` — List configured products
- `config` — Configure product credentials
- `query` — Query product data
- `exec` — Execute product commands

```bash
# List configured products
clawsec product list

# Query WAF attack logs
clawsec product query safeline attack_logs

# Block IP on WAF
clawsec product exec safeline block_ip --ip 1.2.3.4
```

---

### `mcp` - Model Context Protocol Server

Run ClawSec as an MCP server to expose security tools to AI agents.

**Endpoints:**
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/mcp/health` | Health check |
| `GET` | `/mcp/tools` | List available tools |
| `POST` | `/mcp/call` | Execute a tool |

```bash
# Start MCP server on port 8080
clawsec mcp serve --port 8080
```

Integrates with Claude, Cursor, and other MCP-compatible clients.

---

### `workflow` - Automated Penetration Testing Workflows

AI-driven automated penetration testing workflows. The workflow engine uses AI to plan and execute multi-step attack chains, integrating port scanning, PoC execution, and vulnerability verification.

```bash
# Full reconnaissance workflow
clawsec workflow run -t 10.0.0.1 --objective "find all vulnerabilities"

# Targeted exploit chain
clawsec workflow run -t http://target.com --objective "achieve RCE"

# Stealth assessment
clawsec workflow run -t 10.0.0.0/24 --strategy stealth
```

---

## ⚙️ Configuration

ClawSec uses a YAML configuration file located at `~/.clawsec/config.yaml`.

**Example configuration:**

```yaml
# ClawSec Configuration File
# https://github.com/jinyimeng01/clawsec

output_format: text
timeout: 5
threads: 50
rate_limit: 150

# Network settings
user_agent: "ClawSec/0.1.0"
random_ua: false
proxy: "http://127.0.0.1:8080"
force_proxy: false
insecure_ssl: false
follow_redirects: true
max_redirects: 10

# Attack settings
authorized: false
stealth: false

# AI settings
ai:
  enabled: false
  endpoint: ""
  model: "claude-sonnet-4-20250514"
  api_key: ""

# Product configurations
# safeline:
#   url: "https://safeline.example.com"
#   api_key: "your-api-key"
# xray:
#   url: "https://xray.example.com"
#   api_key: "your-api-key"
```

Use `-c, --config` to specify a custom config file:
```bash
clawsec -c /path/to/config.yaml scan port -t 10.0.0.1
```

---

## 📊 Output Formats

ClawSec supports multiple output formats via the `-f, --format` flag:

| Format | Description | Use Case |
|--------|-------------|----------|
| `text` | Human-readable colored table output | Interactive terminal use |
| `json` | Single JSON array of all results | Integration with other tools |
| `jsonl` | One JSON object per line (NDJSON) | Streaming processing |
| `csv` | Comma-separated values | Spreadsheet import |
| `markdown` | Markdown table | Documentation |
| `html` | Dark-themed HTML report | Client deliverables |
| `silent` | No output | Shell scripting |

```bash
# JSON output to file
clawsec scan port -t 10.0.0.1 -f json -o results.json

# HTML report
clawsec scan port -t 10.0.0.1 -f html -o report.html

# Silent mode (exit code only)
clawsec scan port -t 10.0.0.1 -s
```

---

## 🛠️ Development

### Prerequisites

- Go 1.22+
- (Optional) Bun runtime for AI features
- (Optional) Docker for containerized builds

### Project Structure

```
clawsec/
├── cmd/clawsec/          # Main entry point
├── internal/
│   ├── cli/              # Cobra CLI commands
│   ├── config/           # Configuration management
│   ├── logger/           # Structured logging
│   ├── output/           # Output formatting (text/json/csv/html)
│   ├── runner/           # Scan execution runner
│   └── constants/        # Version and build info
├── pkg/
│   ├── engine/
│   │   ├── scanner/      # Port scanner (SYN/Connect/UDP)
│   │   ├── poc/          # PoC engine (Nuclei-compatible)
│   │   ├── brute/        # Brute-force protocols
│   │   └── crawler/      # Web crawler
│   ├── ai/               # AI agent integration
│   ├── mcp/              # MCP server implementation
│   └── products/         # Security product adapters
├── ai-brain/             # TypeScript AI brain (MCP)
├── .github/workflows/    # CI/CD (GitHub Actions)
├── install.sh            # Linux/macOS installer
├── install.ps1           # Windows installer
└── README.md
```

### Build & Test

```bash
# Build
go build -o clawsec ./cmd/clawsec

# Run tests
go test -v ./...

# Run with race detector
go test -race ./...

# Lint
golangci-lint run

# Cross-compile (uses GoReleaser or manual GOOS/GOARCH)
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dist/clawsec-linux-amd64 ./cmd/clawsec
```

### AI Brain Setup

```bash
cd ai-brain
bun install
bun run build
# Set ANTHROPIC_API_KEY environment variable
```

---

## ⚠️ Legal Disclaimer

**This tool is intended for authorized security testing only.**

You must have **explicit written permission** to test any target system. Unauthorized access to computer systems is illegal in most jurisdictions. The authors assume no liability for misuse or damage caused by this program.

**Always ensure:**
- You own the target system, or
- You have explicit written authorization from the owner, or
- You are operating in a controlled lab environment

Use at your own risk.

---

## 📄 License

This project is licensed under the **GPL-3.0 License**. See [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

- [Nuclei](https://github.com/projectdiscovery/nuclei) — YAML template format inspiration
- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [Anthropic](https://anthropic.com) — AI model provider
- [Model Context Protocol](https://modelcontextprotocol.io) — AI integration standard

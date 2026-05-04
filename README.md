# ClawSec 🛡️⚔️

> **Global Top-tier AI-Driven Offensive Security CLI Platform**

[![CI](https://github.com/clawsec/clawsec/actions/workflows/ci.yml/badge.svg)](https://github.com/clawsec/clawsec/actions)
[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-GPL--3.0-blue.svg)](LICENSE)

ClawSec is a unified AI-Native network offensive security testing platform, combining the power of high-performance Go networking with intelligent AI decision-making.

## 🚀 Core Capabilities

| Module | Description | Status |
|--------|-------------|--------|
| **Port Scanner** | SYN/Connect/UDP scanning with adaptive rate control | Phase 2 |
| **PoC Engine** | Nuclei YAML compatible vulnerability verification | Phase 3 |
| **Brute Force** | 10+ protocol password brute-forcing | Phase 4 |
| **AI Agent** | Intelligent target analysis and exploit chain building | Phase 5 |
| **Product Console** | Unified WAF/Scanner/EDR management | Phase 6 |

## 📦 Installation

### From Source

```bash
git clone https://github.com/clawsec/clawsec.git
cd clawsec
go build -o clawsec ./cmd/clawsec
```

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/clawsec/clawsec/releases).

## 🎯 Quick Start

```bash
# Port scan
clawsec scan port -t 10.0.0.0/24 -p 1-65535 --rate 10000

# Run PoC template
clawsec poc run -t CVE-2021-41773.yaml -u http://target.com

# SSH brute force
clawsec brute ssh -t 10.0.0.1 -u root -P passwords.txt

# AI-assisted scanning
clawsec scan -t targets.txt --ai

# Show help
clawsec --help
```

## ⚠️ Legal Disclaimer

**This tool is intended for authorized security testing only.** You must have explicit written permission to test any target system. Unauthorized access to computer systems is illegal. Use at your own risk.

## 📄 License

GPL-3.0 License. See [LICENSE](LICENSE) for details.

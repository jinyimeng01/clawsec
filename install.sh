#!/usr/bin/env bash
# ClawSec Installer for Linux/macOS
# Usage: curl -fsSL https://raw.githubusercontent.com/clawsec/clawsec/main/install.sh | bash
# Or: ./install.sh [-v v0.1.0] [-d ~/.local/bin]

set -e

APP_NAME="clawsec"
REPO="clawsec/clawsec"
VERSION="latest"
INSTALL_DIR="${HOME}/.local/bin"
NO_PATH=false
FORCE=false

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info()  { echo -e "${CYAN}[INFO]${NC} $*"; }
ok()    { echo -e "${GREEN}[OK]  ${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
err()   { echo -e "${RED}[ERR] ${NC} $*"; }

usage() {
    echo "Usage: $0 [-v version] [-d dir] [-n] [-f]"
    echo "  -v    Version to install (default: latest)"
    echo "  -d    Install directory (default: ~/.local/bin)"
    echo "  -n    Do not modify PATH"
    echo "  -f    Force reinstall"
    exit 1
}

while getopts "v:d:nfh" opt; do
    case $opt in
        v) VERSION="$OPTARG" ;;
        d) INSTALL_DIR="$OPTARG" ;;
        n) NO_PATH=true ;;
        f) FORCE=true ;;
        h) usage ;;
        *) usage ;;
    esac
done

get_latest_version() {
    curl -fsSL --max-time 10 "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | \
        grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
}

detect_platform() {
    local os arch
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)
    case "$arch" in
        x86_64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        i386|i686) arch="386" ;;
        *) arch="amd64" ;;
    esac
    case "$os" in
        linux|darwin) echo "${os}_${arch}" ;;
        msys*|mingw*|cygwin*) echo "windows_${arch}" ;;
        *) err "Unsupported OS: $os"; exit 1 ;;
    esac
}

download_binary() {
    local version="$1" out_dir="$2"
    local platform=$(detect_platform)
    local asset="${APP_NAME}_${version}_${platform}.tar.gz"
    local url="https://github.com/${REPO}/releases/download/${version}/${asset}"

    info "Downloading ${asset} ..."
    local tmp="$(mktemp)"
    if curl -fsSL --max-time 120 "$url" -o "$tmp"; then
        tar -xzf "$tmp" -C "$out_dir" 2>/dev/null || tar -xf "$tmp" -C "$out_dir"
        rm -f "$tmp"
        ok "Extracted to ${out_dir}"
    else
        rm -f "$tmp"
        return 1
    fi
}

install_from_source() {
    local out_dir="$1"
    info "Building from source ..."

    if ! command -v go >/dev/null 2>&1; then
        err "Go is not installed. Install Go 1.22+ from https://go.dev/dl/"
        exit 1
    fi

    local gover
    gover=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | head -1)
    info "Go version: ${gover}"

    if [ ! -d "cmd/${APP_NAME}" ]; then
        err "Source code not found. Run this script from the repository root."
        exit 1
    fi

    GOOS=$(uname -s | tr '[:upper:]' '[:lower:]')
    GOARCH=$(uname -m)
    case "$GOARCH" in
        x86_64) GOARCH="amd64" ;;
        aarch64|arm64) GOARCH="arm64" ;;
    esac
    export GOOS GOARCH

    go build -ldflags "-s -w" -o "${out_dir}/${APP_NAME}" "./cmd/${APP_NAME}"
    ok "Built successfully"
}

add_to_path() {
    local dir="$1"
    local shell_rc=""
    case "${SHELL##*/}" in
        bash) shell_rc="${HOME}/.bashrc" ;;
        zsh)  shell_rc="${HOME}/.zshrc" ;;
        fish) shell_rc="${HOME}/.config/fish/config.fish" ;;
        *)    shell_rc="${HOME}/.profile" ;;
    esac

    if [ -f "$shell_rc" ]; then
        if ! grep -q "$dir" "$shell_rc" 2>/dev/null; then
            echo "export PATH=\"${dir}:\$PATH\"" >> "$shell_rc"
            ok "Added ${dir} to PATH in ${shell_rc}"
            info "Run 'source ${shell_rc}' to apply changes"
        else
            info "${dir} already in PATH"
        fi
    fi
}

init_config() {
    local config_dir="${HOME}/.clawsec"
    local cfg_file="${config_dir}/config.yaml"
    if [ ! -f "$cfg_file" ]; then
        mkdir -p "$config_dir"
        cat > "$cfg_file" <<'EOF'
# ClawSec Configuration File
# https://github.com/clawsec/clawsec

output_format: text
timeout: 5
threads: 50
rate_limit: 150

# AI settings
ai:
  enabled: false
  endpoint: ""
  model: "claude-sonnet-4-20250514"
  api_key: ""

# Product configurations (uncomment and configure as needed)
# safeline:
#   url: "https://safeline.example.com"
#   api_key: "your-api-key"
# xray:
#   url: "https://xray.example.com"
#   api_key: "your-api-key"
EOF
        ok "Created default config: ${cfg_file}"
    fi
}

# ============ Main ============

echo ""
echo -e "${GREEN}=============================================${NC}"
echo -e "${GREEN}  ClawSec Installer${NC}"
echo -e "${GREEN}  AI-Native Offensive Security Platform${NC}"
echo -e "${GREEN}=============================================${NC}"
echo ""

EXE_PATH="${INSTALL_DIR}/${APP_NAME}"

# Check existing
if [ -f "$EXE_PATH" ] && [ "$FORCE" = false ]; then
    existing="$($EXE_PATH version 2>/dev/null || true)"
    warn "Already installed: ${existing}"
    printf "Reinstall? [y/N] "
    read -r resp
    if [ "$resp" != "y" ] && [ "$resp" != "Y" ]; then
        info "Installation cancelled."
        exit 0
    fi
fi

mkdir -p "$INSTALL_DIR"

# Resolve version
if [ "$VERSION" = "latest" ]; then
    VERSION=$(get_latest_version)
    if [ -z "$VERSION" ]; then
        warn "Could not fetch latest version from GitHub"
        VERSION="v0.1.0-alpha"
    fi
fi
info "Installing version: ${VERSION}"

# Install
SOURCE_MODE=false
if ! download_binary "$VERSION" "$INSTALL_DIR"; then
    warn "Download failed, falling back to source build..."
    install_from_source "$INSTALL_DIR"
    SOURCE_MODE=true
fi

# Verify
if [ -f "$EXE_PATH" ]; then
    ver_out="$($EXE_PATH version 2>/dev/null || true)"
    ok "Installed: ${ver_out}"
else
    err "Installation failed - binary not found in ${INSTALL_DIR}"
    exit 1
fi

# PATH
if [ "$NO_PATH" = false ]; then
    add_to_path "$INSTALL_DIR"
fi

# Config
init_config

echo ""
ok "Installation complete!"
info "Binary location: ${EXE_PATH}"
info "Config directory: ~/.clawsec"
echo ""
echo -e "${GREEN}Quick start:${NC}"
echo "  clawsec scan port -t 127.0.0.1 -p top100"
echo "  clawsec crawl dir -t http://target.com --ext"
echo "  clawsec poc run -u http://target.com --severity critical,high"
echo ""

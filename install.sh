#!/bin/bash
# ABOUTME: Universal installer script for my-docs
# ABOUTME: Downloads and installs the appropriate release for the current platform

set -euo pipefail

REPO="serialexp/my-docs"
APP_NAME="my-docs"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

info() { echo -e "${GREEN}==>${NC} $1"; }
warn() { echo -e "${YELLOW}warning:${NC} $1"; }
error() { echo -e "${RED}error:${NC} $1" >&2; exit 1; }

# Detect OS and architecture
detect_platform() {
    local os arch

    case "$(uname -s)" in
        Linux*)  os="linux" ;;
        Darwin*) os="darwin" ;;
        MINGW*|MSYS*|CYGWIN*) os="windows" ;;
        *) error "Unsupported operating system: $(uname -s)" ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac

    echo "${os}-${arch}"
}

# Get the latest release version from GitHub
get_latest_version() {
    local version
    version=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" |
              grep '"tag_name":' |
              sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

    if [[ -z "$version" ]]; then
        error "Failed to fetch latest version from GitHub"
    fi

    echo "$version"
}

# Install on Unix-like systems (Linux/macOS)
install_unix() {
    local version="$1"
    local platform="$2"
    local url="https://github.com/${REPO}/releases/download/${version}/${APP_NAME}-${platform}"
    local tmp_dir
    tmp_dir=$(mktemp -d)

    local bin_dir="${HOME}/.local/bin"
    mkdir -p "$bin_dir"

    # Check if ~/.local/bin is in PATH and offer to add it
    if [[ ":$PATH:" != *":${bin_dir}:"* ]]; then
        warn "${bin_dir} is not in your PATH"
        echo ""

        local path_line='export PATH="$HOME/.local/bin:$PATH"'
        local shell_configs=()

        # Detect which shell configs exist
        [[ -f "${HOME}/.zshrc" ]] && shell_configs+=("${HOME}/.zshrc")
        [[ -f "${HOME}/.bashrc" ]] && shell_configs+=("${HOME}/.bashrc")
        [[ -f "${HOME}/.bash_profile" ]] && shell_configs+=("${HOME}/.bash_profile")

        if [[ ${#shell_configs[@]} -gt 0 ]]; then
            echo "  Found shell config(s): ${shell_configs[*]}"
            echo ""
            read -p "  Add ~/.local/bin to PATH in these files? [Y/n] " -n 1 -r
            echo ""

            if [[ ! $REPLY =~ ^[Nn]$ ]]; then
                for rc in "${shell_configs[@]}"; do
                    if ! grep -qF '/.local/bin' "$rc"; then
                        echo "" >> "$rc"
                        echo "$path_line" >> "$rc"
                        info "Added PATH to ${rc}"
                    else
                        info "${rc} already contains .local/bin"
                    fi
                done
                echo ""
                echo "  Restart your shell or run: source ${shell_configs[0]}"
                echo ""
            else
                echo ""
                echo "  You can add it manually later:"
                echo "    $path_line"
                echo ""
            fi
        else
            echo "  No shell config found. Add this to your shell config:"
            echo "    $path_line"
            echo ""
        fi
    fi

    info "Downloading ${APP_NAME} ${version} for ${platform}..."
    curl -sL "$url" -o "${tmp_dir}/${APP_NAME}"

    if [[ ! -f "${tmp_dir}/${APP_NAME}" ]]; then
        error "Failed to download binary"
    fi

    # Make it executable
    chmod +x "${tmp_dir}/${APP_NAME}"

    info "Installing binary to ${bin_dir}..."
    cp "${tmp_dir}/${APP_NAME}" "${bin_dir}/${APP_NAME}"

    rm -rf "$tmp_dir"

    info "Installation complete!"
    echo ""
    echo "  Binary installed to: ${bin_dir}/${APP_NAME}"
    echo ""

    # Ask if user wants to install Claude instructions
    echo "  Would you like to install my-docs instructions into ~/.claude/CLAUDE.md?"
    echo "  This allows Claude Code to use my-docs for documentation lookups."
    echo ""
    read -p "  Install Claude instructions? [Y/n] " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        "${bin_dir}/${APP_NAME}" install
    else
        echo ""
        echo "  Skipped. You can run '${APP_NAME} install' later to add Claude instructions."
    fi

    echo ""
    echo "  Commands:"
    echo "    ${APP_NAME} find <query>                  # Search for repos"
    echo "    ${APP_NAME} search <owner/repo> <pattern> # Search repo contents"
    echo "    ${APP_NAME} cat <owner/repo> <path>       # Read files from repo"
    echo "    ${APP_NAME} help                          # Show all commands"
    echo ""
}

# Install on Windows
install_windows() {
    local version="$1"
    local url="https://github.com/${REPO}/releases/download/${version}/${APP_NAME}-windows-amd64.exe"

    echo ""
    echo "Windows installation via this script is not fully supported."
    echo "Please download manually from:"
    echo "  $url"
    echo ""
    echo "Or use PowerShell:"
    echo "  Invoke-WebRequest -Uri '$url' -OutFile '${APP_NAME}.exe'"
    echo ""
    echo "Then move ${APP_NAME}.exe to a directory in your PATH."
}

main() {
    echo ""
    echo "  ╔═══════════════════════════════════════╗"
    echo "  ║        my-docs Installer              ║"
    echo "  ╚═══════════════════════════════════════╝"
    echo ""

    local platform
    platform=$(detect_platform)
    info "Detected platform: ${platform}"

    local version
    version=$(get_latest_version)
    info "Latest version: ${version}"

    case "$platform" in
        linux-amd64|linux-arm64|darwin-amd64|darwin-arm64)
            install_unix "$version" "$platform"
            ;;
        windows-amd64)
            install_windows "$version"
            ;;
        *)
            error "No installation method for platform: ${platform}"
            ;;
    esac
}

main "$@"

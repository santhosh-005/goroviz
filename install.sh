#!/bin/sh
# install.sh — Cross-platform installer for Goroviz
#
# Usage:
#   curl -sSfL https://raw.githubusercontent.com/santhosh-005/goroviz/main/install.sh | sh
#
# This script:
#   1. Detects your OS and architecture
#   2. Downloads the latest release binary from GitHub
#   3. Installs it to ~/.local/bin (or /usr/local/bin with sudo)
#   4. Verifies it's available on PATH
#
# Supported platforms:
#   - Linux  (amd64, arm64)
#   - macOS  (amd64, arm64)

set -e

REPO="santhosh-005/goroviz"
BINARY_NAME="goroviz"
INSTALL_DIR="${HOME}/.local/bin"

# --- Helper functions ---

info() {
    printf "\033[1;34m==>\033[0m \033[1m%s\033[0m\n" "$1"
}

success() {
    printf "\033[1;32m✓\033[0m %s\n" "$1"
}

error() {
    printf "\033[1;31m✗ Error:\033[0m %s\n" "$1" >&2
    exit 1
}

warn() {
    printf "\033[1;33m!\033[0m %s\n" "$1"
}

# --- Detect OS ---

detect_os() {
    OS="$(uname -s)"
    case "${OS}" in
        Linux*)     OS="linux" ;;
        Darwin*)    OS="darwin" ;;
        *)          error "Unsupported operating system: ${OS}. Goroviz supports Linux and macOS." ;;
    esac
}

# --- Detect Architecture ---

detect_arch() {
    ARCH="$(uname -m)"
    case "${ARCH}" in
        x86_64|amd64)   ARCH="amd64" ;;
        aarch64|arm64)  ARCH="arm64" ;;
        *)              error "Unsupported architecture: ${ARCH}. Goroviz supports amd64 and arm64." ;;
    esac
}

# --- Check for required tools ---

check_dependencies() {
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        error "Either 'curl' or 'wget' is required to download Goroviz."
    fi
    if ! command -v tar >/dev/null 2>&1; then
        error "'tar' is required to extract the release archive."
    fi
}

# --- Download helper (curl or wget) ---

download() {
    url="$1"
    output="$2"
    if command -v curl >/dev/null 2>&1; then
        curl -sSfL -o "${output}" "${url}"
    elif command -v wget >/dev/null 2>&1; then
        wget -q -O "${output}" "${url}"
    fi
}

# --- Get latest release tag from GitHub API ---

get_latest_version() {
    LATEST_URL="https://api.github.com/repos/${REPO}/releases/latest"

    if command -v curl >/dev/null 2>&1; then
        VERSION="$(curl -sSfL "${LATEST_URL}" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/')"
    elif command -v wget >/dev/null 2>&1; then
        VERSION="$(wget -qO- "${LATEST_URL}" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/')"
    fi

    if [ -z "${VERSION}" ]; then
        error "Failed to fetch the latest release version from GitHub.\n  Check your internet connection or visit: https://github.com/${REPO}/releases"
    fi
}

# --- Main installation logic ---

main() {
    printf "\n"
    info "Installing Goroviz..."
    printf "\n"

    check_dependencies
    detect_os
    detect_arch

    success "Detected platform: ${OS}/${ARCH}"

    # Fetch latest version
    info "Fetching latest release version..."
    get_latest_version
    success "Latest version: ${VERSION}"

    # Construct download URL
    # Expected asset name pattern: goroviz_<version>_<os>_<arch>.tar.gz
    VERSION_NO_V="$(echo "${VERSION}" | sed 's/^v//')"
    ARCHIVE_NAME="${BINARY_NAME}_${VERSION_NO_V}_${OS}_${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"

    # Create temp directory for download
    TMP_DIR="$(mktemp -d)"
    trap 'rm -rf "${TMP_DIR}"' EXIT

    # Download the archive
    info "Downloading ${ARCHIVE_NAME}..."
    download "${DOWNLOAD_URL}" "${TMP_DIR}/${ARCHIVE_NAME}"

    if [ ! -f "${TMP_DIR}/${ARCHIVE_NAME}" ]; then
        error "Download failed. The release asset may not exist for your platform.\n  Check available assets at: https://github.com/${REPO}/releases/tag/${VERSION}"
    fi

    success "Downloaded successfully"

    # Extract the binary
    info "Extracting..."
    tar -xzf "${TMP_DIR}/${ARCHIVE_NAME}" -C "${TMP_DIR}"

    # Find the binary (may be in a subdirectory)
    EXTRACTED_BIN="$(find "${TMP_DIR}" -name "${BINARY_NAME}" -type f | head -1)"

    if [ -z "${EXTRACTED_BIN}" ]; then
        error "Expected binary '${BINARY_NAME}' not found in the archive."
    fi

    # Install the binary
    mkdir -p "${INSTALL_DIR}"
    mv "${EXTRACTED_BIN}" "${INSTALL_DIR}/${BINARY_NAME}"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

    success "Installed to ${INSTALL_DIR}/${BINARY_NAME}"

    # Verify PATH
    printf "\n"
    if command -v "${BINARY_NAME}" >/dev/null 2>&1; then
        INSTALLED_VERSION="$(${BINARY_NAME} --help 2>&1 | head -1 || true)"
        success "goroviz is available on your PATH"
        printf "\n"
        info "You're all set! Try it out:"
        printf "\n"
        printf "    goroviz attach localhost:6060\n"
        printf "\n"
    else
        warn "goroviz was installed to ${INSTALL_DIR}, but it's not on your PATH."
        printf "\n"
        printf "  Add it to your PATH by adding this line to your shell config:\n"
        printf "\n"

        SHELL_NAME="$(basename "${SHELL}" 2>/dev/null || echo "sh")"
        case "${SHELL_NAME}" in
            zsh)
                printf "    echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc\n"
                printf "    source ~/.zshrc\n"
                ;;
            bash)
                printf "    echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc\n"
                printf "    source ~/.bashrc\n"
                ;;
            fish)
                printf "    fish_add_path ~/.local/bin\n"
                ;;
            *)
                printf "    export PATH=\"\$HOME/.local/bin:\$PATH\"\n"
                printf "\n"
                printf "  Add the line above to your shell profile (~/.profile, ~/.bashrc, etc.)\n"
                ;;
        esac
        printf "\n"
    fi
}

main

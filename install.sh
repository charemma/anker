#!/bin/sh
set -e

# ikno installer script
# Usage: curl -fsSL https://ikno.charemma.de/install.sh | sh

VERSION="${IKNO_VERSION:-latest}"
INSTALL_DIR="${IKNO_INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux)
        OS="linux"
        ;;
    Darwin)
        OS="darwin"
        ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Get latest version if not specified
if [ "$VERSION" = "latest" ]; then
    VERSION=$(curl -sSL https://api.github.com/repos/charemma/ikno/releases/latest | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        echo "Failed to get latest version"
        exit 1
    fi
fi

TARBALL="ikno_${VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/charemma/ikno/releases/download/v${VERSION}/${TARBALL}"

echo "Downloading ikno v${VERSION} for ${OS}/${ARCH}..."
echo "URL: ${DOWNLOAD_URL}"

# Download and extract
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

if command -v curl > /dev/null 2>&1; then
    curl -sSL "$DOWNLOAD_URL" -o "$TARBALL"
elif command -v wget > /dev/null 2>&1; then
    wget -q "$DOWNLOAD_URL" -O "$TARBALL"
else
    echo "Error: curl or wget is required"
    exit 1
fi

# Extract binary
tar -xzf "$TARBALL" ikno

# Install
echo "Installing to ${INSTALL_DIR}..."
if [ -w "$INSTALL_DIR" ]; then
    mv ikno "$INSTALL_DIR/ikno"
else
    echo "Requesting sudo access to install to ${INSTALL_DIR}..."
    sudo mv ikno "$INSTALL_DIR/ikno"
fi

# Cleanup
cd - > /dev/null
rm -rf "$TMP_DIR"

echo ""
echo "ikno v${VERSION} installed successfully!"
echo "Run 'ikno --version' to verify"
echo ""
echo "Get started:"
echo "  ikno source add git ."
echo "  ikno recap today"

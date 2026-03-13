#!/bin/bash
set -euo pipefail

GO_VERSION="1.24.1"
GO_TARBALL="go${GO_VERSION}.linux-amd64.tar.gz"
GO_URL="https://go.dev/dl/${GO_TARBALL}"

install_go() {
    if command -v go &>/dev/null; then
        echo "Go already installed: $(go version)"
        return 0
    fi

    if [ -x /usr/local/go/bin/go ]; then
        echo "Go found at /usr/local/go/bin/go: $(/usr/local/go/bin/go version)"
        echo "Add to PATH: export PATH=\$PATH:/usr/local/go/bin"
        return 0
    fi

    echo "Installing Go ${GO_VERSION}..."
    curl -fsSL "${GO_URL}" -o "/tmp/${GO_TARBALL}"
    sudo tar -C /usr/local -xzf "/tmp/${GO_TARBALL}"
    rm "/tmp/${GO_TARBALL}"
    export PATH=$PATH:/usr/local/go/bin
    echo "Go installed: $(go version)"
    echo "Add to your shell profile: export PATH=\$PATH:/usr/local/go/bin"
}

install_go

# Ensure go is on PATH for go mod download
export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin

echo "Downloading Go module dependencies..."
go mod download 2>/dev/null || echo "No go.sum yet, dependencies will be fetched on first build."

echo ""
echo "Done. Runtime dependencies (install separately if needed):"
echo "  - kitty (terminal emulator with remote control)"
echo "  - zellij (terminal multiplexer)"
echo "  - git"

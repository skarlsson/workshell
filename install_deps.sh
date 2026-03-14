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

install_runtime_deps() {
    echo ""
    echo "Installing runtime dependencies..."

    # Detect package manager
    if command -v apt-get &>/dev/null; then
        PKG_MGR="apt-get"
        INSTALL="sudo apt-get install -y"
    elif command -v dnf &>/dev/null; then
        PKG_MGR="dnf"
        INSTALL="sudo dnf install -y"
    elif command -v pacman &>/dev/null; then
        PKG_MGR="pacman"
        INSTALL="sudo pacman -S --noconfirm"
    else
        echo "No supported package manager found (apt-get, dnf, pacman)."
        echo "Install manually: kitty zellij git xdotool xprop gdbus"
        return 1
    fi

    echo "Using package manager: ${PKG_MGR}"

    # Required
    for tool in git xdotool; do
        if command -v "$tool" &>/dev/null; then
            echo "  $tool: already installed"
        else
            echo "  Installing $tool..."
            $INSTALL "$tool"
        fi
    done

    # xprop: package name varies
    if command -v xprop &>/dev/null; then
        echo "  xprop: already installed"
    else
        echo "  Installing xprop..."
        case "$PKG_MGR" in
            apt-get) $INSTALL x11-utils ;;
            dnf)     $INSTALL xorg-x11-utils ;;
            pacman)  $INSTALL xorg-xprop ;;
        esac
    fi

    # gdbus: part of glib2
    if command -v gdbus &>/dev/null; then
        echo "  gdbus: already installed"
    else
        echo "  Installing gdbus..."
        case "$PKG_MGR" in
            apt-get) $INSTALL libglib2.0-bin ;;
            dnf)     $INSTALL glib2 ;;
            pacman)  $INSTALL glib2 ;;
        esac
    fi

    # kitty
    if command -v kitty &>/dev/null; then
        echo "  kitty: already installed"
    else
        echo "  Installing kitty..."
        $INSTALL kitty
    fi

    # zellij: not in all repos, try package manager then curl
    if command -v zellij &>/dev/null; then
        echo "  zellij: already installed"
    else
        echo "  Installing zellij..."
        if $INSTALL zellij 2>/dev/null; then
            true
        else
            echo "  zellij not in package repos, installing from GitHub..."
            curl -fsSL https://github.com/zellij-org/zellij/releases/latest/download/zellij-x86_64-unknown-linux-musl.tar.gz | tar xz -C /tmp
            sudo mv /tmp/zellij /usr/local/bin/
            echo "  zellij installed to /usr/local/bin/zellij"
        fi
    fi
}

install_go

# Ensure go is on PATH for go mod download
export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin

echo "Downloading Go module dependencies..."
go mod download 2>/dev/null || echo "No go.sum yet, dependencies will be fetched on first build."

install_runtime_deps

echo ""
echo "Done. Run 'ws doctor' after building to verify all dependencies."

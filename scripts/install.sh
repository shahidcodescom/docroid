#!/bin/sh
set -e

echo "=== Docroid Installer for Termux (Android) ==="

# Check if running in Termux
if [ -z "$PREFIX" ]; then
    echo "[!] Warning: Not running in standard Termux environment."
    INSTALL_DIR="/usr/local/bin"
else
    INSTALL_DIR="$PREFIX/bin"
    echo "[✓] Detected Termux environment: $PREFIX"
fi

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    aarch64|arm64)
        TARGET_ARCH="arm64"
        ;;
    armv7l|armhf)
        TARGET_ARCH="arm"
        ;;
    x86_64|amd64)
        TARGET_ARCH="amd64"
        ;;
    *)
        echo "[!] Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "[✓] Architecture: $ARCH (Target: linux-$TARGET_ARCH)"

# If Go is installed locally, build directly
if command -v go >/dev/null 2>&1; then
    echo "[*] Compiling Docroid from source with Go..."
    go build -ldflags="-s -w" -o "$INSTALL_DIR/docroid" ./cmd/docroid
    chmod +x "$INSTALL_DIR/docroid"
    echo "[✓] Successfully installed docroid to $INSTALL_DIR/docroid"
elif [ -f "./bin/docroid-linux-$TARGET_ARCH" ]; then
    echo "[*] Installing pre-built binary..."
    cp "./bin/docroid-linux-$TARGET_ARCH" "$INSTALL_DIR/docroid"
    chmod +x "$INSTALL_DIR/docroid"
    echo "[✓] Successfully installed docroid to $INSTALL_DIR/docroid"
else
    echo "[*] Pre-built binary not found. Installing Golang to build..."
    if command -v pkg >/dev/null 2>&1; then
        pkg install -y golang
        go build -ldflags="-s -w" -o "$INSTALL_DIR/docroid" ./cmd/docroid
        chmod +x "$INSTALL_DIR/docroid"
        echo "[✓] Successfully installed docroid to $INSTALL_DIR/docroid"
    else
        echo "[!] Please install Go or build the binary first."
        exit 1
    fi
fi

# Check for PRoot in Termux
if ! command -v proot >/dev/null 2>&1; then
    echo "[*] PRoot is recommended for running full Linux rootfs sandboxes."
    if command -v pkg >/dev/null 2>&1; then
        echo "[*] Installing proot via pkg..."
        pkg install -y proot
    fi
fi

echo ""
echo "=== Installation Complete! ==="
"$INSTALL_DIR/docroid" doctor
echo ""
echo "Try running:"
echo "  docroid run -d -n myapp python3 -m http.server 8080"
echo "  docroid ps"

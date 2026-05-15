#!/bin/bash

# Atilgan Quick Install Script
# Inspired by zed.dev and other modern tools.

set -e

REPO="MrSametBurgazoglu/AtilganFileManager"
BINARY_NAME="atilgan"
INSTALL_DIR="$HOME/.local/bin"
ICON_DIR="$HOME/.local/share/icons/hicolor/scalable/apps"
APP_DIR="$HOME/.local/share/applications"
META_DIR="$HOME/.local/share/metainfo"

# 1. Detect OS and Arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$OS" != "linux" ]; then
    echo "Error: Atilgan is currently only supported on Linux."
    exit 1
fi

if [ "$ARCH" == "x86_64" ]; then
    ARCH="amd64"
else
    echo "Error: Architecture $ARCH is not supported yet."
    exit 1
fi

echo "🚀 Installing Atilgan for $OS-$ARCH..."

# 1.5 Install Dependencies
install_dependencies() {
    if command -v apt-get > /dev/null; then
        echo "📦 Detecting Debian/Ubuntu-based system..."
        echo "🔧 Installing dependencies (requires sudo)..."
        sudo apt-get update
        sudo apt-get install -y libgtk-4-1 libadwaita-1-0 libgtksourceview-5-0
    elif command -v dnf > /dev/null; then
        echo "📦 Detecting Fedora-based system..."
        echo "🔧 Installing dependencies (requires sudo)..."
        sudo dnf install -y gtk4 libadwaita gtksourceview5
    else
        echo "⚠️  Could not detect package manager. Please ensure GTK4, Libadwaita, and GtkSourceView 5 are installed."
    fi
}

install_dependencies

# 2. Get latest version from GitHub API
LATEST_RELEASE=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo "Error: Could not fetch latest release. Check your internet connection."
    exit 1
fi

echo "📦 Found version $LATEST_RELEASE"

# 3. Download the archive
URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/AtilganFileManager_${LATEST_RELEASE#v}_linux_amd64.tar.gz"
TEMP_DIR=$(mktemp -d)
curl -L "$URL" -o "$TEMP_DIR/atilgan.tar.gz"

# 4. Extract
tar -xzf "$TEMP_DIR/atilgan.tar.gz" -C "$TEMP_DIR"

# 5. Create directories
mkdir -p "$INSTALL_DIR"
mkdir -p "$ICON_DIR"
mkdir -p "$APP_DIR"
mkdir -p "$META_DIR"

# 6. Install files
mv "$TEMP_DIR/atilgan" "$INSTALL_DIR/atilgan"
cp "$TEMP_DIR/atilgan_icon.svg" "$ICON_DIR/io.github.mrsametburgazoglu.AtilganFileManager.svg"
cp "$TEMP_DIR/io.github.mrsametburgazoglu.AtilganFileManager.desktop" "$APP_DIR/"
cp "$TEMP_DIR/io.github.mrsametburgazoglu.AtilganFileManager.metainfo.xml" "$META_DIR/"

# 7. Update desktop database (optional but recommended)
if command -v update-desktop-database > /dev/null; then
    update-desktop-database "$APP_DIR"
fi

# 8. Success message
echo "✅ Atilgan $LATEST_RELEASE has been installed to $INSTALL_DIR"
echo ""
echo "Note: Make sure $INSTALL_DIR is in your PATH."
echo "You can now run 'atilgan' or find it in your application menu."

rm -rf "$TEMP_DIR"

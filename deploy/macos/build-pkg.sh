#!/bin/bash
#
# Build macOS .pkg installer for The Stem
#
# Usage:
#   ./build-pkg.sh [BINARY_PATH] [VERSION]
#
# Examples:
#   ./build-pkg.sh ./stem-darwin-arm64 0.3.0
#
# Requirements:
#   - Xcode command line tools (pkgbuild, productbuild)
#   - The Stem binary built for macOS
#

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
PKG_ID="com.stem"
PKG_NAME="stem"
INSTALL_LOCATION="/usr/local/stem"
BUILD_DIR="$REPO_ROOT/dist/macos-pkg"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
# shellcheck disable=SC2034  # reserved for future warnings
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

# Find binary
BINARY_PATH=""
if [[ -n "$1" && -f "$1" ]]; then
    BINARY_PATH="$1"
elif [[ -f "$REPO_ROOT/stem-darwin-$(uname -m)" ]]; then
    BINARY_PATH="$REPO_ROOT/stem-darwin-$(uname -m)"
elif [[ -f "$REPO_ROOT/bin/stem" ]]; then
    BINARY_PATH="$REPO_ROOT/bin/stem"
else
    log_error "Cannot find stem binary. Please build first with: make build"
    exit 1
fi

# Get version
VERSION="${2:-}"
if [[ -z "$VERSION" ]]; then
    VERSION=$("$BINARY_PATH" version 2>/dev/null | head -1 || echo "0.0.0")
    VERSION="${VERSION#v}"
fi

# Architecture
ARCH=$(uname -m)
[[ "$ARCH" == "x86_64" ]] && ARCH="amd64"

log_info "Building macOS .pkg installer"
echo "  Binary:       $BINARY_PATH"
echo "  Version:      $VERSION"
echo "  Architecture: $ARCH"
echo

# Clean and create build directory
log_step "1/6 Preparing build directory..."
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR/payload$INSTALL_LOCATION"
mkdir -p "$BUILD_DIR/payload$INSTALL_LOCATION/launchd"
mkdir -p "$BUILD_DIR/payload$INSTALL_LOCATION/configs"
mkdir -p "$BUILD_DIR/payload$INSTALL_LOCATION/logs"
mkdir -p "$BUILD_DIR/payload$INSTALL_LOCATION/data"
mkdir -p "$BUILD_DIR/scripts"
mkdir -p "$BUILD_DIR/resources"

# Copy binary
log_step "2/6 Copying binary..."
cp "$BINARY_PATH" "$BUILD_DIR/payload$INSTALL_LOCATION/$PKG_NAME"
chmod 755 "$BUILD_DIR/payload$INSTALL_LOCATION/$PKG_NAME"

# Copy bundled iperf3 if available
if [[ -f "$REPO_ROOT/bin/iperf3" ]]; then
    cp "$REPO_ROOT/bin/iperf3" "$BUILD_DIR/payload$INSTALL_LOCATION/stem-iperf3"
    chmod 755 "$BUILD_DIR/payload$INSTALL_LOCATION/stem-iperf3"
    log_info "  Bundled: stem-iperf3"
fi

# Copy config
if [[ -f "$REPO_ROOT/deploy/config/stem.yaml" ]]; then
    cp "$REPO_ROOT/deploy/config/stem.yaml" "$BUILD_DIR/payload$INSTALL_LOCATION/configs/"
fi

# Copy launchd plist
log_step "3/6 Copying launchd configuration..."
cp "$REPO_ROOT/deploy/launchd/com.stem.plist" "$BUILD_DIR/payload$INSTALL_LOCATION/launchd/"

# Copy scripts
log_step "4/6 Preparing installation scripts..."
cp "$SCRIPT_DIR/scripts/preinstall" "$BUILD_DIR/scripts/"
cp "$SCRIPT_DIR/scripts/postinstall" "$BUILD_DIR/scripts/"
chmod 755 "$BUILD_DIR/scripts/preinstall"
chmod 755 "$BUILD_DIR/scripts/postinstall"

# Copy resources
cp "$SCRIPT_DIR/resources/welcome.html" "$BUILD_DIR/resources/"
cp "$SCRIPT_DIR/resources/conclusion.html" "$BUILD_DIR/resources/"

# Build component package
log_step "5/6 Building component package..."
pkgbuild \
    --root "$BUILD_DIR/payload" \
    --identifier "$PKG_ID.pkg" \
    --version "$VERSION" \
    --scripts "$BUILD_DIR/scripts" \
    --install-location "/" \
    "$BUILD_DIR/stem-component.pkg"

# Create distribution.xml with version substituted
sed "s/__VERSION__/$VERSION/g" "$SCRIPT_DIR/distribution.xml" > "$BUILD_DIR/distribution.xml"

# Build final product package
log_step "6/6 Building final package..."
PKG_OUTPUT="$REPO_ROOT/dist/stem-${VERSION}-${ARCH}.pkg"
mkdir -p "$(dirname "$PKG_OUTPUT")"

productbuild \
    --distribution "$BUILD_DIR/distribution.xml" \
    --resources "$BUILD_DIR/resources" \
    --package-path "$BUILD_DIR" \
    "$PKG_OUTPUT"

# Clean up intermediate files
rm -rf "$BUILD_DIR"

echo
log_info "Package built successfully: $PKG_OUTPUT"
echo "  Install: sudo installer -pkg $PKG_OUTPUT -target /"
echo

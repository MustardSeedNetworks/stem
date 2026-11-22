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
# Writes dist/stem-<version>-darwin-arm64.pkg. Apple silicon only.
#
# Requirements:
#   - Xcode command line tools (pkgbuild, productbuild)
#   - The Stem binary built for macOS arm64
#

set -e

# cp and pkgbuild otherwise carry extended attributes into the payload as
# AppleDouble ._ files, which the install then drops into /usr/local/stem.
export COPYFILE_DISABLE=1

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
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
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

# Architecture of the binary, not of the build host: the release builds the
# arm64 package from the darwin archive on whatever runner it gets.
ARCH=$(lipo -archs "$BINARY_PATH")
if [[ "$ARCH" != "arm64" ]]; then
    log_error "Stem ships for Apple silicon only; the binary is: $ARCH"
    exit 1
fi

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
mkdir -p "$BUILD_DIR/scripts"
mkdir -p "$BUILD_DIR/resources"

# Copy binary
log_step "2/6 Copying binary..."
cp "$BINARY_PATH" "$BUILD_DIR/payload$INSTALL_LOCATION/$PKG_NAME"
chmod 755 "$BUILD_DIR/payload$INSTALL_LOCATION/$PKG_NAME"

# Sign the daemon. Notarization rejects a package containing an unsigned
# executable, and the three things it checks are all set here: a Developer ID
# signature, the hardened runtime, and a secure timestamp. Skipped only when no
# identity is present, so a local dev build still works; such a package cannot
# be notarized, which is stated rather than discovered at submission.
SIGN_IDENTITY="${STEM_SIGN_IDENTITY:-Developer ID Application}"
if security find-identity -v -p codesigning | grep -q "$SIGN_IDENTITY"; then
    codesign --force --timestamp --options runtime \
        --sign "$SIGN_IDENTITY" \
        "$BUILD_DIR/payload$INSTALL_LOCATION/$PKG_NAME"
    codesign --verify --strict "$BUILD_DIR/payload$INSTALL_LOCATION/$PKG_NAME"
else
    log_warn "No '$SIGN_IDENTITY' identity: shipping an unsigned daemon; this package cannot be notarized."
fi

# The default config is a template; postinstall copies it to stem.yaml only
# when the operator has none, so an upgrade never replaces their file.
cp "$REPO_ROOT/deploy/config/stem.yaml" "$BUILD_DIR/payload$INSTALL_LOCATION/configs/"

# Copy launchd plist
log_step "3/6 Copying launchd configuration..."
cp "$REPO_ROOT/deploy/launchd/com.stem.plist" "$BUILD_DIR/payload$INSTALL_LOCATION/launchd/"
cp "$REPO_ROOT/deploy/launchd/com.stem.update.plist" "$BUILD_DIR/payload$INSTALL_LOCATION/launchd/"

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

# The "Automatic updates" choice: no payload, only the script that turns the
# opt-in on. It sits after the component in distribution.xml, so the binary is
# in place when it runs.
pkgbuild \
    --nopayload \
    --identifier "$PKG_ID.autoupdate.pkg" \
    --version "$VERSION" \
    --scripts "$SCRIPT_DIR/autoupdate-scripts" \
    "$BUILD_DIR/stem-autoupdate.pkg"

# Create distribution.xml with version substituted
sed "s/__VERSION__/$VERSION/g" "$SCRIPT_DIR/distribution.xml" > "$BUILD_DIR/distribution.xml"

# Build final product package
log_step "6/6 Building final package..."
PKG_OUTPUT="$REPO_ROOT/dist/stem-${VERSION}-darwin-${ARCH}.pkg"
mkdir -p "$(dirname "$PKG_OUTPUT")"

# A Developer ID *Installer* certificate is a different certificate from the
# Application one used on the binary. Signing here, rather than with a later
# productsign, keeps the package signed from the moment it exists.
SIGN_PKG=()
if [[ -n "${STEM_INSTALLER_IDENTITY:-}" ]]; then
    SIGN_PKG=(--sign "${STEM_INSTALLER_IDENTITY}" --timestamp)
fi

productbuild \
    --distribution "$BUILD_DIR/distribution.xml" \
    --resources "$BUILD_DIR/resources" \
    --package-path "$BUILD_DIR" \
    "${SIGN_PKG[@]}" \
    "$PKG_OUTPUT"

# Clean up intermediate files
rm -rf "$BUILD_DIR"

if pkgutil --payload-files "$PKG_OUTPUT" | grep '/\._'; then
    log_error "Payload contains AppleDouble ._ files (listed above)"
    exit 1
fi

echo
log_info "Package built successfully: $PKG_OUTPUT"
echo "  Install: double-click the .pkg in Finder, or sudo installer -pkg $PKG_OUTPUT -target /"
echo

#!/usr/bin/env bash
# Installs HazyVPN on this machine: builds the binary, installs the root
# helper, and grants your user passwordless sudo for that helper only.
# Safe to re-run (upgrades in place).
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PREFIX="${PREFIX:-/usr/local}"
BIN_DIR="$PREFIX/bin"
LIB_DIR="$PREFIX/lib/hazyvpn"
VERSION="${HAZYVPN_VERSION:-0.1.0}"

log()  { echo "==> $*"; }
die()  { echo "error: $*" >&2; exit 1; }

[[ $EUID -eq 0 ]] && die "run this as your normal user, not root (it calls sudo itself when needed)"
command -v sudo >/dev/null || die "sudo is required"

install_pkg() {
    # $1 = human name for logging, remaining args = package name(s) per manager
    if command -v apt-get >/dev/null; then
        sudo apt-get update -qq && sudo apt-get install -y "$2"
    elif command -v pacman >/dev/null; then
        sudo pacman -S --needed --noconfirm "$3"
    elif command -v dnf >/dev/null; then
        sudo dnf install -y "$4"
    elif command -v zypper >/dev/null; then
        sudo zypper install -y "$5"
    else
        die "could not detect a package manager; install $1 manually and re-run"
    fi
}

if ! command -v wg-quick >/dev/null; then
    log "wireguard-tools not found; installing"
    install_pkg "wireguard-tools" wireguard-tools wireguard-tools wireguard-tools wireguard-tools
fi

if ! command -v go >/dev/null; then
    log "Go not found; installing a bootstrap toolchain"
    install_pkg "Go" golang-go go golang go
fi

# go.mod pins the exact Go version HazyVPN needs; GOTOOLCHAIN=auto lets an
# older bootstrap `go` download and switch to it automatically.
export GOTOOLCHAIN=auto

log "Building hazyvpn $VERSION"
( cd "$REPO_DIR" && go build -ldflags "-s -w -X main.version=$VERSION" -o "$REPO_DIR/hazyvpn" . )

log "Installing binary to $BIN_DIR/hazyvpn"
sudo install -Dm755 "$REPO_DIR/hazyvpn" "$BIN_DIR/hazyvpn"

log "Installing root helper to $LIB_DIR/helper"
sudo install -Dm755 "$REPO_DIR/hazyvpn-helper" "$LIB_DIR/helper"

TARGET_USER="${SUDO_USER:-$USER}"
log "Granting $TARGET_USER passwordless access to the helper (only)"
sudo install -dm750 /etc/sudoers.d
tmp_sudoers=$(mktemp)
echo "$TARGET_USER ALL=(root) NOPASSWD: $LIB_DIR/helper" > "$tmp_sudoers"
sudo visudo -cf "$tmp_sudoers" || { rm -f "$tmp_sudoers"; die "generated sudoers rule failed validation"; }
sudo install -m440 -o root -g root "$tmp_sudoers" /etc/sudoers.d/hazyvpn
rm -f "$tmp_sudoers"

echo
log "Installed. Try it out:"
echo "    hazyvpn --demo   # tour the UI with fake tunnels, no root needed"
echo "    hazyvpn          # manage your real /etc/wireguard configs"

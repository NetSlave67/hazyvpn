#!/usr/bin/env bash
# Removes HazyVPN and its root helper/sudoers rule. Never touches
# /etc/wireguard — your tunnel configs are left in place.
set -euo pipefail

PREFIX="${PREFIX:-/usr/local}"
BIN_DIR="$PREFIX/bin"
LIB_DIR="$PREFIX/lib/hazyvpn"

log() { echo "==> $*"; }

command -v sudo >/dev/null || { echo "error: sudo is required" >&2; exit 1; }

log "Removing sudoers rule"
sudo rm -f /etc/sudoers.d/hazyvpn

log "Removing root helper"
sudo rm -rf "$LIB_DIR"

log "Removing binary"
sudo rm -f "$BIN_DIR/hazyvpn"

echo "Done. /etc/wireguard was left untouched."

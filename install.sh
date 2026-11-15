#!/usr/bin/env bash
# Install/uninstall dotset. Follows the shared installer contract:
# --user, --system, --prefix, --uninstall, --help; never requires sudo unless needed.
set -euo pipefail

APP_NAME="dotset"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PREFIX="${PREFIX:-$HOME/.local}"
INSTALL_DIR="${INSTALL_DIR:-}"
ACTION="install"

usage() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Options:
  --user            Install to ~/.local/bin (default for non-root users)
  --system          Install to /usr/local/bin (requires sudo when needed)
  --prefix DIR      Install to DIR/bin
  --uninstall       Remove the installed binary
  -h, --help        Show this help

Environment:
  PREFIX            Base prefix for the default install dir (default: ~/.local)
  INSTALL_DIR       Explicit destination directory (overrides PREFIX/bin)
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --user)       INSTALL_DIR="$HOME/.local/bin"; shift ;;
    --system)     INSTALL_DIR="/usr/local/bin"; shift ;;
    --prefix)     INSTALL_DIR="$2/bin"; shift 2 ;;
    --uninstall)  ACTION="uninstall"; shift ;;
    -h|--help)    usage; exit 0 ;;
    *)            echo "Unknown option: $1" >&2; usage >&2; exit 2 ;;
  esac
done

if [[ -z "$INSTALL_DIR" ]]; then
  if [[ $EUID -eq 0 ]]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="$PREFIX/bin"
  fi
fi

TARGET_BIN="$INSTALL_DIR/$APP_NAME"

run_privileged() {
  local target_dir="$1"
  shift
  if [[ -w "$target_dir" || $EUID -eq 0 ]]; then
    "$@"
  elif command -v sudo >/dev/null 2>&1; then
    sudo "$@"
  else
    echo "Error: no write access to $target_dir and sudo is unavailable." >&2
    exit 1
  fi
}

if [[ "$ACTION" == "uninstall" ]]; then
  if [[ -f "$TARGET_BIN" ]]; then
    run_privileged "$INSTALL_DIR" rm -f "$TARGET_BIN"
    echo "Removed $TARGET_BIN"
  else
    echo "Nothing to remove at $TARGET_BIN"
  fi
  exit 0
fi

SOURCE_BIN="$SCRIPT_DIR/bin/$APP_NAME"
if [[ ! -x "$SOURCE_BIN" ]]; then
  if command -v go >/dev/null 2>&1; then
    echo "==> Building $APP_NAME..."
    make -C "$SCRIPT_DIR" build
  else
    echo "Error: $SOURCE_BIN not found and Go is not on PATH." >&2
    exit 1
  fi
fi

mkdir -p "$INSTALL_DIR" 2>/dev/null || run_privileged "$INSTALL_DIR" mkdir -p "$INSTALL_DIR"
run_privileged "$INSTALL_DIR" install -m 755 "$SOURCE_BIN" "$TARGET_BIN"

echo "Installed $APP_NAME in $TARGET_BIN"
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  echo "Warning: $INSTALL_DIR is not on PATH."
  echo "  Add: export PATH=\"\$PATH:$INSTALL_DIR\""
fi

#!/usr/bin/env sh
# Download and install the latest release of dotset from GitHub Releases.
set -eu

REPO="henriquemanduca/dotset"
APP_NAME="dotset"
PREFIX="${PREFIX:-$HOME/.local}"
INSTALL_DIR="${INSTALL_DIR:-}"
VERSION="${VERSION:-}"

usage() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Download and install a prebuilt $APP_NAME binary from GitHub Releases.

Options:
  --version TAG     Install a specific release tag (e.g. v0.1.0)
  --user            Install to ~/.local/bin (default for non-root users)
  --system          Install to /usr/local/bin (uses sudo if needed)
  --prefix DIR      Install to DIR/bin
  -h, --help        Show this help

Environment:
  VERSION           Release tag to install (default: latest)
  PREFIX            Base prefix for the default install dir (default: ~/.local)
  INSTALL_DIR       Explicit destination directory (overrides PREFIX/bin)
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version)
      [ "$#" -ge 2 ] || { echo "Error: --version requires an argument." >&2; exit 2; }
      VERSION="$2"
      shift 2
      ;;
    --user)
      INSTALL_DIR="$HOME/.local/bin"
      shift
      ;;
    --system)
      INSTALL_DIR="/usr/local/bin"
      shift
      ;;
    --prefix)
      [ "$#" -ge 2 ] || { echo "Error: --prefix requires an argument." >&2; exit 2; }
      INSTALL_DIR="$2/bin"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Error: required command '$1' not found." >&2
    exit 1
  fi
}

need_cmd uname
need_cmd tar
need_cmd mktemp

if command -v curl >/dev/null 2>&1; then
  FETCH_TOOL="curl"
elif command -v wget >/dev/null 2>&1; then
  FETCH_TOOL="wget"
else
  echo "Error: curl or wget is required." >&2
  exit 1
fi

detect_os() {
  case "$(uname -s)" in
    Linux)  echo "linux" ;;
    Darwin) echo "darwin" ;;
    *)
      echo "Error: unsupported OS '$(uname -s)'." >&2
      exit 1
      ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *)
      echo "Error: unsupported architecture '$(uname -m)'." >&2
      exit 1
      ;;
  esac
}

resolve_latest_tag() {
  latest_url="https://github.com/$REPO/releases/latest"
  if [ "$FETCH_TOOL" = "curl" ]; then
    final_url="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "$latest_url")"
  else
    final_url="$(wget --spider -S "$latest_url" 2>&1 | awk '/^[[:space:]]*[Ll]ocation:/ {loc=$2} END {print loc}' | tr -d '\r')"
  fi

  tag="${final_url##*/}"
  if [ -z "$tag" ] || [ "$tag" = "latest" ] || [ "$tag" = "releases" ]; then
    echo "Error: could not resolve latest release tag from $latest_url." >&2
    exit 1
  fi
  printf '%s\n' "$tag"
}

download_file() {
  url="$1"
  dest="$2"
  if [ "$FETCH_TOOL" = "curl" ]; then
    curl -fsSL "$url" -o "$dest"
  else
    wget -qO "$dest" "$url"
  fi
}

verify_checksum() {
  checksum_file="$1"
  archive_name="$2"
  expected="$(awk -v file="$archive_name" '$2 == file {print $1}' "$checksum_file")"
  if [ -z "$expected" ]; then
    echo "Error: checksum for $archive_name not found in checksums.txt." >&2
    exit 1
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "$TMP_DIR/$archive_name" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "$TMP_DIR/$archive_name" | awk '{print $1}')"
  else
    echo "Warning: sha256sum/shasum not found; skipping checksum verification." >&2
    return 0
  fi

  if [ "$actual" != "$expected" ]; then
    echo "Error: checksum mismatch for $archive_name." >&2
    echo "  expected: $expected" >&2
    echo "  actual:   $actual" >&2
    exit 1
  fi
}

run_privileged() {
  target_dir="$1"
  shift
  if [ -w "$target_dir" ] || [ "$(id -u)" -eq 0 ]; then
    "$@"
  elif command -v sudo >/dev/null 2>&1; then
    sudo "$@"
  else
    echo "Error: no write access to $target_dir and sudo is unavailable." >&2
    exit 1
  fi
}

if [ -z "$INSTALL_DIR" ]; then
  if [ "$(id -u)" -eq 0 ]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="$PREFIX/bin"
  fi
fi

OS="$(detect_os)"
ARCH="$(detect_arch)"

if [ -z "$VERSION" ]; then
  echo "==> Resolving latest $APP_NAME release..."
  TAG="$(resolve_latest_tag)"
else
  case "$VERSION" in
    v*) TAG="$VERSION" ;;
    *)  TAG="v$VERSION" ;;
  esac
fi

CLEAN_VERSION="${TAG#v}"
ARCHIVE="${APP_NAME}_${CLEAN_VERSION}_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/$REPO/releases/download/$TAG"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT INT TERM

echo "==> Downloading $ARCHIVE ($TAG)..."
download_file "$BASE_URL/$ARCHIVE" "$TMP_DIR/$ARCHIVE"

echo "==> Verifying checksum..."
download_file "$BASE_URL/checksums.txt" "$TMP_DIR/checksums.txt"
verify_checksum "$TMP_DIR/checksums.txt" "$ARCHIVE"

echo "==> Extracting..."
tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR"

if [ ! -f "$TMP_DIR/$APP_NAME" ]; then
  echo "Error: $APP_NAME binary not found inside $ARCHIVE." >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR" 2>/dev/null || run_privileged "$INSTALL_DIR" mkdir -p "$INSTALL_DIR"
run_privileged "$INSTALL_DIR" install -m 755 "$TMP_DIR/$APP_NAME" "$INSTALL_DIR/$APP_NAME"

echo "Installed $APP_NAME $TAG to $INSTALL_DIR/$APP_NAME"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    echo "Warning: $INSTALL_DIR is not on PATH."
    echo "  Add: export PATH=\"\$PATH:$INSTALL_DIR\""
    ;;
esac

#!/usr/bin/env bash
set -euo pipefail

ARCHIVE_PATH="${1:-}"
SERVICE_NAME="${2:-openworld}"
APP_DIR="${3:-/opt/openworld}"
SERVICE_USER="${4:-openworld}"

RELEASES_DIR="$APP_DIR/releases"
CURRENT_LINK="$APP_DIR/current"

if [[ -z "$ARCHIVE_PATH" ]]; then
  echo "Usage: $0 <archive-path> [service-name]"
  exit 1
fi

if [[ ! -f "$ARCHIVE_PATH" ]]; then
  echo "Release archive not found: $ARCHIVE_PATH"
  exit 1
fi

if [[ $EUID -ne 0 ]]; then
  echo "Run as root (or via sudo)."
  exit 1
fi

mkdir -p "$RELEASES_DIR"
RELEASE_ID="$(date -u +%Y%m%d%H%M%S)"
TARGET_DIR="$RELEASES_DIR/$RELEASE_ID"
mkdir -p "$TARGET_DIR"

tar -xzf "$ARCHIVE_PATH" -C "$TARGET_DIR"

if [[ ! -x "$TARGET_DIR/openworld-server" ]]; then
  chmod +x "$TARGET_DIR/openworld-server"
fi

if id "$SERVICE_USER" >/dev/null 2>&1; then
  chown -R "$SERVICE_USER:$SERVICE_USER" "$TARGET_DIR"
fi

ln -sfn "$TARGET_DIR" "$CURRENT_LINK"

if systemctl list-unit-files | grep -q "^${SERVICE_NAME}.service"; then
  systemctl daemon-reload
  systemctl restart "$SERVICE_NAME"
  systemctl --no-pager --full status "$SERVICE_NAME" | head -n 20
else
  echo "Systemd service ${SERVICE_NAME}.service not found; release extracted and symlink updated."
fi

echo "Deployed release: $RELEASE_ID"

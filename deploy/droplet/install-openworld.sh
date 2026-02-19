#!/usr/bin/env bash

set -euo pipefail

APP_DIR="${APP_DIR:-/opt/openworld}"
SERVICE_NAME="${SERVICE_NAME:-openworld}"
SERVICE_USER="${SERVICE_USER:-openworld}"
INSTALL_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

if [[ $EUID -ne 0 ]]; then
  echo "Run as root (or via sudo)."
  exit 1
fi

mkdir -p "$APP_DIR/bin" "$APP_DIR/releases" "$APP_DIR/shared"

if ! id "$SERVICE_USER" >/dev/null 2>&1; then
  useradd --system --home "$APP_DIR" --shell /usr/sbin/nologin "$SERVICE_USER"
fi

install -m 0750 "$INSTALL_DIR/deploy-openworld.sh" "$APP_DIR/bin/deploy-openworld.sh"

if [[ ! -f "$APP_DIR/shared/.env" ]]; then
  install -m 0640 /dev/null "$APP_DIR/shared/.env"
fi

SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
sed \
  -e "s|__APP_DIR__|$APP_DIR|g" \
  -e "s|__SERVICE_NAME__|$SERVICE_NAME|g" \
  -e "s|__SERVICE_USER__|$SERVICE_USER|g" \
  "$INSTALL_DIR/openworld.service" > "$SERVICE_FILE"

chown -R "$SERVICE_USER:$SERVICE_USER" "$APP_DIR/releases" "$APP_DIR/shared"

systemctl daemon-reload
systemctl enable "$SERVICE_NAME"

echo "Installed deployment script and systemd unit at $SERVICE_FILE"
echo "Next: populate $APP_DIR/shared/.env and run your first deploy."

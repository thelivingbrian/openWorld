#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

go run . \
  -style two_teams_512_players \
  -ttl "${TTL_SECONDS:-1800}" \
  -action "${PLAYER_ACTION:-random}" \
  -read="${READ_MESSAGES:-true}"

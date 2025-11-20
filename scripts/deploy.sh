#!/usr/bin/env bash
set -euo pipefail

# Deploy helper:
# Upload a compiled binary to a remote host, back up the current binary,
# install the new one, reload systemd and restart the service.
#
# Usage:
# ./scripts/deploy.sh \
#   --target user@host:/opt/go-gin-gorm-api \
#   --binary ./app \
#   --service go-gin-gorm-api \
#   [--ssh-port 22] [--keep-backups 5]

usage() {
  cat <<'USAGE'
Usage: deploy.sh --target user@host:/path/to/dir [--binary ./app] [--service name] [--ssh-port port] [--keep-backups N] [--no-reload]

Options:
  --target       Required. Remote target in form user@host:/remote/path
  --binary       Local binary to upload (default: ./app)
  --service      systemd service name to restart (default: go-gin-gorm-api)
  --ssh-port     SSH port (default: 22)
  --keep-backups Number of backups to keep on remote (default: 5)
  --no-reload    If set, skip systemd daemon-reload (useful if unit unchanged)
  -h|--help      Show this help
USAGE
}

if [ $# -eq 0 ]; then
  usage
  exit 1
fi

TARGET=""
BINARY="./app"
SERVICE_NAME="go-gin-gorm-api"
SSH_PORT=22
KEEP_BACKUPS=5
NO_RELOAD=0

while [ $# -gt 0 ]; do
  case "$1" in
    --target) TARGET="$2"; shift 2;;
    --binary) BINARY="$2"; shift 2;;
    --service) SERVICE_NAME="$2"; shift 2;;
    --ssh-port) SSH_PORT="$2"; shift 2;;
    --keep-backups) KEEP_BACKUPS="$2"; shift 2;;
    --no-reload) NO_RELOAD=1; shift 1;;
    -h|--help) usage; exit 0;;
    *) echo "Unknown arg: $1" >&2; usage; exit 2;;
  esac
done

if [ ! -f "$BINARY" ]; then
  echo "ERROR: local binary '$BINARY' not found." >&2
  exit 3
fi

if ! command -v scp >/dev/null 2>&1 || ! command -v ssh >/dev/null 2>&1; then
  echo "ERROR: 'scp' and 'ssh' are required locally to run this script." >&2
  exit 4
fi

# Parse TARGET into user, host and path
if [[ "$TARGET" =~ ^([^@]+)@([^:]+):(.+)$ ]]; then
  REMOTE_USER="${BASH_REMATCH[1]}"
  REMOTE_HOST="${BASH_REMATCH[2]}"
  REMOTE_DIR="${BASH_REMATCH[3]}"
else
  echo "ERROR: --target must be in form user@host:/remote/path" >&2
  exit 5
fi

echo "Deploying '$BINARY' to ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR} (service: ${SERVICE_NAME})"

TIMESTAMP=$(date -u +"%Y%m%dT%H%M%SZ")
REMOTE_BIN="${REMOTE_DIR}/$(basename "$BINARY")"
BACKUP_BIN="${REMOTE_BIN}.bak.${TIMESTAMP}"

SSH_BASE=(ssh -p "$SSH_PORT" "${REMOTE_USER}@${REMOTE_HOST}")

echo "Uploading binary..."
scp -P "$SSH_PORT" "$BINARY" "${REMOTE_USER}@${REMOTE_HOST}:$REMOTE_DIR/" 

echo "Backing up existing binary (if exists) on remote..."
"${SSH_BASE[@]}" bash -lc "if [ -f '$REMOTE_BIN' ]; then mv '$REMOTE_BIN' '$BACKUP_BIN'; echo 'Backed up to $BACKUP_BIN'; else echo 'No existing binary to backup'; fi"

echo "Setting permissions and moving new binary into place..."
"${SSH_BASE[@]}" bash -lc "chmod +x '$REMOTE_DIR/$(basename "$BINARY")' && mv -f '$REMOTE_DIR/$(basename "$BINARY")' '$REMOTE_BIN' && chown --no-dereference --quiet $(whoami):$(id -gn) '$REMOTE_BIN' || true"

if [ "$NO_RELOAD" -eq 0 ]; then
  echo "Reloading systemd daemon on remote..."
  "${SSH_BASE[@]}" sudo systemctl daemon-reload || echo "Warning: daemon-reload failed or requires password"
fi

echo "Restarting systemd service '${SERVICE_NAME}' on remote..."
"${SSH_BASE[@]}" sudo systemctl restart "$SERVICE_NAME"

echo "Waiting for service to come up and printing recent logs..."
"${SSH_BASE[@]}" sudo journalctl -u "$SERVICE_NAME" --no-pager -n 50

echo "Pruning old backups (keep last $KEEP_BACKUPS)"
"${SSH_BASE[@]}" bash -lc "ls -1t '${REMOTE_BIN}.bak.'* 2>/dev/null | sed -n '$((KEEP_BACKUPS+1)),999p' | xargs -r rm -f || true"

echo "Deploy finished. If something failed, you can roll back using the backup file(s) listed above."

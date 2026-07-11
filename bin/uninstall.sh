#!/bin/sh
set -eu

PROJECT="gmc-mqtt"
ROOT="${MUSTER_ROOT:-}"
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
CONFIG_DIR="$ROOT/etc/$PROJECT"
INSTALL_DIR="$ROOT/opt/$PROJECT"
SYSTEMD_DIR="$ROOT/etc/systemd/system"
TMP_PARENT="${TMPDIR:-/tmp}"
TMP_DIR=""
TMP_CREATED=0
PURGE=0

# shellcheck disable=SC1091
. "$SCRIPT_DIR/release-transaction.sh"

cleanup() {
  status=$?
  trap - EXIT INT TERM
  release_unlock
  if [ "$TMP_CREATED" = "1" ] && [ -n "$TMP_DIR" ]; then rm -rf "$TMP_DIR"; fi
  exit "$status"
}
trap cleanup EXIT INT TERM

case "${1:-}" in
  '') ;;
  --purge) PURGE=1 ;;
  *) release_die "usage: uninstall.sh [--purge]" ;;
esac

if [ -z "$ROOT" ] && [ "$(id -u)" -ne 0 ]; then
  release_die "uninstall.sh must run as root; use sudo or set MUSTER_ROOT for a staged uninstall"
fi

if [ -z "$ROOT" ] && command -v systemctl >/dev/null 2>&1; then
  systemctl disable --now gmc-mqtt-update.timer >/dev/null 2>&1 || true
fi

old_umask=$(umask)
umask 077
TMP_DIR=$(mktemp -d "$TMP_PARENT/$PROJECT-uninstall.XXXXXX") || {
  umask "$old_umask"
  release_die "could not create private uninstall workspace"
}
umask "$old_umask"
TMP_CREATED=1
release_acquire_lock

if [ -z "$ROOT" ] && command -v systemctl >/dev/null 2>&1; then
  systemctl disable --now \
    gmc-mqtt-collector.service gmc-mqtt-ha-bridge.timer \
    gmc-mqtt-doctor.timer >/dev/null 2>&1 || true
fi

for unit in \
  gmc-mqtt-collector.service \
  gmc-mqtt-ha-bridge.service gmc-mqtt-ha-bridge.timer \
  gmc-mqtt-doctor.service gmc-mqtt-doctor.timer \
  gmc-mqtt-update.service gmc-mqtt-update.timer; do
  rm -f "$SYSTEMD_DIR/$unit"
done

"$SCRIPT_DIR/muster-bootstrap.sh" unregister "$PROJECT"
rm -rf "$INSTALL_DIR"
rm -rf "$ROOT/run/muster/$PROJECT"

if [ "$PURGE" = "1" ]; then
  rm -rf "$CONFIG_DIR" "$ROOT/var/lib/$PROJECT"
else
  printf 'Preserved %s and %s. Pass --purge to remove them.\n' "$CONFIG_DIR" "$ROOT/var/lib/$PROJECT"
fi

if [ -z "$ROOT" ] && command -v systemctl >/dev/null 2>&1; then
  systemctl daemon-reload >/dev/null 2>&1 || true
fi

#!/bin/sh
set -eu

PROJECT="gmc-mqtt"
GITHUB_OWNER="${GITHUB_OWNER:-azide0x37}"
GITHUB_REPO="${GITHUB_REPO:-GMC-MQTT}"
RELEASES_URL="${RELEASES_URL:-https://github.com/$GITHUB_OWNER/$GITHUB_REPO/releases}"
ROOT="${MUSTER_ROOT:-}"
CONFIG_DIR="$ROOT/etc/$PROJECT"
INSTALL_DIR="$ROOT/opt/$PROJECT"
CURRENT_LINK="$INSTALL_DIR/current"
RELEASES_DIR="$INSTALL_DIR/releases"
SYSTEMD_DIR="$ROOT/etc/systemd/system"
REGISTRATION="$ROOT/etc/muster/implementations.d/$PROJECT.json"
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
SRC_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
TMP_PARENT="${TMPDIR:-/tmp}"
TMP_DIR=""
TMP_CREATED=0
VERSION="${MUSTER_VERSION:-}"
RELEASE_DIR=""
TRANSACTION_ACTIVE=0

# shellcheck disable=SC1091
. "$SCRIPT_DIR/release-transaction.sh"

log() { printf '%s\n' "$*"; }

create_private_tmp() {
  old_umask=$(umask)
  umask 077
  TMP_DIR=$(mktemp -d "$TMP_PARENT/$PROJECT-install.XXXXXX") || {
    umask "$old_umask"
    release_die "could not create private install workspace"
  }
  umask "$old_umask"
  TMP_CREATED=1
}

restart_services() {
  [ -z "$ROOT" ] || return 0
  command -v systemctl >/dev/null 2>&1 || return 0
  systemctl daemon-reload >/dev/null 2>&1 || true
  systemctl restart gmc-mqtt-collector.service >/dev/null 2>&1 || true
  systemctl restart gmc-mqtt-ha-bridge.timer >/dev/null 2>&1 || true
}

rollback_transaction() {
  [ "$TRANSACTION_ACTIVE" = "1" ] || return 0
  TRANSACTION_ACTIVE=0
  managed_restore systemd "$SYSTEMD_DIR"
  release_restore_state
  restart_services
}

cleanup() {
  status=$?
  trap - EXIT INT TERM
  rollback_transaction || true
  if [ -n "$RELEASE_STAGE" ]; then
    rm -rf "$RELEASE_STAGE"
    RELEASE_STAGE=""
  fi
  if [ "$TMP_CREATED" = "1" ] && [ -n "$TMP_DIR" ]; then
    rm -rf "$TMP_DIR"
  fi
  release_unlock
  exit "$status"
}
trap cleanup EXIT INT TERM

need_root() {
  if [ -z "$ROOT" ] && [ "$(id -u)" -ne 0 ]; then
    release_die "install.sh must run as root; use sudo or set MUSTER_ROOT for a staged install"
  fi
}

detect_arch() {
  if [ -n "${GMC_INSTALL_ARCH:-}" ]; then
    case "$GMC_INSTALL_ARCH" in
      armv6|armv7|arm64|amd64) printf '%s\n' "$GMC_INSTALL_ARCH"; return ;;
      *) release_die "unsupported GMC_INSTALL_ARCH: $GMC_INSTALL_ARCH" ;;
    esac
  fi
  os_name=$(uname -s 2>/dev/null || true)
  arch_name=$(uname -m 2>/dev/null || true)
  [ "$os_name" = "Linux" ] || release_die "unsupported operating system: $os_name"
  case "$arch_name" in
    armv6l|armv6) printf '%s\n' armv6 ;;
    armv7l|armv7|armv8l) printf '%s\n' armv7 ;;
    aarch64|arm64) printf '%s\n' arm64 ;;
    x86_64|amd64) printf '%s\n' amd64 ;;
    *) release_die "unsupported architecture: $arch_name" ;;
  esac
}

resolve_latest_version() {
  latest_url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "$RELEASES_URL/latest") || release_die "could not resolve latest release"
  printf '%s\n' "$latest_url" | awk -F/ 'END {print $NF}' | sed 's/^v//'
}

prepare_source() {
  if [ -f "$SRC_ROOT/muster.yaml" ] && [ -f "$SRC_ROOT/muster.lock.json" ] && { [ -f "$SRC_ROOT/go.mod" ] || [ -x "$SRC_ROOT/bin/gmc-mqtt" ]; }; then
    VERSION="${VERSION:-$(cat "$SRC_ROOT/VERSION")}"
    release_require_version "$VERSION"
    return
  fi

  command -v curl >/dev/null 2>&1 || release_die "missing required command: curl"
  command -v tar >/dev/null 2>&1 || release_die "missing required command: tar"
  VERSION="${VERSION:-$(resolve_latest_version)}"
  release_require_version "$VERSION"
  arch=$(detect_arch)
  archive_name="${PROJECT}_v${VERSION}_linux_${arch}.tar.gz"
  archive_url="$RELEASES_URL/download/v$VERSION/$archive_name"
  checksums_url="$RELEASES_URL/download/v$VERSION/checksums.txt"
  archive_path="$TMP_DIR/$archive_name"
  checksums_path="$TMP_DIR/checksums.txt"
  curl -fsSL "$checksums_url" -o "$checksums_path" || release_die "could not fetch release checksums"
  curl -fsSL "$archive_url" -o "$archive_path" || release_die "could not fetch release artifact"
  expected=$(awk -v name="$archive_name" '$2 == name {print $1}' "$checksums_path")
  release_require_sha "$expected"
  [ "$(release_sha256 "$archive_path")" = "$expected" ] || release_die "downloaded artifact SHA256 mismatch"
  release_validate_archive "$archive_path" "$VERSION"
  mkdir -p "$TMP_DIR/source"
  tar -xzf "$archive_path" -C "$TMP_DIR/source" --strip-components=1
  SRC_ROOT="$TMP_DIR/source"
}

project_release_valid() {
  directory="$1"
  version="$2"
  release_dir_valid "$directory" "$version" || return 1
  for required in \
    bin/release-transaction.sh bin/muster-observation.sh bin/gmc-mqtt-ha-bridge.sh \
    bin/gmc-mqtt systemd/gmc-mqtt-collector.service systemd/gmc-mqtt-ha-bridge.timer \
    etc/defaults/collector.env etc/defaults/mqtt.env etc/defaults/home-assistant.env; do
    [ -f "$directory/$required" ] && [ ! -L "$directory/$required" ] || return 1
  done
  [ -x "$directory/bin/muster-bootstrap.sh" ] && [ -x "$directory/bin/doctor.sh" ] && \
    [ -x "$directory/bin/gmc-mqtt" ] && [ -x "$directory/bin/gmc-mqtt-ha-bridge.sh" ]
}

stage_release() {
  release_new_stage "$VERSION"
  mkdir -p "$RELEASE_STAGE/bin" "$RELEASE_STAGE/etc/defaults" "$RELEASE_STAGE/systemd" "$RELEASE_STAGE/doc"
  cp "$SRC_ROOT"/bin/*.sh "$RELEASE_STAGE/bin/"
  cp "$SRC_ROOT"/etc/defaults/*.env "$RELEASE_STAGE/etc/defaults/"
  cp "$SRC_ROOT"/systemd/* "$RELEASE_STAGE/systemd/"
  cp "$SRC_ROOT"/VERSION "$SRC_ROOT"/muster.yaml "$SRC_ROOT"/muster.lock.json "$RELEASE_STAGE/"
  cp "$SRC_ROOT"/CHANGELOG.md "$SRC_ROOT"/README.md "$SRC_ROOT"/MUSTER.md "$SRC_ROOT"/RELEASE.md "$SRC_ROOT"/SECURITY.md "$RELEASE_STAGE/doc/"
  if [ -x "$SRC_ROOT/bin/gmc-mqtt" ]; then
    cp "$SRC_ROOT/bin/gmc-mqtt" "$RELEASE_STAGE/bin/gmc-mqtt"
  elif [ -f "$SRC_ROOT/go.mod" ]; then
    (cd "$SRC_ROOT" && go build -trimpath -o "$RELEASE_STAGE/bin/gmc-mqtt" ./cmd)
  else
    release_die "release source has no collector binary"
  fi
  chmod 0755 "$RELEASE_STAGE/bin"/*
  project_release_valid "$RELEASE_STAGE" "$VERSION" || release_die "staged release $VERSION failed project validation"
  release_publish_stage "$VERSION"
  RELEASE_DIR="$RELEASES_DIR/$VERSION"
}

seed_config() {
  mkdir -p "$CONFIG_DIR"
  for file in collector.env mqtt.env home-assistant.env; do
    if [ -f "$CONFIG_DIR/$file" ]; then
      log "Preserving existing $CONFIG_DIR/$file"
    else
      cp "$RELEASE_DIR/etc/defaults/$file" "$CONFIG_DIR/$file"
      if [ "$file" = "mqtt.env" ]; then chmod 0600 "$CONFIG_DIR/$file"; else chmod 0644 "$CONFIG_DIR/$file"; fi
      log "Installed $CONFIG_DIR/$file"
    fi
  done
}

validate_registration() {
  if [ -n "$ROOT" ]; then
    "$ROOT/usr/local/bin/muster" --root "$ROOT" validate
  else
    /usr/local/bin/muster validate
  fi
}

activate_release() {
  old_release="$TMP_DIR/no-old"
  if [ "$HAD_CURRENT" = "1" ]; then old_release="$INSTALL_DIR/$PREVIOUS_TARGET"; fi
  managed_snapshot systemd "$SYSTEMD_DIR" "$RELEASE_DIR/systemd" "$old_release/systemd"
  TRANSACTION_ACTIVE=1
  managed_apply systemd "$SYSTEMD_DIR" "$RELEASE_DIR/systemd"
  release_switch_current "$VERSION"
  if ! "$RELEASE_DIR/bin/muster-bootstrap.sh" register "$PROJECT"; then
    rollback_transaction
    release_die "Muster registration failed; restored the prior release transaction"
  fi
  if ! validate_registration; then
    rollback_transaction
    release_die "Muster graph validation failed; restored the prior release transaction"
  fi
  TRANSACTION_ACTIVE=0
}

enable_systemd() {
  [ -z "$ROOT" ] || return 0
  command -v systemctl >/dev/null 2>&1 || return 0
  systemctl daemon-reload
  systemctl enable --now gmc-mqtt-collector.service
  systemctl enable --now gmc-mqtt-ha-bridge.timer gmc-mqtt-doctor.timer gmc-mqtt-update.timer
}

need_root
create_private_tmp
release_acquire_lock
prepare_source
stage_release
seed_config
"$RELEASE_DIR/bin/muster-bootstrap.sh" ensure
release_snapshot_state
activate_release
enable_systemd
log "$PROJECT $VERSION installed"

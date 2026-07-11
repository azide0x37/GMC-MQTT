#!/bin/sh
set -eu

PROJECT="gmc-mqtt"
ROOT="${MUSTER_ROOT:-}"
DEFAULTS_DIR="${DEFAULTS_DIR:-$ROOT/opt/$PROJECT/current/etc/defaults}"
CONFIG_DIR="${CONFIG_DIR:-$ROOT/etc/$PROJECT}"
STATE_DIR="${GMC_RUNTIME_DIR:-$ROOT/run/muster/$PROJECT}"
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
JSON_ONLY=0
CONFIG_READY=0

for argument in "$@"; do
  case "$argument" in
    --runtime) ;;
    --json) JSON_ONLY=1 ;;
    *) printf '%s\n' "usage: doctor.sh [--runtime] [--json]" >&2; exit 2 ;;
  esac
done

# shellcheck disable=SC1091
. "$SCRIPT_DIR/muster-observation.sh"

if [ -n "${MUSTER_DOCTOR_OUTPUT:-}" ]; then
  OBSERVATION_FILE="$MUSTER_DOCTOR_OUTPUT"
elif [ -n "$ROOT" ]; then
  OBSERVATION_FILE="$ROOT/run/muster/$PROJECT/observations/doctor.json"
else
  OBSERVATION_FILE="/run/muster/$PROJECT/observations/doctor.json"
fi

muster_observation_begin "$PROJECT" "doctor" "$OBSERVATION_FILE" "$JSON_ONLY"

load_env_file() {
  file="$1"
  if [ -f "$file" ]; then
    set -a
    # shellcheck disable=SC1090
    . "$file"
    set +a
  fi
}

root_path() {
  path="$1"
  case "$path" in
    /*)
      if [ -n "$ROOT" ] && [ "${path#"$ROOT"/}" = "$path" ]; then
        printf '%s%s\n' "$ROOT" "$path"
      else
        printf '%s\n' "$path"
      fi
      ;;
    *) printf '%s\n' "$path" ;;
  esac
}

positive_int() {
  case "$1" in
    ''|*[!0-9]*|0) return 1 ;;
    *) return 0 ;;
  esac
}

config_value() {
  name="$1"
  value="$2"
  if [ -n "$value" ]; then
    muster_check healthy "config-$name" "config $name is set"
  else
    muster_check unhealthy "config-$name" "config $name is required"
    CONFIG_READY=0
  fi
}

check_config_files() {
  CONFIG_READY=1
  for file in collector.env mqtt.env home-assistant.env; do
    if [ -f "$CONFIG_DIR/$file" ]; then
      muster_check healthy "config-file-$file" "configuration file is present: $file"
    else
      muster_check unhealthy "config-file-$file" "missing configuration file: $CONFIG_DIR/$file"
      CONFIG_READY=0
    fi
  done

  config_value GMC_SERIAL_DEVICE "$GMC_SERIAL_DEVICE"
  config_value GMC_MQTT_HOST "$GMC_MQTT_HOST"
  config_value GMC_STATE_TOPIC "$GMC_STATE_TOPIC"
  if positive_int "$GMC_BAUD_RATE"; then
    muster_check healthy config-GMC_BAUD_RATE "GMC_BAUD_RATE is a positive integer"
  else
    muster_check unhealthy config-GMC_BAUD_RATE "GMC_BAUD_RATE must be a positive integer"
    CONFIG_READY=0
  fi
  if positive_int "$GMC_QUERY_INTERVAL"; then
    muster_check healthy config-GMC_QUERY_INTERVAL "GMC_QUERY_INTERVAL is a positive integer"
  else
    muster_check unhealthy config-GMC_QUERY_INTERVAL "GMC_QUERY_INTERVAL must be a positive integer"
    CONFIG_READY=0
  fi
  if positive_int "$GMC_MQTT_PORT"; then
    muster_check healthy config-GMC_MQTT_PORT "GMC_MQTT_PORT is a positive integer"
  else
    muster_check unhealthy config-GMC_MQTT_PORT "GMC_MQTT_PORT must be a positive integer"
    CONFIG_READY=0
  fi
  muster_check healthy config-secrets "MQTT password is configured outside inspector metadata and redacted from doctor output"
}

check_unit() {
  unit="$1"
  if [ -f "$ROOT/etc/systemd/system/$unit" ] || { [ -z "$ROOT" ] && [ -f "/etc/systemd/system/$unit" ]; }; then
    muster_check healthy "unit-$unit" "systemd unit present: $unit"
  else
    muster_check unhealthy "unit-$unit" "systemd unit missing: $unit"
  fi
}

check_state_file() {
  if [ ! -f "$GMC_STATE_PATH" ]; then
    muster_check unhealthy collector-state "collector state is missing: $GMC_STATE_PATH"
    return
  fi
  if grep -q '"timestamp"' "$GMC_STATE_PATH" && grep -q '"healthy"' "$GMC_STATE_PATH"; then
    muster_check healthy collector-state "collector state contains timestamp and health fields"
  else
    muster_check unhealthy collector-state "collector state is malformed or incomplete"
  fi
}

check_runtime() {
  if [ -n "$ROOT" ]; then
    muster_check unknown runtime "staged root: live serial, systemd, and MQTT checks skipped"
    return
  fi
  if [ "$CONFIG_READY" -ne 1 ]; then
    muster_check unknown runtime "live checks skipped because required configuration is incomplete"
    return
  fi
  if command -v systemctl >/dev/null 2>&1; then
    muster_check healthy command-systemctl "command exists: systemctl"
  else
    muster_check unhealthy command-systemctl "missing command: systemctl"
  fi
  if command -v mosquitto_pub >/dev/null 2>&1; then
    muster_check healthy command-mosquitto_pub "command exists: mosquitto_pub"
  else
    muster_check unhealthy command-mosquitto_pub "missing command: mosquitto_pub"
  fi
  if [ -e "$GMC_SERIAL_DEVICE" ]; then
    muster_check healthy serial-device "serial device exists: $GMC_SERIAL_DEVICE"
  else
    muster_check unhealthy serial-device "serial device is missing: $GMC_SERIAL_DEVICE"
  fi
  if systemctl is-active --quiet gmc-mqtt-collector.service; then
    muster_check healthy collector-active "gmc-mqtt-collector.service is active"
  else
    muster_check unhealthy collector-active "gmc-mqtt-collector.service is not active"
  fi
  if systemctl is-active --quiet gmc-mqtt-ha-bridge.timer; then
    muster_check healthy bridge-timer-active "gmc-mqtt-ha-bridge.timer is active"
  else
    muster_check unhealthy bridge-timer-active "gmc-mqtt-ha-bridge.timer is not active"
  fi
}

load_env_file "$DEFAULTS_DIR/collector.env"
load_env_file "$DEFAULTS_DIR/mqtt.env"
load_env_file "$DEFAULTS_DIR/home-assistant.env"
load_env_file "$CONFIG_DIR/collector.env"
load_env_file "$CONFIG_DIR/mqtt.env"
load_env_file "$CONFIG_DIR/home-assistant.env"

GMC_SERIAL_DEVICE="${GMC_SERIAL_DEVICE:-}"
GMC_BAUD_RATE="${GMC_BAUD_RATE:-}"
GMC_QUERY_INTERVAL="${GMC_QUERY_INTERVAL:-}"
GMC_MQTT_HOST="${GMC_MQTT_HOST:-}"
GMC_MQTT_PORT="${GMC_MQTT_PORT:-}"
GMC_STATE_TOPIC="${GMC_STATE_TOPIC:-${GMC_PUBLISH_TOPIC:-}}"
GMC_STATE_PATH=$(root_path "${GMC_STATE_PATH:-/run/muster/$PROJECT/state.json}")

check_config_files
for unit in \
  gmc-mqtt-collector.service \
  gmc-mqtt-ha-bridge.service gmc-mqtt-ha-bridge.timer \
  gmc-mqtt-doctor.service gmc-mqtt-doctor.timer \
  gmc-mqtt-update.service gmc-mqtt-update.timer; do
  check_unit "$unit"
done
check_state_file
check_runtime

if muster_observation_emit; then
  exit 0
fi
exit 1

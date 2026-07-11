#!/bin/sh
set -eu

PROJECT="gmc-mqtt"
ROOT="${MUSTER_ROOT:-}"
APPLY=0
DISCOVER=0
ONCE=0

usage() {
  printf '%s\n' "Usage: gmc-mqtt-ha-bridge.sh [--apply] [--discover] [--once]"
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --apply) APPLY=1 ;;
    --discover) DISCOVER=1 ;;
    --once) ONCE=1 ;;
    -h|--help) usage; exit 0 ;;
    *) printf '%s\n' "unknown argument: $1" >&2; exit 2 ;;
  esac
  shift
done

if [ "$DISCOVER" = "0" ] && [ "$ONCE" = "0" ]; then
  ONCE=1
fi

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
  root="$1"
  path="$2"
  case "$path" in
    /*)
      if [ -n "$root" ] && [ "${path#"$root"/}" = "$path" ]; then
        printf '%s%s\n' "$root" "$path"
      else
        printf '%s\n' "$path"
      fi
      ;;
    *) printf '%s\n' "$path" ;;
  esac
}

if [ "$APPLY" = "1" ]; then
  PATH_ROOT="$ROOT"
  DEFAULTS_DIR="${DEFAULTS_DIR:-$ROOT/opt/$PROJECT/current/etc/defaults}"
  CONFIG_DIR="${CONFIG_DIR:-$ROOT/etc/$PROJECT}"
  STATE_DIR="${GMC_RUNTIME_DIR:-$ROOT/run/muster/$PROJECT}"
  VERSION_FILE="${VERSION_FILE:-$ROOT/opt/$PROJECT/current/VERSION}"
else
  MOCK_ROOT="${MUSTER_MOCK_ROOT:-${TMPDIR:-/tmp}/$PROJECT-mock}"
  PATH_ROOT="$MOCK_ROOT"
  DEFAULTS_DIR="${DEFAULTS_DIR:-$MOCK_ROOT/opt/$PROJECT/current/etc/defaults}"
  CONFIG_DIR="${CONFIG_DIR:-$MOCK_ROOT/etc/$PROJECT}"
  STATE_DIR="${GMC_RUNTIME_DIR:-$MOCK_ROOT/run/muster/$PROJECT}"
  VERSION_FILE="${VERSION_FILE:-VERSION}"
fi

load_env_file "$DEFAULTS_DIR/collector.env"
load_env_file "$DEFAULTS_DIR/mqtt.env"
load_env_file "$DEFAULTS_DIR/home-assistant.env"
load_env_file "$CONFIG_DIR/collector.env"
load_env_file "$CONFIG_DIR/mqtt.env"
load_env_file "$CONFIG_DIR/home-assistant.env"

GMC_MQTT_ENABLE="${GMC_MQTT_ENABLE:-true}"
GMC_MQTT_HOST="${GMC_MQTT_HOST:-mainsail}"
GMC_MQTT_PORT="${GMC_MQTT_PORT:-1883}"
GMC_MQTT_USERNAME="${GMC_MQTT_USERNAME:-}"
GMC_MQTT_PASSWORD="${GMC_MQTT_PASSWORD:-}"
GMC_MQTT_PUBLISH_TIMEOUT_SECONDS="${GMC_MQTT_PUBLISH_TIMEOUT_SECONDS:-5}"
GMC_STATE_TOPIC="${GMC_STATE_TOPIC:-${GMC_PUBLISH_TOPIC:-homeassistant/sensor/gmc300/state}}"
GMC_ENABLE_DISCOVERY="${GMC_ENABLE_DISCOVERY:-true}"
GMC_DISCOVERY_PREFIX="${GMC_DISCOVERY_PREFIX:-homeassistant}"
GMC_DEVICE_ID="${GMC_DEVICE_ID:-gmc300_001}"
GMC_DEVICE_NAME="${GMC_DEVICE_NAME:-GMC-300}"
GMC_DEVICE_MANUFACTURER="${GMC_DEVICE_MANUFACTURER:-AzideMakes}"
GMC_DEVICE_MODEL="${GMC_DEVICE_MODEL:-GMC-300}"
GMC_DEVICE_SW_VERSION="${GMC_DEVICE_SW_VERSION:-1.0}"
GMC_DEVICE_SERIAL="${GMC_DEVICE_SERIAL:-gmc300_001}"
GMC_DEVICE_HW_VERSION="${GMC_DEVICE_HW_VERSION:-1.0}"
GMC_ORIGIN_NAME="${GMC_ORIGIN_NAME:-gmc2mqtt}"
GMC_ORIGIN_SW="${GMC_ORIGIN_SW:-1.0}"
GMC_ORIGIN_URL="${GMC_ORIGIN_URL:-https://azidemakes.com/support}"
GMC_STATE_PATH="$(root_path "$PATH_ROOT" "${GMC_STATE_PATH:-$STATE_DIR/state.json}")"
GMC_HA_OUTBOX_DIR="${GMC_HA_OUTBOX_DIR:-$STATE_DIR/ha-mqtt-outbox}"
MOSQUITTO_PUB="${MOSQUITTO_PUB:-mosquitto_pub}"
PUBLISH_FAILED=0

mkdir -p "$STATE_DIR" "$GMC_HA_OUTBOX_DIR"

topic_file() {
  printf '%s' "$1" | tr '/+' '__'
}

json_string() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

bool_enabled() {
  case "$1" in
    true|TRUE|1|yes|YES|on|ON) return 0 ;;
    *) return 1 ;;
  esac
}

run_mosquitto_pub() {
  if command -v timeout >/dev/null 2>&1; then
    timeout "$GMC_MQTT_PUBLISH_TIMEOUT_SECONDS" "$MOSQUITTO_PUB" "$@"
  else
    "$MOSQUITTO_PUB" "$@"
  fi
}

publish() {
  topic="$1"
  payload="$2"
  retain="${3:-0}"
  file="$GMC_HA_OUTBOX_DIR/$(topic_file "$topic").json"
  printf '%s\n' "$payload" > "$file"
  printf '%s\t%s\tretain=%s\n' "$topic" "$file" "$retain" >> "$GMC_HA_OUTBOX_DIR/topics.log"

  if [ "$APPLY" = "1" ] && bool_enabled "$GMC_MQTT_ENABLE"; then
    if ! command -v "$MOSQUITTO_PUB" >/dev/null 2>&1; then
      printf '%s\n' "mqtt publish failed: $MOSQUITTO_PUB is not installed or not executable" >&2
      PUBLISH_FAILED=1
      return 0
    fi
    args="-h $GMC_MQTT_HOST -p $GMC_MQTT_PORT -t $topic -m $payload"
    if [ -n "$GMC_MQTT_USERNAME" ]; then
      if [ "$retain" = "1" ]; then
        run_mosquitto_pub -h "$GMC_MQTT_HOST" -p "$GMC_MQTT_PORT" -u "$GMC_MQTT_USERNAME" -P "$GMC_MQTT_PASSWORD" -t "$topic" -m "$payload" -r || PUBLISH_FAILED=1
      else
        run_mosquitto_pub -h "$GMC_MQTT_HOST" -p "$GMC_MQTT_PORT" -u "$GMC_MQTT_USERNAME" -P "$GMC_MQTT_PASSWORD" -t "$topic" -m "$payload" || PUBLISH_FAILED=1
      fi
    elif [ "$retain" = "1" ]; then
      run_mosquitto_pub -h "$GMC_MQTT_HOST" -p "$GMC_MQTT_PORT" -t "$topic" -m "$payload" -r || PUBLISH_FAILED=1
    else
      run_mosquitto_pub -h "$GMC_MQTT_HOST" -p "$GMC_MQTT_PORT" -t "$topic" -m "$payload" || PUBLISH_FAILED=1
    fi
    if [ "$PUBLISH_FAILED" = "1" ]; then
      printf '%s\n' "mqtt publish failed: $args" >&2
    fi
  fi
}

json_field() {
  key="$1"
  default="$2"
  if [ -f "$GMC_STATE_PATH" ]; then
    value=$(sed -n "s/.*\"$key\"[[:space:]]*:[[:space:]]*\"\\([^\"]*\\)\".*/\\1/p" "$GMC_STATE_PATH" | head -n 1)
    if [ -n "$value" ]; then
      printf '%s\n' "$value"
      return 0
    fi
  fi
  printf '%s\n' "$default"
}

json_number() {
  key="$1"
  default="$2"
  if [ -f "$GMC_STATE_PATH" ]; then
    value=$(sed -n "s/.*\"$key\"[[:space:]]*:[[:space:]]*\\([-0-9.][0-9.]*\\).*/\\1/p" "$GMC_STATE_PATH" | head -n 1)
    if [ -n "$value" ]; then
      printf '%s\n' "$value"
      return 0
    fi
  fi
  printf '%s\n' "$default"
}

sensor_payload() {
  entity="$1"
  name="$2"
  template="$3"
  unit="$4"
  device_class="$5"
  state_class="$6"
  icon="$7"
  unique_id="${GMC_DEVICE_ID}_${entity}"
  payload=$(printf '{"device":{"identifiers":["%s"],"name":"%s","manufacturer":"%s","model":"%s","sw_version":"%s","serial_number":"%s","hw_version":"%s"},"origin":{"name":"%s","sw_version":"%s","support_url":"%s"},"name":"%s","unique_id":"%s","state_topic":"%s","value_template":"%s"' \
    "$(json_string "$GMC_DEVICE_ID")" "$(json_string "$GMC_DEVICE_NAME")" "$(json_string "$GMC_DEVICE_MANUFACTURER")" "$(json_string "$GMC_DEVICE_MODEL")" "$(json_string "$GMC_DEVICE_SW_VERSION")" "$(json_string "$GMC_DEVICE_SERIAL")" "$(json_string "$GMC_DEVICE_HW_VERSION")" \
    "$(json_string "$GMC_ORIGIN_NAME")" "$(json_string "$GMC_ORIGIN_SW")" "$(json_string "$GMC_ORIGIN_URL")" "$(json_string "$name")" "$(json_string "$unique_id")" "$(json_string "$GMC_STATE_TOPIC")" "$(json_string "$template")")
  if [ -n "$unit" ]; then
    payload="$payload,\"unit_of_measurement\":\"$(json_string "$unit")\""
  fi
  if [ -n "$device_class" ]; then
    payload="$payload,\"device_class\":\"$(json_string "$device_class")\""
  fi
  if [ -n "$state_class" ]; then
    payload="$payload,\"state_class\":\"$(json_string "$state_class")\""
  fi
  if [ -n "$icon" ]; then
    payload="$payload,\"icon\":\"$(json_string "$icon")\""
  fi
  printf '%s}\n' "$payload"
}

publish_discovery() {
  bool_enabled "$GMC_ENABLE_DISCOVERY" || return 0
  publish "$GMC_DISCOVERY_PREFIX/sensor/${GMC_DEVICE_ID}_cpm/config" "$(sensor_payload cpm CPM '{{ value_json.cpm }}' CPM '' measurement 'mdi:counter')" 1
  publish "$GMC_DISCOVERY_PREFIX/sensor/${GMC_DEVICE_ID}_battery/config" "$(sensor_payload battery 'Battery Voltage' '{{ value_json.battery }}' V voltage measurement '')" 1
  publish "$GMC_DISCOVERY_PREFIX/sensor/${GMC_DEVICE_ID}_version/config" "$(sensor_payload version 'Firmware Version' '{{ value_json.version }}' '' '' '' 'mdi:chip')" 1
  publish "$GMC_DISCOVERY_PREFIX/sensor/${GMC_DEVICE_ID}_serial/config" "$(sensor_payload serial 'Serial Number' '{{ value_json.serial }}' '' '' '' 'mdi:identifier')" 1
  publish "$GMC_DISCOVERY_PREFIX/sensor/${GMC_DEVICE_ID}_uptime/config" "$(sensor_payload uptime Uptime '{{ value_json.uptime }}' s duration measurement '')" 1
  publish "$GMC_DISCOVERY_PREFIX/sensor/${GMC_DEVICE_ID}_usv/config" "$(sensor_payload usv 'uSv/h' '{{ value_json.usv }}' 'uSv/h' '' measurement 'mdi:radioactive')" 1
  publish "$GMC_DISCOVERY_PREFIX/sensor/${GMC_DEVICE_ID}_mr/config" "$(sensor_payload mr 'mR/h' '{{ value_json.mr }}' 'mR/h' '' measurement 'mdi:radioactive')" 1
}

if [ ! -f "$GMC_STATE_PATH" ]; then
  printf '%s\n' "missing collector state: $GMC_STATE_PATH" >&2
  exit 1
fi

cpm=$(json_number cpm 0)
battery=$(json_number battery 0)
version=$(json_field version "$(cat "$VERSION_FILE" 2>/dev/null || printf '%s\n' unknown)")
model=$(json_field model "$GMC_DEVICE_MODEL")
serial=$(json_field serial "$GMC_DEVICE_SERIAL")
uptime=$(json_number uptime 0)
usv=$(json_number usv 0)
mr=$(json_number mr 0)
timestamp=$(json_field timestamp "")
healthy=$(json_field healthy true)
error=$(json_field error "")
state_payload=$(printf '{"cpm":%s,"battery":%s,"version":"%s","model":"%s","serial":"%s","uptime":%s,"usv":%s,"mr":%s,"timestamp":"%s","healthy":%s,"error":"%s"}' \
  "$cpm" "$battery" "$(json_string "$version")" "$(json_string "$model")" "$(json_string "$serial")" "$uptime" "$usv" "$mr" "$(json_string "$timestamp")" "$healthy" "$(json_string "$error")")

if [ "$DISCOVER" = "1" ] || [ "$ONCE" = "1" ]; then
  publish_discovery
fi
publish "$GMC_STATE_TOPIC" "$state_payload" 0
printf '%s\n' "$state_payload" > "$STATE_DIR/ha-mqtt-state.json"

if [ "$PUBLISH_FAILED" = "1" ]; then
  exit 1
fi

printf '%s\n' "ok: $PROJECT Home Assistant MQTT bridge updated"

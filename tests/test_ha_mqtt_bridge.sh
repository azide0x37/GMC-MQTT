#!/bin/sh
set -eu

ROOT="$(mktemp -d)"
trap 'rm -rf "$ROOT"' EXIT INT TERM

mkdir -p "$ROOT/run/muster/gmc-mqtt" "$ROOT/etc/gmc-mqtt"
cat > "$ROOT/run/muster/gmc-mqtt/state.json" <<'JSON'
{"cpm":123,"battery":3.7,"version":"1.0","model":"GMC-300","serial":"gmc300_001","uptime":42,"usv":0.7011,"mr":0.0701,"timestamp":"2026-06-15T12:00:00Z","healthy":true}
JSON

MUSTER_MOCK_ROOT="$ROOT" DEFAULTS_DIR="$PWD/etc/defaults" CONFIG_DIR="$ROOT/etc/gmc-mqtt" sh bin/gmc-mqtt-ha-bridge.sh --once > "$ROOT/bridge.out"
grep -q "ok: gmc-mqtt Home Assistant MQTT bridge updated" "$ROOT/bridge.out"
test -s "$ROOT/run/muster/gmc-mqtt/ha-mqtt-state.json"
grep -q '"cpm":123' "$ROOT/run/muster/gmc-mqtt/ha-mqtt-state.json"
grep -q 'homeassistant/sensor/gmc300_001_cpm/config' "$ROOT/run/muster/gmc-mqtt/ha-mqtt-outbox/topics.log"
grep -q 'homeassistant/sensor/gmc300/state' "$ROOT/run/muster/gmc-mqtt/ha-mqtt-outbox/topics.log"
grep -q '"device_class":"voltage"' "$ROOT/run/muster/gmc-mqtt/ha-mqtt-outbox/homeassistant_sensor_gmc300_001_battery_config.json"

rm -rf "$ROOT/run/muster/gmc-mqtt/ha-mqtt-outbox"
cat > "$ROOT/etc/gmc-mqtt/home-assistant.env" <<'EOF'
GMC_ENABLE_DISCOVERY=false
EOF
MUSTER_MOCK_ROOT="$ROOT" DEFAULTS_DIR="$PWD/etc/defaults" CONFIG_DIR="$ROOT/etc/gmc-mqtt" sh bin/gmc-mqtt-ha-bridge.sh --once > "$ROOT/bridge-disabled.out"
grep -q 'homeassistant/sensor/gmc300/state' "$ROOT/run/muster/gmc-mqtt/ha-mqtt-outbox/topics.log"
if grep -q '/config' "$ROOT/run/muster/gmc-mqtt/ha-mqtt-outbox/topics.log"; then
  printf '%s\n' "discovery was published while disabled" >&2
  exit 1
fi

cat > "$ROOT/etc/gmc-mqtt/home-assistant.env" <<'EOF'
GMC_ENABLE_DISCOVERY=true
EOF
if MUSTER_ROOT="$ROOT" DEFAULTS_DIR="$PWD/etc/defaults" CONFIG_DIR="$ROOT/etc/gmc-mqtt" MOSQUITTO_PUB="$ROOT/missing-mosquitto-pub" sh bin/gmc-mqtt-ha-bridge.sh --apply --once > "$ROOT/apply.out" 2> "$ROOT/apply.err"; then
  printf '%s\n' "enabled apply publish unexpectedly succeeded with missing mosquitto_pub" >&2
  exit 1
fi
grep -q "mqtt publish failed" "$ROOT/apply.err"

printf '%s\n' "ha mqtt bridge ok"

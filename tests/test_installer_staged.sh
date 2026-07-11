#!/bin/sh
set -eu

ROOT="$(mktemp -d)"
trap 'rm -rf "$ROOT"' EXIT INT TERM

FAKE_MUSTER="$ROOT/fake-muster"
cat > "$FAKE_MUSTER" <<'EOF'
#!/bin/sh
case "${1:-}" in
  version) printf '%s\n' "${MUSTER_CLI_VERSION:-0.1.0}" ;;
  --root) shift 2; [ "${1:-}" = "validate" ] && printf '%s\n' "PASS: staged graph accepted" ;;
  validate) printf '%s\n' "PASS: graph accepted" ;;
  *) printf '%s\n' "fake muster: unsupported arguments: $*" >&2; exit 2 ;;
esac
EOF
chmod 0755 "$FAKE_MUSTER"

MUSTER_ROOT="$ROOT" MUSTER_CLI_SOURCE="$FAKE_MUSTER" MUSTER_CLI_VERSION=0.1.0 sh bin/install.sh > "$ROOT/install.out"
test -L "$ROOT/opt/gmc-mqtt/current"
test "$(readlink "$ROOT/opt/gmc-mqtt/current")" = "releases/$(cat VERSION)"
test -x "$ROOT/opt/gmc-mqtt/current/bin/gmc-mqtt"
test -x "$ROOT/opt/gmc-mqtt/current/bin/gmc-mqtt-ha-bridge.sh"
test -x "$ROOT/opt/muster/current/bin/muster"
test "$(readlink "$ROOT/usr/local/bin/muster")" = "../../../opt/muster/current/bin/muster"
test -f "$ROOT/etc/muster/implementations.d/gmc-mqtt.json"
grep -q 'implementation:gmc-mqtt' "$ROOT/etc/muster/implementations.d/gmc-mqtt.json"
test -f "$ROOT/opt/gmc-mqtt/current/muster.lock.json"
test ! -w "$ROOT/opt/gmc-mqtt/current/muster.yaml"
test -f "$ROOT/etc/gmc-mqtt/collector.env"
test -f "$ROOT/etc/gmc-mqtt/mqtt.env"
test -f "$ROOT/etc/gmc-mqtt/home-assistant.env"
test "$(stat -f %Lp "$ROOT/etc/gmc-mqtt/mqtt.env" 2>/dev/null || stat -c %a "$ROOT/etc/gmc-mqtt/mqtt.env")" = "600"
test -f "$ROOT/etc/systemd/system/gmc-mqtt-collector.service"
test -f "$ROOT/etc/systemd/system/gmc-mqtt-ha-bridge.timer"

printf '%s\n' "GMC_MQTT_HOST=preserved" >> "$ROOT/etc/gmc-mqtt/mqtt.env"
MUSTER_ROOT="$ROOT" MUSTER_CLI_SOURCE="$FAKE_MUSTER" MUSTER_CLI_VERSION=0.1.0 sh bin/install.sh > "$ROOT/reinstall.out"
grep -q preserved "$ROOT/etc/gmc-mqtt/mqtt.env"

mkdir -p "$ROOT/run/muster/gmc-mqtt"
printf '{"cpm":1,"battery":3.7,"version":"1","model":"GMC-300","serial":"s","uptime":1,"usv":0.0057,"mr":0.0006,"timestamp":"2026-06-15T12:00:00Z","healthy":true}\n' > "$ROOT/run/muster/gmc-mqtt/state.json"
MUSTER_ROOT="$ROOT" sh "$ROOT/opt/gmc-mqtt/current/bin/doctor.sh" > "$ROOT/doctor.out"
OBSERVATION="$ROOT/run/muster/gmc-mqtt/observations/doctor.json"
test -s "$OBSERVATION"
grep -q '"schema":"muster.observation/v1"' "$OBSERVATION"
grep -q '"component":"doctor"' "$OBSERVATION"
if grep -q 'GMC_MQTT_PASSWORD' "$OBSERVATION"; then
  printf '%s\n' "doctor observation exposed secret metadata" >&2
  exit 1
fi

MUSTER_ROOT="$ROOT" sh "$ROOT/opt/gmc-mqtt/current/bin/uninstall.sh" > "$ROOT/uninstall.out"
test ! -e "$ROOT/etc/muster/implementations.d/gmc-mqtt.json"
test ! -e "$ROOT/opt/gmc-mqtt"
test -x "$ROOT/opt/muster/current/bin/muster"
test -f "$ROOT/etc/gmc-mqtt/collector.env"

printf '%s\n' "staged installer ok"

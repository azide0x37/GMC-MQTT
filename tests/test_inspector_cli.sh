#!/bin/sh
set -eu

: "${MUSTER_CLI_SOURCE:?MUSTER_CLI_SOURCE must point to a tested Muster CLI}"
: "${MUSTER_CLI_VERSION:?MUSTER_CLI_VERSION must identify the tested Muster CLI}"

ROOT=$(mktemp -d)
trap 'rm -rf "$ROOT"' EXIT INT TERM

"$MUSTER_CLI_SOURCE" compile muster.yaml "$ROOT/muster.lock.json" >/dev/null
cmp muster.lock.json "$ROOT/muster.lock.json"

MUSTER_ROOT="$ROOT" MUSTER_CLI_SOURCE="$MUSTER_CLI_SOURCE" MUSTER_CLI_VERSION="$MUSTER_CLI_VERSION" sh bin/install.sh >/dev/null
CLI="$ROOT/usr/local/bin/muster"
"$CLI" --root "$ROOT" validate >/dev/null
"$CLI" --root "$ROOT" list --json | grep -q 'implementation:gmc-mqtt'
"$CLI" --root "$ROOT" inspect component:gmc-mqtt:state --json | grep -q 'component:gmc-mqtt:state'
"$CLI" --root "$ROOT" inspect action:gmc-mqtt:doctor.run --json | grep -q 'requires_root'
"$CLI" --root "$ROOT" explain pattern:gmc-mqtt:T2R6.home-assistant-mqtt-bridge | grep -q 'Home Assistant MQTT Bridge'

printf '%s\n' "inspector cli compatibility ok"

#!/bin/sh
set -eu

test -f muster.yaml
test -f muster.lock.json

for required in \
  "schema: 2" \
  "framework: Muster" \
  "name: gmc-mqtt" \
  "manager: systemd" \
  "update_polling: systemd-timer" \
  "required_tool: uv" \
  "primary: T2R6.home-assistant-mqtt-bridge" \
  "verified_head: ea6d02aaa6860e5102a760473b2ffe9b90d13c75" \
  "component:gmc-mqtt:doctor" \
  "action:gmc-mqtt:doctor.run" \
  "component:gmc-mqtt:state"; do
  grep -q "$required" muster.yaml
done

manifest_digest=$(sed -n 's/.*"manifest_sha256": "\([0-9a-f]*\)".*/\1/p' muster.lock.json | head -n 1)
if command -v sha256sum >/dev/null 2>&1; then
  actual_digest=$(sha256sum muster.yaml | awk '{print $1}')
else
  actual_digest=$(shasum -a 256 muster.yaml | awk '{print $1}')
fi
test "$manifest_digest" = "$actual_digest"
grep -q '"schema": "muster.lock/v1"' muster.lock.json
grep -q '"version": "0.2.0"' muster.lock.json
grep -q '"id": "implementation:gmc-mqtt"' muster.lock.json

awk '
function finish_component() {
  if (!in_component) return
  if (!has_what || !has_why) {
    printf "component %s is missing literate.what or literate.why\n", component_id > "/dev/stderr"
    failures++
  }
}
$0 == "  components:" { in_components = 1; next }
in_components && $0 == "  edges:" { finish_component(); in_component = 0; in_components = 0; next }
in_components && /^    - id: / {
  finish_component()
  component_id = $0
  sub(/^    - id: /, "", component_id)
  in_component = 1
  has_what = 0
  has_why = 0
  next
}
in_component && /^        what:[[:space:]]*[^[:space:]]/ { has_what = 1 }
in_component && /^        why:[[:space:]]*[^[:space:]]/ { has_why = 1 }
END { if (in_components) finish_component(); exit failures != 0 }
' muster.yaml

printf '%s\n' "muster yaml and lock ok"

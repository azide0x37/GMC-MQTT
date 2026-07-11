#!/bin/sh
set -eu

ROOT=$(mktemp -d)
WORK=$(mktemp -d)
trap 'rm -rf "$ROOT" "$WORK"' EXIT INT TERM

FAKE_MUSTER="$ROOT/fake-muster"
cat > "$FAKE_MUSTER" <<'EOF'
#!/bin/sh
case "${1:-}" in
  version) printf '%s\n' "${MUSTER_CLI_VERSION:-0.1.0}" ;;
  --root) shift 2; [ "${1:-}" = "validate" ] && exit 0 ;;
  validate) exit 0 ;;
  *) exit 2 ;;
esac
EOF
chmod 0755 "$FAKE_MUSTER"

MUSTER_ROOT="$ROOT" MUSTER_CLI_SOURCE="$FAKE_MUSTER" MUSTER_CLI_VERSION=0.1.0 sh bin/install.sh >/dev/null
CURRENT="$ROOT/opt/gmc-mqtt/current"
REGISTRATION="$ROOT/etc/muster/implementations.d/gmc-mqtt.json"
PREVIOUS_TARGET=$(readlink "$CURRENT")
cp "$REGISTRATION" "$WORK/registration.before"

NEXT="$WORK/gmc-mqtt-0.2.1"
cp -R "$ROOT/opt/gmc-mqtt/releases/0.2.0" "$NEXT"
chmod -R u+w "$NEXT"
printf '%s\n' 0.2.1 > "$NEXT/VERSION"
sed 's/"version": "0.2.0"/"version": "0.2.1"/g' "$NEXT/muster.lock.json" > "$NEXT/muster.lock.json.new"
mv "$NEXT/muster.lock.json.new" "$NEXT/muster.lock.json"
cat > "$NEXT/bin/doctor.sh" <<'EOF'
#!/bin/sh
printf '%s\n' "intentional doctor failure" >&2
exit 1
EOF
chmod 0755 "$NEXT/bin/doctor.sh"
tar -C "$WORK" -czf "$WORK/gmc-mqtt-0.2.1.tar.gz" gmc-mqtt-0.2.1

if MUSTER_ROOT="$ROOT" MUSTER_UPDATE_TARBALL="$WORK/gmc-mqtt-0.2.1.tar.gz" MUSTER_VERSION=0.2.1 "$CURRENT/bin/update.sh" > "$WORK/update.out" 2>&1; then
  printf '%s\n' "update unexpectedly succeeded despite doctor failure" >&2
  exit 1
fi

test "$(readlink "$CURRENT")" = "$PREVIOUS_TARGET"
cmp "$WORK/registration.before" "$REGISTRATION"
test -x "$ROOT/opt/muster/current/bin/muster"
test ! -e "$ROOT/var/lock/muster/gmc-mqtt.release.lock"
grep -q "rolling back" "$WORK/update.out"

printf '%s\n' "update registration rollback ok"

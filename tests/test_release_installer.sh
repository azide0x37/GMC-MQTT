#!/bin/sh
set -eu

: "${MUSTER_CLI_SOURCE:?MUSTER_CLI_SOURCE must point to a tested Muster CLI}"
: "${MUSTER_CLI_VERSION:?MUSTER_CLI_VERSION must identify the tested Muster CLI}"

VERSION=$(cat VERSION)
TAG="v$VERSION"
ASSETS="dist/$TAG"
test -x "$ASSETS/install.sh"
test -s "$ASSETS/checksums.txt"

ROOT=$(mktemp -d)
RELEASES=$(mktemp -d)
trap 'rm -rf "$ROOT" "$RELEASES"' EXIT INT TERM
mkdir -p "$RELEASES/download/$TAG"
cp "$ASSETS"/*.tar.gz "$ASSETS/checksums.txt" "$RELEASES/download/$TAG/"

MUSTER_ROOT="$ROOT" \
MUSTER_VERSION="$VERSION" \
MUSTER_CLI_SOURCE="$MUSTER_CLI_SOURCE" \
MUSTER_CLI_VERSION="$MUSTER_CLI_VERSION" \
GMC_INSTALL_ARCH=amd64 \
RELEASES_URL="file://$RELEASES" \
sh "$ASSETS/install.sh" >/dev/null

test -L "$ROOT/opt/gmc-mqtt/current"
test "$(readlink "$ROOT/opt/gmc-mqtt/current")" = "releases/$VERSION"
test -x "$ROOT/opt/gmc-mqtt/current/bin/gmc-mqtt"
test -f "$ROOT/etc/muster/implementations.d/gmc-mqtt.json"
"$ROOT/usr/local/bin/muster" --root "$ROOT" validate >/dev/null

printf '%s\n' "standalone release installer ok"

#!/bin/sh
set -eu

usage() {
  printf '%s\n' "usage: sh packaging/build-release.sh <version>"
}

if [ "${1-}" = "" ]; then
  usage >&2
  exit 1
fi

INPUT_VERSION="$1"
VERSION=${INPUT_VERSION#v}
RELEASE_TAG="v$VERSION"
ROOT_DIR=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
OUTPUT_DIR="$ROOT_DIR/dist/$RELEASE_TAG"
TMP_DIR=$(mktemp -d "${TMPDIR:-/tmp}/gmc-mqtt-release.XXXXXX")
trap 'rm -rf "$TMP_DIR"' EXIT INT TERM HUP
export GOCACHE="${GOCACHE:-$ROOT_DIR/.cache/go-build}"
export GOMODCACHE="${GOMODCACHE:-$ROOT_DIR/.cache/go-mod}"

mkdir -p "$OUTPUT_DIR" "$GOCACHE" "$GOMODCACHE"

if command -v git >/dev/null 2>&1; then
  COMMIT=$(git -C "$ROOT_DIR" rev-parse --short=12 HEAD 2>/dev/null || echo "unknown")
  COMMIT_DATE=$(git -C "$ROOT_DIR" log -1 --format=%cI HEAD 2>/dev/null || echo "unknown")
  if git -C "$ROOT_DIR" diff --quiet --ignore-submodules HEAD -- 2>/dev/null; then
    TREE_STATE="clean"
  else
    TREE_STATE="dirty"
  fi
else
  COMMIT="unknown"
  COMMIT_DATE="unknown"
  TREE_STATE="unknown"
fi

LD_FLAGS="-s -w -X main.Version=$VERSION -X main.Commit=$COMMIT -X main.CommitDate=$COMMIT_DATE -X main.TreeState=$TREE_STATE"

stage_common() {
  stage_dir="$1"
  mkdir -p "$stage_dir/bin" "$stage_dir/etc/defaults" "$stage_dir/systemd"
  cp "$ROOT_DIR"/bin/*.sh "$stage_dir/bin/"
  cp "$ROOT_DIR"/etc/defaults/*.env "$stage_dir/etc/defaults/"
  cp "$ROOT_DIR"/systemd/* "$stage_dir/systemd/"
  cp "$ROOT_DIR"/CHANGELOG.md "$ROOT_DIR"/README.md "$ROOT_DIR"/MUSTER.md "$ROOT_DIR"/RELEASE.md "$ROOT_DIR"/SECURITY.md "$ROOT_DIR"/muster.yaml "$ROOT_DIR"/muster.lock.json "$stage_dir/"
  printf '%s\n' "$VERSION" > "$stage_dir/VERSION"
  chmod 0755 "$stage_dir/bin"/*.sh
}

build_target() {
  goos="$1"
  goarch="$2"
  goarm="$3"
  archive_arch="$4"
  stage_parent="$TMP_DIR/$archive_arch"
  stage_dir="$stage_parent/gmc-mqtt-$VERSION"
  archive_path="$OUTPUT_DIR/gmc-mqtt_${RELEASE_TAG}_${goos}_${archive_arch}.tar.gz"

  stage_common "$stage_dir"
  if [ -n "$goarm" ]; then
    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" GOARM="$goarm" go build -trimpath -ldflags "$LD_FLAGS" -o "$stage_dir/bin/gmc-mqtt" ./cmd
  else
    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -trimpath -ldflags "$LD_FLAGS" -o "$stage_dir/bin/gmc-mqtt" ./cmd
  fi
  chmod 0755 "$stage_dir/bin/gmc-mqtt"
  tar -C "$stage_parent" -czf "$archive_path" "gmc-mqtt-$VERSION"
}

build_target linux arm 6 armv6
build_target linux arm 7 armv7
build_target linux arm64 "" arm64
build_target linux amd64 "" amd64

: > "$OUTPUT_DIR/checksums.txt"
for archive in "$OUTPUT_DIR"/*.tar.gz; do
  archive_name=$(basename "$archive")
  if command -v sha256sum >/dev/null 2>&1; then
    checksum=$(sha256sum "$archive" | awk '{print $1}')
  else
    checksum=$(shasum -a 256 "$archive" | awk '{print $1}')
  fi
  printf "%s  %s\n" "$checksum" "$archive_name" >> "$OUTPUT_DIR/checksums.txt"
done

{ cat "$ROOT_DIR/bin/release-transaction.sh"; printf 'MUSTER_STANDALONE=1\n'; awk 'NR > 1 && $0 != "# shellcheck disable=SC1091" && $0 != ". \"$SCRIPT_DIR/release-transaction.sh\"" { print }' "$ROOT_DIR/bin/install.sh"; } > "$OUTPUT_DIR/install.sh"
chmod 0755 "$OUTPUT_DIR/install.sh"
printf '%s\n' "release artifacts written to $OUTPUT_DIR"

#!/bin/sh
set -eu

OUT_DIR="${1:-dist/rendered-systemd}"
mkdir -p "$OUT_DIR"
cp systemd/*.service systemd/*.timer "$OUT_DIR/"
printf '%s\n' "Rendered gmc-mqtt systemd units into $OUT_DIR"

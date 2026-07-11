#!/bin/sh
set -eu

for file in bin/*.sh packaging/*.sh; do
  [ -f "$file" ] || continue
  sh -n "$file"
done
printf '%s\n' "shell syntax ok"

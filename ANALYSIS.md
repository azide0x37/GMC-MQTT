# Analysis Checklist (Muster-Native Snapshot)

Scoring: `2 = complete`, `1 = partial`, `0 = missing`.

## Core Functionality

1. [x] GMC serial protocol collector (Score: 2)
   - `gmc/gmc.go` owns serial commands and response parsing.
2. [x] Atomic local state snapshot (Score: 2)
   - `cmd/state.go` writes `/run/muster/gmc-mqtt/state.json`.
3. [x] Durable event ledger (Score: 2)
   - Collector appends JSONL events under `/var/lib/gmc-mqtt/ledger.jsonl`.
4. [x] MQTT removed from Go (Score: 2)
   - The Go module no longer depends on Paho MQTT or TOML config.
5. [x] Home Assistant MQTT bridge (Score: 2)
   - `bin/gmc-mqtt-ha-bridge.sh` owns discovery and state publish.

## Muster Contract

1. [x] Runtime under `/opt/gmc-mqtt/releases/<version>` with `/opt/gmc-mqtt/current`.
2. [x] Config seeded under `/etc/gmc-mqtt/*.env`.
3. [x] systemd owns collector, bridge, doctor, and update lifecycle.
4. [x] Timers own repeated bridge publish, doctor checks, and update checks.
5. [x] Lifecycle scripts exist for install, update, uninstall, doctor, and unit rendering.
6. [x] Schema-2 manifest and deterministic lock expose the implementation to the shared CLI/TUI inspector.
7. [x] Install/update atomically manage units, active release, and implementation registration.
8. [x] Doctor emits `muster.observation/v1` evidence at the stable inspector path.

## Quality & Ops

1. [x] Go tests cover collector config, serial parsing, state calculations, state writes, and ledger writes.
2. [x] Shell tests cover bridge rendering, disabled discovery, missing publish dependency, staged install idempotence, shared core bootstrap, registration, structured doctor evidence, and lock integrity.
3. [x] Packaging creates a Muster tarball and legacy multi-arch release archives.

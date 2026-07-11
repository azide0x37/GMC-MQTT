# GMC-MQTT Muster Contract

GMC-MQTT is a Muster service appliance for a GMC-300 series counter.

## Boundary

- Go owns serial polling, GMC protocol parsing, radiation calculations, and atomic local state writes.
- Muster scripts own MQTT publishing, Home Assistant discovery, health checks, install/update/uninstall, and rollback.
- No `config.toml` is required. Editable deployment config lives under `/etc/gmc-mqtt/`.
- `muster.yaml` schema 2 and `muster.lock.json` expose the installed component graph to the shared Muster inspector.
- Doctor is ordinary `muster.observation/v1` evidence, not application-specific inspector code.

## Patterns

- `T2R6.home-assistant-mqtt-bridge`: state and discovery are published by the bridge.
- `T2C5.local-sidecar-bridge`: local collector facts are exported outward.
- `C1.service-capsule`: systemd owns collector and bridge execution.
- `C2.persistent-tick`: systemd timers own bridge, doctor, and update cadence.
- `C5.failure-ratchet`: failures leave files in `/run/muster/gmc-mqtt` and `/var/lib/gmc-mqtt`.
- `C6.lifecycle-capsule`: lifecycle scripts are the install/update/uninstall boundary.
- `R4.state-ledger`: runtime snapshot and durable event ledger are explicit artifacts.

Pattern coverage is intentionally partial because T2R6 telemetry and discovery are implemented while bounded MQTT control ingestion remains disabled.

## Inspector Contract

- The first installation bootstraps `/opt/muster/current` and the managed `/usr/local/bin/muster` symlink.
- GMC-MQTT owns only `/etc/muster/implementations.d/gmc-mqtt.json`; uninstall never removes the shared core.
- Initial inspection reads the manifest, lock, metadata, systemd state, and existing observations without sourcing env files or using the network.
- `action:gmc-mqtt:doctor.run` is an explicit root-required action.
- Install and update serialize changes to systemd units, `current`, and registration, and restore all three on failed validation.

## Paths

- Runtime code: `/opt/gmc-mqtt/releases/<version>/`
- Active release: `/opt/gmc-mqtt/current`
- Config: `/etc/gmc-mqtt/*.env`
- Runtime state: `/run/muster/gmc-mqtt/state.json`
- Durable ledger: `/var/lib/gmc-mqtt/ledger.jsonl`
- Doctor observation: `/run/muster/gmc-mqtt/observations/doctor.json`
- Inspector registration: `/etc/muster/implementations.d/gmc-mqtt.json`

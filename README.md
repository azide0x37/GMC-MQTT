# GMC-MQTT

GMC-MQTT is a Muster-native Linux service appliance for a GMC-300 series counter. The Go collector reads the counter over USB serial and writes local state. Muster-owned scripts publish that state and Home Assistant discovery over MQTT.

There is no required `config.toml`. Deployment configuration is seeded under `/etc/gmc-mqtt/` as shell-readable env files.

## Architecture

- `gmc-mqtt-collector.service`: runs `/opt/gmc-mqtt/current/bin/gmc-mqtt` and writes `/run/muster/gmc-mqtt/state.json`.
- `gmc-mqtt-ha-bridge.timer`: runs the Home Assistant MQTT bridge every second.
- `gmc-mqtt-ha-bridge.service`: publishes retained discovery payloads and non-retained state through `mosquitto_pub`.
- `gmc-mqtt-doctor.timer`: periodically writes structured health evidence for the Muster inspector.
- `gmc-mqtt-update.timer`: performs serialized, rollback-aware release checks.

## Installation

For a Raspberry Pi install from a published release:

```sh
curl -fsSL https://github.com/azide0x37/GMC-MQTT/releases/latest/download/install.sh | sudo sh
```

For a staged local install:

```sh
MUSTER_ROOT="$(mktemp -d)" sh bin/install.sh
```

The installer writes immutable runtime code to `/opt/gmc-mqtt/releases/<version>/`, updates `/opt/gmc-mqtt/current`, seeds `/etc/gmc-mqtt/*.env` only if missing, installs systemd units, bootstraps the shared Muster core when needed, and atomically registers GMC-MQTT under `/etc/muster/implementations.d/`.

## Muster Inspector

The installed schema-2 declaration and digest-bound lock project GMC-MQTT into the shared host inspector. Ordinary inspection is read-only and does not source the application env files or access the network.

```sh
muster list
muster status gmc-mqtt
muster inspect component:gmc-mqtt:state
muster explain pattern:gmc-mqtt:T2R6.home-assistant-mqtt-bridge
sudo muster doctor gmc-mqtt
```

Doctor writes `muster.observation/v1` evidence to `/run/muster/gmc-mqtt/observations/doctor.json`. The explicit doctor action requires root because it updates root-owned runtime evidence; browsing existing evidence does not.

## Configuration

Edit these files after deployment:

- `/etc/gmc-mqtt/collector.env`
- `/etc/gmc-mqtt/mqtt.env`
- `/etc/gmc-mqtt/home-assistant.env`

Production-ready example:

```sh
GMC_SERIAL_DEVICE=/dev/ttyUSB0
GMC_BAUD_RATE=115200
GMC_QUERY_INTERVAL=1

GMC_MQTT_HOST=mainsail
GMC_MQTT_PORT=1883
GMC_STATE_TOPIC=homeassistant/sensor/gmc300/state
GMC_CONFIG_TOPIC=gmc/config/temp
GMC_PERMANENT_CONFIG_TOPIC=gmc/config/permanent

GMC_ENABLE_DISCOVERY=true
GMC_DISCOVERY_PREFIX=homeassistant
GMC_DEVICE_ID=gmc300_001
GMC_DEVICE_NAME=GMC-300
GMC_DEVICE_MANUFACTURER=AzideMakes
GMC_DEVICE_MODEL=GMC-300
GMC_DEVICE_SW_VERSION=1.0
GMC_DEVICE_SERIAL=gmc300_001
GMC_DEVICE_HW_VERSION=1.0
GMC_ORIGIN_NAME=gmc2mqtt
GMC_ORIGIN_SW=1.0
GMC_ORIGIN_URL=https://azidemakes.com/support
```

Restart after changing config:

```sh
sudo systemctl restart gmc-mqtt-collector.service
sudo systemctl restart gmc-mqtt-ha-bridge.service
```

## Local Development

```sh
go test ./...
make test
make package
```

When Go commands need a workspace-local cache:

```sh
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test ./...
```

## Home Assistant

The bridge publishes discovery entities for:

- CPM
- Battery voltage
- Firmware version
- Serial number
- Uptime
- `uSv/h`
- `mR/h`

Discovery topics use:

```text
<GMC_DISCOVERY_PREFIX>/sensor/<GMC_DEVICE_ID>_<entity>/config
```

State is published to `GMC_STATE_TOPIC`.

## Self-Certification

| Muster requirement | Evidence |
| --- | --- |
| systemd owns lifecycle | `systemd/gmc-mqtt-collector.service`, `systemd/gmc-mqtt-ha-bridge.service` |
| timers own repeated checks | `systemd/gmc-mqtt-ha-bridge.timer`, `systemd/gmc-mqtt-doctor.timer`, `systemd/gmc-mqtt-update.timer` |
| config under `/etc/<project>/` | `bin/install.sh` seeds `/etc/gmc-mqtt/*.env` |
| runtime under `/opt/<project>/releases/<version>/` | `bin/install.sh` installs into `/opt/gmc-mqtt/releases/<version>/` |
| `/opt/<project>/current` active link | `bin/install.sh` switches `/opt/gmc-mqtt/current` |
| doctor exists | `bin/doctor.sh` |
| install/update/uninstall exist | `bin/install.sh`, `bin/update.sh`, `bin/uninstall.sh` |
| shared inspector bootstrap and registration | `bin/muster-bootstrap.sh`, `/etc/muster/implementations.d/gmc-mqtt.json` |
| schema-2 component graph | `muster.yaml` |
| deterministic installed projection | `muster.lock.json` |
| structured doctor evidence | `bin/muster-observation.sh`, `/run/muster/gmc-mqtt/observations/doctor.json` |
| transaction and registration rollback | `bin/release-transaction.sh`, `bin/install.sh`, `bin/update.sh` |
| MPL patterns documented at a verified commit | `muster.yaml`, `MUSTER.md` |
| tests cover core behavior | `go test ./...`, `tests/*.sh` |

## License

MIT License

# AzideMakes GMC-MQTT Configurator Spec

## Purpose
This page lives on `azidemakes.com` and exists to generate a copyable Raspberry Pi install command for GMC-MQTT. It does not install anything directly, it does not host release binaries, and it does not include implementation code in this phase.

The page must always output a fully runnable command in this shape:

```sh
curl -fsSL https://azidemakes.com/install/gmc-mqtt.sh | sudo sh ...
```

## Information Architecture
- Hero: short product statement, one-sentence install promise, and a primary callout that the command is intended for Raspberry Pi OS.
- Configurator: two sections, `Basic Setup` and `Advanced Options`.
- Live Command Preview: sticky on desktop, pinned below the form on mobile.
- Install Notes: three short notes covering `sudo`, supported architectures, and where the config/service files are written.
- Troubleshooting Footer: links or short copy for MQTT host issues, serial permissions, and how to rerun the installer.

## Layout
- Desktop:
  - Two-column layout.
  - Left column contains the form.
  - Right column contains the command preview card, copy button, defaults summary, and generated file locations.
- Mobile:
  - Single-column stacked layout.
  - Command preview appears immediately after the hero and again at the bottom as a persistent summary bar with a copy action.
- Visual direction:
  - Industrial lab aesthetic rather than generic SaaS.
  - Use strong contrast, clean technical typography, and visible grouping between MQTT, device, and service settings.
  - Avoid dark-only presentation; the page must feel readable in bright workshop lighting conditions.

## Form Structure
### Basic Setup
- `serial_device`
  - Control: text input
  - Default: `/dev/ttyUSB0`
  - Help text: `USB serial device exposed by the Geiger counter.`
- `GMC_MQTT_HOST`
  - Control: text input
  - Default: `localhost`
  - Help text: `Hostname or IP address for the MQTT broker.`
- `mqtt_port`
  - Control: numeric input
  - Default: `1883`
  - Help text: `Standard unencrypted MQTT port.`
- `query_interval`
  - Control: numeric input
  - Default: `10`
  - Help text: `Seconds between device polls.`
- `GMC_STATE_TOPIC`
  - Control: text input
  - Default: `gmc/state`
  - Help text: `MQTT topic where the state payload is published.`
- `enable_discovery`
  - Control: segmented toggle
  - Default: `true`
  - Help text: `Publishes Home Assistant discovery entities automatically.`
- `discovery_prefix`
  - Control: text input
  - Default: `homeassistant`
  - Help text: `Discovery topic prefix used by Home Assistant.`
- `device_name`
  - Control: text input
  - Default: `GMC-300`
  - Help text: `Friendly device name shown in Home Assistant.`
- `device_id`
  - Control: text input with placeholder state
  - Default state: empty field labeled `Auto (gmc300_<hostname>)`
  - Help text: `Leave empty to derive a stable ID from the Raspberry Pi hostname.`

### Advanced Options
- MQTT group:
  - `config_topic` default `gmc/config/temp`
  - `permanent_config_topic` default `gmc/config/permanent`
- Serial and service group:
  - `baud_rate` default `9600`
  - `serial_group` default `dialout`
  - `service_name` default `gmc-mqtt`
  - `service_user` default `gmc-mqtt`
- Device metadata group:
  - `device_manufacturer` default `GQ Electronics`
  - `device_model` default `GMC-300S`
  - `device_sw_version` default `unknown`
  - `device_serial` default `unknown`
  - `device_hw_version` default `unknown`
- Origin metadata group:
  - `origin_name` default `GMC-MQTT`
  - `origin_sw` default `installed release version`
  - `origin_url` default `https://github.com/azide0x37/GMC-MQTT`

## Fixed Values Shown as Read-Only
- Active binary path: `/opt/gmc-mqtt/current/bin/gmc-mqtt`
- Active bridge path: `/opt/gmc-mqtt/current/bin/gmc-mqtt-ha-bridge.sh`
- Config files: `/etc/gmc-mqtt/collector.env`, `/etc/gmc-mqtt/mqtt.env`, `/etc/gmc-mqtt/home-assistant.env`
- Service unit paths: `/etc/systemd/system/gmc-mqtt-collector.service`, `/etc/systemd/system/gmc-mqtt-ha-bridge.service`
- Runtime state directory: `/run/muster/gmc-mqtt`
- Durable data directory: `/var/lib/gmc-mqtt`

These values appear in a read-only summary card titled `What the installer creates`.

## Interaction Rules
- The command preview updates on every change with no submit step.
- Default-valued flags are omitted unless the field has been manually edited back to a default and the user explicitly toggles `Show explicit defaults`.
- Boolean values are rendered explicitly when changed from the default.
- Empty `device_id` never emits a `--device-id` flag.
- The copy action copies only the shell command, not explanatory text.
- A secondary action labeled `Reset to defaults` restores all default field values and the command preview.

## Validation and Error Copy
- `GMC_MQTT_HOST`
  - Empty state copy: `Enter the MQTT broker hostname or IP address.`
- `mqtt_port`
  - Invalid state copy: `MQTT port must be a positive integer.`
- `query_interval`
  - Invalid state copy: `Query interval must be at least 1 second.`
- `serial_device`
  - Empty state copy: `Serial device path is required.`
- Generic inline validation behavior:
  - Validate on blur and on copy.
  - Do not block typing with aggressive formatting.
  - Disable the copy action only when the command would be invalid.

## Command Preview Behavior
- Command preview card title: `Install Command`
- Subtitle: `Run this on the Raspberry Pi that is connected to the counter.`
- Desktop behavior:
  - Sticky card while the form scrolls.
  - Includes the rendered command, copy button, and a `Copied` confirmation state.
- Mobile behavior:
  - Full-width preview below the form.
  - A compact sticky footer exposes `Copy Command`.
- Provide a short line below the command:
  - `Downloads the correct release from GitHub Releases and installs a systemd service.`

## Empty, Loading, and Failure States
- Initial state:
  - Render with defaults immediately. No loading spinner.
- Copy success:
  - Show `Command copied` for 2 seconds.
- Copy failure:
  - Show `Copy failed. Select and copy the command manually.`
- Unsupported browser fallback:
  - Keep the command text selectable at all times.

## Content and Tone
- Copy should sound technical and direct.
- Avoid marketing-heavy language in the form itself.
- Use short helper text focused on install consequences, not feature promotion.

## Non-Goals
- No HTML, CSS, JavaScript, or framework implementation.
- No API design.
- No binary hosting.
- No auth or user accounts.

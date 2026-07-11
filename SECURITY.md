# Security

GMC-MQTT stores deployment configuration under `/etc/gmc-mqtt/`.

- `/etc/gmc-mqtt/mqtt.env` is installed with mode `0600` because it may contain broker credentials.
- Read-only Muster inspection consumes configuration metadata, not environment-file contents.
- The explicit doctor action may source configuration to perform checks, requires root, and emits only bounded summaries without secret values.
- Registration paths are constrained under `/opt/gmc-mqtt/` and the installed lock digest must match `muster.yaml`.
- Release archives reject path traversal and link entries before extraction.
- MQTT controls remain disabled; arbitrary remote config mutation is not implemented.

Report issues through the repository issue tracker or the support URL configured in `GMC_ORIGIN_URL`.

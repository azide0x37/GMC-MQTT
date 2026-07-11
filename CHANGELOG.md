# Changelog

## 0.2.0

- Adopt Muster schema 2 and a deterministic `muster.lock.json`.
- Bootstrap and register with the shared Muster CLI server inspector.
- Emit structured doctor observations under `/run/muster/gmc-mqtt/observations/`.
- Serialize release, unit, and registration changes with rollback on validation or doctor failure.
- Harden immutable release validation and archive extraction.

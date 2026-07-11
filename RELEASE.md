# Release Process

1. Update `VERSION` and `CHANGELOG.md`.
2. Compile `muster.yaml` with the release Muster CLI and commit the resulting `muster.lock.json`.
3. Run `make test`.
4. Run `make package`.
5. Run `sh packaging/build-release.sh v<version>` for the Linux architecture matrix.
6. Publish archives, SHA256 files, `checksums.txt`, and the standalone `install.sh`.

The public install command is:

```sh
curl -fsSL https://github.com/azide0x37/GMC-MQTT/releases/latest/download/install.sh | sudo sh
```

The lock manifest digest and locked version must match the release exactly. Install and update validate this before immutable publication. The updater changes units, active release, and inspector registration as one serialized transaction and restores all three when graph validation or doctor evidence rejects a release.

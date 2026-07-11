SHELL := /bin/sh
PROJECT := gmc-mqtt
VERSION := $(shell cat VERSION)
DIST := dist
PACKAGE_ROOT := $(DIST)/$(PROJECT)-$(VERSION)
TARBALL := $(DIST)/$(PROJECT)-$(VERSION).tar.gz

.PHONY: test go-test shell-test render-units package clean doctor

test: go-test shell-test

go-test:
	go test ./...

shell-test:
	sh tests/test_shell_syntax.sh
	sh tests/test_ha_mqtt_bridge.sh
	sh tests/test_installer_staged.sh
	sh tests/test_update_registration_rollback.sh
	sh tests/test_muster_yaml.sh
	@if [ -n "$${MUSTER_CLI_SOURCE:-}" ]; then sh tests/test_inspector_cli.sh; else printf '%s\n' "inspector cli compatibility skipped: MUSTER_CLI_SOURCE not set"; fi
	@if [ -n "$${MUSTER_CLI_SOURCE:-}" ] && [ -d "dist/v$(VERSION)" ]; then sh tests/test_release_installer.sh; else printf '%s\n' "release installer test skipped: build release assets and set MUSTER_CLI_SOURCE"; fi

doctor:
	sh bin/doctor.sh

render-units:
	sh bin/render-units.sh

package: clean
	mkdir -p "$(PACKAGE_ROOT)/bin" "$(PACKAGE_ROOT)/etc/defaults" "$(PACKAGE_ROOT)/systemd"
	go build -trimpath -o "$(PACKAGE_ROOT)/bin/gmc-mqtt" ./cmd
	cp bin/*.sh "$(PACKAGE_ROOT)/bin/"
	cp etc/defaults/*.env "$(PACKAGE_ROOT)/etc/defaults/"
	cp systemd/* "$(PACKAGE_ROOT)/systemd/"
	cp CHANGELOG.md README.md MUSTER.md RELEASE.md SECURITY.md VERSION muster.yaml muster.lock.json go.mod go.sum "$(PACKAGE_ROOT)/"
	chmod 0755 "$(PACKAGE_ROOT)/bin"/*
	COPYFILE_DISABLE=1 tar --no-xattrs -C "$(DIST)" -czf "$(TARBALL)" "$(PROJECT)-$(VERSION)"
	if command -v sha256sum >/dev/null 2>&1; then sha256sum "$(TARBALL)" | awk '{print $$1}' > "$(TARBALL).sha256"; else shasum -a 256 "$(TARBALL)" | awk '{print $$1}' > "$(TARBALL).sha256"; fi
	{ cat bin/release-transaction.sh; printf 'MUSTER_STANDALONE=1\n'; awk 'NR > 1 && $$0 != "# shellcheck disable=SC1091" && $$0 != ". \"$$SCRIPT_DIR/release-transaction.sh\"" { print }' bin/install.sh; } > "$(DIST)/install.sh"
	chmod 0755 "$(DIST)/install.sh"
	printf '{\n' > "$(DIST)/manifest.json"
	printf '  "project": "%s",\n' "$(PROJECT)" >> "$(DIST)/manifest.json"
	printf '  "version": "%s",\n' "$(VERSION)" >> "$(DIST)/manifest.json"
	printf '  "artifact": "%s",\n' "$(PROJECT)-$(VERSION).tar.gz" >> "$(DIST)/manifest.json"
	printf '  "artifact_url": "https://github.com/azide0x37/GMC-MQTT/releases/download/v%s/%s",\n' "$(VERSION)" "$(PROJECT)-$(VERSION).tar.gz" >> "$(DIST)/manifest.json"
	printf '  "sha256": "%s",\n' "$$(cat "$(TARBALL).sha256")" >> "$(DIST)/manifest.json"
	printf '  "installer": "install.sh"\n' >> "$(DIST)/manifest.json"
	printf '}\n' >> "$(DIST)/manifest.json"

clean:
	rm -rf "$(DIST)"

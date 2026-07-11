package config

import (
	"os"
	"testing"
)

func clearEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		key := key
		original, ok := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Unsetenv(%s) failed: %v", key, err)
		}
		t.Cleanup(func() {
			if ok {
				_ = os.Setenv(key, original)
			} else {
				_ = os.Unsetenv(key)
			}
		})
	}
}

func TestLoadConfigFromEnvUsesDefaults(t *testing.T) {
	clearEnv(t, "GMC_SERIAL_DEVICE", "GMC_BAUD_RATE", "GMC_QUERY_INTERVAL", "GMC_STATE_PATH", "GMC_LEDGER_PATH")
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv failed: %v", err)
	}
	if cfg.SerialDevice != DefaultSerialDevice {
		t.Fatalf("SerialDevice = %q", cfg.SerialDevice)
	}
	if cfg.BaudRate != DefaultBaudRate {
		t.Fatalf("BaudRate = %d", cfg.BaudRate)
	}
	if cfg.QueryInterval != DefaultQueryInterval {
		t.Fatalf("QueryInterval = %d", cfg.QueryInterval)
	}
	if cfg.StatePath != DefaultStatePath {
		t.Fatalf("StatePath = %q", cfg.StatePath)
	}
}

func TestLoadConfigFromEnvOverridesDefaults(t *testing.T) {
	t.Setenv("GMC_SERIAL_DEVICE", "/dev/ttyAMA0")
	t.Setenv("GMC_BAUD_RATE", "9600")
	t.Setenv("GMC_QUERY_INTERVAL", "5")
	t.Setenv("GMC_STATE_PATH", "/tmp/gmc/state.json")
	t.Setenv("GMC_LEDGER_PATH", "/tmp/gmc/ledger.jsonl")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv failed: %v", err)
	}
	if cfg.SerialDevice != "/dev/ttyAMA0" {
		t.Fatalf("SerialDevice = %q", cfg.SerialDevice)
	}
	if cfg.BaudRate != 9600 {
		t.Fatalf("BaudRate = %d", cfg.BaudRate)
	}
	if cfg.QueryInterval != 5 {
		t.Fatalf("QueryInterval = %d", cfg.QueryInterval)
	}
	if cfg.StatePath != "/tmp/gmc/state.json" {
		t.Fatalf("StatePath = %q", cfg.StatePath)
	}
}

func TestValidateConfigRejectsInvalidNumbers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.QueryInterval = 0
	if err := ValidateConfig(&cfg); err == nil {
		t.Fatalf("expected validation error")
	}
}

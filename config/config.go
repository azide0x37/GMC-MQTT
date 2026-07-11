package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	DefaultSerialDevice  = "/dev/ttyUSB0"
	DefaultBaudRate      = 115200
	DefaultQueryInterval = 1
	DefaultStatePath     = "/run/muster/gmc-mqtt/state.json"
	DefaultLedgerPath    = "/var/lib/gmc-mqtt/ledger.jsonl"
)

// Config holds collector-only runtime configuration. MQTT and Home Assistant
// settings are owned by the Muster bridge scripts.
type Config struct {
	SerialDevice  string
	BaudRate      int
	QueryInterval int
	StatePath     string
	LedgerPath    string
}

func DefaultConfig() Config {
	return Config{
		SerialDevice:  DefaultSerialDevice,
		BaudRate:      DefaultBaudRate,
		QueryInterval: DefaultQueryInterval,
		StatePath:     DefaultStatePath,
		LedgerPath:    DefaultLedgerPath,
	}
}

func LoadConfigFromEnv() (Config, error) {
	cfg := DefaultConfig()

	cfg.SerialDevice = stringFromEnv("GMC_SERIAL_DEVICE", cfg.SerialDevice)
	cfg.BaudRate = intFromEnv("GMC_BAUD_RATE", cfg.BaudRate)
	cfg.QueryInterval = intFromEnv("GMC_QUERY_INTERVAL", cfg.QueryInterval)
	cfg.StatePath = stringFromEnv("GMC_STATE_PATH", cfg.StatePath)
	cfg.LedgerPath = stringFromEnv("GMC_LEDGER_PATH", cfg.LedgerPath)

	if err := ValidateConfig(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func ValidateConfig(cfg *Config) error {
	if cfg.SerialDevice == "" {
		return fmt.Errorf("GMC_SERIAL_DEVICE is required")
	}
	if cfg.BaudRate <= 0 {
		return fmt.Errorf("GMC_BAUD_RATE must be > 0")
	}
	if cfg.QueryInterval <= 0 {
		return fmt.Errorf("GMC_QUERY_INTERVAL must be > 0")
	}
	if cfg.StatePath == "" {
		return fmt.Errorf("GMC_STATE_PATH is required")
	}
	if cfg.LedgerPath == "" {
		return fmt.Errorf("GMC_LEDGER_PATH is required")
	}
	return nil
}

func stringFromEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func intFromEnv(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

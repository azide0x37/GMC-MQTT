package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config holds the configuration loaded from the TOML file.
type Config struct {
	SerialDevice  string `toml:"serial_device"`
	BaudRate      int    `toml:"baud_rate"`
	MQTTHost      string `toml:"mqtt_host"`
	MQTTPort      int    `toml:"mqtt_port"`
	QueryInterval int    `toml:"query_interval"` // seconds between queries
	PublishTopic  string `toml:"publish_topic"`  // legacy alias for state_topic

	ConfigTopic          string `toml:"config_topic"`
	PermanentConfigTopic string `toml:"permanent_config_topic"`

	// Discovery settings for Home Assistant MQTT Discovery
	EnableDiscovery    bool   `toml:"enable_discovery"`
	DiscoveryPrefix    string `toml:"discovery_prefix"` // default "homeassistant"
	DeviceID           string `toml:"device_id"`        // unique device id, e.g., "gmc300_001"
	DeviceName         string `toml:"device_name"`      // e.g., "GMC-300"
	DeviceManufacturer string `toml:"device_manufacturer"`
	DeviceModel        string `toml:"device_model"`
	DeviceSWVersion    string `toml:"device_sw_version"`
	DeviceSerial       string `toml:"device_serial"`
	DeviceHWVersion    string `toml:"device_hw_version"`

	// Origin information for discovery
	OriginName string `toml:"origin_name"`
	OriginSW   string `toml:"origin_sw"`
	OriginURL  string `toml:"origin_url"`

	// StateTopic: where sensor state messages are published
	StateTopic string `toml:"state_topic"`
}

func normalizeConfig(cfg *Config) {
	if cfg.StateTopic == "" && cfg.PublishTopic != "" {
		cfg.StateTopic = cfg.PublishTopic
	}
	if cfg.StateTopic != "" && cfg.PublishTopic != cfg.StateTopic {
		cfg.PublishTopic = cfg.StateTopic
	}
	if cfg.DiscoveryPrefix == "" {
		cfg.DiscoveryPrefix = "homeassistant"
	}
}

func ValidateConfig(cfg *Config) error {
	if cfg.SerialDevice == "" {
		return fmt.Errorf("serial_device is required")
	}
	if cfg.MQTTHost == "" {
		return fmt.Errorf("mqtt_host is required")
	}
	if cfg.MQTTPort <= 0 {
		return fmt.Errorf("mqtt_port must be > 0")
	}
	if cfg.QueryInterval <= 0 {
		return fmt.Errorf("query_interval must be > 0")
	}
	if cfg.StateTopic == "" {
		return fmt.Errorf("state_topic is required")
	}
	return nil
}

// LoadConfig loads the configuration from the given TOML file.
func LoadConfig(path string) (*Config, error) {
	var cfg Config

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err := toml.DecodeReader(file, &cfg); err != nil {
		return nil, err
	}
	normalizeConfig(&cfg)
	if err := ValidateConfig(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

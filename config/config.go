package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

// Config holds the configuration loaded from the TOML file.
type Config struct {
	SerialDevice         string `toml:"serial_device"`
	BaudRate             int    `toml:"baud_rate"`
	MQTTHost             string `toml:"mqtt_host"`
	MQTTPort             int    `toml:"mqtt_port"`
	QueryInterval        int    `toml:"query_interval"` // seconds between queries
	PublishTopic         string `toml:"publish_topic"`
	ConfigTopic          string `toml:"config_topic"`
	PermanentConfigTopic string `toml:"permanent_config_topic"`
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
	return &cfg, nil
}

package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file for testing.
	tmpFile, err := os.CreateTemp("", "config*.toml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	configContent := `
serial_device = "/dev/ttyUSB0"
baud_rate = 9600
mqtt_host = "localhost"
mqtt_port = 1883
query_interval = 10
state_topic = "gmc/state"
config_topic = "gmc/config/temp"
permanent_config_topic = "gmc/config/permanent"
`
	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	if cfg.MQTTHost != "localhost" {
		t.Errorf("Expected MQTTHost to be 'localhost', got %s", cfg.MQTTHost)
	}
	if cfg.StateTopic != "gmc/state" {
		t.Errorf("Expected StateTopic to be 'gmc/state', got %s", cfg.StateTopic)
	}
	if cfg.PublishTopic != "gmc/state" {
		t.Errorf("Expected PublishTopic to mirror StateTopic, got %s", cfg.PublishTopic)
	}
}

func TestValidateConfigRequiresFields(t *testing.T) {
	cfg := &Config{}
	if err := ValidateConfig(cfg); err == nil {
		t.Fatalf("expected validation error for missing fields")
	}
}

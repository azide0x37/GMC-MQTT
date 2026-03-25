package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
)

type UpdateSummary struct {
	Changed []string

	RestartRequired bool

	QueryIntervalChanged bool
	OldQueryInterval     int
	NewQueryInterval     int

	StateTopicChanged bool
	OldStateTopic     string
	NewStateTopic     string

	ConfigTopicChanged bool
	OldConfigTopic     string
	NewConfigTopic     string

	PermanentConfigTopicChanged bool
	OldPermanentConfigTopic     string
	NewPermanentConfigTopic     string

	SerialDeviceChanged bool
	OldSerialDevice     string
	NewSerialDevice     string

	BaudRateChanged bool
	OldBaudRate     int
	NewBaudRate     int

	MQTTHostChanged bool
	OldMQTTHost     string
	NewMQTTHost     string

	MQTTPortChanged bool
	OldMQTTPort     int
	NewMQTTPort     int
}

type Manager struct {
	mu      sync.RWMutex
	cfg     Config
	path    string
	updates chan UpdateSummary
}

// ConfigUpdate represents a partial configuration update.
type ConfigUpdate struct {
	SerialDevice         *string `json:"serial_device"`
	BaudRate             *int    `json:"baud_rate"`
	MQTTHost             *string `json:"mqtt_host"`
	MQTTPort             *int    `json:"mqtt_port"`
	QueryInterval        *int    `json:"query_interval"`
	PublishTopic         *string `json:"publish_topic"`
	ConfigTopic          *string `json:"config_topic"`
	PermanentConfigTopic *string `json:"permanent_config_topic"`
	StateTopic           *string `json:"state_topic"`

	EnableDiscovery    *bool   `json:"enable_discovery"`
	DiscoveryPrefix    *string `json:"discovery_prefix"`
	DeviceID           *string `json:"device_id"`
	DeviceName         *string `json:"device_name"`
	DeviceManufacturer *string `json:"device_manufacturer"`
	DeviceModel        *string `json:"device_model"`
	DeviceSWVersion    *string `json:"device_sw_version"`
	DeviceSerial       *string `json:"device_serial"`
	DeviceHWVersion    *string `json:"device_hw_version"`

	OriginName *string `json:"origin_name"`
	OriginSW   *string `json:"origin_sw"`
	OriginURL  *string `json:"origin_url"`
}

func NewManager(cfg *Config, path string) *Manager {
	cpy := *cfg
	normalizeConfig(&cpy)
	return &Manager{
		cfg:     cpy,
		path:    path,
		updates: make(chan UpdateSummary, 10),
	}
}

func (m *Manager) Get() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

func (m *Manager) Updates() <-chan UpdateSummary {
	return m.updates
}

func (m *Manager) ApplyJSONUpdate(payload []byte, permanent bool) (UpdateSummary, error) {
	var update ConfigUpdate
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&update); err != nil {
		return UpdateSummary{}, fmt.Errorf("invalid config JSON: %w", err)
	}
	return m.ApplyUpdate(update, permanent)
}

func (m *Manager) ApplyUpdate(update ConfigUpdate, permanent bool) (UpdateSummary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if update.QueryInterval != nil && *update.QueryInterval <= 0 {
		return UpdateSummary{}, fmt.Errorf("query_interval must be > 0")
	}
	if update.MQTTPort != nil && *update.MQTTPort <= 0 {
		return UpdateSummary{}, fmt.Errorf("mqtt_port must be > 0")
	}
	if update.BaudRate != nil && *update.BaudRate <= 0 {
		return UpdateSummary{}, fmt.Errorf("baud_rate must be > 0")
	}

	before := m.cfg

	if update.SerialDevice != nil {
		m.cfg.SerialDevice = *update.SerialDevice
	}
	if update.BaudRate != nil {
		m.cfg.BaudRate = *update.BaudRate
	}
	if update.MQTTHost != nil {
		m.cfg.MQTTHost = *update.MQTTHost
	}
	if update.MQTTPort != nil {
		m.cfg.MQTTPort = *update.MQTTPort
	}
	if update.QueryInterval != nil {
		m.cfg.QueryInterval = *update.QueryInterval
	}
	if update.PublishTopic != nil {
		m.cfg.PublishTopic = *update.PublishTopic
		if update.StateTopic == nil {
			m.cfg.StateTopic = *update.PublishTopic
		}
	}
	if update.ConfigTopic != nil {
		m.cfg.ConfigTopic = *update.ConfigTopic
	}
	if update.PermanentConfigTopic != nil {
		m.cfg.PermanentConfigTopic = *update.PermanentConfigTopic
	}
	if update.StateTopic != nil {
		m.cfg.StateTopic = *update.StateTopic
	}
	if update.EnableDiscovery != nil {
		m.cfg.EnableDiscovery = *update.EnableDiscovery
	}
	if update.DiscoveryPrefix != nil {
		m.cfg.DiscoveryPrefix = *update.DiscoveryPrefix
	}
	if update.DeviceID != nil {
		m.cfg.DeviceID = *update.DeviceID
	}
	if update.DeviceName != nil {
		m.cfg.DeviceName = *update.DeviceName
	}
	if update.DeviceManufacturer != nil {
		m.cfg.DeviceManufacturer = *update.DeviceManufacturer
	}
	if update.DeviceModel != nil {
		m.cfg.DeviceModel = *update.DeviceModel
	}
	if update.DeviceSWVersion != nil {
		m.cfg.DeviceSWVersion = *update.DeviceSWVersion
	}
	if update.DeviceSerial != nil {
		m.cfg.DeviceSerial = *update.DeviceSerial
	}
	if update.DeviceHWVersion != nil {
		m.cfg.DeviceHWVersion = *update.DeviceHWVersion
	}
	if update.OriginName != nil {
		m.cfg.OriginName = *update.OriginName
	}
	if update.OriginSW != nil {
		m.cfg.OriginSW = *update.OriginSW
	}
	if update.OriginURL != nil {
		m.cfg.OriginURL = *update.OriginURL
	}

	normalizeConfig(&m.cfg)

	summary := diffConfigs(before, m.cfg)
	if len(summary.Changed) == 0 {
		return summary, nil
	}

	if permanent {
		if err := m.saveLocked(); err != nil {
			m.cfg = before
			return summary, err
		}
	}

	select {
	case m.updates <- summary:
	default:
		// Drop if the channel is full to avoid blocking MQTT handlers.
	}

	return summary, nil
}

func (m *Manager) saveLocked() error {
	if m.path == "" {
		return fmt.Errorf("config path is empty")
	}

	dir := filepath.Dir(m.path)
	tmpFile, err := os.CreateTemp(dir, "config-*.toml")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	encoder := toml.NewEncoder(tmpFile)
	if err := encoder.Encode(m.cfg); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	return os.Rename(tmpFile.Name(), m.path)
}

func diffConfigs(before, after Config) UpdateSummary {
	summary := UpdateSummary{}

	if before.QueryInterval != after.QueryInterval {
		summary.QueryIntervalChanged = true
		summary.OldQueryInterval = before.QueryInterval
		summary.NewQueryInterval = after.QueryInterval
		summary.Changed = append(summary.Changed, "query_interval")
	}
	if before.StateTopic != after.StateTopic {
		summary.StateTopicChanged = true
		summary.OldStateTopic = before.StateTopic
		summary.NewStateTopic = after.StateTopic
		summary.Changed = append(summary.Changed, "state_topic")
	}
	if before.ConfigTopic != after.ConfigTopic {
		summary.ConfigTopicChanged = true
		summary.OldConfigTopic = before.ConfigTopic
		summary.NewConfigTopic = after.ConfigTopic
		summary.Changed = append(summary.Changed, "config_topic")
	}
	if before.PermanentConfigTopic != after.PermanentConfigTopic {
		summary.PermanentConfigTopicChanged = true
		summary.OldPermanentConfigTopic = before.PermanentConfigTopic
		summary.NewPermanentConfigTopic = after.PermanentConfigTopic
		summary.Changed = append(summary.Changed, "permanent_config_topic")
	}
	if before.SerialDevice != after.SerialDevice {
		summary.SerialDeviceChanged = true
		summary.OldSerialDevice = before.SerialDevice
		summary.NewSerialDevice = after.SerialDevice
		summary.Changed = append(summary.Changed, "serial_device")
	}
	if before.BaudRate != after.BaudRate {
		summary.BaudRateChanged = true
		summary.OldBaudRate = before.BaudRate
		summary.NewBaudRate = after.BaudRate
		summary.Changed = append(summary.Changed, "baud_rate")
	}
	if before.MQTTHost != after.MQTTHost {
		summary.MQTTHostChanged = true
		summary.OldMQTTHost = before.MQTTHost
		summary.NewMQTTHost = after.MQTTHost
		summary.Changed = append(summary.Changed, "mqtt_host")
	}
	if before.MQTTPort != after.MQTTPort {
		summary.MQTTPortChanged = true
		summary.OldMQTTPort = before.MQTTPort
		summary.NewMQTTPort = after.MQTTPort
		summary.Changed = append(summary.Changed, "mqtt_port")
	}
	if before.PublishTopic != after.PublishTopic {
		summary.Changed = append(summary.Changed, "publish_topic")
	}
	if before.EnableDiscovery != after.EnableDiscovery {
		summary.Changed = append(summary.Changed, "enable_discovery")
	}
	if before.DiscoveryPrefix != after.DiscoveryPrefix {
		summary.Changed = append(summary.Changed, "discovery_prefix")
	}
	if before.DeviceID != after.DeviceID {
		summary.Changed = append(summary.Changed, "device_id")
	}
	if before.DeviceName != after.DeviceName {
		summary.Changed = append(summary.Changed, "device_name")
	}
	if before.DeviceManufacturer != after.DeviceManufacturer {
		summary.Changed = append(summary.Changed, "device_manufacturer")
	}
	if before.DeviceModel != after.DeviceModel {
		summary.Changed = append(summary.Changed, "device_model")
	}
	if before.DeviceSWVersion != after.DeviceSWVersion {
		summary.Changed = append(summary.Changed, "device_sw_version")
	}
	if before.DeviceSerial != after.DeviceSerial {
		summary.Changed = append(summary.Changed, "device_serial")
	}
	if before.DeviceHWVersion != after.DeviceHWVersion {
		summary.Changed = append(summary.Changed, "device_hw_version")
	}
	if before.OriginName != after.OriginName {
		summary.Changed = append(summary.Changed, "origin_name")
	}
	if before.OriginSW != after.OriginSW {
		summary.Changed = append(summary.Changed, "origin_sw")
	}
	if before.OriginURL != after.OriginURL {
		summary.Changed = append(summary.Changed, "origin_url")
	}

	summary.RestartRequired = summary.SerialDeviceChanged || summary.BaudRateChanged || summary.MQTTHostChanged || summary.MQTTPortChanged
	return summary
}

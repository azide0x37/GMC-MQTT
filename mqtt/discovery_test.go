package mqtt

import (
	"encoding/json"
	"testing"

	"github.com/azide0x37/gmc-mqtt/config"
)

func TestDiscoveryTopic(t *testing.T) {
	cfg := &config.Config{DiscoveryPrefix: "homeassistant"}
	topic := DiscoveryTopic(cfg, "gmc300_cpm")
	if topic != "homeassistant/sensor/gmc300_cpm/config" {
		t.Fatalf("Unexpected topic: %s", topic)
	}
}

func TestMarshalDiscoveryPayload(t *testing.T) {
	payload := SensorDiscoveryPayload{
		Device:            map[string]interface{}{"identifiers": []string{"gmc300_001"}},
		Origin:            map[string]interface{}{"name": "GMC-MQTT"},
		StateClass:        "measurement",
		UnitOfMeasurement: "CPM",
		ValueTemplate:     "{{ value_json.cpm }}",
		UniqueID:          "gmc300_cpm",
		StateTopic:        "gmc/state",
		QoS:               1,
	}

	data, err := MarshalDiscoveryPayload(payload)
	if err != nil {
		t.Fatalf("MarshalDiscoveryPayload failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to decode payload: %v", err)
	}
	if decoded["state_topic"] != "gmc/state" {
		t.Fatalf("Expected state_topic=gmc/state, got %v", decoded["state_topic"])
	}
	if decoded["unique_id"] != "gmc300_cpm" {
		t.Fatalf("Expected unique_id=gmc300_cpm, got %v", decoded["unique_id"])
	}
}

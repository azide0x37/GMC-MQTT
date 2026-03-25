package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/azide0x37/gmc-mqtt/config"
)

// SensorDiscoveryPayload represents the payload for MQTT discovery for a sensor.
type SensorDiscoveryPayload struct {
	Device            map[string]interface{} `json:"device"`
	Origin            map[string]interface{} `json:"origin"`
	DeviceClass       string                 `json:"device_class,omitempty"`
	StateClass        string                 `json:"state_class,omitempty"`
	UnitOfMeasurement string                 `json:"unit_of_measurement,omitempty"`
	ValueTemplate     string                 `json:"value_template,omitempty"`
	UniqueID          string                 `json:"unique_id,omitempty"`
	Name              *string                `json:"name"` // if nil, Home Assistant derives the name from device info
	StateTopic        string                 `json:"state_topic"`
	QoS               int                    `json:"qos"`
}

// DiscoveryTopic builds the Home Assistant discovery topic for a sensor.
func DiscoveryTopic(cfg *config.Config, discoveryID string) string {
	return fmt.Sprintf("%s/sensor/%s/config", cfg.DiscoveryPrefix, discoveryID)
}

// MarshalDiscoveryPayload encodes the discovery payload as JSON.
func MarshalDiscoveryPayload(sensorPayload SensorDiscoveryPayload) ([]byte, error) {
	return json.Marshal(sensorPayload)
}

// PublishDiscovery publishes a discovery message for a given sensor to the MQTT broker.
func PublishDiscovery(client mqtt.Client, cfg *config.Config, discoveryID string, sensorPayload SensorDiscoveryPayload) error {
	// Discovery topic format: <discovery_prefix>/<component>/<object_id>/config; here component is "sensor".
	topic := DiscoveryTopic(cfg, discoveryID)
	payloadBytes, err := MarshalDiscoveryPayload(sensorPayload)
	if err != nil {
		return err
	}
	log.Printf("Publishing discovery message to topic %s: %s", topic, string(payloadBytes))
	token := client.Publish(topic, byte(sensorPayload.QoS), true, payloadBytes)
	if !token.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("timeout publishing discovery message")
	}
	return token.Error()
}

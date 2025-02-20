package mqtt

import (
	"fmt"
	"log"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/azide0x37/gmc-mqtt/config"
)

// NewMQTTClient creates and returns an MQTT client based on the provided configuration.
func NewMQTTClient(cfg *config.Config) (mqtt.Client, error) {
	brokerURL := fmt.Sprintf("tcp://%s:%d", cfg.MQTTHost, cfg.MQTTPort)
	opts := mqtt.NewClientOptions().AddBroker(brokerURL)
	opts.SetClientID("gmc-mqtt-client-" + strconv.Itoa(int(time.Now().UnixNano())))
	// Optional: set additional options such as auto-reconnect.
	opts.OnConnect = func(c mqtt.Client) {
		log.Println("MQTT connected")
	}
	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		return nil, token.Error()
	}
	return client, nil
}

// ConfigMessageHandler handles temporary configuration update messages.
func ConfigMessageHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Temporary configuration update received on topic %s: %s", msg.Topic(), string(msg.Payload()))
	// TODO: Implement temporary configuration update logic (apply changes in-memory)
}

// PermanentConfigMessageHandler handles permanent configuration update messages.
func PermanentConfigMessageHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Permanent configuration update received on topic %s: %s", msg.Topic(), string(msg.Payload()))
	// TODO: Implement permanent configuration update logic (e.g., update config file)
}

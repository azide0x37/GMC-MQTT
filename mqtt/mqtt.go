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
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(2 * time.Second)
	opts.SetMaxReconnectInterval(30 * time.Second)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(5 * time.Second)
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
func ConfigMessageHandler(manager *config.Manager) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("Temporary configuration update received on topic %s: %s", msg.Topic(), string(msg.Payload()))
		summary, err := manager.ApplyJSONUpdate(msg.Payload(), false)
		if err != nil {
			log.Printf("Temporary configuration update failed: %v", err)
			return
		}
		if len(summary.Changed) == 0 {
			log.Printf("Temporary configuration update made no changes.")
			return
		}
		log.Printf("Temporary configuration updated: %v", summary.Changed)
		if summary.RestartRequired {
			log.Printf("Restart required for changes: %v", summary.Changed)
		}
	}
}

// PermanentConfigMessageHandler handles permanent configuration update messages.
func PermanentConfigMessageHandler(manager *config.Manager) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("Permanent configuration update received on topic %s: %s", msg.Topic(), string(msg.Payload()))
		summary, err := manager.ApplyJSONUpdate(msg.Payload(), true)
		if err != nil {
			log.Printf("Permanent configuration update failed: %v", err)
			return
		}
		if len(summary.Changed) == 0 {
			log.Printf("Permanent configuration update made no changes.")
			return
		}
		log.Printf("Permanent configuration updated and persisted: %v", summary.Changed)
		if summary.RestartRequired {
			log.Printf("Restart required for changes: %v", summary.Changed)
		}
	}
}

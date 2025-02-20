package main

import (
	"flag"
	"log"
	"time"

	"github.com/azide0x37/gmc-mqtt/config"
	"github.com/azide0x37/gmc-mqtt/gmc"
	"github.com/azide0x37/gmc-mqtt/mqtt"
)

func main() {
	// Parse command-line flag for the configuration file path.
	configPath := flag.String("config", "config.toml", "Path to TOML configuration file")
	flag.Parse()

	// Load configuration.
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Println("Configuration loaded successfully.")

	// Initialize MQTT client.
	mqttClient, err := mqtt.NewMQTTClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create MQTT client: %v", err)
	}
	log.Println("Connected to MQTT broker.")

	// Subscribe to configuration update topics.
	if token := mqttClient.Subscribe(cfg.ConfigTopic, 0, mqtt.ConfigMessageHandler); token.Wait() && token.Error() != nil {
		log.Printf("Error subscribing to temp config topic: %v", token.Error())
	} else {
		log.Printf("Subscribed to temporary config topic: %s", cfg.ConfigTopic)
	}

	if token := mqttClient.Subscribe(cfg.PermanentConfigTopic, 0, mqtt.PermanentConfigMessageHandler); token.Wait() && token.Error() != nil {
		log.Printf("Error subscribing to permanent config topic: %v", token.Error())
	} else {
		log.Printf("Subscribed to permanent config topic: %s", cfg.PermanentConfigTopic)
	}

	// Initialize the GMC device over serial.
	gmcDevice, err := gmc.NewGMC(cfg.SerialDevice, cfg.BaudRate)
	if err != nil {
		log.Fatalf("Failed to open serial device: %v", err)
	}
	defer gmcDevice.Close()
	log.Printf("Connected to GMC device on %s.", cfg.SerialDevice)

	// Main loop: query the device and publish data at the specified interval.
	ticker := time.NewTicker(time.Duration(cfg.QueryInterval) * time.Second)
	defer ticker.Stop()

	log.Println("Starting main loop.")
	for {
		<-ticker.C

		// Query device for data.
		data, err := gmcDevice.Query()
		if err != nil {
			log.Printf("Error querying GMC device: %v", err)
			continue
		}
		log.Printf("Data received: %s", data)

		// Publish data to MQTT.
		token := mqttClient.Publish(cfg.PublishTopic, 0, false, data)
		token.Wait()
		if token.Error() != nil {
			log.Printf("Error publishing data: %v", token.Error())
		} else {
			log.Printf("Data published to topic %s.", cfg.PublishTopic)
		}
	}
}

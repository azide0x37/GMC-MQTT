package main

import (
	"encoding/json"
	"flag"
	"log"
	"time"

	"github.com/azide0x37/gmc-mqtt/config"
	"github.com/azide0x37/gmc-mqtt/gmc"
	"github.com/azide0x37/gmc-mqtt/mqtt"
)

type deviceInfo struct {
	model    string
	firmware string
	serial   string
}

func openGMCDevice(cfg config.Config) (*gmc.GMC, deviceInfo, error) {
	device, err := gmc.NewGMC(cfg.SerialDevice, cfg.BaudRate)
	if err != nil {
		return nil, deviceInfo{}, err
	}

	info := deviceInfo{}
	model, fw, err := device.QueryVersion()
	if err != nil {
		log.Printf("Error querying version: %v", err)
	} else {
		info.model = model
		info.firmware = fw
		log.Printf("Device model: %s, firmware: %s", model, fw)
	}
	serialStr, err := device.QuerySerial()
	if err != nil {
		log.Printf("Error querying serial: %v", err)
	} else {
		info.serial = serialStr
		log.Printf("Device serial: %s", serialStr)
	}
	return device, info, nil
}

func subscribeTopic(client mqtt.Client, topic string, handler mqtt.MessageHandler, label string) {
	if topic == "" {
		return
	}
	if token := client.Subscribe(topic, 0, handler); token.Wait() && token.Error() != nil {
		log.Printf("Error subscribing to %s topic %s: %v", label, topic, token.Error())
		return
	}
	log.Printf("Subscribed to %s topic: %s", label, topic)
}

func unsubscribeTopic(client mqtt.Client, topic string, label string) {
	if topic == "" {
		return
	}
	if token := client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		log.Printf("Error unsubscribing from %s topic %s: %v", label, topic, token.Error())
		return
	}
	log.Printf("Unsubscribed from %s topic: %s", label, topic)
}

func main() {
	configPath := flag.String("config", "config.toml", "Path to TOML configuration file")
	flag.Parse()

	// Load configuration.
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Println("Configuration loaded successfully.")

	cfgManager := config.NewManager(cfg, *configPath)
	cfgSnapshot := cfgManager.Get()

	// Initialize MQTT client.
	mqttClient, err := mqtt.NewMQTTClient(&cfgSnapshot)
	if err != nil {
		log.Fatalf("Failed to create MQTT client: %v", err)
	}
	log.Println("Connected to MQTT broker.")

	tempHandler := mqtt.ConfigMessageHandler(cfgManager)
	permHandler := mqtt.PermanentConfigMessageHandler(cfgManager)

	// Subscribe to configuration update topics.
	subscribeTopic(mqttClient, cfgSnapshot.ConfigTopic, tempHandler, "temporary config")
	subscribeTopic(mqttClient, cfgSnapshot.PermanentConfigTopic, permHandler, "permanent config")

	// Publish discovery messages if enabled.
	if cfgSnapshot.EnableDiscovery {
		// Define explicit sensor names.
		nameCPM := "CPM"
		nameBattery := "Battery Voltage"
		nameVersion := "Firmware Version"
		nameSerial := "Serial Number"
		nameUptime := "Uptime"
		nameUsv := "µSv/h"
		nameMr := "mR/h"

		deviceInfo := map[string]interface{}{
			"identifiers":   []string{cfgSnapshot.DeviceID},
			"name":          cfgSnapshot.DeviceName,
			"manufacturer":  cfgSnapshot.DeviceManufacturer,
			"model":         cfgSnapshot.DeviceModel,
			"sw_version":    cfgSnapshot.DeviceSWVersion,
			"serial_number": cfgSnapshot.DeviceSerial,
			"hw_version":    cfgSnapshot.DeviceHWVersion,
		}
		originInfo := map[string]interface{}{
			"name":        cfgSnapshot.OriginName,
			"sw_version":  cfgSnapshot.OriginSW,
			"support_url": cfgSnapshot.OriginURL,
		}

		// Discovery for CPM sensor.
		cpmPayload := mqtt.SensorDiscoveryPayload{
			Device:            deviceInfo,
			Origin:            originInfo,
			StateClass:        "measurement",
			UnitOfMeasurement: "CPM",
			ValueTemplate:     "{{ value_json.cpm }}",
			UniqueID:          "gmc300_cpm",
			Name:              &nameCPM,
			StateTopic:        cfgSnapshot.StateTopic,
			QoS:               2,
		}
		if err := mqtt.PublishDiscovery(mqttClient, &cfgSnapshot, "gmc300_cpm", cpmPayload); err != nil {
			log.Printf("Error publishing CPM discovery: %v", err)
		}

		// Discovery for Battery sensor.
		batteryPayload := mqtt.SensorDiscoveryPayload{
			Device:            deviceInfo,
			Origin:            originInfo,
			DeviceClass:       "battery",
			UnitOfMeasurement: "V",
			ValueTemplate:     "{{ value_json.battery }}",
			UniqueID:          "gmc300_battery",
			Name:              &nameBattery,
			StateTopic:        cfgSnapshot.StateTopic,
			QoS:               2,
		}
		if err := mqtt.PublishDiscovery(mqttClient, &cfgSnapshot, "gmc300_battery", batteryPayload); err != nil {
			log.Printf("Error publishing battery discovery: %v", err)
		}

		// Discovery for Firmware Version sensor.
		versionPayload := mqtt.SensorDiscoveryPayload{
			Device:        deviceInfo,
			Origin:        originInfo,
			ValueTemplate: "{{ value_json.version }}",
			UniqueID:      "gmc300_version",
			Name:          &nameVersion,
			StateTopic:    cfgSnapshot.StateTopic,
			QoS:           2,
		}
		if err := mqtt.PublishDiscovery(mqttClient, &cfgSnapshot, "gmc300_version", versionPayload); err != nil {
			log.Printf("Error publishing version discovery: %v", err)
		}

		// Discovery for Serial Number sensor.
		serialPayload := mqtt.SensorDiscoveryPayload{
			Device:        deviceInfo,
			Origin:        originInfo,
			ValueTemplate: "{{ value_json.serial }}",
			UniqueID:      "gmc300_serial",
			Name:          &nameSerial,
			StateTopic:    cfgSnapshot.StateTopic,
			QoS:           2,
		}
		if err := mqtt.PublishDiscovery(mqttClient, &cfgSnapshot, "gmc300_serial", serialPayload); err != nil {
			log.Printf("Error publishing serial discovery: %v", err)
		}

		// Discovery for Uptime sensor.
		uptimePayload := mqtt.SensorDiscoveryPayload{
			Device:            deviceInfo,
			Origin:            originInfo,
			UnitOfMeasurement: "s",
			ValueTemplate:     "{{ value_json.uptime }}",
			UniqueID:          "gmc300_uptime",
			Name:              &nameUptime,
			StateTopic:        cfgSnapshot.StateTopic,
			QoS:               2,
		}
		if err := mqtt.PublishDiscovery(mqttClient, &cfgSnapshot, "gmc300_uptime", uptimePayload); err != nil {
			log.Printf("Error publishing uptime discovery: %v", err)
		}

		// Discovery for uSv/h sensor.
		usvPayload := mqtt.SensorDiscoveryPayload{
			Device:            deviceInfo,
			Origin:            originInfo,
			StateClass:        "measurement",
			UnitOfMeasurement: "µSv/h",
			ValueTemplate:     "{{ value_json.usv }}",
			UniqueID:          "gmc300_usv",
			Name:              &nameUsv,
			StateTopic:        cfgSnapshot.StateTopic,
			QoS:               2,
		}
		if err := mqtt.PublishDiscovery(mqttClient, &cfgSnapshot, "gmc300_usv", usvPayload); err != nil {
			log.Printf("Error publishing uSv/h discovery: %v", err)
		}

		// Discovery for mR/h sensor.
		mrPayload := mqtt.SensorDiscoveryPayload{
			Device:            deviceInfo,
			Origin:            originInfo,
			StateClass:        "measurement",
			UnitOfMeasurement: "mR/h",
			ValueTemplate:     "{{ value_json.mr }}",
			UniqueID:          "gmc300_mr",
			Name:              &nameMr,
			StateTopic:        cfgSnapshot.StateTopic,
			QoS:               2,
		}
		if err := mqtt.PublishDiscovery(mqttClient, &cfgSnapshot, "gmc300_mr", mrPayload); err != nil {
			log.Printf("Error publishing mR/h discovery: %v", err)
		}
	}

	// Initialize the GMC device over serial.
	gmcDevice, info, err := openGMCDevice(cfgSnapshot)
	if err != nil {
		log.Fatalf("Failed to open serial device: %v", err)
	}
	defer gmcDevice.Close()
	log.Printf("Connected to GMC device on %s.", cfgSnapshot.SerialDevice)

	model := info.model
	fw := info.firmware
	serialStr := info.serial

	// Record start time for uptime calculation.
	startTime := time.Now()

	interval := cfgSnapshot.QueryInterval
	if interval <= 0 {
		interval = 10
		log.Printf("Invalid query_interval; defaulting to %d seconds.", interval)
	}

	// Main loop: query CPM and voltage, compute additional metrics, and publish combined state.
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	log.Println("Starting main loop.")
	for {
		select {
		case <-ticker.C:
			cfgSnapshot = cfgManager.Get()

			cpm, err := gmcDevice.QueryCPM()
			if err != nil {
				log.Printf("Error querying CPM: %v", err)
				_ = gmcDevice.Close()
				time.Sleep(2 * time.Second)
				gmcDevice, info, err = openGMCDevice(cfgSnapshot)
				if err != nil {
					log.Printf("Failed to reopen serial device: %v", err)
					continue
				}
				model = info.model
				fw = info.firmware
				serialStr = info.serial
				continue
			}
			voltage, err := gmcDevice.QueryVoltage()
			if err != nil {
				log.Printf("Error querying battery voltage: %v", err)
				_ = gmcDevice.Close()
				time.Sleep(2 * time.Second)
				gmcDevice, info, err = openGMCDevice(cfgSnapshot)
				if err != nil {
					log.Printf("Failed to reopen serial device: %v", err)
					continue
				}
				model = info.model
				fw = info.firmware
				serialStr = info.serial
				continue
			}
			uptime := int(time.Since(startTime).Seconds())

			statePayload, err := json.Marshal(buildState(cpm, voltage, fw, model, serialStr, uptime))
			if err != nil {
				log.Printf("Error marshalling sensor state: %v", err)
				continue
			}
			if cfgSnapshot.StateTopic == "" {
				log.Printf("State topic is empty; skipping publish.")
				continue
			}
			token := mqttClient.Publish(cfgSnapshot.StateTopic, 0, false, statePayload)
			token.Wait()
			if token.Error() != nil {
				log.Printf("Error publishing sensor state: %v", token.Error())
			} else {
				log.Printf("Sensor state published to topic %s.", cfgSnapshot.StateTopic)
			}
		case update := <-cfgManager.Updates():
			if update.QueryIntervalChanged {
				interval = update.NewQueryInterval
				if interval <= 0 {
					interval = 10
					log.Printf("Invalid query_interval; defaulting to %d seconds.", interval)
				}
				ticker.Stop()
				ticker = time.NewTicker(time.Duration(interval) * time.Second)
				log.Printf("Query interval updated to %d seconds.", interval)
			}
			if update.ConfigTopicChanged {
				unsubscribeTopic(mqttClient, update.OldConfigTopic, "temporary config")
				subscribeTopic(mqttClient, update.NewConfigTopic, tempHandler, "temporary config")
			}
			if update.PermanentConfigTopicChanged {
				unsubscribeTopic(mqttClient, update.OldPermanentConfigTopic, "permanent config")
				subscribeTopic(mqttClient, update.NewPermanentConfigTopic, permHandler, "permanent config")
			}
			if update.SerialDeviceChanged || update.BaudRateChanged {
				log.Printf("Serial settings changed; reconnecting to device.")
				_ = gmcDevice.Close()
				cfgSnapshot = cfgManager.Get()
				gmcDevice, info, err = openGMCDevice(cfgSnapshot)
				if err != nil {
					log.Printf("Failed to reopen serial device: %v", err)
					continue
				}
				model = info.model
				fw = info.firmware
				serialStr = info.serial
			}
			if update.RestartRequired {
				log.Printf("Some configuration changes require a restart to take full effect.")
			}
		}
	}
}

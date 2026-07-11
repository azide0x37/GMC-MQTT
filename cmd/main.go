package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/azide0x37/gmc-mqtt/config"
	"github.com/azide0x37/gmc-mqtt/gmc"
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

func collectOnce(device *gmc.GMC, cfg config.Config, info deviceInfo, startTime time.Time) error {
	uptime := int(time.Since(startTime).Seconds())
	cpm, err := device.QueryCPM()
	if err != nil {
		state := errorState(fmt.Sprintf("query cpm: %v", err), info.firmware, info.model, info.serial, uptime)
		_ = writeState(cfg.StatePath, state)
		_ = appendLedger(cfg.LedgerPath, LedgerEvent{Event: "collector_error", Message: state.Error})
		return err
	}
	voltage, err := device.QueryVoltage()
	if err != nil {
		state := errorState(fmt.Sprintf("query voltage: %v", err), info.firmware, info.model, info.serial, uptime)
		_ = writeState(cfg.StatePath, state)
		_ = appendLedger(cfg.LedgerPath, LedgerEvent{Event: "collector_error", Message: state.Error})
		return err
	}

	state := buildState(cpm, voltage, info.firmware, info.model, info.serial, uptime)
	if err := writeState(cfg.StatePath, state); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	return nil
}

func main() {
	showVersion := flag.Bool("version", false, "Print version information and exit")
	runOnce := flag.Bool("once", false, "Poll once, write state, and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(versionString())
		return
	}

	cfg, err := config.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to load collector configuration: %v", err)
	}
	log.Printf("Collector configuration loaded: serial=%s baud=%d interval=%ds state=%s", cfg.SerialDevice, cfg.BaudRate, cfg.QueryInterval, cfg.StatePath)

	startTime := time.Now()
	var device *gmc.GMC
	var info deviceInfo
	defer func() {
		if device != nil {
			_ = device.Close()
		}
	}()

	connect := func() bool {
		if device != nil {
			_ = device.Close()
			device = nil
		}
		opened, openedInfo, err := openGMCDevice(cfg)
		if err != nil {
			message := fmt.Sprintf("open serial device: %v", err)
			log.Printf("Error: %s", message)
			uptime := int(time.Since(startTime).Seconds())
			_ = writeState(cfg.StatePath, errorState(message, info.firmware, info.model, info.serial, uptime))
			_ = appendLedger(cfg.LedgerPath, LedgerEvent{Event: "collector_error", Message: message})
			return false
		}
		device = opened
		info = openedInfo
		_ = appendLedger(cfg.LedgerPath, LedgerEvent{Event: "collector_connected", Message: cfg.SerialDevice})
		return true
	}

	if !connect() {
		if *runOnce {
			log.Fatalf("Failed to connect to GMC device")
		}
	}

	if *runOnce {
		if device == nil {
			log.Fatalf("GMC device is not connected")
		}
		if err := collectOnce(device, cfg, info, startTime); err != nil {
			log.Fatalf("Collector run failed: %v", err)
		}
		return
	}

	ticker := time.NewTicker(time.Duration(cfg.QueryInterval) * time.Second)
	defer ticker.Stop()

	for {
		if device == nil {
			time.Sleep(time.Duration(cfg.QueryInterval) * time.Second)
			connect()
			continue
		}
		if err := collectOnce(device, cfg, info, startTime); err != nil {
			log.Printf("Collector run failed: %v", err)
			_ = device.Close()
			device = nil
			continue
		}
		<-ticker.C
	}
}

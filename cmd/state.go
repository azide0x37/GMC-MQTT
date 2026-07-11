package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"
)

type StateSnapshot struct {
	CPM       uint16  `json:"cpm"`
	Battery   float64 `json:"battery"`
	Version   string  `json:"version"`
	Model     string  `json:"model"`
	Serial    string  `json:"serial"`
	Uptime    int     `json:"uptime"`
	USV       float64 `json:"usv"`
	MR        float64 `json:"mr"`
	Timestamp string  `json:"timestamp"`
	Healthy   bool    `json:"healthy"`
	Error     string  `json:"error,omitempty"`
}

type LedgerEvent struct {
	Timestamp string `json:"timestamp"`
	Event     string `json:"event"`
	Message   string `json:"message,omitempty"`
}

func buildState(cpm uint16, voltage float64, fw, model, serial string, uptime int) StateSnapshot {
	usv := math.Round(float64(cpm)*0.0057*10000) / 10000
	mr := math.Round(float64(cpm)*0.00057*10000) / 10000

	return StateSnapshot{
		CPM:       cpm,
		Battery:   voltage,
		Version:   fw,
		Model:     model,
		Serial:    serial,
		Uptime:    uptime,
		USV:       usv,
		MR:        mr,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Healthy:   true,
	}
}

func errorState(message string, fw, model, serial string, uptime int) StateSnapshot {
	return StateSnapshot{
		Version:   fw,
		Model:     model,
		Serial:    serial,
		Uptime:    uptime,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Healthy:   false,
		Error:     message,
	}
}

func writeState(path string, state StateSnapshot) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".state-*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	encoder := json.NewEncoder(tmp)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(state); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func appendLedger(path string, event LedgerEvent) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(file, string(data)); err != nil {
		return err
	}
	return nil
}

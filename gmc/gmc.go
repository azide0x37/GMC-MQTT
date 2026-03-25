package gmc

import (
	"fmt"
	"io"
	"log"
	"time"

	"go.bug.st/serial"
)

// GMC represents a connection to a GMC‑300 device via serial.
type GMC struct {
	port serial.Port
}

// SensorData holds sensor values from the device.
type SensorData struct {
	CPM     uint16  `json:"cpm"`
	Battery float64 `json:"battery"`
}

// NewGMC opens a connection to the GMC device on the specified serial port with the given baud rate.
func NewGMC(portName string, baud int) (*GMC, error) {
	mode := &serial.Mode{
		BaudRate: baud,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, err
	}
	if err := port.SetReadTimeout(2 * time.Second); err != nil {
		_ = port.Close()
		return nil, fmt.Errorf("failed to set read timeout: %w", err)
	}
	// Give the port a moment to stabilize.
	time.Sleep(100 * time.Millisecond)
	return &GMC{port: port}, nil
}

// sendCommand flushes the input buffer, writes a command, and then reads exactly expected bytes from the response.
func (g *GMC) sendCommand(cmd string, expected int) ([]byte, error) {
	// Flush any stale data (such as heartbeat messages) from the input buffer.
	if flusher, ok := g.port.(interface{ ResetInputBuffer() error }); ok {
		if err := flusher.ResetInputBuffer(); err != nil {
			log.Printf("Warning: failed to flush input buffer: %v", err)
		}
	}
	log.Printf("Sending command: %s", cmd)
	_, err := g.port.Write([]byte(cmd))
	if err != nil {
		return nil, fmt.Errorf("failed to write command %s: %w", cmd, err)
	}
	// For commands that expect no response.
	if expected == 0 {
		return nil, nil
	}
	buf := make([]byte, expected)
	n, err := io.ReadFull(g.port, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read response for %s: %w", cmd, err)
	}
	if n != expected {
		return nil, fmt.Errorf("expected %d bytes, got %d", expected, n)
	}
	log.Printf("Received response for %s: %x", cmd, buf)
	return buf, nil
}

// QueryVersion sends the <GETVER>> command and returns the hardware model and firmware version.
// The response is 14 ASCII characters: first 7 bytes are the model, next 7 are the firmware version.
func (g *GMC) QueryVersion() (string, string, error) {
	data, err := g.sendCommand("<GETVER>>", 14)
	if err != nil {
		return "", "", err
	}
	model := string(data[:7])
	firmware := string(data[7:])
	return model, firmware, nil
}

// QueryCPM sends the <GETCPM>> command and returns the CPM as a 16-bit unsigned integer.
func (g *GMC) QueryCPM() (uint16, error) {
	data, err := g.sendCommand("<GETCPM>>", 2)
	if err != nil {
		return 0, err
	}
	// Interpret the two binary bytes in big-endian order.
	cpm := uint16(data[0])<<8 | uint16(data[1])
	return cpm, nil
}

// QueryVoltage sends the <GETVOLT>> command and returns the battery voltage in volts.
func (g *GMC) QueryVoltage() (float64, error) {
	data, err := g.sendCommand("<GETVOLT>>", 1)
	if err != nil {
		return 0, err
	}
	voltage := float64(data[0]) / 10.0
	return voltage, nil
}

func decodeASCIIOrHex(data []byte) string {
	for _, b := range data {
		if b < 0x20 || b > 0x7E {
			// Return hex, e.g. "00FF1A"
			return fmt.Sprintf("%X", data)
		}
	}
	return string(data)
}

// QuerySerial sends the <GETSERIAL>> command and returns the serial number as a string.
func (g *GMC) QuerySerial() (string, error) {
	data, err := g.sendCommand("<GETSERIAL>>", 7)
	if err != nil {
		return "", err
	}
	return decodeASCIIOrHex(data), nil
}

// Close closes the serial port.
func (g *GMC) Close() error {
	return g.port.Close()
}

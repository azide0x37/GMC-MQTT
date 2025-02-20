package gmc

import (
	"bufio"
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
)

// GMC represents a connection to a GMC-300s device via serial.
type GMC struct {
	port *serial.Port
}

// NewGMC opens a connection to the GMC device on the specified serial port with the given baud rate.
func NewGMC(portName string, baud int) (*GMC, error) {
	c := &serial.Config{Name: portName, Baud: baud, ReadTimeout: 5 * time.Second}
	port, err := serial.OpenPort(c)
	if err != nil {
		return nil, err
	}
	return &GMC{port: port}, nil
}

// Query sends a command to the GMC device and returns its response.
// For this initial build, it sends a simple "READ" command followed by a newline.
func (g *GMC) Query() (string, error) {
	command := "READ\n"
	_, err := g.port.Write([]byte(command))
	if err != nil {
		return "", fmt.Errorf("failed to write to serial port: %w", err)
	}

	// Read the response until a newline is received.
	reader := bufio.NewReader(g.port)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read from serial port: %w", err)
	}
	log.Printf("Raw response from GMC: %s", response)
	return response, nil
}

// Close closes the serial port connection.
func (g *GMC) Close() error {
	return g.port.Close()
}

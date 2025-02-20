package gmc

import (
	"bufio"
	"fmt"
	"log"

	"go.bug.st/serial"
)

// GMC represents a connection to a GMC-300s device via serial.
type GMC struct {
	port serial.Port
}

// NewGMC opens a connection to the GMC device on the specified serial port with the given baud rate.
func NewGMC(portName string, baud int) (*GMC, error) {
	mode := &serial.Mode{
		BaudRate: baud,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, err
	}
	return &GMC{port: port}, nil
}

// Query sends a command to the GMC device and returns its response.
// According to the GMC-300s documentation, use "READALL\n" to retrieve all data.
func (g *GMC) Query() (string, error) {
	// Log that we are sending a query command.
	log.Println("Sending query command to GMC-300s.")
	// Use the full query command from the GMC-300s documentation.
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
	// Log the raw response for debugging.
	log.Printf("Raw response from GMC-300s: %s", response)
	return response, nil
}

// Close closes the serial port connection.
func (g *GMC) Close() error {
	return g.port.Close()
}

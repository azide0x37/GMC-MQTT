package gmc

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"

	"go.bug.st/serial"
)

type fakePort struct {
	readBuf    *bytes.Buffer
	written    bytes.Buffer
	resetInput bool
}

func newFakePort(read []byte) *fakePort {
	return &fakePort{readBuf: bytes.NewBuffer(read)}
}

func (f *fakePort) SetMode(mode *serial.Mode) error { return nil }

func (f *fakePort) Read(p []byte) (int, error) {
	if f.readBuf.Len() == 0 {
		return 0, io.EOF
	}
	return f.readBuf.Read(p)
}

func (f *fakePort) Write(p []byte) (int, error) { return f.written.Write(p) }

func (f *fakePort) Drain() error { return nil }

func (f *fakePort) ResetInputBuffer() error {
	f.resetInput = true
	return nil
}

func (f *fakePort) ResetOutputBuffer() error { return nil }

func (f *fakePort) SetDTR(dtr bool) error { return nil }

func (f *fakePort) SetRTS(rts bool) error { return nil }

func (f *fakePort) GetModemStatusBits() (*serial.ModemStatusBits, error) {
	return &serial.ModemStatusBits{}, nil
}

func (f *fakePort) SetReadTimeout(t time.Duration) error { return nil }

func (f *fakePort) Close() error { return nil }

func (f *fakePort) Break(time.Duration) error { return nil }

func TestQueryCPM(t *testing.T) {
	port := newFakePort([]byte{0x01, 0x02})
	device := &GMC{port: port}

	cpm, err := device.QueryCPM()
	if err != nil {
		t.Fatalf("QueryCPM failed: %v", err)
	}
	if cpm != 0x0102 {
		t.Fatalf("expected 0x0102, got %x", cpm)
	}
	if port.written.String() != "<GETCPM>>" {
		t.Fatalf("unexpected command: %q", port.written.String())
	}
	if !port.resetInput {
		t.Fatalf("expected ResetInputBuffer to be called")
	}
}

func TestQueryVoltage(t *testing.T) {
	port := newFakePort([]byte{37})
	device := &GMC{port: port}

	voltage, err := device.QueryVoltage()
	if err != nil {
		t.Fatalf("QueryVoltage failed: %v", err)
	}
	if voltage != 3.7 {
		t.Fatalf("expected 3.7, got %v", voltage)
	}
}

func TestQueryVersion(t *testing.T) {
	port := newFakePort([]byte("MODEL01FWVER01"))
	device := &GMC{port: port}

	model, fw, err := device.QueryVersion()
	if err != nil {
		t.Fatalf("QueryVersion failed: %v", err)
	}
	if model != "MODEL01" {
		t.Fatalf("expected model MODEL01, got %s", model)
	}
	if fw != "FWVER01" {
		t.Fatalf("expected firmware FWVER01, got %s", fw)
	}
}

func TestQuerySerialASCII(t *testing.T) {
	port := newFakePort([]byte("ABC1234"))
	device := &GMC{port: port}

	serialStr, err := device.QuerySerial()
	if err != nil {
		t.Fatalf("QuerySerial failed: %v", err)
	}
	if serialStr != "ABC1234" {
		t.Fatalf("expected ABC1234, got %s", serialStr)
	}
}

func TestQuerySerialHexFallback(t *testing.T) {
	port := newFakePort([]byte{0x00, 0xFF, 0x10, 0x20, 0x30, 0x40, 0x50})
	device := &GMC{port: port}

	serialStr, err := device.QuerySerial()
	if err != nil {
		t.Fatalf("QuerySerial failed: %v", err)
	}
	if serialStr != "00FF1020304050" {
		t.Fatalf("expected hex fallback, got %s", serialStr)
	}
}

func TestSendCommandReadError(t *testing.T) {
	port := newFakePort(nil)
	device := &GMC{port: port}

	_, err := device.sendCommand("<GETCPM>>", 2)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}

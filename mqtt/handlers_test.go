package mqtt

import (
	"path/filepath"
	"testing"

	"github.com/azide0x37/gmc-mqtt/config"
)

type fakeMessage struct {
	topic   string
	payload []byte
}

func (m fakeMessage) Duplicate() bool { return false }
func (m fakeMessage) Qos() byte       { return 0 }
func (m fakeMessage) Retained() bool  { return false }
func (m fakeMessage) Topic() string   { return m.topic }
func (m fakeMessage) MessageID() uint16 {
	return 0
}
func (m fakeMessage) Payload() []byte { return m.payload }
func (m fakeMessage) Ack()            {}

func TestConfigMessageHandlerUpdatesConfig(t *testing.T) {
	cfg := &config.Config{QueryInterval: 10, StateTopic: "gmc/state", PublishTopic: "gmc/state"}
	mgr := config.NewManager(cfg, filepath.Join(t.TempDir(), "config.toml"))
	handler := ConfigMessageHandler(mgr)

	handler(nil, fakeMessage{topic: "gmc/config/temp", payload: []byte(`{"query_interval":15}`)})

	updated := mgr.Get()
	if updated.QueryInterval != 15 {
		t.Fatalf("Expected QueryInterval=15, got %d", updated.QueryInterval)
	}
}

func TestConfigMessageHandlerRejectsInvalidPayload(t *testing.T) {
	cfg := &config.Config{QueryInterval: 10, StateTopic: "gmc/state", PublishTopic: "gmc/state"}
	mgr := config.NewManager(cfg, filepath.Join(t.TempDir(), "config.toml"))
	handler := ConfigMessageHandler(mgr)

	handler(nil, fakeMessage{topic: "gmc/config/temp", payload: []byte(`{"query_interval":0}`)})

	updated := mgr.Get()
	if updated.QueryInterval != 10 {
		t.Fatalf("Expected QueryInterval to remain 10, got %d", updated.QueryInterval)
	}
}

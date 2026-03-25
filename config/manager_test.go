package config

import (
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestApplyJSONUpdateTemporary(t *testing.T) {
	cfg := &Config{
		QueryInterval: 10,
		StateTopic:    "gmc/state",
		PublishTopic:  "gmc/state",
	}
	mgr := NewManager(cfg, filepath.Join(t.TempDir(), "config.toml"))

	summary, err := mgr.ApplyJSONUpdate([]byte(`{"query_interval":5,"state_topic":"gmc/new"}`), false)
	if err != nil {
		t.Fatalf("ApplyJSONUpdate failed: %v", err)
	}
	if !summary.QueryIntervalChanged {
		t.Fatalf("Expected query interval change")
	}
	if !summary.StateTopicChanged {
		t.Fatalf("Expected state topic change")
	}

	updated := mgr.Get()
	if updated.QueryInterval != 5 {
		t.Fatalf("Expected QueryInterval=5, got %d", updated.QueryInterval)
	}
	if updated.StateTopic != "gmc/new" {
		t.Fatalf("Expected StateTopic=gmc/new, got %s", updated.StateTopic)
	}
	if updated.PublishTopic != "gmc/new" {
		t.Fatalf("Expected PublishTopic to mirror StateTopic, got %s", updated.PublishTopic)
	}
}

func TestApplyJSONUpdatePermanentWritesFile(t *testing.T) {
	cfg := &Config{
		QueryInterval: 10,
		StateTopic:    "gmc/state",
		PublishTopic:  "gmc/state",
	}
	path := filepath.Join(t.TempDir(), "config.toml")
	mgr := NewManager(cfg, path)

	_, err := mgr.ApplyJSONUpdate([]byte(`{"state_topic":"gmc/perm"}`), true)
	if err != nil {
		t.Fatalf("ApplyJSONUpdate failed: %v", err)
	}

	var loaded Config
	if _, err := toml.DecodeFile(path, &loaded); err != nil {
		t.Fatalf("Failed to decode saved config: %v", err)
	}
	normalizeConfig(&loaded)
	if loaded.StateTopic != "gmc/perm" {
		t.Fatalf("Expected persisted StateTopic=gmc/perm, got %s", loaded.StateTopic)
	}
	if loaded.PublishTopic != "gmc/perm" {
		t.Fatalf("Expected persisted PublishTopic=gmc/perm, got %s", loaded.PublishTopic)
	}
}

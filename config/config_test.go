package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestCurrentReturnsIndependentSnapshot(t *testing.T) {
	publish(Config{
		App:   AppConf{Mode: "development"},
		Slack: SlackConf{Channels: map[string]string{"critical": "C1"}},
		Discord: DiscordConf{
			Channels:     map[string]string{"critical": "D1"},
			MentionRoles: []string{"role-1"},
		},
	}, false)

	first := Current()
	first.Slack.Channels["critical"] = "modified"
	first.Discord.Channels["critical"] = "modified"
	first.Discord.MentionRoles[0] = "modified"

	current := Current()
	if current.Slack.Channels["critical"] != "C1" {
		t.Fatalf("Slack channel = %q, want %q", current.Slack.Channels["critical"], "C1")
	}
	if current.Discord.Channels["critical"] != "D1" {
		t.Fatalf("Discord channel = %q, want %q", current.Discord.Channels["critical"], "D1")
	}
	if current.Discord.MentionRoles[0] != "role-1" {
		t.Fatalf("Discord mention role = %q, want %q", current.Discord.MentionRoles[0], "role-1")
	}
}

func TestLoadConfigFromFileKeepsCurrentSnapshotWhenParsingFails(t *testing.T) {
	publish(Config{App: AppConf{Version: "stable"}}, false)

	configPath := filepath.Join(t.TempDir(), "invalid.yaml")
	if err := os.WriteFile(configPath, []byte("app: ["), 0o600); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}

	if err := LoadConfigFromFile(configPath); err == nil {
		t.Fatal("LoadConfigFromFile() error = nil, want parse error")
	}
	if got := Current().App.Version; got != "stable" {
		t.Fatalf("current version = %q, want %q", got, "stable")
	}
}

func TestCurrentIsSafeDuringConcurrentPublish(t *testing.T) {
	const iterations = 1000
	var writers sync.WaitGroup
	writers.Add(1)
	go func() {
		defer writers.Done()
		for i := 0; i < iterations; i++ {
			publish(Config{App: AppConf{Version: fmt.Sprintf("v%d", i)}}, false)
		}
	}()

	var readers sync.WaitGroup
	for i := 0; i < 4; i++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for j := 0; j < iterations; j++ {
				_ = Current().App.Version
			}
		}()
	}

	writers.Wait()
	readers.Wait()
}

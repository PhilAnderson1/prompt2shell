package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	config, err := parseConfig(strings.NewReader("# comment\nendpoint=https://example.com/v1/chat/completions\nmodel=test-model\napi_token=secret\n"))
	if err != nil {
		t.Fatal(err)
	}
	if config.Endpoint != "https://example.com/v1/chat/completions" || config.Model != "test-model" || config.APIToken != "secret" {
		t.Fatalf("unexpected config: %#v", config)
	}
}

func TestParseConfigRejectsInvalidInput(t *testing.T) {
	tests := []string{
		"model=test\n",
		"endpoint=ftp://example.com\nmodel=test\n",
		"endpoint=https://example.com\nmodel=test\nunknown=value\n",
		"endpoint=https://one.example\nendpoint=https://two.example\nmodel=test\n",
	}
	for _, input := range tests {
		if _, err := parseConfig(strings.NewReader(input)); err == nil {
			t.Errorf("expected error for %q", input)
		}
	}
}

func TestLoadConfigUsesFirstExistingFile(t *testing.T) {
	directory := t.TempDir()
	userConfig := filepath.Join(directory, "user.conf")
	systemConfig := filepath.Join(directory, "system.conf")
	if err := os.WriteFile(userConfig, []byte("endpoint=https://user.example/v1/chat/completions\nmodel=user-model\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(systemConfig, []byte("endpoint=https://system.example/v1/chat/completions\nmodel=system-model\n"), 0600); err != nil {
		t.Fatal(err)
	}
	config, err := loadConfigFrom([]string{userConfig, systemConfig})
	if err != nil {
		t.Fatal(err)
	}
	if config.Model != "user-model" {
		t.Fatalf("got model %q, want user-model", config.Model)
	}
}

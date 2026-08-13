package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	config, err := parseConfig(strings.NewReader("# comment\nendpoint=https://example.com/v1/chat/completions\nmodel=test-model\napi_token=secret\napi_type=openrouter\nreasoning_effort=low\n"))
	if err != nil {
		t.Fatal(err)
	}
	if config.Endpoint != "https://example.com/v1/chat/completions" || config.Model != "test-model" || config.APIToken != "secret" || config.APIType != "openrouter" || config.ReasoningEffort != "low" {
		t.Fatalf("unexpected config: %#v", config)
	}
}

func TestParseConfigRejectsInvalidInput(t *testing.T) {
	tests := []string{
		"model=test\n",
		"endpoint=ftp://example.com\nmodel=test\n",
		"endpoint=https://example.com\nmodel=test\nunknown=value\n",
		"endpoint=https://one.example\nendpoint=https://two.example\nmodel=test\n",
		"endpoint=https://example.com\nmodel=test\napi_type=unknown\n",
		"endpoint=https://example.com\nmodel=test\nreasoning_effort=extreme\n",
	}
	for _, input := range tests {
		if _, err := parseConfig(strings.NewReader(input)); err == nil {
			t.Errorf("expected error for %q", input)
		}
	}
}

func TestParseConfigDefaultsAPITypeToGeneric(t *testing.T) {
	config, err := parseConfig(strings.NewReader("endpoint=https://example.com\nmodel=test\n"))
	if err != nil {
		t.Fatal(err)
	}
	if config.APIType != "generic" {
		t.Fatalf("got API type %q, want generic", config.APIType)
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

func TestConfigPathBesideExecutable(t *testing.T) {
	want := filepath.Join("somewhere", "bin", "prompt2shell.conf")
	got := configPathBesideExecutable(filepath.Join("somewhere", "bin", "p2s"))
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

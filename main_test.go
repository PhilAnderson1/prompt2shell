package main

import (
	"strings"
	"testing"
)

func TestDecodeCommand(t *testing.T) {
	got, err := decodeCommand(`{"command":"Get-ChildItem"}`)
	if err != nil {
		t.Fatal(err)
	}
	if got.Command != "Get-ChildItem" {
		t.Fatalf("got %q, want Get-ChildItem", got.Command)
	}
}

func TestDecodeCommandRejectsMalformedShapes(t *testing.T) {
	invalid := []string{
		`"Get-Date"`,
		`"command": "Get-ChildItem"`,
		"```json\n{\"command\":\"Get-ChildItem\"}\n```",
		`<think>  </think>  "command": "Get-Date"`,
		`{"command":"Get-Date","explanation":"extra"}`,
		`{"command":"Get-Date"} trailing text`,
	}
	for _, content := range invalid {
		if _, err := decodeCommand(content); err == nil {
			t.Errorf("expected %q to be rejected", content)
		}
	}
}

func TestBuildChatRequestUsesProviderControl(t *testing.T) {
	tests := []struct {
		apiType       string
		wantReasoning bool
		wantTemplate  bool
	}{
		{"generic", false, false},
		{"openrouter", true, false},
		{"llamacpp", false, true},
	}
	for _, test := range tests {
		payload := buildChatRequest(Config{Model: "test", APIType: test.apiType}, "hello")
		if (payload.Reasoning != nil) != test.wantReasoning {
			t.Errorf("%s reasoning present = %v, want %v", test.apiType, payload.Reasoning != nil, test.wantReasoning)
		}
		if (payload.ChatTemplateKwargs != nil) != test.wantTemplate {
			t.Errorf("%s template kwargs present = %v, want %v", test.apiType, payload.ChatTemplateKwargs != nil, test.wantTemplate)
		}
	}
}

func TestConfirm(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"\n", true},
		{"yes\n", false},
		{" \n", false},
		{"", false},
		{"\x1b", false},
	}
	for _, test := range tests {
		got, err := confirm(strings.NewReader(test.input))
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("confirm(%q) = %v, want %v", test.input, got, test.want)
		}
	}
}

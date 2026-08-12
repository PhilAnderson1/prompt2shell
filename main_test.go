package main

import (
	"strings"
	"testing"
)

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

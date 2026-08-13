package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateOpenAICommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("unexpected authorization header: %q", request.Header.Get("Authorization"))
		}
		var body map[string]interface{}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["input"] != "list files" || body["instructions"] != platformSystemPrompt {
			t.Errorf("unexpected input or instructions: %#v", body)
		}
		reasoning, ok := body["reasoning"].(map[string]interface{})
		if !ok || reasoning["effort"] != "none" {
			t.Errorf("unexpected reasoning: %#v", body["reasoning"])
		}
		text, ok := body["text"].(map[string]interface{})
		if !ok {
			t.Fatalf("missing text configuration: %#v", body["text"])
		}
		format, ok := text["format"].(map[string]interface{})
		if !ok || format["type"] != "json_schema" || format["strict"] != true {
			t.Errorf("unexpected structured output format: %#v", text["format"])
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"status":"completed","output":[{"type":"message","content":[{"type":"output_text","text":"{\"command\":\"Get-ChildItem\"}"}]}]}`))
	}))
	defer server.Close()

	command, err := generateCommand(Config{Endpoint: server.URL, Model: "test", APIToken: "secret", APIType: "openai", ReasoningEffort: "none"}, "list files")
	if err != nil {
		t.Fatal(err)
	}
	if command != "Get-ChildItem" {
		t.Fatalf("got %q, want Get-ChildItem", command)
	}
}

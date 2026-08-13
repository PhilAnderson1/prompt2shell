package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type openAIRequest struct {
	Model           string                 `json:"model"`
	Instructions    string                 `json:"instructions"`
	Input           string                 `json:"input"`
	MaxOutputTokens int                    `json:"max_output_tokens"`
	Reasoning       map[string]string      `json:"reasoning,omitempty"`
	Text            map[string]interface{} `json:"text"`
}

type openAIResponse struct {
	Status string `json:"status"`
	Output []struct {
		Type    string `json:"type"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func commandSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type": "string", "minLength": 1, "pattern": `^[^\r\n\x00]+$`,
			},
		},
		"required": []string{"command"}, "additionalProperties": false,
	}
}

func generateOpenAICommand(config Config, instruction string) (string, error) {
	var reasoning map[string]string
	if config.ReasoningEffort != "" {
		reasoning = map[string]string{"effort": config.ReasoningEffort}
	}
	payload := openAIRequest{
		Model: config.Model, Instructions: platformSystemPrompt, Input: instruction,
		MaxOutputTokens: 300,
		Reasoning:       reasoning,
		Text: map[string]interface{}{
			"format": map[string]interface{}{
				"type": "json_schema", "name": "shell_command", "strict": true,
				"schema": commandSchema(),
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequest(http.MethodPost, config.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	if config.APIToken != "" {
		request.Header.Set("Authorization", "Bearer "+config.APIToken)
	}
	response, err := (&http.Client{Timeout: 5 * time.Minute}).Do(request)
	if err != nil {
		return "", fmt.Errorf("AI request failed: %w", err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read AI response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("AI endpoint returned %s: %s", response.Status, strings.TrimSpace(string(responseBody)))
	}
	var result openAIResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", fmt.Errorf("decode AI response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("OpenAI response failed: %s", result.Error.Message)
	}
	for _, output := range result.Output {
		if output.Type != "message" {
			continue
		}
		for _, content := range output.Content {
			if content.Type != "output_text" {
				continue
			}
			generated, err := decodeCommand(content.Text)
			if err != nil {
				return "", err
			}
			command := strings.TrimSpace(generated.Command)
			if strings.ContainsAny(command, "\r\n\x00") {
				return "", errors.New("AI returned a command containing invalid control characters")
			}
			return command, nil
		}
	}
	return "", fmt.Errorf("OpenAI returned no output text (status=%s)", result.Status)
}

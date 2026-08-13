package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type chatRequest struct {
	Model              string            `json:"model"`
	Messages           []message         `json:"messages"`
	Temperature        float64           `json:"temperature"`
	MaxTokens          int               `json:"max_tokens"`
	ResponseFormat     map[string]any    `json:"response_format"`
	ChatTemplateKwargs map[string]bool   `json:"chat_template_kwargs,omitempty"`
	Reasoning          map[string]string `json:"reasoning,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
}

type commandResponse struct {
	Command string `json:"command"`
}

func buildChatRequest(config Config, instruction string) chatRequest {
	payload := chatRequest{
		Model: config.Model,
		Messages: []message{
			{Role: "system", Content: platformSystemPrompt},
			{Role: "user", Content: instruction},
		},
		Temperature: 0.1,
		MaxTokens:   300,
		ResponseFormat: map[string]any{
			"type": "json_object",
			"schema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":      "string",
						"minLength": 1,
						"pattern":   `^[^\r\n\x00]+$`,
					},
				},
				"required":             []string{"command"},
				"additionalProperties": false,
			},
		},
	}
	switch config.APIType {
	case "openrouter":
		payload.Reasoning = map[string]string{"effort": "none"}
	case "llamacpp":
		payload.ChatTemplateKwargs = map[string]bool{"enable_thinking": false}
	}
	return payload
}

func generateCommand(config Config, instruction string) (string, error) {
	if config.APIType == "openai" {
		return generateOpenAICommand(config, instruction)
	}
	payload := buildChatRequest(config, instruction)
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	request, err := http.NewRequest(http.MethodPost, config.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	if config.APIToken != "" {
		request.Header.Set("Authorization", "Bearer "+config.APIToken)
	}

	response, err := client.Do(request)
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

	var completion chatResponse
	if err := json.Unmarshal(responseBody, &completion); err != nil {
		return "", fmt.Errorf("decode AI response: %w", err)
	}
	if len(completion.Choices) == 0 {
		return "", errors.New("AI endpoint returned no choices")
	}

	generated, err := decodeCommand(completion.Choices[0].Message.Content)
	if err != nil {
		return "", err
	}
	command := strings.TrimSpace(generated.Command)
	if command == "" {
		return "", errors.New("AI returned an empty command")
	}
	if strings.ContainsAny(command, "\r\n\x00") {
		return "", errors.New("AI returned a command containing invalid control characters")
	}
	return command, nil
}

func decodeCommand(content string) (commandResponse, error) {
	var generated commandResponse
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&generated); err != nil {
		return commandResponse{}, fmt.Errorf("decode generated command: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("multiple JSON values")
		}
		return commandResponse{}, fmt.Errorf("decode generated command: %w", err)
	}
	if strings.TrimSpace(generated.Command) == "" {
		return commandResponse{}, errors.New("AI returned an empty command")
	}
	return generated, nil
}

func confirm(reader io.Reader) (bool, error) {
	return readConfirmation(reader)
}

func readByteConfirmation(reader io.Reader) (bool, error) {
	var key [1]byte
	_, err := io.ReadFull(reader, key[:])
	if err != nil {
		if errors.Is(err, io.EOF) {
			return false, nil
		}
		return false, err
	}
	return key[0] == (byte)(10) || key[0] == (byte)(13), nil
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: %s <describe the command you want>", executableName)
	}

	config, err := loadConfig()
	if err != nil {
		return err
	}
	command, err := generateCommand(config, strings.Join(args, " "))
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, command)
	fmt.Fprint(stderr, "Press Enter to run, or Esc to abort: ")

	restore, err := enableSingleKeyInput(stdin)
	if err != nil {
		return fmt.Errorf("configure terminal: %w", err)
	}
	ok, err := confirm(stdin)
	restore()
	fmt.Fprintln(stderr)
	if err != nil {
		return fmt.Errorf("read confirmation: %w", err)
	}
	if !ok {
		return errors.New("aborted")
	}

	return executeCommand(command, stdin, stdout, stderr)
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			os.Exit(exitError.ExitCode())
		}
		fmt.Fprintln(os.Stderr, executableName+":", err)
		os.Exit(1)
	}
}

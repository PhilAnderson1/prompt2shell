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
	"syscall"
	"time"
	"unsafe"
)

const (
	endpoint = "https://yourendpoint.host:8081/v1/chat/completions"
	model    = "Qwen3.6-35B-A3B"
	apiToken = "your-api-token-goes-here"
)

const systemPrompt = `Convert the user's natural-language request into one Linux shell command.
Return JSON with exactly one string field named "command".
The command must work from the user's current directory unless they ask otherwise.
Prefer standard Linux utilities, preserve paths and constraints precisely, and do not add sudo unless explicitly requested.
Return only the command itself in the field: no Markdown, explanation, prompt prefix, or newline.
If the request is ambiguous, choose the safest non-destructive interpretation.`

type chatRequest struct {
	Model              string            `json:"model"`
	Messages           []message         `json:"messages"`
	Temperature        float64           `json:"temperature"`
	MaxTokens          int               `json:"max_tokens"`
	ResponseFormat     map[string]string `json:"response_format"`
	ChatTemplateKwargs map[string]bool   `json:"chat_template_kwargs"`
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

func generateCommand(instruction string) (string, error) {
	payload := chatRequest{
		Model: model,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: instruction},
		},
		Temperature:        0.1,
		MaxTokens:          300,
		ResponseFormat:     map[string]string{"type": "json_object"},
		ChatTemplateKwargs: map[string]bool{"enable_thinking": false},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+apiToken)

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

	var generated commandResponse
	if err := json.Unmarshal([]byte(completion.Choices[0].Message.Content), &generated); err != nil {
		return "", fmt.Errorf("decode generated command: %w", err)
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

func enableSingleKeyInput(reader io.Reader) (func(), error) {
	file, ok := reader.(*os.File)
	if !ok {
		return func() {}, nil
	}
	fd := file.Fd()
	var original syscall.Termios
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(&original)), 0, 0, 0)
	if errno != 0 {
		if errno == syscall.ENOTTY {
			return func() {}, nil
		}
		return nil, errno
	}
	modified := original
	modified.Lflag &^= syscall.ICANON | syscall.ECHO
	modified.Cc[syscall.VMIN] = 1
	modified.Cc[syscall.VTIME] = 0
	_, _, errno = syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TCSETS, uintptr(unsafe.Pointer(&modified)), 0, 0, 0)
	if errno != 0 {
		return nil, errno
	}
	return func() {
		syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TCSETS, uintptr(unsafe.Pointer(&original)), 0, 0, 0)
	}, nil
}

func confirm(reader io.Reader) (bool, error) {
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
		return errors.New("usage: s <describe the command you want>")
	}

	command, err := generateCommand(strings.Join(args, " "))
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

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	child := exec.Command(shell, "-c", command)
	child.Stdin = stdin
	child.Stdout = stdout
	child.Stderr = stderr
	return child.Run()
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "s:", err)
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(1)
	}
}

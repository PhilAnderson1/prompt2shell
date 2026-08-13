package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Endpoint        string
	Model           string
	APIToken        string
	APIType         string
	ReasoningEffort string
}

func loadConfig() (Config, error) {
	paths, err := configPaths()
	if err != nil {
		return Config{}, err
	}
	return loadConfigFrom(paths)
}

func executableConfigPath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("find executable path: %w", err)
	}
	return configPathBesideExecutable(executable), nil
}

func configPathBesideExecutable(executable string) string {
	return filepath.Join(filepath.Dir(executable), "prompt2shell.conf")
}

func loadConfigFrom(paths []string) (Config, error) {
	for _, path := range paths {
		file, err := os.Open(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return Config{}, fmt.Errorf("open configuration %s: %w", path, err)
		}
		config, parseErr := parseConfig(file)
		closeErr := file.Close()
		if parseErr != nil {
			return Config{}, fmt.Errorf("parse configuration %s: %w", path, parseErr)
		}
		if closeErr != nil {
			return Config{}, fmt.Errorf("close configuration %s: %w", path, closeErr)
		}
		return config, nil
	}
	return Config{}, fmt.Errorf("configuration not found; checked %s", strings.Join(paths, ", "))
}

func parseConfig(reader io.Reader) (Config, error) {
	config := Config{APIType: "generic"}
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(reader)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !found || key == "" {
			return Config{}, fmt.Errorf("line %d: expected key=value", lineNumber)
		}
		if seen[key] {
			return Config{}, fmt.Errorf("line %d: duplicate key %q", lineNumber, key)
		}
		seen[key] = true
		switch key {
		case "endpoint":
			config.Endpoint = value
		case "model":
			config.Model = value
		case "api_token":
			config.APIToken = value
		case "api_type":
			config.APIType = value
		case "reasoning_effort":
			config.ReasoningEffort = value
		default:
			return Config{}, fmt.Errorf("line %d: unknown key %q", lineNumber, key)
		}
	}
	if err := scanner.Err(); err != nil {
		return Config{}, err
	}
	if config.Endpoint == "" || config.Model == "" {
		return Config{}, errors.New("endpoint and model are required")
	}
	validAPITypes := map[string]bool{
		"generic": true, "openrouter": true, "llamacpp": true, "openai": true,
	}
	if !validAPITypes[config.APIType] {
		return Config{}, errors.New("api_type must be one of: generic, openrouter, llamacpp, openai")
	}
	validReasoningEfforts := map[string]bool{
		"": true, "none": true, "minimal": true, "low": true,
		"medium": true, "high": true, "xhigh": true, "max": true,
	}
	if !validReasoningEfforts[config.ReasoningEffort] {
		return Config{}, errors.New("reasoning_effort must be one of: none, minimal, low, medium, high, xhigh, max")
	}
	parsedURL, err := url.Parse(config.Endpoint)
	if err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "https" && parsedURL.Scheme != "http") {
		return Config{}, errors.New("endpoint must be a valid HTTP or HTTPS URL")
	}
	return config, nil
}

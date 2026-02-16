package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".claude-print-config.json"

// Config represents the claude-print configuration settings.
type Config struct {
	ClaudePath       string `json:"claudePath"`
	DefaultVerbosity string `json:"defaultVerbosity"`
	ColorEnabled     bool   `json:"colorEnabled"`
	EmojiEnabled     bool   `json:"emojiEnabled"`
}

// DefaultConfig returns a Config with sensible default values.
func DefaultConfig() Config {
	return Config{
		ClaudePath:       "",
		DefaultVerbosity: "normal",
		ColorEnabled:     true,
		EmojiEnabled:     true,
	}
}

// getConfigPath returns the full path to the config file in the user's home directory.
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, configFileName), nil
}

// LoadConfig reads the config from ~/.claude-print-config.json.
// If the file doesn't exist, it returns a default config.
// If the file exists but contains invalid JSON, it returns an error.
func LoadConfig() (Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}
		return DefaultConfig(), fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	return cfg, nil
}

// SaveConfig writes the config to ~/.claude-print-config.json.
func SaveConfig(cfg Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidatePath checks if the given path points to a valid executable file.
// It returns an error if the path doesn't exist or is a directory.
func ValidatePath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("Claude CLI not found at %s. Please update ~/%s or delete it to auto-detect", path, configFileName)
		}
		return fmt.Errorf("failed to check path %s: %w", path, err)
	}

	if info.IsDir() {
		return fmt.Errorf("Claude CLI not found at %s. Please update ~/%s or delete it to auto-detect", path, configFileName)
	}

	return nil
}

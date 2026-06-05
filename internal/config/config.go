package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getConfigPath() (string, error) {
	HomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("there was an error getting the home directory: %w", err)
	}
	return filepath.Join(HomeDir, ".gatorconfig.json"), nil
}

func Read() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		fmt.Printf("There was an error getting the config path: %v\n", err)
		return nil, err
	}
	file, err := os.Open(configPath)
	if err != nil {
		fmt.Printf("There was an error opening the config file: %v\n", err)
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		fmt.Printf("There was an error decoding the config file: %v\n", err)
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) SetUser(username string) error {
	c.CurrentUserName = username
	if err := c.Write(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

func (c *Config) Write() error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("there was an error getting the config path: %w", err)
	}
	file, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("there was an error opening the config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("there was an error encoding the config file: %w", err)
	}
	return nil
}

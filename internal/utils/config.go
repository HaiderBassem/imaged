package utils

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigManager handles configuration loading and saving
type ConfigManager struct {
	configPath string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
	}
}

// LoadConfig loads configuration from YAML file
func (cm *ConfigManager) LoadConfig(cfg interface{}) error {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return err
	}

	return nil
}

// SaveConfig saves configuration to YAML file
func (cm *ConfigManager) SaveConfig(cfg interface{}) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(cm.configPath, data, 0644)
}

// ConfigExists checks if configuration file exists
func (cm *ConfigManager) ConfigExists() bool {
	_, err := os.Stat(cm.configPath)
	return !os.IsNotExist(err)
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "imaged.yaml"
	}
	return filepath.Join(homeDir, ".config", "imaged", "config.yaml")
}

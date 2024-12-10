package control

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	SyncInterval int `mapstructure:"sync_interval"`
	Account      struct {
		Name           string `mapstructure:"name"`
		PrivateKeyPath string `mapstructure:"private_key_path"`
	} `mapstructure:"account"`
	Domains []string `mapstructure:"domains"`
}

func LoadConfig(config *Config) error {
	viper.SetConfigType("yaml")

	// Check if CONFIG_PATH environment variable is set
	configPath := os.Getenv("CONFIG_PATH")
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// Fallback to local directory and default config.yaml
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
	}

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal the config into the provided struct
	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Check for PRIVATE_KEY_PATH environment variable
	if keyPathEnv := os.Getenv("PRIVATE_KEY_PATH"); keyPathEnv != "" {
		config.Account.PrivateKeyPath = keyPathEnv
	} else if config.Account.PrivateKeyPath == "" {
		// Fallback to the same directory as the config file
		configDir := viper.ConfigFileUsed()
		if configDir != "" {
			config.Account.PrivateKeyPath = fmt.Sprintf("%s/%s", configDir, "private.key")
		}
	}

	return nil
}

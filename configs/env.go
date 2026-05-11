package configs

import (
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
)

var Envs EnvsConfig

// Config holds all application configuration loaded from environment variables.
type EnvsConfig struct {
	// Database
	DatabaseHost     string `mapstructure:"DB_HOST"`
	DatabasePort     string `mapstructure:"DB_PORT"`
	DatabaseUser     string `mapstructure:"DB_USER"`
	DatabasePassword string `mapstructure:"DB_PASSWORD"`
	DatabaseName     string `mapstructure:"DB_NAME"`
	DatabaseSSLMode  string `mapstructure:"DB_SSLMODE"`

	// LLM / AI Provider
	LLMApiKey      string  `mapstructure:"LLM_API_KEY"`
	LLMModel       string  `mapstructure:"LLM_MODEL"`
	LLMBaseURL     string  `mapstructure:"LLM_BASE_URL"`
	LLMMaxTokens   int     `mapstructure:"LLM_MAX_TOKENS"`
	LLMTemperature float64 `mapstructure:"LLM_TEMPERATURE"`

	// Server
	ServerPort string `mapstructure:"SERVER_PORT"`
}

// LoadConfig reads configuration from a .env file and environment variables,
// then unmarshals everything into a Config struct.
//
// It searches for a file named ".env" starting from the current directory
// and walking up. Environment variables set at the OS level take precedence
// over values in the .env file.
func LoadConfig() error {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Warn(".env file not found, falling back to OS environment variables")
		} else {
			return fmt.Errorf("failed to read .env config: %w", err)
		}
	}

	if err := viper.Unmarshal(&Envs); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply defaults for fields that were not set.
	if Envs.DatabasePort == "" {
		Envs.DatabasePort = "5432"
	}
	if Envs.DatabaseSSLMode == "" {
		Envs.DatabaseSSLMode = "disable"
	}
	if Envs.ServerPort == "" {
		Envs.ServerPort = "3000"
	}

	return nil
}

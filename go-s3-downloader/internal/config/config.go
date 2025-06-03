package config

import (
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	Environment string `json:"environment"`
	Region      string `json:"region"`
	Bucket      string `json:"bucket"`
	Prefix      string `json:"prefix"`
	Profile     string `json:"profile"` // AWS profile name to use
}

// NewConfig creates a new configuration with the specified environment
// If profile is empty, it will default to the value of env
func NewConfig(env string, profile ...string) *Config {
	// Default region
	region := "eu-west-1" // Default region

	// Set profile to environment value by default
	profileValue := env

	// If profile was explicitly passed, use that instead
	if len(profile) > 0 && profile[0] != "" {
		profileValue = profile[0]
	}

	// Different defaults based on environment
	if env == "prod" {
		region = "us-east-1" // Use different region for prod
	}

	return &Config{
		Environment: env,
		Region:      region,
		Profile:     profileValue,
	}
}

func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Validate() error {
	// Environment can be any non-empty string now
	if c.Environment == "" {
		return errors.New("environment cannot be empty")
	}
	return nil
}

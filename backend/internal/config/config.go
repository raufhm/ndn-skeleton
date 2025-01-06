package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Environment string         `yaml:"environment"`
	Server      ServerConfig   `yaml:"server"`
	Database    DatabaseConfig `yaml:"database"`
	JWT         JWTConfig      `yaml:"jwt"`
	NewRelic    NewRelicConfig `yaml:"newrelic"`
	Logger      LoggerConfig   `yaml:"logger"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            string `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	SSLMode         string `yaml:"sslmode"`
	MaxOpenConns    int    `yaml:"maxOpenConns"`
	MaxIdleConns    int    `yaml:"maxIdleConns"`
	ConnMaxLifetime int    `yaml:"connMaxLifetime"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
}

type NewRelicConfig struct {
	AppName                  string `yaml:"app_name"`
	LicenseKey               string `yaml:"license_key"`
	Enabled                  bool   `yaml:"enabled"`
	DistributedTracerEnabled bool   `yaml:"distributed_tracer_enabled"`
}

type LoggerConfig struct {
	Level    string `yaml:"level"`
	Encoding string `yaml:"encoding"`
}

func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

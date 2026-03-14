package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Store StoreConfig `yaml:"store"`
}

type StoreConfig struct {
	Type  string      `yaml:"type"`
	MySQL MySQLConfig `yaml:"mysql"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func (c *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
	)
}

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg.applyEnvOverrides()

	return &cfg, nil
}

func (c *Config) applyEnvOverrides() {
	if storeType := os.Getenv("STORE_TYPE"); storeType != "" {
		c.Store.Type = storeType
	}
	if host := os.Getenv("MYSQL_HOST"); host != "" {
		c.Store.MySQL.Host = host
	}
	if port := os.Getenv("MYSQL_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &c.Store.MySQL.Port)
	}
	if user := os.Getenv("MYSQL_USER"); user != "" {
		c.Store.MySQL.User = user
	}
	if password := os.Getenv("MYSQL_PASSWORD"); password != "" {
		c.Store.MySQL.Password = password
	}
	if database := os.Getenv("MYSQL_DATABASE"); database != "" {
		c.Store.MySQL.Database = database
	}
}

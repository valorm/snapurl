// internal/config/config.go
package config

import (
    "os"
    "path/filepath"
    "strconv"
    "strings"

    "gopkg.in/yaml.v3"
)

type Config struct {
    Port      string   `yaml:"port"`
    DBPath    string   `yaml:"db_path"`
    RateLimit int      `yaml:"rate_limit"`
    APIKeys   []string `yaml:"api_keys"`
}

func LoadConfig() (*Config, error) {
    cfg := &Config{}

    // 1) Read YAML file
    yamlPath := filepath.Join("config", "default.yaml")
    data, err := os.ReadFile(yamlPath)
    if err != nil {
        return nil, err
    }
    if err := yaml.Unmarshal(data, cfg); err != nil {
        return nil, err
    }

    // 2) Environment overrides
    if port := os.Getenv("PORT"); port != "" {
        cfg.Port = port
    }
    if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
        cfg.DBPath = dbPath
    }
    if rl := os.Getenv("RATE_LIMIT"); rl != "" {
        if v, err := strconv.Atoi(rl); err == nil {
            cfg.RateLimit = v
        }
    }
    if keys := os.Getenv("API_KEYS"); keys != "" {
        cfg.APIKeys = strings.Split(keys, ",")
    }

    return cfg, nil
}

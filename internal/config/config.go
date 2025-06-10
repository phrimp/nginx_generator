package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type ServerConfig struct {
	Listen     string `json:"listen" yaml:"listen"`
	ServerName string `json:"server_name" yaml:"server_name"`
	Root       string `json:"root" yaml:"root"`
	Index      string `json:"index" yaml:"index"`
	ProxyPass  string `json:"proxy_pass" yaml:"proxy_pass"`
	ProxyPort  string `json:"proxy_port" yaml:"proxy_port"`
}

func Load(filepath string) (*ServerConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg ServerConfig
	ext := strings.ToLower(filepath[strings.LastIndex(filepath, ".")+1:])

	switch ext {
	case "json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case "yaml", "yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	if cfg.Listen == "" {
		cfg.Listen = "80"
	}
	if cfg.Index == "" && cfg.Root != "" {
		cfg.Index = "index.html"
	}

	return &cfg, nil
}

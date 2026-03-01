package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	ChainLength         int      `json:"chain_length"`
	RotateMin           int      `json:"rotate_min"`
	RotateMax           int      `json:"rotate_max"`
	TorEnabled          bool     `json:"tor_enabled"`
	TorPort             int      `json:"tor_port"`
	TorControlPort      int      `json:"tor_control_port"`
	LocalPort           int      `json:"local_port"`
	HealthCheckInterval int      `json:"health_check_interval"`
	ProxyTimeout        int      `json:"proxy_timeout"`
	MaxRetries          int      `json:"max_retries"`
	SelectedCountries   []string `json:"selected_countries"`
	ExcludedCountries   []string `json:"excluded_countries"`
	Protocols           []string `json:"protocols"`
	AutoUpdateProxy     bool     `json:"auto_update_proxy"`
	LogLevel            string   `json:"log_level"`
	DebugMode           bool     `json:"debug_mode"`
}

const configFile = "config/config.json"

func DefaultConfig() *Config {
	return &Config{
		ChainLength:         5,
		RotateMin:           10,
		RotateMax:           25,
		TorEnabled:          true,
		TorPort:             9050,
		TorControlPort:      9051,
		LocalPort:           1080,
		HealthCheckInterval: 30,
		ProxyTimeout:        10,
		MaxRetries:          3,
		SelectedCountries:   []string{"USA", "Germany", "Japan", "Brazil", "Singapore", "Netherlands", "Canada", "India"},
		ExcludedCountries:   []string{"China", "Russia", "Iran"},
		Protocols:           []string{"SOCKS5", "HTTP"},
		AutoUpdateProxy:     true,
		LogLevel:            "info",
		DebugMode:           false,
	}
}

func LoadConfig() (*Config, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return DefaultConfig(), err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return DefaultConfig(), err
	}

	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	// Create config directory if not exists
	os.MkdirAll("config", 0755)
	
	return ioutil.WriteFile(configFile, data, 0644)
}

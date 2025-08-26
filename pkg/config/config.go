package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port    int           `yaml:"port"`
		Timeout time.Duration `yaml:"timeout"`
	} `yaml:"server"`

	ExternalAPI struct {
		BaseURL string        `yaml:"base_url"`
		APIKey  string        `yaml:"api_key"`
		Timeout time.Duration `yaml:"timeout"`
	} `yaml:"external_api"`

	Cache struct {
		TTL int `yaml:"ttl"`
	} `yaml:"cache"`
}

func Load(path string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

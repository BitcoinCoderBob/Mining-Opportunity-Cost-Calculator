package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	PriceDataKrakenPath   string `yaml:"priceDataKrakenPath"`
	MessariUrl            string `yaml:"messariUrl"`
	BlockchainInfoUrl     string `yaml:"blockchainInfoUrl"`
	SlushPoolUrl          string `yaml:"slushPoolUrl"`
	PriceDataCoinbasePath string `yaml:"priceDataCoinbasePath"`
	DataPlotFileName      string `yaml:"dataPlotFileName"`
}

func New(filepath string) (*Config, error) {
	fd, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening the config file: %w", err)
	}

	var cfg Config
	if err := yaml.NewDecoder(fd).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("error decoding the config: %w", err)
	}

	return &cfg, nil
}

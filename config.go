package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port      int    `json:"port"`
	ChatType  string `json:"chatType"`
	APIURL    string `json:"apiURL"`
	ModelsURL string `json:"modelsURL"`
	APIKey    string `json:"apiKey"`
	Mock      bool   `json:"mock"`
	// Model        string            `json:"model"`
	DifyAppMap   map[string]string `json:"difyAppMap"`
	DifyTokenUrl string            `json:"difyTokenUrl"`
	DifyToken    string            `json:"-"`
}

func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var config Config

	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	XConfig = &config
	return &config, nil
}

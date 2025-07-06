package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port       int    `json:"port"`
	OpenaiPort int    `json:"openaiPort"`
	ChatType   string `json:"chatType"`
	APIURL     string `json:"apiURL"`
	APIURLProd string `json:"apiURLProd"`
	ModelsURL  string `json:"modelsURL"`
	APIKey     string `json:"apiKey"`
	Debug      bool   `json:"debug"`
	Mock       bool   `json:"mock"`
	// Model        string            `json:"model"`
	DifyAppMap       map[string]string `json:"difyAppMap"`
	DifyAppMapProd   map[string]string `json:"difyAppMapProd"`
	DifyTokenUrl     string            `json:"difyTokenUrl"`
	DifyTokenUrlProd string            `json:"difyTokenUrlProd"`
	Mapping          map[string]string `json:"mapping"`
	DifyTokenMap     map[string]string `json:"-"`
	IsProd           bool              `json:"-"`
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

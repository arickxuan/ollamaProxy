package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port       int    `json:"port"`
	OpenaiPort int    `json:"openaiPort"`
	ChatType   string `json:"chatType"`
	BaseUrl    string `json:"baseUrl"`
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
	ProxyMapping     map[string]string `json:"proxyMapping"`
	DifyTokenMap     map[string]string `json:"-"`
	IsProd           bool              `json:"-"`
	CAFile           string            `json:"caFile"`
	CAKeyFile        string            `json:"caKeyFile"`
	Domain           string            `json:"domain"`
	DomainPemFile    string            `json:"domainPemFile"`
	DomainKeyFile    string            `json:"domainKeyFile"`
	IsTls            bool              `json:"isTls"`
	OSSConfig        OSSConfig         `json:"oss"`
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

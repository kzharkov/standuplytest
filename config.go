package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	ApiKey   string `json:"api_key"`
	ConfigDB string `json:"config_db"`
}

func LoadConfig() (Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return Config{}, err
	}

	configRaw, err := ioutil.ReadAll(file)
	if err != nil {
		return Config{}, err
	}

	config := Config{}

	err = json.Unmarshal(configRaw, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

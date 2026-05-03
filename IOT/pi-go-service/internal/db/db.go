package db

import (
	"encoding/json"
	"os"
)

type Config struct {
	CentralDeviceID    string `json:"central_device_id"`
	CentralDeviceSecret string `json:"central_device_secret"`
	SkipFieldMode      bool   `json:"skip_field_mode"`
}

var configPath string
var config Config

func LoadConfig(path string) error {
	configPath = path

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			config = Config{}
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &config)
}

func GetHubID() string {
	return config.CentralDeviceID
}

func GetHubSecret() string {
	return config.CentralDeviceSecret
}

func GetSkipFieldMode() bool {
	return config.SkipFieldMode
}

func SaveConfig(key, value string) error {
	switch key {
	case "central_device_id":
		config.CentralDeviceID = value
	case "central_device_secret":
		config.CentralDeviceSecret = value
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func ResetConfig() error {
	config = Config{}
	data, _ := json.MarshalIndent(config, "", "  ")
	return os.WriteFile(configPath, data, 0644)
}
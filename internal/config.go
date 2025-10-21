package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DB_URL            string
	Current_User_Name string
}

func (c *Config) SetUser(current_user_name string) {
	c.Current_User_Name = current_user_name
	Write(c)
}
func (c *Config) GetCurrentUser() string {
	return c.Current_User_Name
}
func (c *Config) SetDBUrl(url string) {
	c.DB_URL = url
	Write(c)
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(homeDir, configFileName)

	return configPath, nil
}

func Read() (*Config, error) {
	var cfg Config
	configPath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	//fmt.Println(cfg.DB_URL)
	//fmt.Println(cfg.Current_User_Name)

	return &cfg, nil
}

func Write(cfg *Config) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(configPath, data, 0777); err != nil {
		return err
	}
	fmt.Println("Config updated successfully at:", configPath)
	return nil
}

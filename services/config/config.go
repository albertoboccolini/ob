package config

import (
	"log"
	"os"
	"path/filepath"
)

func GetConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	configDir := filepath.Join(home, ".config", "ob")
	return configDir
}

func CreateConfigDir() {
	if err := os.MkdirAll(GetConfigDir(), 0755); err != nil {
		log.Fatal("Error creating config directory:", err)
		os.Exit(1)
	}
}

var (
	configDir  = GetConfigDir()
	PidFile    = filepath.Join(configDir, "ob.pid")
	LogFile    = filepath.Join(configDir, "ob.log")
	ConfigFile = filepath.Join(configDir, "vault.path")
)

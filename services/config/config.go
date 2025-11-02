package config

import (
	"log"
	"os"
	"path/filepath"
)

var (
	configDir  string
	pidFile    string
	logFile    string
	configFile string
	OB_VERSION = "0.0.2"
)

func InitConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	configDir = filepath.Join(home, ".config", "ob")
	pidFile = filepath.Join(configDir, "ob.pid")
	logFile = filepath.Join(configDir, "ob.log")
	configFile = filepath.Join(configDir, "vault.path")
}

func CreateConfigDir() {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Fatal("Error creating config directory:", err)
	}
}

func GetPidFile() string {
	return pidFile
}

func GetLogFile() string {
	return logFile
}

func GetConfigFile() string {
	return configFile
}

package sync

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"ob/services/config"
	"ob/services/git"
)

func syncToRemote(vaultPath string) {
	err := git.PullIfNeeded(vaultPath)
	if err != nil {
		log.Println("Error syncing vault:", err)
		return
	}
}

func syncVault(vaultPath string) {
	hasChanges, err := git.HasUncommittedChanges(vaultPath)
	if err != nil {
		log.Println("Error checking for uncommitted changes:", err)
		return
	}

	if hasChanges {
		err = git.CommitChanges(vaultPath)
		if err != nil {
			log.Println("Error committing changes:", err)
			return
		}

		log.Println("Changes committed successfully.")
		return
	}
}

func ManualSync() {
	log.Println("Manual sync...")
	data, err := os.ReadFile(config.GetConfigFile())
	if err != nil {
		log.Fatal("Error reading vault path from config:", err)
	}

	vaultPath := strings.TrimSpace(string(data))
	syncVault(vaultPath)
	syncToRemote(vaultPath)
	fmt.Println("Manual sync completed")
}

func RunDaemon() {
	data, err := os.ReadFile(config.GetConfigFile())
	if err != nil {
		log.Fatal("Error reading vault path from config:", err)
	}

	vaultPath := strings.TrimSpace(string(data))
	config.CreateConfigDir()

	f, err := os.OpenFile(config.GetLogFile(), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error opening log file:", err)
	}

	defer f.Close()
	log.SetOutput(f)

	pidFile := config.GetPidFile()
	pid := os.Getpid()

	err = os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		log.Println("Warning: could not write PID file:", err)
	}

	defer os.Remove(pidFile)

	syncToRemoteTicker := time.NewTicker(12 * time.Hour)
	syncVaultTicker := time.NewTicker(1 * time.Minute)
	defer syncToRemoteTicker.Stop()
	defer syncVaultTicker.Stop()

	log.Println("Started sync operations...")

	syncVault(vaultPath)
	syncToRemote(vaultPath)

	for {
		select {
		case <-syncVaultTicker.C:
			syncVault(vaultPath)
		case <-syncToRemoteTicker.C:
			syncToRemote(vaultPath)
		}
	}
}

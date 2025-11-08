package sync

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"ob/services/config"
	"ob/services/git"
)

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

func completeSync(vaultPath string, manualSync ...bool) {
	threshold := 25
	err := git.PullIfNeeded(vaultPath)
	if err != nil {
		log.Println("Error syncing vault:", err)
		return
	}

	syncVault(vaultPath)

	isManualSync := len(manualSync) > 0 && manualSync[0]

	if isManualSync {
		threshold = 0
	}

	err = git.SquashAndPushIfNeeded(vaultPath, threshold)
	if err != nil {
		log.Println("Error syncing vault:", err)
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
	completeSync(vaultPath, true)
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
	completeSync(vaultPath)

	for {
		select {
		case <-syncVaultTicker.C:
			syncVault(vaultPath)
		case <-syncToRemoteTicker.C:
			completeSync(vaultPath)
		}
	}
}

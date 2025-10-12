package sync

import (
	"log"
	"ob/services/config"
	"ob/services/git"
	"os"
	"strconv"
	"time"
)

func syncToRemote(vaultPath string) error {
	err := git.PullIfNeeded(vaultPath)
	if err != nil {
		return err
	}

	log.Println("Sync to remote successful.")
	return nil
}

func syncVault(vaultPath string) error {
	hasChanges, err := git.HasUncommittedChanges(vaultPath)
	if err != nil {
		return err
	}

	if hasChanges {
		err = git.CommitChanges(vaultPath)
		if err != nil {
			return err
		}

		log.Println("Changes committed successfully.")
		return nil
	}

	return nil
}

func RunDaemon() {
	data, err := os.ReadFile(config.ConfigFile)
	if err != nil {
		log.Fatal("Error reading vault path from config:", err)
	}

	vaultPath := string(data)
	config.CreateConfigDir()

	f, err := os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error opening log file:", err)
	}

	defer f.Close()
	log.SetOutput(f)

	pid := os.Getpid()
	err = os.WriteFile(config.PidFile, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		log.Println("Warning: could not write PID file:", err)
	}

	defer os.Remove(config.PidFile)

	syncToRemoteTicker := time.NewTicker(12 * time.Hour)
	syncVaultTicker := time.NewTicker(1 * time.Minute)
	defer syncToRemoteTicker.Stop()
	defer syncVaultTicker.Stop()

	log.Println("Starting sync operations...")
	for {
		select {
		case <-syncToRemoteTicker.C:
			go func() {
				err := syncToRemote(vaultPath)
				if err != nil {
					log.Println("Error syncing to remote:", err)
				}
			}()
		case <-syncVaultTicker.C:
			go func() {
				err := syncVault(vaultPath)
				if err != nil {
					log.Println("Error syncing vault:", err)
				}
			}()
		}
	}
}

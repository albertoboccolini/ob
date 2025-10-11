package main

import "fmt"
import "time"
import "os"
import "os/exec"
import "flag"
import "strconv"
import "log"
import "path/filepath"
import "ob/services"

func getConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	configDir := filepath.Join(home, ".config", "ob")
	return configDir
}

var (
	configDir  = getConfigDir()
	pidFile    = filepath.Join(configDir, "ob.pid")
	logFile    = filepath.Join(configDir, "ob.log")
	configFile = filepath.Join(configDir, "vault.path")
)

func CreateConfigDir() {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Fatal("Error creating config directory:", err)
		os.Exit(1)
	}
}

func SyncToRemote(vaultPath string) error {
	err := git.PullIfNeeded(vaultPath)
	if err != nil {
		return err
	}

	log.Println("Sync to remote successful.")
	return nil
}

func SyncVault(vaultPath string) error {
	hasChanges, err := git.HasUncommittedChanges(vaultPath)
	if err != nil {
		log.Println("Error:", err)
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

func runDaemon() {
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal("Error reading vault path from config:", err)
	}
	vaultPath := string(data)

	CreateConfigDir()

	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error opening log file:", err)
	}
	defer f.Close()
	log.SetOutput(f)

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

	log.Println("Starting sync operations...")

	go func() {
		err := SyncToRemote(vaultPath)
		if err != nil {
			log.Println("Error syncing to remote:", err)
		}
	}()

	go func() {
		err := SyncVault(vaultPath)
		if err != nil {
			log.Println("Error syncing vault:", err)
		}
	}()

	for {
		select {
		case <-syncToRemoteTicker.C:
			go func() {
				err := SyncToRemote(vaultPath)
				if err != nil {
					log.Println("Error syncing to remote:", err)
				}
			}()
		case <-syncVaultTicker.C:
			go func() {
				err := SyncVault(vaultPath)
				if err != nil {
					log.Println("Error syncing vault:", err)
				}
			}()
		}
	}
}

func startSync(vaultPath string) {
	if _, err := os.Stat(pidFile); err == nil {
		fmt.Println("Sync is already running")
		os.Exit(1)
	}

	CreateConfigDir()

	err := os.WriteFile(configFile, []byte(vaultPath), 0644)
	if err != nil {
		fmt.Println("Error saving vault path:", err)
		os.Exit(1)
	}

	// Fork process
	cmd := exec.Command(os.Args[0], "--daemon")
	cmd.Start()

	fmt.Println("Sync started successfully")
	fmt.Printf("PID: %d\n", cmd.Process.Pid)
	fmt.Printf("Vault: %s\n", vaultPath)
	fmt.Printf("Logs: %s\n", logFile)
}

func stopSync() {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		fmt.Println("No running instance found")
		os.Exit(1)
	}

	os.Remove(pidFile) // Remove potentially stale PID file
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		fmt.Println("Invalid PID file")
		os.Exit(1)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Println("Process not found")
	}

	err = process.Kill()
	if err != nil {
		fmt.Println("Warning:", err)
	}

	fmt.Println("Sync stopped")
}

func main() {
	daemon := flag.Bool("daemon", false, "Run as daemon")
	flag.Parse()

	if *daemon {
		runDaemon()
		return
	}

	if flag.NArg() < 1 {
		fmt.Println("Usage: ob <command>")
		fmt.Println("Commands:")
		fmt.Println("  start <vault-path>    Start the sync operations")
		fmt.Println("  stop                  Stop the sync operations")
		os.Exit(1)
	}

	command := flag.Arg(0)

	switch command {
	case "start":
		if flag.NArg() < 2 {
			fmt.Println("Error: vault path is required")
			fmt.Println("Usage: ob start <vault-path>")
			os.Exit(1)
		}
		vaultPath := flag.Arg(1)
		startSync(vaultPath)
	case "stop":
		stopSync()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

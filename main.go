package main

import "fmt"
import "time"
import "os"
import "os/exec"
import "flag"
import "strconv"
import "syscall"
import "log"
import "path/filepath"
import "ob/services"

var (
	configDir = filepath.Join(os.Getenv("HOME"), ".config", "ob")
	pidFile   = filepath.Join(configDir, "ob.pid")
	logFile   = filepath.Join(configDir, "ob.log")
)

func SyncToRemote() error {
	err := git.PullIfNeeded()
	if err != nil {
		return err
	}

	log.Println("Sync to remote successful.")
	return nil
}

func SyncVault() error {
	hasChanges, err := git.HasUncommittedChanges()
	if err != nil {
		log.Println("Error:", err)
		return err
	}

	if hasChanges {
		err = git.CommitChanges()
		if err != nil {
			return err
		}
		log.Println("Changes committed successfully.")
		return nil
	}

	return nil
}

func runDaemon() {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Fatal("Error creating config directory:", err)
	}

	// Setup logging to file
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error opening log file:", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// Save PID
	pid := os.Getpid()
	err = os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		log.Println("Warning: could not write PID file:", err)
	}
	defer os.Remove(pidFile)

	syncToRemoteTicker := time.NewTicker(1 * time.Hour)
	syncVaultTicker := time.NewTicker(30 * time.Second)
	defer syncToRemoteTicker.Stop()
	defer syncVaultTicker.Stop()

	log.Println("Starting sync operations...")

	go func() {
		err := SyncToRemote()
		if err != nil {
			log.Println("Error syncing to remote:", err)
		}
	}()

	go func() {
		err := SyncVault()
		if err != nil {
			log.Println("Error syncing vault:", err)
		}
	}()

	for {
		select {
		case <-syncToRemoteTicker.C:
			log.Println("Executing SyncToRemote...")
			go func() {
				err := SyncToRemote()
				if err != nil {
					log.Println("Error syncing to remote:", err)
				}
			}()
		case <-syncVaultTicker.C:
			log.Println("Executing SyncVault...")
			go func() {
				err := SyncVault()
				if err != nil {
					log.Println("Error syncing vault:", err)
				}
			}()
		}
	}
}

func startSync() {
	// Check if already running
	if _, err := os.Stat(pidFile); err == nil {
		fmt.Println("Sync is already running")
		os.Exit(1)
	}

	// Fork process
	cmd := exec.Command(os.Args[0], "--daemon")
	cmd.Start()

	fmt.Println("Sync started successfully")
	fmt.Printf("PID: %d\n", cmd.Process.Pid)
	fmt.Printf("Logs: %s\n", logFile)
}

func stopSync() {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		fmt.Println("No running instance found")
		os.Exit(1)
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		fmt.Println("Invalid PID file")
		os.Exit(1)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Println("Process not found")
		os.Exit(1)
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		fmt.Println("Error stopping process:", err)
		os.Exit(1)
	}

	os.Remove(pidFile)
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
		fmt.Println("  start    Start the sync operations")
		fmt.Println("  stop     Stop the sync operations")
		os.Exit(1)
	}

	command := flag.Arg(0)

	switch command {
	case "start":
		startSync()
	case "stop":
		stopSync()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

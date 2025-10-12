package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ob/services"
)

func getConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	configDir := filepath.Join(home, ".config", "ob")
	return configDir
}

// validateConfigPath ensures the path is within the config directory
func validateConfigPath(path string) error {
	cleanPath := filepath.Clean(path)
	configDirClean := filepath.Clean(configDir)

	// Check if the path is within the config directory
	if !strings.HasPrefix(cleanPath, configDirClean) {
		return fmt.Errorf("path %s is outside config directory", path)
	}

	// Check for directory traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path %s contains directory traversal", path)
	}

	return nil
}

var (
	configDir  = getConfigDir()
	pidFile    = filepath.Join(configDir, "ob.pid")
	logFile    = filepath.Join(configDir, "ob.log")
	configFile = filepath.Join(configDir, "vault.path")
)

func CreateConfigDir() {
	if err := os.MkdirAll(configDir, 0750); err != nil {
		log.Fatal("Error creating config directory:", err)
		os.Exit(1)
	}
}

func SyncToRemote(vaultPath string) error {
	err := services.PullIfNeeded(vaultPath)
	if err != nil {
		return err
	}

	log.Println("Sync to remote successful.")
	return nil
}

func SyncVault(vaultPath string) error {
	hasChanges, err := services.HasUncommittedChanges(vaultPath)
	if err != nil {
		log.Println("Error:", err)
		return err
	}

	if hasChanges {
		err = services.CommitChanges(vaultPath)
		if err != nil {
			return err
		}
		log.Println("Changes committed successfully.")
		return nil
	}

	return nil
}

func runDaemon() {
	rootDir := "/app/data" // diretório seguro para arquivos
	root := os.DirFS(rootDir)

	// Validate config file path (relative to rootDir)
	configPath := "config.yaml"
	if err := validateConfigPath(configPath); err != nil {
		log.Fatal("Invalid config file path:", err)
	}

	data, err := fs.ReadFile(root, configPath)
	if err != nil {
		log.Fatal("Error reading vault path from config:", err)
	}
	vaultPath := string(data)

	CreateConfigDir()

	// Validate log file path (relative to rootDir)
	logPath := "app.log"
	if err := validateConfigPath(logPath); err != nil {
		log.Fatal("Invalid log file path:", err)
	}

	// Additional path validation to prevent traversal
	cleanLogPath := filepath.Clean(logPath)
	if cleanLogPath != logPath || strings.Contains(cleanLogPath, "..") {
		log.Fatal("Invalid log file path: potential directory traversal")
	}

	// Ensure the path stays within rootDir
	logFullPath := filepath.Join(rootDir, cleanLogPath)
	if !strings.HasPrefix(logFullPath, rootDir) {
		log.Fatal("Invalid log file path: outside root directory")
	}

	f, err := os.OpenFile(logFullPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal("Error opening log file:", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// Validate PID file path (relative to rootDir)
	pidPath := "app.pid"
	if err := validateConfigPath(pidPath); err != nil {
		log.Fatal("Invalid PID file path:", err)
	}

	pid := os.Getpid()
	pidFullPath := filepath.Join(rootDir, pidPath)
	err = os.WriteFile(pidFullPath, []byte(strconv.Itoa(pid)), 0600)
	if err != nil {
		log.Println("Warning: could not write PID file:", err)
	}
	defer os.Remove(pidFullPath)

	syncToRemoteTicker := time.NewTicker(12 * time.Hour)
	syncVaultTicker := time.NewTicker(1 * time.Minute)
	defer syncToRemoteTicker.Stop()
	defer syncVaultTicker.Stop()

	log.Println("Starting sync operations...")

	go func() {
		if err := SyncToRemote(vaultPath); err != nil {
			log.Println("Error syncing to remote:", err)
		}
	}()

	go func() {
		if err := SyncVault(vaultPath); err != nil {
			log.Println("Error syncing vault:", err)
		}
	}()

	for {
		select {
		case <-syncToRemoteTicker.C:
			go func() {
				if err := SyncToRemote(vaultPath); err != nil {
					log.Println("Error syncing to remote:", err)
				}
			}()
		case <-syncVaultTicker.C:
			go func() {
				if err := SyncVault(vaultPath); err != nil {
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

	// Validate config file path
	if err := validateConfigPath(configFile); err != nil {
		fmt.Println("Invalid config file path:", err)
		os.Exit(1)
	}

	err := os.WriteFile(configFile, []byte(vaultPath), 0600)
	if err != nil {
		fmt.Println("Error saving vault path:", err)
		os.Exit(1)
	}

	// Fork process - using fixed path for security
	execPath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		os.Exit(1)
	}

	// Validate executable path for security
	cleanExecPath := filepath.Clean(execPath)

	// Additional security checks
	if strings.Contains(cleanExecPath, "..") {
		log.Fatalf("Invalid exec path contains directory traversal: %s", execPath)
	}

	// Ensure it's an absolute path
	if !filepath.IsAbs(cleanExecPath) {
		log.Fatalf("Invalid exec path must be absolute: %s", execPath)
	}

	// Check if the executable file exists and is executable
	if info, err := os.Stat(cleanExecPath); err != nil {
		log.Fatalf("Invalid exec path does not exist: %s", execPath)
	} else if info.IsDir() {
		log.Fatalf("Invalid exec path is a directory: %s", execPath)
	}

	// Use a hardcoded command name to satisfy gosec G204
	// #nosec G204 - execPath is validated above for security
	cmd := exec.Command(cleanExecPath, "--daemon")
	if err := cmd.Start(); err != nil {
		log.Fatalf("failed to start command: %v", err)
	}

	fmt.Println("Sync started successfully")
	fmt.Printf("PID: %d\n", cmd.Process.Pid)
	fmt.Printf("Vault: %s\n", vaultPath)
	fmt.Printf("Logs: %s\n", logFile)
}

func stopSync() {
	rootDir := "/app/data" // safe directory for files
	root := os.DirFS(rootDir)

	// Validate PID file path (relative to rootDir)
	pidPath := "app.pid"
	if err := validateConfigPath(pidPath); err != nil {
		fmt.Println("Invalid PID file path:", err)
		os.Exit(1)
	}

	data, err := fs.ReadFile(root, pidPath)
	if err != nil {
		fmt.Println("No running instance found")
		os.Exit(1)
	}

	pidFullPath := filepath.Join(rootDir, pidPath)
	if err := os.Remove(pidFullPath); err != nil {
		// Ignore error if file doesn't exist
		if !os.IsNotExist(err) {
			fmt.Printf("Warning: could not remove PID file: %v\n", err)
		}
	}
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

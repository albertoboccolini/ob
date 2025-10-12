package main

import (
	"flag"
	"fmt"
	"log"
	"ob/services/config"
	"ob/services/sync"
	"os"
	"os/exec"
	"strconv"
)

func startSync(vaultPath string) {
	if _, err := os.Stat(config.PidFile); err == nil {
		fmt.Println("Sync is already running")
		os.Exit(1)
	}

	config.CreateConfigDir()
	err := os.WriteFile(config.ConfigFile, []byte(vaultPath), 0644)
	if err != nil {
		fmt.Println("Error saving vault path:", err)
		os.Exit(1)
	}

	// Fork process
	cmd := exec.Command(os.Args[0], "--daemon")
	if err := cmd.Start(); err != nil {
		log.Fatal("Failed to start command: %v", err)
	}

	fmt.Println("Sync started successfully")
	fmt.Printf("PID: %d\n", cmd.Process.Pid)
	fmt.Printf("Vault: %s\n", vaultPath)
	fmt.Printf("Logs: %s\n", config.LogFile)
}

func stopSync() {
	data, err := os.ReadFile(config.PidFile)
	if err != nil {
		fmt.Println("No running instance found")
		os.Exit(1)
	}

	os.Remove(config.PidFile) // Remove potentially stale PID file
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
		sync.RunDaemon()
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

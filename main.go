package main

import (
	"flag"
	"fmt"
	"log"
	"ob/services/boot"
	"ob/services/config"
	"ob/services/sync"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"gopkg.in/natefinch/lumberjack.v2"
)

const obStopMessage = "Sync stopped"

func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func startSync(vaultPath string) {
	pidFile := config.GetPidFile()
	if data, err := os.ReadFile(pidFile); err == nil {
		pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err == nil && isProcessRunning(pid) {
			fmt.Println("Sync is already running")
			os.Exit(1)
		}
		// Remove stale PID file
		if err := os.Remove(pidFile); err != nil {
			fmt.Printf("Failed to remove stale PID file %s: %v", pidFile, err)
			os.Exit(1)
		}
	}

	// Fork process
	cmd := exec.Command(os.Args[0], "--daemon")
	cmd.Start()

	fmt.Println("Sync started successfully")
	fmt.Printf("PID: %d\n", cmd.Process.Pid)
	fmt.Printf("Vault: %s\n", vaultPath)
	fmt.Printf("Logs: %s\n", config.GetLogFile())
}

func stopSync() {
	pidFile := config.GetPidFile()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		fmt.Println("No running instance found")
		os.Exit(1)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		fmt.Println("Invalid PID file")
		os.Exit(1)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Println("Process not found")
		os.Exit(1)
	}

	err = process.Kill()
	if err != nil {
		fmt.Println("Error stopping process:", err)
		os.Exit(1)
	}

	log.Println(obStopMessage)
	fmt.Println(obStopMessage)
}

func main() {
	config.InitConfig()
	logger := &lumberjack.Logger{
		Filename:   config.GetLogFile(),
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}
	log.SetOutput(logger)
	defer logger.Close()

	daemon := flag.Bool("daemon", false, "Run as daemon")
	var version bool
	flag.BoolVar(&version, "version", false, "Show version information")
	flag.BoolVar(&version, "v", false, "Show version information (shorthand)")
	flag.Parse()

	if version {
		fmt.Printf("v%s\n", config.OB_VERSION)
		return
	}

	if *daemon {
		sync.RunDaemon()
		return
	}

	if flag.NArg() < 1 {
		fmt.Println("Usage: ob <command>")
		fmt.Println("\nCommands:")
		fmt.Println("  start <vault-path>    Start the sync operations")
		fmt.Println("  stop                  Stop the sync operations")
		fmt.Println("  boot <enable|disable> Enable or disable ob to start on boot")
		fmt.Println("  sync                  Trigger a manual sync")
		fmt.Println("  version               Show the version information")
		fmt.Println("\nFlags:")
		fmt.Println("  -v, --version         Show the version information")
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
	case "sync":
		sync.ManualSync()
		fmt.Println("Manual sync completed")
	case "stop":
		stopSync()
	case "boot":
		boot.HandleBootCommand()
	case "version":
		fmt.Printf("v%s\n", config.OB_VERSION)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

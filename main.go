package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/albertoboccolini/ob/services/boot"
	"github.com/albertoboccolini/ob/services/config"
	"github.com/albertoboccolini/ob/services/git"
	"github.com/albertoboccolini/ob/services/sync"

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

func printLogs() {
	logFile := config.GetLogFile()
	file, err := os.Open(logFile)
	if err != nil {
		fmt.Printf("Error reading log file: %v\n", err)
		return
	}
	defer file.Close()

	_, err = io.Copy(os.Stdout, file)
	if err != nil {
		fmt.Printf("Error printing log file: %v\n", err)
	}
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

func isSyncRunning() bool {
	data, err := os.ReadFile(config.GetPidFile())
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false
	}
	return isProcessRunning(pid)
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
		fmt.Println("  squash <num>          Merge last N squashed commits into a single commit")
		fmt.Println("  status                Show the sync status and other useful information")
		fmt.Println("  logs                  Display the logs")
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
	case "squash":
		if flag.NArg() < 2 {
			fmt.Println("Error: number of commits is required")
			fmt.Println("Usage: ob squash <num>")
			os.Exit(1)
		}
		numCommits, err := strconv.Atoi(flag.Arg(1))
		if err != nil || numCommits < 1 {
			fmt.Println("Error: invalid number of commits")
			os.Exit(1)
		}
		data, err := os.ReadFile(config.GetConfigFile())
		if err != nil {
			fmt.Println("Error reading vault path from config:", err)
			os.Exit(1)
		}
		vaultPath := strings.TrimSpace(string(data))
		err = git.SquashCommits(vaultPath, numCommits)
		if err != nil {
			fmt.Println("Error squashing commits:", err)
			os.Exit(1)
		}
		fmt.Println("Commits squashed successfully")
	case "stop":
		stopSync()
	case "boot":
		boot.HandleBootCommand()
	case "logs":
		printLogs()
	case "status":
		data, err := os.ReadFile(config.GetConfigFile())
		if err != nil {
			log.Fatal("Error reading vault path from config:", err)
		}

		vaultPath := strings.TrimSpace(string(data))
		commits, err := git.GetCommitsDifference(vaultPath)

		if err != nil {
			log.Fatal("Error checking sync status:", err)
		}

		lastLocalCommit, err := git.GetLastCommitTime(vaultPath, "HEAD")
		if err != nil {
			log.Fatal("Error getting last local commit time:", err)
		}

		lastRemoteCommit, err := git.GetLastCommitTime(vaultPath, "origin/main")
		if err != nil {
			log.Fatal("Error getting last remote commit time:", err)
		}

		remoteCommits, err := git.GetRemoteCommitCount(vaultPath)
		if err != nil {
			log.Fatal("Error getting remote commit count:", err)
		}

		fmt.Printf("Sync is running: %t\n", isSyncRunning())
		fmt.Printf("Boot enabled: %t\n", boot.IsBootEnabled())
		fmt.Printf("Last local commit: %s\n", lastLocalCommit.Local().Format("02/01/2006 15:04"))
		fmt.Printf("Last remote commit: %s\n", lastRemoteCommit.Local().Format("02/01/2006 15:04"))
		fmt.Printf("Commits ahead of remote: %d\n", commits)
		fmt.Printf("Total remote commits: %d\n", remoteCommits)
		fmt.Printf("Vault: %s\n", vaultPath)
		fmt.Printf("Logs: %s\n", config.GetLogFile())
	case "version":
		fmt.Printf("v%s\n", config.OB_VERSION)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

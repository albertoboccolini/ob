package boot

import (
	"fmt"
	"log"
	"ob/services/config"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const bootDisableMessage = "Boot disabled successfully"
const bootEnableMessage = "Boot enabled successfully"

const serviceTemplate = `[Unit]
Description=ob Service
After=network-online.target ssh-agent.service
Wants=network-online.target ssh-agent.service

[Service]
Type=simple
ExecStart=%s --daemon
Restart=on-failure
RestartSec=10
Environment="PATH=%s"
Environment="HOME=%s"
Environment="VAULT_PATH=%s"
Environment="SSH_AUTH_SOCK=%s"
Environment="GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=yes"
StandardOutput=journal
StandardError=journal
ExecStartPre=/bin/sh -c 'for i in $(seq 1 30); do ping -c1 -W1 8.8.8.8 && exit 0; sleep 1; done; exit 1'

[Install]
WantedBy=default.target
`

func getServicePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("failed to get home directory: %v", err)
		os.Exit(1)
	}
	return filepath.Join(home, ".config", "systemd", "user", "ob.service")
}

func getExecutablePath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(executable)
}

func runSystemctl(args ...string) error {
	fullArgs := append([]string{"--user"}, args...)
	cmd := exec.Command("systemctl", fullArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func enableBootService(vaultPath string) error {
	execPath, err := getExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		pathEnv = "/usr/local/bin:/usr/bin:/bin"
	}

	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	if sshAuthSock == "" {
		sshAuthSock = filepath.Join(home, ".ssh", "agent.sock")
	}

	serviceContent := fmt.Sprintf(serviceTemplate,
		execPath,
		pathEnv,
		home,
		vaultPath,
		sshAuthSock,
	)

	serviceDir := filepath.Dir(getServicePath())
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return fmt.Errorf("failed to create systemd directory: %w", err)
	}

	if err := os.WriteFile(getServicePath(), []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	if err := runSystemctl("daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	if err := runSystemctl("enable", "ob.service"); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	log.Println(bootEnableMessage)
	fmt.Println(bootEnableMessage)
	return nil
}

func disableBootService() error {
	servicePath := getServicePath()

	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		return fmt.Errorf("boot is not configured")
	}

	// We disable directly to leave ob active
	if err := runSystemctl("disable", "ob.service"); err != nil {
		return fmt.Errorf("failed to disable service: %w", err)
	}

	if err := os.Remove(servicePath); err != nil {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	if err := runSystemctl("daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	log.Println(bootDisableMessage)
	fmt.Println(bootDisableMessage)
	return nil
}

func IsBootEnabled() bool {
	servicePath := getServicePath()
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func HandleBootCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Error: boot subcommand required (enable|disable)")
		os.Exit(1)
	}

	bootCmd := os.Args[2]
	switch bootCmd {
	case "enable":
		data, err := os.ReadFile(config.GetConfigFile())
		if err != nil {
			fmt.Println("Error: no vault configured. Run 'ob start <vault-path>' first")
			os.Exit(1)
		}

		vaultPath := strings.TrimSpace(string(data))
		if vaultPath == "" {
			fmt.Println("Error: vault path is empty in config")
			os.Exit(1)
		}

		if err := enableBootService(vaultPath); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "disable":
		if err := disableBootService(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown boot command: %s\n", bootCmd)
		os.Exit(1)
	}
}

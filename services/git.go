// Package services provides git operations for the ob tool.
package services

import (
	"os/exec"
	"strings"
)

func IssueCommand(command string, args []string) ([]string, error) {
	cmd := exec.Command(command, args...)

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		return []string{}, nil
	}

	lines := strings.Split(output, "\n")
	return lines, nil
}

func HasUncommittedChanges(vaultPath string) (bool, error) {
	lines, err := IssueCommand("git", []string{"-C", vaultPath, "status", "--porcelain"})
	if err != nil {
		return false, err
	}

	hasChanges := false
	for _, line := range lines {
		if line != "" {
			hasChanges = true
			break
		}
	}

	return hasChanges, nil
}

func CommitChanges(vaultPath string) error {
	_, err := IssueCommand("git", []string{"-C", vaultPath, "add", "."})
	if err != nil {
		return err
	}

	_, err = IssueCommand("git", []string{"-C", vaultPath, "commit", "-m", "Auto commit by ob"})
	if err != nil {
		return err
	}

	return nil
}

func PushChanges(vaultPath string) error {
	_, err := IssueCommand("git", []string{"-C", vaultPath, "push", "origin", "main"})
	if err != nil {
		return err
	}

	return nil
}

func PullIfNeeded(vaultPath string) error {
	_, err := IssueCommand("git", []string{"-C", vaultPath, "fetch", "origin", "main"})
	if err != nil {
		return err
	}

	lines, err := IssueCommand("git", []string{"-C", vaultPath, "rev-list", "--count", "HEAD..origin/main"})
	if err != nil {
		return err
	}

	if len(lines) > 0 && lines[0] != "0" {
		_, err = IssueCommand("git", []string{"-C", vaultPath, "pull", "-X", "theirs", "origin", "main"})
		if err != nil {
			return err
		}
		return nil
	}

	// Check if local is ahead and push if needed
	lines, err = IssueCommand("git", []string{"-C", vaultPath, "rev-list", "--count", "origin/main..HEAD"})
	if err != nil {
		return err
	}

	if len(lines) > 0 && lines[0] != "0" {
		// Squash commits and push
		_, err = IssueCommand("git", []string{"-C", vaultPath, "reset", "--soft", "origin/main"})
		if err != nil {
			return err
		}

		_, err = IssueCommand("git", []string{"-C", vaultPath, "commit", "-m", "Squashed commits by ob"})
		if err != nil {
			return err
		}

		err = PushChanges(vaultPath)
		if err != nil {
			return err
		}
	}

	return nil
}

package git

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func issueCommand(command string, args []string) ([]string, error) {
	cmd := exec.Command(command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return []string{}, nil
	}

	lines := strings.Split(output, "\n")
	return lines, nil
}

func GetCommitsDifference(vaultPath string) (int, error) {
	lines, err := issueCommand("git", []string{"-C", vaultPath, "log", "--oneline", "origin/main..HEAD"})
	if err != nil {
		return 0, err
	}

	return len(lines), nil
}

func HasUncommittedChanges(vaultPath string) (bool, error) {
	lines, err := issueCommand("git", []string{"-C", vaultPath, "status", "--porcelain"})
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
	_, err := issueCommand("git", []string{"-C", vaultPath, "add", "."})
	if err != nil {
		return err
	}

	_, err = issueCommand("git", []string{"-C", vaultPath, "commit", "-m", "Auto commit by ob"})
	if err != nil {
		return err
	}

	return nil
}

func PushChanges(vaultPath string) error {
	_, err := issueCommand("git", []string{"-C", vaultPath, "push", "origin", "main"})
	if err != nil {
		return err
	}

	return nil
}

func PullIfNeeded(vaultPath string) error {
	_, err := issueCommand("git", []string{"-C", vaultPath, "fetch", "origin", "main"})
	if err != nil {
		return err
	}

	lines, err := issueCommand("git", []string{"-C", vaultPath, "rev-list", "--count", "HEAD..origin/main"})
	if err != nil {
		return err
	}

	if len(lines) > 0 && lines[0] != "0" {
		_, err = issueCommand("git", []string{"-C", vaultPath, "pull", "-X", "theirs", "origin", "main"})
		if err != nil {
			return err
		}
		log.Println("Pulled changes from remote.")
	}

	return nil
}

func SquashAndPushIfNeeded(vaultPath string, commitThreshold int) error {
	_, err := issueCommand("git", []string{"-C", vaultPath, "fetch", "origin", "main"})
	if err != nil {
		return err
	}

	lines, err := issueCommand("git", []string{"-C", vaultPath, "rev-list", "--count", "origin/main..HEAD"})
	if err != nil {
		return err
	}

	if len(lines) == 0 {
		return nil
	}

	numCommits, err := strconv.Atoi(lines[0])
	if err != nil {
		return err
	}

	if numCommits == 0 || numCommits < commitThreshold {
		return nil
	}

	// Squash commits and push
	_, err = issueCommand("git", []string{"-C", vaultPath, "reset", "--soft", "origin/main"})
	if err != nil {
		return err
	}

	_, err = issueCommand("git", []string{"-C", vaultPath, "commit", "-m", "Squashed " + strconv.Itoa(numCommits) + " commits by ob"})
	if err != nil {
		return err
	}

	err = PushChanges(vaultPath)
	if err != nil {
		return err
	}

	log.Println("Sync to remote successful.")
	return nil
}

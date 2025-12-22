package git

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
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

func GetLastCommitTime(vaultPath string, ref string) (time.Time, error) {
	lines, err := issueCommand("git", []string{"-C", vaultPath, "log", "-1", "--format=%ct", ref})
	if err != nil {
		return time.Time{}, err
	}

	if len(lines) == 0 {
		return time.Time{}, fmt.Errorf("no commits found for ref: %s", ref)
	}

	timestamp, err := strconv.ParseInt(lines[0], 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(timestamp, 0), nil
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

func SquashCommits(vaultPath string, numCommits int) error {
	_, err := issueCommand("git", []string{"-C", vaultPath, "fetch", "origin", "main"})
	if err != nil {
		return err
	}

	localCommits, err := GetCommitsDifference(vaultPath)
	if err != nil {
		return err
	}

	if localCommits > 0 {
		_, err = issueCommand("git", []string{"-C", vaultPath, "reset", "--soft", "origin/main"})
		if err != nil {
			return err
		}

		_, err = issueCommand("git", []string{"-C", vaultPath, "commit", "-m", "Squashed " + strconv.Itoa(localCommits) + " commits by ob"})
		if err != nil {
			return err
		}

		err = PushChanges(vaultPath)
		if err != nil {
			return err
		}

		log.Println("Local commits squashed and pushed.")
	}

	lines, err := issueCommand("git", []string{"-C", vaultPath, "log", "--oneline", "-" + strconv.Itoa(numCommits)})
	if err != nil {
		return err
	}

	totalCommits := 0
	for _, line := range lines {
		if strings.Contains(line, "Squashed") && strings.Contains(line, "commits by ob") {
			parts := strings.Split(line, " ")
			for i, part := range parts {
				if part == "Squashed" && i+1 < len(parts) {
					count, err := strconv.Atoi(parts[i+1])
					if err == nil {
						totalCommits += count
					}
					break
				}
			}
		}
	}

	if totalCommits == 0 {
		return fmt.Errorf("no squashed commits found in the last %d commits. Ensure you're targeting commits with the format 'Squashed N commits by ob'", numCommits)
	}

	_, err = issueCommand("git", []string{"-C", vaultPath, "reset", "--soft", "HEAD~" + strconv.Itoa(numCommits)})
	if err != nil {
		return err
	}

	// Ensure numSquashedCommits does not exceed the number of available commits.
	lines, err = issueCommand("git", []string{"-C", vaultPath, "rev-list", "--count", "HEAD"})
	if err != nil {
		return err
	}

	if len(lines) == 0 {
		return fmt.Errorf("unable to determine commit count")
	}

	commitCount, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return fmt.Errorf("invalid commit count: %w", err)
	}

	if numCommits > commitCount {
		return fmt.Errorf("cannot squash %d commits; repository only has %d commits", numCommits, commitCount)
	}

	_, err = issueCommand("git", []string{"-C", vaultPath, "commit", "-m", "Squashed " + strconv.Itoa(totalCommits) + " commits by ob"})
	if err != nil {
		return err
	}

	_, err = issueCommand("git", []string{"-C", vaultPath, "push", "--force", "origin", "main"})
	if err != nil {
		return err
	}

	log.Printf("Squashed %d squash-commits representing %d total commits into one.", numCommits, totalCommits)
	return nil
}

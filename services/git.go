package git

import "os/exec"
import "strings"

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

func HasUncommittedChanges() (bool, error) {
	lines, err := IssueCommand("git", []string{"-C", "/home/albertoboccolini/Documenti/Obsidian/debian", "status", "--porcelain"})
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

func CommitChanges() error {
	_, err := IssueCommand("git", []string{"-C", "/home/albertoboccolini/Documenti/Obsidian/debian", "add", "."})
	if err != nil {
		return err
	}

	_, err = IssueCommand("git", []string{"-C", "/home/albertoboccolini/Documenti/Obsidian/debian", "commit", "-m", "Auto commit by ob"})
	if err != nil {
		return err
	}

	return nil
}

func PushChanges() error {
	_, err := IssueCommand("git", []string{"-C", "/home/albertoboccolini/Documenti/Obsidian/debian", "push", "origin", "main"})
	if err != nil {
		return err
	}

	return nil
}

func PullIfNeeded() error {
	_, err := IssueCommand("git", []string{"-C", "/home/albertoboccolini/Documenti/Obsidian/debian", "fetch", "origin", "main"})
	if err != nil {
		return err
	}

	lines, err := IssueCommand("git", []string{"-C", "/home/albertoboccolini/Documenti/Obsidian/debian", "rev-list", "--count", "HEAD..origin/main"})
	if err != nil {
		return err
	}

	if len(lines) > 0 && lines[0] != "0" {
		_, err = IssueCommand("git", []string{"-C", "/home/albertoboccolini/Documenti/Obsidian/debian", "pull", "-X", "theirs", "origin", "main"})
		if err != nil {
			return err
		}
		return nil
	}

	// Check if local is ahead and push if needed
	lines, err = IssueCommand("git", []string{"-C", "/home/albertoboccolini/Documenti/Obsidian/debian", "rev-list", "--count", "origin/main..HEAD"})
	if err != nil {
		return err
	}

	if len(lines) > 0 && lines[0] != "0" {
		// Squash commits and push
		_, err = IssueCommand("git", []string{"-C", "/home/albertoboccolini/Documenti/Obsidian/debian", "reset", "--soft", "origin/main"})
		if err != nil {
			return err
		}

		_, err = IssueCommand("git", []string{"-C", "/home/albertoboccolini/Documenti/Obsidian/debian", "commit", "-m", "Squashed commits by ob"})
		if err != nil {
			return err
		}

		err = PushChanges()
		if err != nil {
			return err
		}
	}

	return nil
}

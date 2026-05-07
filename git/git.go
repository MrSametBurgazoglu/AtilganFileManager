package git

import (
	"bufio"
	"os/exec"
	"path/filepath"
	"strings"
)

type Status string

const (
	StatusUntracked Status = "untracked"
	StatusModified  Status = "modified"
	StatusAdded     Status = "added"
	StatusDeleted   Status = "deleted"
	StatusRenamed   Status = "renamed"
	StatusCopied    Status = "copied"
	StatusUnchanged Status = "unchanged"
)

type GitManager struct {
	Statuses map[string]Status
}

func NewGitManager() *GitManager {
	return &GitManager{
		Statuses: make(map[string]Status),
	}
}

func (gm *GitManager) Refresh(repoPath string) error {
	// Clear current statuses
	gm.Statuses = make(map[string]Status)

	// Check if it's a git repo
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		return err
	}

	// Get status
	cmd = exec.Command("git", "-C", repoPath, "status", "--porcelain", "-uall")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 4 {
			continue
		}

		statusChar := line[:2]
		filePath := line[3:]

		// Handle renamed files (e.g., "R  old -> new")
		if strings.Contains(filePath, " -> ") {
			parts := strings.Split(filePath, " -> ")
			filePath = parts[1]
		}

		status := StatusUnchanged
		switch statusChar[0] {
		case 'M', ' ':
			if statusChar[1] == 'M' {
				status = StatusModified
			} else if statusChar[0] == 'M' {
				status = StatusAdded // Staged modification is often treated as "added" or just "staged"
			}
		case 'A':
			status = StatusAdded
		case 'D':
			status = StatusDeleted
		case 'R':
			status = StatusRenamed
		case 'C':
			status = StatusCopied
		case '?':
			status = StatusUntracked
		}
		
		if statusChar[1] == 'M' {
			status = StatusModified
		}

		gm.Statuses[filepath.Join(repoPath, filePath)] = status
	}

	return cmd.Wait()
}

func (gm *GitManager) GetStatus(path string) Status {
	if status, ok := gm.Statuses[path]; ok {
		return status
	}
	return StatusUnchanged
}

package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// CommitManager handles git commit operations
type CommitManager struct {
	repoPath string
}

// NewCommitManager creates a new CommitManager
func NewCommitManager(repoPath string) *CommitManager {
	return &CommitManager{repoPath: repoPath}
}

// HasChanges checks if there are uncommitted changes
func (cm *CommitManager) HasChanges() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = cm.repoPath
	output, err := cmd.CombinedOutput()
	return err == nil && strings.TrimSpace(string(output)) != ""
}

// StageAll stages all changes
func (cm *CommitManager) StageAll() error {
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = cm.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stage changes: %w, output: %s", err, string(output))
	}
	return nil
}

// Commit creates a commit with the given message
func (cm *CommitManager) Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = cm.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to commit: %w, output: %s", err, string(output))
	}
	return nil
}

// GetLastCommitHash returns the hash of the last commit
func (cm *CommitManager) GetLastCommitHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = cm.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get last commit: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

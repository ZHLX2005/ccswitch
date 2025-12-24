package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// RebaseManager handles git rebase operations
type RebaseManager struct {
	repoPath string
}

// NewRebaseManager creates a new RebaseManager
func NewRebaseManager(repoPath string) *RebaseManager {
	return &RebaseManager{repoPath: repoPath}
}

// RebaseCommit rebases a specific commit onto the current branch
// Returns (success, conflictDetected, error)
func (rm *RebaseManager) RebaseCommit(commitHash string) (bool, bool, error) {
	// Perform rebase
	rebaseCmd := exec.Command("git", "rebase", commitHash)
	rebaseCmd.Dir = rm.repoPath
	output, err := rebaseCmd.CombinedOutput()

	if err != nil {
		outputStr := string(output)
		// Check if it's a conflict error
		if strings.Contains(outputStr, "conflict") || strings.Contains(outputStr, "CONFLICT") ||
			strings.Contains(outputStr, "Failed to merge") {
			// Auto-abort on conflict
			_ = rm.AbortRebase()
			return false, true, fmt.Errorf("rebase conflict detected, auto-aborted")
		}
		return false, false, fmt.Errorf("rebase failed: %w, output: %s", err, outputStr)
	}

	return true, false, nil
}

// AbortRebase aborts the current rebase
func (rm *RebaseManager) AbortRebase() error {
	cmd := exec.Command("git", "rebase", "--abort")
	cmd.Dir = rm.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to abort rebase: %w, output: %s", err, string(output))
	}
	return nil
}

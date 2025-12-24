package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetRepoName returns the repository name from the current directory
func GetRepoName(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	repoPath := strings.TrimSpace(string(output))
	return filepath.Base(repoPath), nil
}

// GetMainRepoPath returns the path to the main repository (not worktree)
func GetMainRepoPath(dir string) (string, error) {
	// First get the common git directory
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	gitDir := strings.TrimSpace(string(output))

	// If gitDir is just ".git", we're in the main repo already
	if gitDir == ".git" {
		cmd = exec.Command("git", "rev-parse", "--show-toplevel")
		cmd.Dir = dir
		output, err = cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(output)), nil
	}

	// The main repo path is the parent of the .git directory
	mainPath := filepath.Dir(gitDir)

	// If the path ends with .git, it's already correct
	// If not, we might be in the main repo already
	if !strings.HasSuffix(gitDir, ".git") {
		// We're likely in a bare repository or the main repo
		cmd = exec.Command("git", "rev-parse", "--show-toplevel")
		cmd.Dir = dir
		output, err = cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		mainPath = strings.TrimSpace(string(output))
	}

	return mainPath, nil
}

// IsGitRepository checks if the directory is a git repository
func IsGitRepository(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	if err == nil {
		return true
	}

	// Check if we're in a worktree or subdirectory
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	err = cmd.Run()
	return err == nil
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// HasUncommittedChanges checks if a worktree has uncommitted changes
func HasUncommittedChanges(dir string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return err == nil && strings.TrimSpace(string(output)) != ""
}

// GetCommitCountDifference returns the number of commits the worktree branch
// is ahead (+) or behind (-) relative to the base branch.
// Positive values = ahead, Negative = behind, Zero = same
func GetCommitCountDifference(worktreePath, baseBranch string) (int, error) {
	// Get ahead count: commits in worktree that are not in baseBranch
	aheadCmd := exec.Command("git", "rev-list", "--count", baseBranch+"..HEAD")
	aheadCmd.Dir = worktreePath
	aheadOutput, err := aheadCmd.CombinedOutput()
	if err != nil {
		return 0, err
	}
	ahead := strings.TrimSpace(string(aheadOutput))

	// Get behind count: commits in baseBranch that are not in worktree
	behindCmd := exec.Command("git", "rev-list", "--count", "HEAD.."+baseBranch)
	behindCmd.Dir = worktreePath
	behindOutput, err := behindCmd.CombinedOutput()
	if err != nil {
		return 0, err
	}
	behind := strings.TrimSpace(string(behindOutput))

	// Parse counts (default to 0 if empty)
	aheadCount := 0
	behindCount := 0
	if ahead != "" {
		fmt.Sscanf(ahead, "%d", &aheadCount)
	}
	if behind != "" {
		fmt.Sscanf(behind, "%d", &behindCount)
	}

	// Return net difference (positive = ahead, negative = behind)
	return aheadCount - behindCount, nil
}

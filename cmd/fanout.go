package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/ksred/ccswitch/internal/git"
	"github.com/ksred/ccswitch/internal/session"
	"github.com/ksred/ccswitch/internal/ui"
	"github.com/spf13/cobra"
)

func newFanoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fanout",
		Short: "Propagate current branch commits to all other worktrees",
		Long: `Synchronously rebase all worktree branches onto the current branch.

This is useful for synchronizing all feature branches with a core business branch.

Safety checks before fanout:
  1. No other worktree has uncommitted changes
  2. No other worktree is ahead of current branch
  3. Auto-abort on any conflict

Examples:
  ccswitch fanout    # Interactive confirmation and fanout`,
		Run: fanoutBranches,
	}

	return cmd
}

func fanoutBranches(cmd *cobra.Command, args []string) {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		ui.Error("✗ Failed to get current directory")
		return
	}

	// Create session manager
	manager := session.NewManager(currentDir)

	// Get current branch (source branch)
	currentBranch, err := manager.GetCurrentBranch()
	if err != nil {
		ui.Errorf("✗ Failed to get current branch: %v", err)
		return
	}

	// Get all worktrees
	worktreeManager := git.NewWorktreeManager(currentDir)
	worktrees, err := worktreeManager.List()
	if err != nil {
		ui.Errorf("✗ Failed to list worktrees: %v", err)
		return
	}

	// Filter out current directory and find target worktrees
	var targetWorktrees []git.Worktree
	for _, wt := range worktrees {
		// Skip current directory and empty branches
		if wt.Path != currentDir && wt.Branch != "" && wt.Branch != currentBranch {
			targetWorktrees = append(targetWorktrees, wt)
		}
	}

	if len(targetWorktrees) == 0 {
		ui.Info("No other worktrees found to fanout to")
		return
	}

	ui.Title("Fanout Plan")
	ui.Infof("Source: %s (current branch)", currentBranch)
	ui.Infof("Targets: %d worktree(s)", len(targetWorktrees))
	fmt.Println()

	// Color definitions
	yellow := color.New(color.FgYellow, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	green := color.New(color.FgGreen)

	// Safety checks
	var unsafeWorktrees []string
	var safeWorktrees []git.Worktree

	for _, wt := range targetWorktrees {
		// Check 1: Uncommitted changes
		if git.HasUncommittedChanges(wt.Path) {
			yellow.Printf("  ● %s (%s)\n", wt.Branch, wt.Path)
			fmt.Println("     ⚠ Has uncommitted changes - cannot fanout")
			unsafeWorktrees = append(unsafeWorktrees, wt.Branch)
			continue
		}

		// Check 2: Branch is ahead of current
		diff, err := git.GetCommitCountDifference(wt.Path, currentBranch)
		if err != nil {
			ui.Errorf("  ✗ %s: failed to check status - %v", wt.Branch, err)
			unsafeWorktrees = append(unsafeWorktrees, wt.Branch)
			continue
		}

		if diff > 0 {
			red.Printf("  ↑ %s (%s)\n", wt.Branch, wt.Path)
			fmt.Printf("     ⚠ Ahead of %s by %d commit(s) - cannot fanout\n", currentBranch, diff)
			unsafeWorktrees = append(unsafeWorktrees, wt.Branch)
			continue
		}

		// Safe to fanout
		safeWorktrees = append(safeWorktrees, wt)
		green.Printf("  ○ %s (%s)\n", wt.Branch, wt.Path)
		if diff < 0 {
			fmt.Printf("     Behind by %d commit(s)\n", -diff)
		} else {
			fmt.Println("     Up to date")
		}
	}

	fmt.Println()

	// Check if any worktree is unsafe
	if len(unsafeWorktrees) > 0 {
		ui.Errorf("✗ Cannot fanout: %d worktree(s) failed safety checks", len(unsafeWorktrees))
		ui.Info("Please fix the issues above before running fanout")
		return
	}

	if len(safeWorktrees) == 0 {
		ui.Info("No worktrees to fanout to")
		return
	}

	// Confirm with user
	ui.Title("Ready to Fanout")
	ui.Warningf("This will rebase %d worktree(s) onto %s", len(safeWorktrees), currentBranch)
	ui.Info("Worktrees will be preserved after successful fanout")
	fmt.Println()
	fmt.Print("Continue? (yes/no): ")

	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "yes" {
		ui.Info("Fanout cancelled")
		return
	}

	// Perform fanout
	ui.Title("Fanout Progress")
	fmt.Println()

	successCount := 0
	for _, wt := range safeWorktrees {
		ui.Infof("Rebasing %s onto %s...", wt.Branch, currentBranch)

		// Perform rebase directly in the worktree
		success, hasConflict, errMsg := rebaseWorktree(wt.Path, currentBranch)

		if errMsg != nil {
			if hasConflict {
				ui.Errorf("  ✗ Conflict detected, auto-aborted")
				ui.Errorf("✗ Fanout stopped at %s due to conflict", wt.Branch)
				ui.Info("Please resolve conflicts manually before continuing")
				return
			}
			ui.Errorf("  ✗ Failed: %v", errMsg)
			ui.Errorf("✗ Fanout stopped at %s", wt.Branch)
			return
		}

		if !success {
			ui.Errorf("  ✗ Rebase failed")
			return
		}

		ui.Successf("  ✓ Success")
		successCount++
	}

	// Summary
	fmt.Println()
	ui.Title("Fanout Complete")
	ui.Successf("✓ Successfully fanned out to %d worktree(s)", successCount)
	if successCount > 0 {
		ui.Infof("All worktrees are now synchronized with %s", currentBranch)
	}
}

// rebaseWorktree rebases a worktree onto the specified branch
func rebaseWorktree(worktreePath, branch string) (success, conflict bool, err error) {
	// Perform rebase
	rebaseCmd := exec.Command("git", "rebase", branch)
	rebaseCmd.Dir = worktreePath
	output, e := rebaseCmd.CombinedOutput()

	if e != nil {
		outputStr := string(output)
		// Check if it's a conflict error
		if strings.Contains(outputStr, "conflict") || strings.Contains(outputStr, "CONFLICT") ||
			strings.Contains(outputStr, "Failed to merge") {
			// Auto-abort on conflict
			abortRebaseInWorktree(worktreePath)
			return false, true, fmt.Errorf("rebase conflict detected, auto-aborted")
		}
		return false, false, fmt.Errorf("rebase failed: %w, output: %s", e, outputStr)
	}

	return true, false, nil
}

// abortRebaseInWorktree aborts an ongoing rebase in a worktree
func abortRebaseInWorktree(worktreePath string) {
	cmd := exec.Command("git", "rebase", "--abort")
	cmd.Dir = worktreePath
	_ = cmd.Run()
}

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/ksred/ccswitch/internal/git"
	"github.com/ksred/ccswitch/internal/session"
	"github.com/ksred/ccswitch/internal/ui"
	"github.com/spf13/cobra"
)

func newRebaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rebase [worktree-path|branch-name]",
		Short: "Commit changes in a worktree and rebase to current branch",
		Long: `Commit changes in a worktree and rebase it to the current branch.

This allows you to quickly merge work from any worktree into your current branch.
Works with both ccswitch sessions and manually created git worktrees.

The rebase will:
1. Prompt for a commit message
2. Stage and commit all changes in the worktree
3. Rebase the commit onto the current branch
4. Automatically abort if conflicts are detected

Examples:
  ccswitch rebase                    # Interactive selection from all worktrees
  ccswitch rebase /path/to/worktree  # Rebase specific worktree by path
  ccswitch rebase feature-branch     # Rebase worktree by branch name`,
		Args: cobra.MaximumNArgs(1),
		Run:  rebaseSession,
	}

	return cmd
}

func rebaseSession(cmd *cobra.Command, args []string) {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		ui.Error("✗ Failed to get current directory")
		return
	}

	// Create session manager
	manager := session.NewManager(currentDir)

	// Get current branch (target branch for rebase)
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

	if len(worktrees) == 0 {
		ui.Info("No worktrees found")
		return
	}

	// Determine which worktree to rebase
	var targetWorktree *git.Worktree
	if len(args) > 0 {
		target := args[0]
		// Check if it's a path
		if filepath.IsAbs(target) || strings.HasPrefix(target, ".") || strings.HasPrefix(target, "~") {
			// Find by path
			for _, wt := range worktrees {
				if wt.Path == target || strings.HasSuffix(wt.Path, target) {
					targetWorktree = &wt
					break
				}
			}
		} else {
			// Find by branch name
			for _, wt := range worktrees {
				if wt.Branch == target {
					targetWorktree = &wt
					break
				}
			}
		}
		if targetWorktree == nil {
			ui.Errorf("✗ Worktree '%s' not found", target)
			ui.Info("Available worktrees:")
			for _, wt := range worktrees {
				name := getWorktreeDisplayName(wt, currentDir)
				fmt.Printf("  %s (%s)\n", name, wt.Branch)
				fmt.Printf("    Path: %s\n", wt.Path)
			}
			return
		}
	} else {
		// Interactive selection
		targetWorktree = selectWorktreeForRebase(worktrees, currentDir)
		if targetWorktree == nil {
			return // User quit
		}
	}

	// Skip if trying to rebase current branch onto itself
	if targetWorktree.Branch == currentBranch {
		ui.Errorf("✗ Cannot rebase %s onto itself", currentBranch)
		return
	}

	// Get commit message
	displayName := getWorktreeDisplayName(*targetWorktree, currentDir)
	ui.Infof("Rebasing %s onto %s", displayName, currentBranch)
	fmt.Println()

	commitMessage := promptForCommitMessage()
	if commitMessage == "" {
		ui.Error("✗ Commit message cannot be empty")
		return
	}

	// Perform commit and rebase
	ui.Info("Committing changes...")
	if err := manager.CommitAndRebaseSession(targetWorktree.Path, commitMessage); err != nil {
		ui.Errorf("✗ Failed: %v", err)
		return
	}

	ui.Successf("✓ Successfully rebased %s onto %s", displayName, currentBranch)
	ui.Infof("Worktree preserved at: %s", targetWorktree.Path)
}

func selectWorktreeForRebase(worktrees []git.Worktree, currentDir string) *git.Worktree {
	// Filter out current directory and main worktree
	var availableWorktrees []git.Worktree

	for _, wt := range worktrees {
		// Skip current directory and empty branches (detached HEAD)
		if wt.Path != currentDir && wt.Branch != "" {
			availableWorktrees = append(availableWorktrees, wt)
		}
	}

	if len(availableWorktrees) == 0 {
		ui.Info("No worktrees available for rebase")
		return nil
	}

	// Get current branch for comparison
	currentBranch, _ := git.GetCurrentBranch(currentDir)

	// Color definitions
	yellow := color.New(color.FgYellow, color.Bold)
	green := color.New(color.FgGreen)
	gray := color.New(color.FgHiBlack)

	// Show numbered list
	ui.Title("Select worktree to rebase:")
	fmt.Println()

	for i, wt := range availableWorktrees {
		name := getWorktreeDisplayName(wt, currentDir)

		// Determine status and color
		var statusColor *color.Color
		var statusIcon string

		if git.HasUncommittedChanges(wt.Path) {
			// Has uncommitted changes - Yellow
			statusColor = yellow
			statusIcon = "●"
		} else if diff, err := git.GetCommitCountDifference(wt.Path, currentBranch); err == nil && diff > 0 {
			// Ahead of current branch - Green
			statusColor = green
			statusIcon = "↑"
		} else {
			// Behind or same - Gray/Default
			statusColor = gray
			statusIcon = "○"
		}

		// Print with status color
		statusColor.Printf("  %d. %s %s (%s)\n", i+1, statusIcon, name, wt.Branch)
		fmt.Printf("     Path: %s\n", wt.Path)
	}

	fmt.Println()
	fmt.Print("Enter number (or q to quit): ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return nil
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "q" || input == "" {
		return nil
	}

	// Parse number
	var choice int
	if _, err := fmt.Sscanf(input, "%d", &choice); err != nil || choice < 1 || choice > len(availableWorktrees) {
		ui.Error("✗ Invalid selection")
		return nil
	}

	return &availableWorktrees[choice-1]
}

// getWorktreeDisplayName returns a friendly name for the worktree
func getWorktreeDisplayName(wt git.Worktree, currentDir string) string {
	// Check if it's a ccswitch session
	if strings.Contains(wt.Path, ".ccswitch/worktrees/") {
		parts := strings.Split(wt.Path, string(filepath.Separator))
		for i, part := range parts {
			if part == ".ccswitch" && i+2 < len(parts) {
				// Return just the session name
				return parts[i+2]
			}
		}
	}

	// For non-ccswitch worktrees, use branch name or basename
	if wt.Branch != "" {
		return wt.Branch
	}
	return filepath.Base(wt.Path)
}

func promptForCommitMessage() string {
	fmt.Print("Enter commit message: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return ""
	}
	return strings.TrimSpace(scanner.Text())
}

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ksred/ccswitch/internal/config"
	"github.com/ksred/ccswitch/internal/errors"
	"github.com/ksred/ccswitch/internal/git"
	"github.com/ksred/ccswitch/internal/session"
	"github.com/ksred/ccswitch/internal/ui"
	"github.com/ksred/ccswitch/internal/utils"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new session",
		Run:   createSession,
	}
}

func createSession(cmd *cobra.Command, args []string) {
	scanner := bufio.NewScanner(os.Stdin)

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		ui.Error("✗ Failed to get current directory")
		return
	}

	// Check if we're in a git repository root
	mainRepoPath, err := git.GetMainRepoPath(currentDir)
	if err != nil {
		ui.Errorf("✗ Failed to get git repository: %v", err)
		return
	}

	// Check if current directory is the git root
	if currentDir != mainRepoPath {
		// We're in a subdirectory
		ui.Warningf("You are currently in a subdirectory of the git repository")
		ui.Infof("Current directory: %s", currentDir)
		ui.Infof("Git root: %s", mainRepoPath)
		fmt.Println()

		// Ask user if they want to continue
		fmt.Print("Do you want to create a session from this subdirectory? (yes/no): ")

		if !scanner.Scan() {
			ui.Info("Session creation cancelled")
			return
		}

		answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if answer != "yes" && answer != "y" {
			ui.Info("Session creation cancelled")
			ui.Info("Tip: Navigate to the git repository root first:")
			fmt.Printf("  cd %s\n", mainRepoPath)
			return
		}
	}

	// Create session manager
	manager := session.NewManager(currentDir)

	// Get description from user
	fmt.Print(ui.TitleStyle.Render("🚀 What are you working on? "))

	if !scanner.Scan() {
		return
	}

	description := strings.TrimSpace(scanner.Text())
	if description == "" {
		ui.Error("✗ Description cannot be empty")
		return
	}

	// Create the session
	if err := manager.CreateSession(description); err != nil {
		ui.Errorf("✗ %s", err)

		// Provide helpful tips based on error
		hint := errors.ErrorHint(err)
		if hint != "" {
			ui.Infof("  Tip: %s", hint)
		}

		// Special handling for branch exists error
		if errors.IsBranchExists(err) {
			cfg, _ := config.Load()
			branchName := cfg.Branch.Prefix + utils.Slugify(description)
			ui.Infof("  Branch: %s", branchName)
		}
		return
	}

	// Success!
	sessionName := utils.Slugify(description)
	cfg, _ := config.Load()
	branchName := cfg.Branch.Prefix + sessionName
	repoName := filepath.Base(mainRepoPath)

	// Get the full worktree path
	homeDir, _ := os.UserHomeDir()
	worktreePath := filepath.Join(homeDir, ".ccswitch", "worktrees", repoName, sessionName)

	ui.Successf("✓ Created session: %s", sessionName)
	ui.Infof("Branch: %s", branchName)
	ui.Infof("Location: ~/.ccswitch/worktrees/%s/%s", repoName, sessionName)

	// Output the cd command for the shell wrapper to execute on a separate line
	fmt.Printf("\ncd %s\n", worktreePath)

	// If shell integration is not active, show a helpful message
	if !utils.IsShellIntegrationActive() {
		fmt.Println()
		ui.Info("💡 Note: Shell integration is not active.")
		ui.Info(utils.GetShellIntegrationInstructions())
	}
}

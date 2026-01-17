package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ksred/ccswitch/internal/session"
	"github.com/ksred/ccswitch/internal/ui"
	"github.com/spf13/cobra"
)

func newWorkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "work <command> [args...]",
		Short: "Execute a command in a selected worktree session",
		Long: `Execute a command in a selected worktree session.

This command will:
1. Show an interactive list of all sessions
2. Let you select which session to work in
3. Execute the specified command in that session's directory

Examples:
  ccswitch work make build
  ccswitch work npm test
  ccswitch work python script.py`,
		Args: cobra.MinimumNArgs(1),
		Run:  workCommand,
	}
}

func workCommand(cmd *cobra.Command, args []string) {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		ui.Error("✗ Failed to get current directory")
		return
	}

	// Create session manager
	manager := session.NewManager(currentDir)

	// Get sessions
	sessions, err := manager.ListSessions()
	if err != nil {
		ui.Errorf("✗ Failed to list sessions: %v", err)
		return
	}

	if len(sessions) == 0 {
		ui.Info("No active sessions")
		return
	}

	// Use interactive selector
	selector := ui.NewSessionSelector(sessions)
	p := tea.NewProgram(selector)

	if _, err := p.Run(); err != nil {
		ui.Errorf("✗ Failed to run selector: %v", err)
		return
	}

	if selector.IsQuit() {
		return
	}

	selected := selector.GetSelected()
	if selected == nil {
		return
	}

	// Build the command to execute
	commandName := args[0]
	var commandArgs []string
	if len(args) > 1 {
		commandArgs = args[1:]
	}

	// Execute the command in the selected session directory
	ui.Infof("→ Executing in session '%s': %s %s", selected.Name, commandName, strings.Join(commandArgs, " "))
	ui.Infof("  Location: %s", selected.Path)
	fmt.Println()

	err = executeInDir(selected.Path, commandName, commandArgs)
	if err != nil {
		ui.Errorf("✗ Command execution failed: %v", err)
		os.Exit(1)
	}
}

// executeInDir executes a command in the specified directory
func executeInDir(dir, command string, args []string) error {
	// Create the command
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// For Windows, we might need to use shell execution for some commands
	if runtime.GOOS == "windows" {
		// Check if this is a shell built-in or batch file
		if isShellBuiltin(command) {
			return executeViaShell(dir, command, args)
		}
	}

	return cmd.Run()
}

// isShellBuiltin checks if a command is a shell built-in (Windows)
func isShellBuiltin(command string) bool {
	shellBuiltins := map[string]bool{
		"dir": true, "cd": true, "echo": true, "type": true,
		"set": true, "if": true, "for": true, "call": true,
	}
	return shellBuiltins[strings.ToLower(command)]
}

// executeViaShell executes a command via the system shell
func executeViaShell(dir, command string, args []string) error {
	var shellCmd []string

	if runtime.GOOS == "windows" {
		shellCmd = []string{"cmd", "/c"}
	} else {
		shellCmd = []string{"/bin/sh", "-c"}
	}

	// Build the full command string
	fullCommand := command
	if len(args) > 0 {
		fullCommand += " " + strings.Join(args, " ")
	}

	shellCmd = append(shellCmd, fullCommand)

	cmd := exec.Command(shellCmd[0], shellCmd[1:]...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

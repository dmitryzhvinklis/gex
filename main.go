package main

import (
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"strings"
	"syscall"

	"gex/internal/cli"
	"gex/internal/config"
	"gex/internal/core"
	"gex/internal/executor"
	"gex/internal/readline"
	"gex/internal/shell"
	"gex/internal/ui"
)

const (
	// Shell version
	VERSION = "1.0.0"
	// Shell name
	SHELL_NAME = "gex"
)

func main() {
	// Initialize signal handling
	setupSignalHandling()

	// Initialize configuration
	cfg := config.New()

	// Initialize shell components
	session := shell.NewSession(cfg)
	executor := executor.New(session)
	reader := readline.New(session)

	// Initialize command pool for performance
	core.InitializePool()

	// Initialize color config
	colorConfig := ui.DefaultColorConfig()

	// Print welcome message
	printWelcome()

	// Get user info for prompt
	currentUser, _ := user.Current()
	hostname, _ := os.Hostname()
	username := "user"
	if currentUser != nil {
		username = currentUser.Username
	}

	// Main REPL loop
	for {
		// Create dynamic colorful prompt
		cwd, _ := os.Getwd()
		prompt := colorConfig.FormatPrompt(username, hostname, cwd, SHELL_NAME)
		reader.SetPrompt(prompt)

		// Read input with readline support
		input, err := reader.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("\nExiting...")
				break
			}
			continue
		}

		// Skip empty lines
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Add to history
		session.AddHistory(input)

		// Parse and execute command
		cmd, err := cli.Parse(input)
		if err != nil {
			ui.PrintError(fmt.Sprintf("Parse error: %v", err))
			continue
		}

		// Execute command
		if err := executor.Execute(cmd); err != nil {
			if err.Error() == "exit" {
				break
			}
			ui.PrintError(fmt.Sprintf("%v", err))
		}
	}
}

func setupSignalHandling() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nInterrupt received, exiting...")
		os.Exit(0)
	}()
}

func printWelcome() {
	ui.PrintWelcome(SHELL_NAME, VERSION)
}

package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gex/internal/cli"
	"gex/internal/config"
	"gex/internal/core"
	"gex/internal/executor"
	"gex/internal/readline"
	"gex/internal/shell"
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

	// Print welcome message
	printWelcome()

	// Main REPL loop
	for {
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
			fmt.Printf("Parse error: %v\n", err)
			continue
		}

		// Execute command
		if err := executor.Execute(cmd); err != nil {
			if err.Error() == "exit" {
				break
			}
			fmt.Printf("Error: %v\n", err)
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
	fmt.Printf("Welcome to %s v%s - High-Performance Linux Shell\n", SHELL_NAME, VERSION)
	fmt.Println("Type 'help' for available commands")
	fmt.Println()
}

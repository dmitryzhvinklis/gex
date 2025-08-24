package executor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"gex/internal/builtin"
	"gex/internal/cli"
	"gex/internal/shell"
)

// Executor handles command execution with high performance
type Executor struct {
	session *shell.Session
	mutex   sync.RWMutex
}

// New creates a new executor instance
func New(session *shell.Session) *Executor {
	return &Executor{
		session: session,
	}
}

// Execute executes a parsed command
func (e *Executor) Execute(cmd *cli.Command) error {
	if cmd == nil {
		return errors.New("nil command")
	}

	// Handle pipes
	if len(cmd.Pipes) > 0 {
		return e.executePipeline(cmd)
	}

	// Handle single command
	return e.executeSingle(cmd)
}

// executeSingle executes a single command
func (e *Executor) executeSingle(cmd *cli.Command) error {
	// Expand aliases
	cli.ExpandAliases(cmd, e.session.GetAliases())

	// Check if it's a built-in command
	if cli.IsBuiltin(cmd.Name) {
		return e.executeBuiltin(cmd)
	}

	// Execute external command
	return e.executeExternal(cmd)
}

// executeBuiltin executes a built-in command
func (e *Executor) executeBuiltin(cmd *cli.Command) error {
	switch cmd.Name {
	case "cd":
		return builtin.Cd(cmd.Args, e.session)
	case "pwd":
		return builtin.Pwd(cmd.Args)
	case "echo":
		return builtin.Echo(cmd.Args)
	case "exit":
		return builtin.Exit(cmd.Args)
	case "help":
		return builtin.Help(cmd.Args)
	case "history":
		return builtin.History(cmd.Args, e.session)
	case "alias":
		return builtin.Alias(cmd.Args, e.session)
	case "unalias":
		return builtin.Unalias(cmd.Args, e.session)
	case "env":
		return builtin.Env(cmd.Args)
	case "export":
		return builtin.Export(cmd.Args)
	case "which":
		return builtin.Which(cmd.Args)
	case "type":
		return builtin.Type(cmd.Args, e.session)
	default:
		return fmt.Errorf("unknown built-in command: %s", cmd.Name)
	}
}

// executeExternal executes an external command
func (e *Executor) executeExternal(cmd *cli.Command) error {
	// Find the executable
	execPath, err := e.findExecutable(cmd.Name)
	if err != nil {
		return fmt.Errorf("command not found: %s", cmd.Name)
	}

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the command
	execCmd := exec.CommandContext(ctx, execPath, cmd.Args...)

	// Set environment
	execCmd.Env = os.Environ()

	// Set working directory
	execCmd.Dir = e.session.GetWorkingDir()

	// Handle redirections
	if err := e.setupRedirections(execCmd, cmd.Redirect); err != nil {
		return err
	}

	// Execute command
	if cmd.Background {
		return e.executeBackground(execCmd)
	}

	return e.executeForeground(execCmd)
}

// executeForeground executes a command in the foreground
func (e *Executor) executeForeground(cmd *exec.Cmd) error {
	// Set up default I/O if not redirected
	if cmd.Stdin == nil {
		cmd.Stdin = os.Stdin
	}
	if cmd.Stdout == nil {
		cmd.Stdout = os.Stdout
	}
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Wait for completion
	return cmd.Wait()
}

// executeBackground executes a command in the background
func (e *Executor) executeBackground(cmd *exec.Cmd) error {
	// Set up default I/O for background processes
	if cmd.Stdin == nil {
		cmd.Stdin = nil // No stdin for background processes
	}
	if cmd.Stdout == nil {
		cmd.Stdout = os.Stdout
	}
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	fmt.Printf("[%d] %d\n", 1, cmd.Process.Pid) // Job number and PID

	// Don't wait - let it run in background
	go func() {
		cmd.Wait()
		fmt.Printf("[%d] Done\n", 1)
	}()

	return nil
}

// executePipeline executes a pipeline of commands
func (e *Executor) executePipeline(cmd *cli.Command) error {
	commands := []*cli.Command{cmd}
	commands = append(commands, cmd.Pipes...)

	// Create pipes between commands
	var pipes []*os.File
	var readers []*os.File

	for i := 0; i < len(commands)-1; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			return err
		}
		pipes = append(pipes, w)
		readers = append(readers, r)
	}

	// Defer closing all pipes
	defer func() {
		for _, p := range pipes {
			p.Close()
		}
		for _, r := range readers {
			r.Close()
		}
	}()

	// Start all commands
	var cmds []*exec.Cmd
	var wg sync.WaitGroup

	for i, command := range commands {
		// Handle built-in commands in pipeline
		if cli.IsBuiltin(command.Name) {
			return errors.New("built-in commands in pipelines not yet supported")
		}

		execPath, err := e.findExecutable(command.Name)
		if err != nil {
			return fmt.Errorf("command not found: %s", command.Name)
		}

		execCmd := exec.Command(execPath, command.Args...)
		execCmd.Env = os.Environ()
		execCmd.Dir = e.session.GetWorkingDir()

		// Set up stdin
		if i == 0 {
			// First command - stdin from terminal
			execCmd.Stdin = os.Stdin
		} else {
			// Middle/last commands - stdin from previous pipe
			execCmd.Stdin = readers[i-1]
		}

		// Set up stdout
		if i == len(commands)-1 {
			// Last command - stdout to terminal
			execCmd.Stdout = os.Stdout
		} else {
			// First/middle commands - stdout to next pipe
			execCmd.Stdout = pipes[i]
		}

		execCmd.Stderr = os.Stderr
		cmds = append(cmds, execCmd)
	}

	// Start all commands
	for _, execCmd := range cmds {
		if err := execCmd.Start(); err != nil {
			return err
		}
		wg.Add(1)
		go func(c *exec.Cmd) {
			defer wg.Done()
			c.Wait()
		}(execCmd)
	}

	// Wait for all commands to complete
	wg.Wait()

	return nil
}

// setupRedirections sets up input/output redirections
func (e *Executor) setupRedirections(cmd *exec.Cmd, redirect *cli.Redirect) error {
	if redirect == nil {
		return nil
	}

	switch redirect.Type {
	case cli.RedirectOut:
		file, err := os.Create(redirect.Target)
		if err != nil {
			return err
		}
		cmd.Stdout = file

	case cli.RedirectAppend:
		file, err := os.OpenFile(redirect.Target, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		cmd.Stdout = file

	case cli.RedirectIn:
		file, err := os.Open(redirect.Target)
		if err != nil {
			return err
		}
		cmd.Stdin = file

	case cli.RedirectErr:
		file, err := os.Create(redirect.Target)
		if err != nil {
			return err
		}
		cmd.Stderr = file

	case cli.RedirectBoth:
		file, err := os.Create(redirect.Target)
		if err != nil {
			return err
		}
		cmd.Stdout = file
		cmd.Stderr = file
	}

	return nil
}

// findExecutable finds an executable in PATH
func (e *Executor) findExecutable(name string) (string, error) {
	// If it's an absolute or relative path, check directly
	if strings.Contains(name, "/") {
		if filepath.IsAbs(name) {
			if e.isExecutable(name) {
				return name, nil
			}
		} else {
			// Relative path
			fullPath := filepath.Join(e.session.GetWorkingDir(), name)
			if e.isExecutable(fullPath) {
				return fullPath, nil
			}
		}
		return "", errors.New("not found")
	}

	// Search in PATH
	path := os.Getenv("PATH")
	if path == "" {
		path = "/usr/local/bin:/usr/bin:/bin"
	}

	for _, dir := range strings.Split(path, ":") {
		if dir == "" {
			continue
		}

		fullPath := filepath.Join(dir, name)
		if e.isExecutable(fullPath) {
			return fullPath, nil
		}
	}

	return "", errors.New("not found")
}

// isExecutable checks if a file is executable
func (e *Executor) isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check if it's a regular file and executable
	if !info.Mode().IsRegular() {
		return false
	}

	// Check execute permission
	return info.Mode()&0111 != 0
}

// InterruptRunning interrupts any running foreground process
func (e *Executor) InterruptRunning() error {
	// This would be implemented with process group management
	// For now, just return nil
	return nil
}

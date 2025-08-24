package builtin

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gex/internal/cli"
	"gex/internal/shell"
)

// Cd changes the current working directory
func Cd(args []string, session *shell.Session) error {
	var target string

	if len(args) == 0 {
		// No arguments - go to home directory
		home := os.Getenv("HOME")
		if home == "" {
			return fmt.Errorf("HOME environment variable not set")
		}
		target = home
	} else {
		target = args[0]
	}

	// Handle special cases
	if target == "-" {
		// Go to previous directory
		prev := session.GetPreviousDir()
		if prev == "" {
			return fmt.Errorf("no previous directory")
		}
		target = prev
		fmt.Println(target) // Print the directory we're going to
	}

	// Expand ~ to home directory
	if strings.HasPrefix(target, "~/") {
		home := os.Getenv("HOME")
		if home == "" {
			return fmt.Errorf("HOME environment variable not set")
		}
		target = home + target[1:]
	}

	// Change directory
	if err := os.Chdir(target); err != nil {
		return err
	}

	// Update session state
	oldDir := session.GetWorkingDir()
	newDir, err := os.Getwd()
	if err != nil {
		return err
	}

	session.SetWorkingDir(newDir)
	session.SetPreviousDir(oldDir)

	return nil
}

// Pwd prints the current working directory
func Pwd(args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println(wd)
	return nil
}

// Echo displays text
func Echo(args []string) error {
	output := strings.Join(args, " ")
	fmt.Println(output)
	return nil
}

// Exit exits the shell
func Exit(args []string) error {
	code := 0
	if len(args) > 0 {
		if c, err := strconv.Atoi(args[0]); err == nil {
			code = c
		}
	}

	if code != 0 {
		os.Exit(code)
	}

	return fmt.Errorf("exit")
}

// Help displays help information
func Help(args []string) error {
	if len(args) == 0 {
		// General help
		fmt.Println("Gex Shell - High-Performance Linux Shell")
		fmt.Println()
		fmt.Println("Built-in commands:")

		builtins := cli.GetAllBuiltins()
		for name, info := range builtins {
			fmt.Printf("  %-12s %s\n", name, info.Description)
		}

		fmt.Println()
		fmt.Println("Use 'help <command>' for specific command help")
		return nil
	}

	// Specific command help
	cmdName := args[0]
	info := cli.GetCommandInfo(cmdName)

	fmt.Printf("Command: %s\n", info.Name)
	fmt.Printf("Description: %s\n", info.Description)
	fmt.Printf("Usage: %s\n", info.Usage)

	return nil
}

// History displays command history
func History(args []string, session *shell.Session) error {
	history := session.GetHistory()

	limit := len(history)
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
			limit = n
		}
	}

	start := 0
	if len(history) > limit {
		start = len(history) - limit
	}

	for i := start; i < len(history); i++ {
		fmt.Printf("%4d  %s\n", i+1, history[i])
	}

	return nil
}

// Alias manages command aliases
func Alias(args []string, session *shell.Session) error {
	if len(args) == 0 {
		// Display all aliases
		aliases := session.GetAliases()
		for name, value := range aliases {
			fmt.Printf("%s='%s'\n", name, value)
		}
		return nil
	}

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			// Set alias
			parts := strings.SplitN(arg, "=", 2)
			name := parts[0]
			value := parts[1]

			// Remove quotes if present
			if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}

			session.SetAlias(name, value)
		} else {
			// Display specific alias
			if value, exists := session.GetAliases()[arg]; exists {
				fmt.Printf("%s='%s'\n", arg, value)
			} else {
				fmt.Printf("alias: %s: not found\n", arg)
			}
		}
	}

	return nil
}

// Unalias removes aliases
func Unalias(args []string, session *shell.Session) error {
	if len(args) == 0 {
		return fmt.Errorf("unalias: usage: unalias name [name ...]")
	}

	for _, name := range args {
		session.RemoveAlias(name)
	}

	return nil
}

// Env displays or sets environment variables
func Env(args []string) error {
	if len(args) == 0 {
		// Display all environment variables
		for _, env := range os.Environ() {
			fmt.Println(env)
		}
		return nil
	}

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			// Set environment variable
			parts := strings.SplitN(arg, "=", 2)
			os.Setenv(parts[0], parts[1])
		} else {
			// Display specific variable
			if value := os.Getenv(arg); value != "" {
				fmt.Printf("%s=%s\n", arg, value)
			}
		}
	}

	return nil
}

// Export exports environment variables
func Export(args []string) error {
	if len(args) == 0 {
		// Display all exported variables (same as env for now)
		return Env(args)
	}

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			// Set and export
			parts := strings.SplitN(arg, "=", 2)
			os.Setenv(parts[0], parts[1])
		} else {
			// Export existing variable
			if value := os.Getenv(arg); value != "" {
				os.Setenv(arg, value)
			}
		}
	}

	return nil
}

// Which locates a command
func Which(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("which: usage: which command [command ...]")
	}

	path := os.Getenv("PATH")
	if path == "" {
		path = "/usr/local/bin:/usr/bin:/bin"
	}

	for _, cmd := range args {
		found := false

		// Check if it's a built-in
		if cli.IsBuiltin(cmd) {
			fmt.Printf("%s: shell built-in command\n", cmd)
			found = true
			continue
		}

		// Search in PATH
		for _, dir := range strings.Split(path, ":") {
			if dir == "" {
				continue
			}

			fullPath := dir + "/" + cmd
			if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
				// Check if executable
				if info.Mode()&0111 != 0 {
					fmt.Println(fullPath)
					found = true
					break
				}
			}
		}

		if !found {
			fmt.Printf("%s not found\n", cmd)
		}
	}

	return nil
}

// Type displays information about command type
func Type(args []string, session *shell.Session) error {
	if len(args) == 0 {
		return fmt.Errorf("type: usage: type command [command ...]")
	}

	for _, cmd := range args {
		// Check aliases first
		if alias, exists := session.GetAliases()[cmd]; exists {
			fmt.Printf("%s is aliased to `%s'\n", cmd, alias)
			continue
		}

		// Check built-ins
		if cli.IsBuiltin(cmd) {
			fmt.Printf("%s is a shell builtin\n", cmd)
			continue
		}

		// Check PATH
		path := os.Getenv("PATH")
		if path == "" {
			path = "/usr/local/bin:/usr/bin:/bin"
		}

		found := false
		for _, dir := range strings.Split(path, ":") {
			if dir == "" {
				continue
			}

			fullPath := dir + "/" + cmd
			if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
				if info.Mode()&0111 != 0 {
					fmt.Printf("%s is %s\n", cmd, fullPath)
					found = true
					break
				}
			}
		}

		if !found {
			fmt.Printf("%s: not found\n", cmd)
		}
	}

	return nil
}

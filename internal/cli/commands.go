package cli

import (
	"strings"
)

// CommandType represents the type of command
type CommandType int

const (
	CommandBuiltin CommandType = iota
	CommandExternal
	CommandAlias
)

// CommandInfo contains metadata about a command
type CommandInfo struct {
	Name        string
	Type        CommandType
	Description string
	Usage       string
}

// IsBuiltin checks if a command is a built-in command
func IsBuiltin(name string) bool {
	_, exists := builtinCommands[name]
	return exists
}

// GetCommandInfo returns information about a command
func GetCommandInfo(name string) *CommandInfo {
	if info, exists := builtinCommands[name]; exists {
		return info
	}

	// For external commands, return basic info
	return &CommandInfo{
		Name:        name,
		Type:        CommandExternal,
		Description: "External command",
		Usage:       name + " [options] [arguments]",
	}
}

// GetAllBuiltins returns all built-in commands
func GetAllBuiltins() map[string]*CommandInfo {
	result := make(map[string]*CommandInfo)
	for name, info := range builtinCommands {
		result[name] = info
	}
	return result
}

// Built-in commands registry
var builtinCommands = map[string]*CommandInfo{
	"cd": {
		Name:        "cd",
		Type:        CommandBuiltin,
		Description: "Change the current directory",
		Usage:       "cd [directory]",
	},
	"pwd": {
		Name:        "pwd",
		Type:        CommandBuiltin,
		Description: "Print the current working directory",
		Usage:       "pwd",
	},
	"echo": {
		Name:        "echo",
		Type:        CommandBuiltin,
		Description: "Display a line of text",
		Usage:       "echo [text...]",
	},
	"exit": {
		Name:        "exit",
		Type:        CommandBuiltin,
		Description: "Exit the shell",
		Usage:       "exit [code]",
	},
	"help": {
		Name:        "help",
		Type:        CommandBuiltin,
		Description: "Display help information",
		Usage:       "help [command]",
	},
	"history": {
		Name:        "history",
		Type:        CommandBuiltin,
		Description: "Display command history",
		Usage:       "history [n]",
	},
	"alias": {
		Name:        "alias",
		Type:        CommandBuiltin,
		Description: "Create or display aliases",
		Usage:       "alias [name[=value]...]",
	},
	"unalias": {
		Name:        "unalias",
		Type:        CommandBuiltin,
		Description: "Remove aliases",
		Usage:       "unalias name...",
	},
	"env": {
		Name:        "env",
		Type:        CommandBuiltin,
		Description: "Display or set environment variables",
		Usage:       "env [name[=value]...]",
	},
	"export": {
		Name:        "export",
		Type:        CommandBuiltin,
		Description: "Export environment variables",
		Usage:       "export name[=value]...",
	},
	"which": {
		Name:        "which",
		Type:        CommandBuiltin,
		Description: "Locate a command",
		Usage:       "which command...",
	},
	"type": {
		Name:        "type",
		Type:        CommandBuiltin,
		Description: "Display information about command type",
		Usage:       "type command...",
	},
}

// ExpandAliases expands aliases in the command
func ExpandAliases(cmd *Command, aliases map[string]string) {
	if alias, exists := aliases[cmd.Name]; exists {
		// Simple alias expansion - split on whitespace
		parts := strings.Fields(alias)
		if len(parts) > 0 {
			cmd.Name = parts[0]
			if len(parts) > 1 {
				// Prepend alias arguments to existing arguments
				cmd.Args = append(parts[1:], cmd.Args...)
			}
		}
	}
}

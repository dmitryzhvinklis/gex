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
	// Basic shell commands
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

	// File operations
	"ls": {
		Name:        "ls",
		Type:        CommandBuiltin,
		Description: "List directory contents",
		Usage:       "ls [options] [files...]",
	},
	"mkdir": {
		Name:        "mkdir",
		Type:        CommandBuiltin,
		Description: "Create directories",
		Usage:       "mkdir [options] directory...",
	},
	"rmdir": {
		Name:        "rmdir",
		Type:        CommandBuiltin,
		Description: "Remove empty directories",
		Usage:       "rmdir directory...",
	},
	"rm": {
		Name:        "rm",
		Type:        CommandBuiltin,
		Description: "Remove files and directories",
		Usage:       "rm [options] file...",
	},
	"cp": {
		Name:        "cp",
		Type:        CommandBuiltin,
		Description: "Copy files and directories",
		Usage:       "cp [options] source... destination",
	},
	"mv": {
		Name:        "mv",
		Type:        CommandBuiltin,
		Description: "Move/rename files and directories",
		Usage:       "mv source... destination",
	},
	"touch": {
		Name:        "touch",
		Type:        CommandBuiltin,
		Description: "Create empty files or update timestamps",
		Usage:       "touch file...",
	},

	// Text operations
	"cat": {
		Name:        "cat",
		Type:        CommandBuiltin,
		Description: "Display file contents",
		Usage:       "cat [file...]",
	},
	"head": {
		Name:        "head",
		Type:        CommandBuiltin,
		Description: "Display first lines of files",
		Usage:       "head [options] [file...]",
	},
	"tail": {
		Name:        "tail",
		Type:        CommandBuiltin,
		Description: "Display last lines of files",
		Usage:       "tail [options] [file...]",
	},
	"wc": {
		Name:        "wc",
		Type:        CommandBuiltin,
		Description: "Count lines, words, and characters",
		Usage:       "wc [options] [file...]",
	},
	"grep": {
		Name:        "grep",
		Type:        CommandBuiltin,
		Description: "Search for patterns in files",
		Usage:       "grep [options] pattern [file...]",
	},
	"sort": {
		Name:        "sort",
		Type:        CommandBuiltin,
		Description: "Sort lines in files",
		Usage:       "sort [options] [file...]",
	},

	// System operations
	"ps": {
		Name:        "ps",
		Type:        CommandBuiltin,
		Description: "Display running processes",
		Usage:       "ps [options]",
	},
	"kill": {
		Name:        "kill",
		Type:        CommandBuiltin,
		Description: "Send signals to processes",
		Usage:       "kill [signal] pid...",
	},
	"df": {
		Name:        "df",
		Type:        CommandBuiltin,
		Description: "Display filesystem disk space usage",
		Usage:       "df [options] [filesystem...]",
	},
	"du": {
		Name:        "du",
		Type:        CommandBuiltin,
		Description: "Display directory disk usage",
		Usage:       "du [options] [directory...]",
	},
	"free": {
		Name:        "free",
		Type:        CommandBuiltin,
		Description: "Display memory usage",
		Usage:       "free [options]",
	},
	"uptime": {
		Name:        "uptime",
		Type:        CommandBuiltin,
		Description: "Display system uptime",
		Usage:       "uptime",
	},
	"uname": {
		Name:        "uname",
		Type:        CommandBuiltin,
		Description: "Display system information",
		Usage:       "uname [options]",
	},

	// Search operations
	"find": {
		Name:        "find",
		Type:        CommandBuiltin,
		Description: "Search for files and directories",
		Usage:       "find [path] [options]",
	},
	"locate": {
		Name:        "locate",
		Type:        CommandBuiltin,
		Description: "Find files by name in database",
		Usage:       "locate pattern",
	},

	// Permission operations
	"chmod": {
		Name:        "chmod",
		Type:        CommandBuiltin,
		Description: "Change file permissions",
		Usage:       "chmod [options] mode file...",
	},
	"chown": {
		Name:        "chown",
		Type:        CommandBuiltin,
		Description: "Change file ownership",
		Usage:       "chown [options] owner[:group] file...",
	},
	"chgrp": {
		Name:        "chgrp",
		Type:        CommandBuiltin,
		Description: "Change group ownership",
		Usage:       "chgrp [options] group file...",
	},

	// Network operations
	"ping": {
		Name:        "ping",
		Type:        CommandBuiltin,
		Description: "Send ICMP echo requests",
		Usage:       "ping [options] host",
	},
	"wget": {
		Name:        "wget",
		Type:        CommandBuiltin,
		Description: "Download files from web",
		Usage:       "wget [options] URL",
	},
	"curl": {
		Name:        "curl",
		Type:        CommandBuiltin,
		Description: "Transfer data from/to servers",
		Usage:       "curl [options] URL",
	},
	"netstat": {
		Name:        "netstat",
		Type:        CommandBuiltin,
		Description: "Display network connections",
		Usage:       "netstat [options]",
	},

	// Archive operations
	"tar": {
		Name:        "tar",
		Type:        CommandBuiltin,
		Description: "Archive files",
		Usage:       "tar [options] archive files...",
	},
	"gzip": {
		Name:        "gzip",
		Type:        CommandBuiltin,
		Description: "Compress files",
		Usage:       "gzip [options] file...",
	},
	"gunzip": {
		Name:        "gunzip",
		Type:        CommandBuiltin,
		Description: "Decompress gzip files",
		Usage:       "gunzip [options] file...",
	},
	"zip": {
		Name:        "zip",
		Type:        CommandBuiltin,
		Description: "Create zip archives",
		Usage:       "zip [options] archive files...",
	},
	"unzip": {
		Name:        "unzip",
		Type:        CommandBuiltin,
		Description: "Extract zip archives",
		Usage:       "unzip [options] archive",
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

# Gex Shell

A Linux shell written in Go

## Quick Start

### Installation

```bash
# Quick install (recommended)
curl -fsSL https://raw.githubusercontent.com/dmitryzhvinklis/gex/main/install.sh | bash

# Or manual installation
git clone https://github.com/dmitryzhvinklis/gex.git
cd gex
make install
```

### Usage

```bash
# Start the shell
gex

# Or set as default shell
chsh -s $(which gex)
```

## Built-in Commands

| Command | Description |
|---------|-------------|
| `cd [dir]` | Change directory |
| `pwd` | Print working directory |
| `echo [text]` | Display text |
| `exit [code]` | Exit shell |
| `help [cmd]` | Show help |
| `history [n]` | Show command history |
| `alias [name=value]` | Manage aliases |
| `unalias [name]` | Remove aliases |
| `env [var=value]` | Environment variables |
| `export [var=value]` | Export variables |
| `which [cmd]` | Locate command |
| `type [cmd]` | Command type info |

## Advanced Features

### Command Line Editing

- **Arrow Keys**: Navigate through command history
- **Ctrl+A**: Beginning of line
- **Ctrl+E**: End of line  
- **Ctrl+L**: Clear screen
- **Ctrl+C**: Cancel current command
- **Ctrl+D**: Exit shell or delete character
- **Tab**: Auto-completion

### Pipes and Redirection

```bash
# Pipes
ls -la | grep txt | sort

# Output redirection
echo "hello" > file.txt
echo "world" >> file.txt

# Input redirection
sort < file.txt

# Error redirection
command 2> error.log

# Combined redirection
command &> output.log
```

### Background Jobs

```bash
# Run in background
long_running_command &

# Job control (basic)
jobs
fg %1
```

### Aliases

```bash
# Create aliases
alias ll='ls -la'
alias grep='grep --color=auto'

# List aliases
alias

# Remove aliases
unalias ll
```

## Configuration

Configuration file: `~/.config/gex/config.json`

```json
{
  "history_limit": 1000,
  "prompt": "gex> ",
  "aliases": {
    "ll": "ls -la",
    "la": "ls -la"
  },
  "auto_complete": true,
  "color_output": true,
  "tab_completion": true,
  "history_search": true,
  "case_sensitive": false,
  "max_jobs": 10,
  "timeout_seconds": 30
}
```

### Benchmarks

```bash
# Run performance benchmarks
make benchmark

# Profile CPU usage
make profile

# Profile memory usage  
make memprofile
```

## Building from Source

### Requirements

- Go 1.21 or later
- Linux (tested on Ubuntu, Debian, CentOS, Arch)

## License

MIT License - see LICENSE file for details.

---
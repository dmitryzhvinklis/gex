#!/bin/bash

# Gex Shell Installer
# High-performance Linux shell written in Go

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GEX_VERSION="1.0.0"
GEX_REPO="git@github.com:dmitryzhvinklis/gex.git"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/gex"
SHELL_CONFIG_FILES=("$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile")

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_requirements() {
    log_info "Checking system requirements..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21 or later."
        log_info "Visit: https://golang.org/doc/install"
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | grep -oP 'go\d+\.\d+' | sed 's/go//')
    if [ "$(printf '%s\n' "1.21" "$GO_VERSION" | sort -V | head -n1)" != "1.21" ]; then
        log_error "Go version 1.21 or later is required. Found: $GO_VERSION"
        exit 1
    fi
    
    # Check if git is installed
    if ! command -v git &> /dev/null; then
        log_error "Git is not installed. Please install git."
        exit 1
    fi
    
    log_success "System requirements satisfied"
}

install_gex() {
    log_info "Installing Gex shell..."
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # Clone or download source code
    if [ -n "$GEX_LOCAL_PATH" ]; then
        log_info "Using local source code from: $GEX_LOCAL_PATH"
        cp -r "$GEX_LOCAL_PATH" ./gex
    elif [ -f "$(dirname "$0")/go.mod" ]; then
        log_info "Using local source code..."
        cp -r "$(dirname "$0")" ./gex
    else
        log_info "Cloning from repository..."
        git clone "$GEX_REPO" gex
    fi
    
    cd gex
    
    # Build the binary
    log_info "Building Gex..."
    go mod tidy
    go build -ldflags="-s -w" -o gex .
    
    # Install binary
    log_info "Installing binary to $INSTALL_DIR..."
    sudo cp gex "$INSTALL_DIR/gex"
    sudo chmod +x "$INSTALL_DIR/gex"
    
    # Clean up
    cd /
    rm -rf "$TEMP_DIR"
    
    log_success "Gex binary installed successfully"
}

setup_config() {
    log_info "Setting up configuration..."
    
    # Create config directory
    mkdir -p "$CONFIG_DIR"
    
    # Create default configuration
    cat > "$CONFIG_DIR/config.json" << EOF
{
  "history_limit": 1000,
  "prompt": "gex> ",
  "aliases": {
    "ll": "ls -la",
    "la": "ls -la",
    "l": "ls -l",
    "..": "cd ..",
    "...": "cd ../..",
    "grep": "grep --color=auto",
    "fgrep": "fgrep --color=auto",
    "egrep": "egrep --color=auto"
  },
  "auto_complete": true,
  "color_output": true,
  "tab_completion": true,
  "history_search": true,
  "case_sensitive": false,
  "max_jobs": 10,
  "timeout_seconds": 30
}
EOF
    
    log_success "Configuration created at $CONFIG_DIR/config.json"
}

setup_shell_integration() {
    log_info "Setting up shell integration..."
    
    # Add gex to PATH if not already there
    if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
        for config_file in "${SHELL_CONFIG_FILES[@]}"; do
            if [ -f "$config_file" ]; then
                if ! grep -q "export PATH.*$INSTALL_DIR" "$config_file"; then
                    echo "" >> "$config_file"
                    echo "# Gex shell integration" >> "$config_file"
                    echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$config_file"
                    log_info "Added $INSTALL_DIR to PATH in $config_file"
                fi
            fi
        done
    fi
    
    # Create gex startup script
    cat > "$CONFIG_DIR/startup.sh" << 'EOF'
#!/bin/bash
# Gex shell startup script

# Set up environment
export GEX_CONFIG_DIR="$HOME/.config/gex"

# Load custom aliases and functions
if [ -f "$GEX_CONFIG_DIR/aliases.sh" ]; then
    source "$GEX_CONFIG_DIR/aliases.sh"
fi

# Load custom functions
if [ -f "$GEX_CONFIG_DIR/functions.sh" ]; then
    source "$GEX_CONFIG_DIR/functions.sh"
fi
EOF
    
    chmod +x "$CONFIG_DIR/startup.sh"
    
    log_success "Shell integration configured"
}

create_desktop_entry() {
    log_info "Creating desktop entry..."
    
    DESKTOP_DIR="$HOME/.local/share/applications"
    mkdir -p "$DESKTOP_DIR"
    
    cat > "$DESKTOP_DIR/gex.desktop" << EOF
[Desktop Entry]
Name=Gex Shell
Comment=High-Performance Linux Shell
Exec=gnome-terminal -e gex
Icon=terminal
Type=Application
Categories=System;TerminalEmulator;
Keywords=shell;terminal;command;
EOF
    
    log_success "Desktop entry created"
}

setup_completions() {
    log_info "Setting up shell completions..."
    
    # Create completion scripts for different shells
    mkdir -p "$CONFIG_DIR/completions"
    
    # Bash completion
    cat > "$CONFIG_DIR/completions/gex.bash" << 'EOF'
# Gex shell completion for bash
_gex_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # Basic command completion
    opts="cd pwd echo exit help history alias unalias env export which type"
    
    if [[ ${cur} == -* ]]; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi
    
    # File completion for most commands
    COMPREPLY=( $(compgen -f -- ${cur}) )
}

complete -F _gex_completion gex
EOF
    
    # Zsh completion
    cat > "$CONFIG_DIR/completions/_gex" << 'EOF'
#compdef gex
# Gex shell completion for zsh

_gex() {
    local context state line
    
    _arguments -C \
        '1:command:_gex_commands' \
        '*:file:_files'
}

_gex_commands() {
    local commands
    commands=(
        'cd:Change directory'
        'pwd:Print working directory'
        'echo:Display text'
        'exit:Exit shell'
        'help:Show help'
        'history:Show command history'
        'alias:Manage aliases'
        'unalias:Remove aliases'
        'env:Environment variables'
        'export:Export variables'
        'which:Locate command'
        'type:Command type info'
    )
    
    _describe 'commands' commands
}

_gex "$@"
EOF
    
    log_success "Shell completions configured"
}

verify_installation() {
    log_info "Verifying installation..."
    
    # Check if gex is in PATH
    if ! command -v gex &> /dev/null; then
        log_error "Gex is not in PATH. Please restart your terminal or run: source ~/.bashrc"
        return 1
    fi
    
    # Check if gex runs
    if ! gex --version &> /dev/null; then
        log_error "Gex binary is not working correctly"
        return 1
    fi
    
    log_success "Installation verified successfully"
}

show_post_install() {
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                    Installation Complete!                    ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}Gex Shell v$GEX_VERSION has been successfully installed!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Restart your terminal or run: source ~/.bashrc"
    echo "  2. Start Gex by typing: gex"
    echo "  3. Type 'help' in Gex to see available commands"
    echo "  4. Edit ~/.config/gex/config.json to customize settings"
    echo ""
    echo "Configuration files:"
    echo "  • Config: ~/.config/gex/config.json"
    echo "  • Startup: ~/.config/gex/startup.sh"
    echo "  • Completions: ~/.config/gex/completions/"
    echo ""
    echo "For more information, visit: $GEX_REPO"
    echo ""
}

uninstall_gex() {
    log_info "Uninstalling Gex..."
    
    # Remove binary
    if [ -f "$INSTALL_DIR/gex" ]; then
        sudo rm "$INSTALL_DIR/gex"
        log_success "Removed gex binary"
    fi
    
    # Remove configuration (ask user)
    read -p "Remove configuration files? [y/N]: " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$CONFIG_DIR"
        log_success "Removed configuration files"
    fi
    
    # Remove desktop entry
    if [ -f "$HOME/.local/share/applications/gex.desktop" ]; then
        rm "$HOME/.local/share/applications/gex.desktop"
        log_success "Removed desktop entry"
    fi
    
    log_success "Gex uninstalled successfully"
}

# Main installation flow
main() {
    case "${1:-install}" in
        "install")
            echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
            echo -e "${BLUE}║                     Gex Shell Installer                      ║${NC}"
            echo -e "${BLUE}║                   v$GEX_VERSION - High Performance                   ║${NC}"
            echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
            echo ""
            
            check_requirements
            install_gex
            setup_config
            setup_shell_integration
            create_desktop_entry
            setup_completions
            verify_installation
            show_post_install
            ;;
        "uninstall")
            uninstall_gex
            ;;
        "update")
            log_info "Updating Gex..."
            install_gex
            log_success "Gex updated successfully"
            ;;
        *)
            echo "Usage: $0 [install|uninstall|update]"
            echo ""
            echo "Commands:"
            echo "  install   - Install Gex shell (default)"
            echo "  uninstall - Remove Gex shell"
            echo "  update    - Update Gex to latest version"
            exit 1
            ;;
    esac
}

# Run installer
main "$@"

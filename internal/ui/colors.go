package ui

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ANSI color codes
const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
	Dim   = "\033[2m"

	// Regular colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bright colors
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// Background colors
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

// Rainbow colors for dynamic prompt
var RainbowColors = []string{
	Red, Green, Yellow, Blue, Magenta, Cyan,
	BrightRed, BrightGreen, BrightYellow, BrightBlue, BrightMagenta, BrightCyan,
}

// File type colors
var FileTypeColors = map[string]string{
	// Directories
	"directory": Bold + Blue,

	// Executables
	"executable": Bold + Green,

	// Archives
	".tar": Red,
	".gz":  Red,
	".zip": Red,
	".rar": Red,
	".7z":  Red,
	".bz2": Red,
	".xz":  Red,

	// Images
	".jpg":  Magenta,
	".jpeg": Magenta,
	".png":  Magenta,
	".gif":  Magenta,
	".bmp":  Magenta,
	".svg":  Magenta,
	".ico":  Magenta,

	// Videos
	".mp4": BrightMagenta,
	".avi": BrightMagenta,
	".mkv": BrightMagenta,
	".mov": BrightMagenta,
	".wmv": BrightMagenta,

	// Audio
	".mp3":  Cyan,
	".wav":  Cyan,
	".flac": Cyan,
	".ogg":  Cyan,

	// Documents
	".pdf":  BrightRed,
	".doc":  BrightRed,
	".docx": BrightRed,
	".txt":  White,
	".md":   BrightWhite,
	".rtf":  BrightRed,

	// Code files
	".go":   BrightCyan,
	".py":   BrightYellow,
	".js":   BrightYellow,
	".ts":   BrightBlue,
	".html": BrightRed,
	".css":  BrightMagenta,
	".java": BrightRed,
	".c":    BrightBlue,
	".cpp":  BrightBlue,
	".h":    BrightCyan,
	".rs":   BrightRed,
	".php":  BrightBlue,
	".rb":   BrightRed,
	".sh":   BrightGreen,
	".bash": BrightGreen,
	".zsh":  BrightGreen,
	".fish": BrightGreen,

	// Config files
	".json":   BrightYellow,
	".yaml":   BrightYellow,
	".yml":    BrightYellow,
	".xml":    BrightYellow,
	".toml":   BrightYellow,
	".ini":    BrightYellow,
	".conf":   BrightYellow,
	".config": BrightYellow,

	// System files
	".log":   BrightBlack,
	".tmp":   BrightBlack,
	".cache": BrightBlack,
	".pid":   BrightBlack,
	".lock":  BrightBlack,

	// Special files
	"Makefile":   BrightGreen,
	"Dockerfile": BrightBlue,
	"README":     BrightCyan,
	"LICENSE":    BrightYellow,
	".gitignore": BrightBlack,
	".gitconfig": BrightBlack,
	".env":       BrightYellow,
}

// ColorConfig holds color configuration
type ColorConfig struct {
	Enabled      bool
	PromptColors bool
	FileColors   bool
	OutputColors bool
	currentColor int
}

// DefaultColorConfig returns default color configuration
func DefaultColorConfig() *ColorConfig {
	return &ColorConfig{
		Enabled:      true,
		PromptColors: true,
		FileColors:   true,
		OutputColors: true,
		currentColor: 0,
	}
}

// IsColorSupported checks if terminal supports colors
func IsColorSupported() bool {
	term := os.Getenv("TERM")
	if term == "" {
		return false
	}

	// Check for common color-supporting terminals
	colorTerms := []string{
		"xterm", "xterm-256color", "screen", "screen-256color",
		"tmux", "tmux-256color", "rxvt", "gnome", "konsole",
	}

	for _, colorTerm := range colorTerms {
		if strings.Contains(term, colorTerm) {
			return true
		}
	}

	// Check COLORTERM environment variable
	if os.Getenv("COLORTERM") != "" {
		return true
	}

	return false
}

// Colorize wraps text with color codes
func Colorize(text, color string) string {
	if !IsColorSupported() {
		return text
	}
	return color + text + Reset
}

// GetFileColor returns appropriate color for a file
func GetFileColor(filename string, isDir bool, isExecutable bool) string {
	if !IsColorSupported() {
		return ""
	}

	if isDir {
		return FileTypeColors["directory"]
	}

	if isExecutable {
		return FileTypeColors["executable"]
	}

	// Check by exact filename
	if color, exists := FileTypeColors[filename]; exists {
		return color
	}

	// Check by extension
	ext := strings.ToLower(filepath.Ext(filename))
	if color, exists := FileTypeColors[ext]; exists {
		return color
	}

	// Default color for regular files
	return White
}

// ColorizeFilename returns colored filename
func ColorizeFilename(filename string, isDir bool, isExecutable bool) string {
	color := GetFileColor(filename, isDir, isExecutable)
	return Colorize(filename, color)
}

// GetNextPromptColor returns next color for dynamic prompt
func (c *ColorConfig) GetNextPromptColor() string {
	if !c.Enabled || !c.PromptColors || !IsColorSupported() {
		return ""
	}

	color := RainbowColors[c.currentColor]
	c.currentColor = (c.currentColor + 1) % len(RainbowColors)
	return color
}

// GetRandomPromptColor returns random color for prompt
func GetRandomPromptColor() string {
	if !IsColorSupported() {
		return ""
	}

	rand.Seed(time.Now().UnixNano())
	return RainbowColors[rand.Intn(len(RainbowColors))]
}

// FormatPrompt creates a colorful prompt
func (c *ColorConfig) FormatPrompt(username, hostname, cwd, shellName string) string {
	if !c.Enabled || !IsColorSupported() {
		return fmt.Sprintf("%s> ", shellName)
	}

	// Get current color for this prompt
	promptColor := c.GetNextPromptColor()

	// Create colorful prompt parts
	var parts []string

	if username != "" {
		parts = append(parts, Colorize(username, BrightGreen))
	}

	if hostname != "" {
		parts = append(parts, Colorize("@"+hostname, BrightCyan))
	}

	if cwd != "" {
		// Shorten path if too long
		shortCwd := cwd
		if len(cwd) > 30 {
			shortCwd = "..." + cwd[len(cwd)-27:]
		}
		parts = append(parts, Colorize(":"+shortCwd, BrightBlue))
	}

	// Add shell name with dynamic color
	shellPart := Colorize(shellName, promptColor)

	// Combine parts
	prompt := strings.Join(parts, "")
	if len(parts) > 0 {
		prompt += " "
	}
	prompt += shellPart + Colorize("> ", BrightWhite)

	return prompt
}

// PrintSuccess prints success message in green
func PrintSuccess(message string) {
	fmt.Printf("%s✅ %s%s\n", BrightGreen, message, Reset)
}

// PrintError prints error message in red
func PrintError(message string) {
	fmt.Printf("%s❌ %s%s\n", BrightRed, message, Reset)
}

// PrintWarning prints warning message in yellow
func PrintWarning(message string) {
	fmt.Printf("%s⚠️  %s%s\n", BrightYellow, message, Reset)
}

// PrintInfo prints info message in blue
func PrintInfo(message string) {
	fmt.Printf("%s💡 %s%s\n", BrightBlue, message, Reset)
}

// PrintHeader prints a colorful header
func PrintHeader(title string) {
	border := strings.Repeat("═", len(title)+4)
	fmt.Printf("%s╔%s╗%s\n", BrightCyan, border, Reset)
	fmt.Printf("%s║  %s%s%s  ║%s\n", BrightCyan, Bold+BrightWhite, title, BrightCyan, Reset)
	fmt.Printf("%s╚%s╝%s\n", BrightCyan, border, Reset)
}

// Rainbow effect for text
func Rainbow(text string) string {
	if !IsColorSupported() {
		return text
	}

	var result strings.Builder
	for i, char := range text {
		color := RainbowColors[i%len(RainbowColors)]
		result.WriteString(color + string(char))
	}
	result.WriteString(Reset)
	return result.String()
}

// Gradient effect for text
func Gradient(text string, startColor, endColor string) string {
	if !IsColorSupported() {
		return text
	}

	// Simple gradient - alternate between start and end colors
	var result strings.Builder
	for i, char := range text {
		if i%2 == 0 {
			result.WriteString(startColor + string(char))
		} else {
			result.WriteString(endColor + string(char))
		}
	}
	result.WriteString(Reset)
	return result.String()
}

// PrintWelcome prints colorful welcome message
func PrintWelcome(shellName, version string) {
	if !IsColorSupported() {
		fmt.Printf("Welcome to %s v%s - High-Performance Linux Shell\n", shellName, version)
		fmt.Println("Type 'help' for available commands")
		fmt.Println()
		return
	}

	// Colorful welcome
	fmt.Printf("%sWelcome to %s%s v%s%s - %s%s\n",
		BrightCyan, Rainbow(shellName), BrightMagenta, version,
		BrightGreen, Gradient("High-Performance Linux Shell", BrightYellow, BrightRed), Reset)
	fmt.Printf("%sType %s'help'%s for available commands\n",
		BrightBlue, BrightYellow, BrightBlue)
	fmt.Printf("%s\n", Reset)
}

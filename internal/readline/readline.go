package readline

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"gex/internal/shell"
)

// Readline provides advanced line editing capabilities
type Readline struct {
	session    *shell.Session
	reader     *bufio.Reader
	history    []string
	historyPos int
	line       []rune
	cursor     int
	prompt     string
}

// New creates a new readline instance
func New(session *shell.Session) *Readline {
	return &Readline{
		session:    session,
		reader:     bufio.NewReader(os.Stdin),
		history:    make([]string, 0),
		historyPos: -1,
		line:       make([]rune, 0),
		cursor:     0,
		prompt:     "gex> ",
	}
}

// ReadLine reads a line with advanced editing features
func (r *Readline) ReadLine() (string, error) {
	// Check if stdin is a terminal
	if !isTerminal() {
		return r.readSimple()
	}

	// Set terminal to raw mode for advanced editing
	oldState, err := setRawMode()
	if err != nil {
		return r.readSimple()
	}
	defer restoreTerminal(oldState)

	return r.readAdvanced()
}

// readSimple reads a line without advanced features (for non-terminals)
func (r *Readline) readSimple() (string, error) {
	fmt.Print(r.prompt)
	line, err := r.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\n\r"), nil
}

// readAdvanced reads a line with advanced editing features
func (r *Readline) readAdvanced() (string, error) {
	r.line = r.line[:0]
	r.cursor = 0
	r.historyPos = -1

	r.displayPrompt()

	for {
		char, err := r.readChar()
		if err != nil {
			return "", err
		}

		switch char {
		case '\r', '\n':
			// Enter - submit line
			fmt.Print("\r\n")
			result := string(r.line)
			if result != "" {
				r.addToHistory(result)
			}
			return result, nil

		case '\x03': // Ctrl+C
			fmt.Print("^C\r\n")
			r.line = r.line[:0]
			r.cursor = 0
			r.displayPrompt()

		case '\x04': // Ctrl+D (EOF)
			if len(r.line) == 0 {
				return "", fmt.Errorf("EOF")
			}
			// Delete character at cursor
			r.deleteChar()

		case '\x08', '\x7f': // Backspace or DEL
			r.backspace()

		case '\x09': // Tab - autocomplete
			r.autoComplete()

		case '\x0c': // Ctrl+L - clear screen
			r.clearScreen()

		case '\x01': // Ctrl+A - beginning of line
			r.moveToBeginning()

		case '\x05': // Ctrl+E - end of line
			r.moveToEnd()

		case '\x02': // Ctrl+B - move left
			r.moveLeft()

		case '\x06': // Ctrl+F - move right
			r.moveRight()

		case '\x0e': // Ctrl+N - next history
			r.nextHistory()

		case '\x10': // Ctrl+P - previous history
			r.prevHistory()

		case '\x0b': // Ctrl+K - kill to end of line
			r.killToEnd()

		case '\x15': // Ctrl+U - kill entire line
			r.killLine()

		case '\x17': // Ctrl+W - kill word backward
			r.killWordBackward()

		case '\x1b': // ESC - handle escape sequences
			if err := r.handleEscapeSequence(); err != nil {
				return "", err
			}

		default:
			if char >= 32 && char < 127 {
				// Printable character
				r.insertChar(rune(char))
			}
		}
	}
}

// readChar reads a single character
func (r *Readline) readChar() (byte, error) {
	var buf [1]byte
	_, err := os.Stdin.Read(buf[:])
	return buf[0], err
}

// handleEscapeSequence handles escape sequences (arrow keys, etc.)
func (r *Readline) handleEscapeSequence() error {
	char, err := r.readChar()
	if err != nil {
		return err
	}

	if char == '[' {
		char, err = r.readChar()
		if err != nil {
			return err
		}

		switch char {
		case 'A': // Up arrow - previous history
			r.prevHistory()
		case 'B': // Down arrow - next history
			r.nextHistory()
		case 'C': // Right arrow
			r.moveRight()
		case 'D': // Left arrow
			r.moveLeft()
		case 'H': // Home
			r.moveToBeginning()
		case 'F': // End
			r.moveToEnd()
		case '3': // Delete key
			if char, err := r.readChar(); err == nil && char == '~' {
				r.deleteChar()
			}
		}
	}

	return nil
}

// Movement and editing functions
func (r *Readline) insertChar(char rune) {
	// Insert character at cursor position
	r.line = append(r.line, 0)
	copy(r.line[r.cursor+1:], r.line[r.cursor:])
	r.line[r.cursor] = char
	r.cursor++
	r.redrawLine()
}

func (r *Readline) backspace() {
	if r.cursor > 0 {
		copy(r.line[r.cursor-1:], r.line[r.cursor:])
		r.line = r.line[:len(r.line)-1]
		r.cursor--
		r.redrawLine()
	}
}

func (r *Readline) deleteChar() {
	if r.cursor < len(r.line) {
		copy(r.line[r.cursor:], r.line[r.cursor+1:])
		r.line = r.line[:len(r.line)-1]
		r.redrawLine()
	}
}

func (r *Readline) moveLeft() {
	if r.cursor > 0 {
		r.cursor--
		fmt.Print("\x1b[D")
	}
}

func (r *Readline) moveRight() {
	if r.cursor < len(r.line) {
		r.cursor++
		fmt.Print("\x1b[C")
	}
}

func (r *Readline) moveToBeginning() {
	if r.cursor > 0 {
		fmt.Printf("\x1b[%dD", r.cursor)
		r.cursor = 0
	}
}

func (r *Readline) moveToEnd() {
	if r.cursor < len(r.line) {
		fmt.Printf("\x1b[%dC", len(r.line)-r.cursor)
		r.cursor = len(r.line)
	}
}

func (r *Readline) killToEnd() {
	if r.cursor < len(r.line) {
		r.line = r.line[:r.cursor]
		r.redrawLine()
	}
}

func (r *Readline) killLine() {
	r.line = r.line[:0]
	r.cursor = 0
	r.redrawLine()
}

func (r *Readline) killWordBackward() {
	if r.cursor == 0 {
		return
	}

	// Find start of current word
	pos := r.cursor - 1
	for pos > 0 && r.line[pos] == ' ' {
		pos--
	}
	for pos > 0 && r.line[pos] != ' ' {
		pos--
	}
	if r.line[pos] == ' ' {
		pos++
	}

	// Remove the word
	copy(r.line[pos:], r.line[r.cursor:])
	r.line = r.line[:len(r.line)-(r.cursor-pos)]
	r.cursor = pos
	r.redrawLine()
}

// History navigation
func (r *Readline) prevHistory() {
	history := r.session.GetHistory()
	if len(history) == 0 {
		return
	}

	if r.historyPos == -1 {
		r.historyPos = len(history) - 1
	} else if r.historyPos > 0 {
		r.historyPos--
	}

	if r.historyPos >= 0 && r.historyPos < len(history) {
		r.line = []rune(history[r.historyPos])
		r.cursor = len(r.line)
		r.redrawLine()
	}
}

func (r *Readline) nextHistory() {
	history := r.session.GetHistory()
	if len(history) == 0 || r.historyPos == -1 {
		return
	}

	r.historyPos++
	if r.historyPos >= len(history) {
		r.historyPos = -1
		r.line = r.line[:0]
		r.cursor = 0
	} else {
		r.line = []rune(history[r.historyPos])
		r.cursor = len(r.line)
	}
	r.redrawLine()
}

func (r *Readline) addToHistory(line string) {
	r.session.AddHistory(line)
}

// Display functions
func (r *Readline) displayPrompt() {
	fmt.Print(r.prompt)
}

func (r *Readline) redrawLine() {
	// Clear current line
	fmt.Print("\r\x1b[K")

	// Print prompt and line
	fmt.Print(r.prompt)
	fmt.Print(string(r.line))

	// Move cursor to correct position
	if r.cursor < len(r.line) {
		fmt.Printf("\x1b[%dD", len(r.line)-r.cursor)
	}
}

func (r *Readline) clearScreen() {
	fmt.Print("\x1b[2J\x1b[H")
	r.displayPrompt()
	fmt.Print(string(r.line))
	if r.cursor < len(r.line) {
		fmt.Printf("\x1b[%dD", len(r.line)-r.cursor)
	}
}

// Autocompletion
func (r *Readline) autoComplete() {
	if len(r.line) == 0 {
		return
	}

	// Get current word
	wordStart := r.cursor
	for wordStart > 0 && r.line[wordStart-1] != ' ' {
		wordStart--
	}

	if wordStart >= r.cursor {
		return
	}

	word := string(r.line[wordStart:r.cursor])

	// Get completions
	completions := r.getCompletions(word)
	if len(completions) == 0 {
		return
	}

	if len(completions) == 1 {
		// Single completion - insert it
		completion := completions[0]
		suffix := completion[len(word):]
		for _, char := range suffix {
			r.insertChar(char)
		}
	} else {
		// Multiple completions - show them
		fmt.Print("\r\n")
		for _, completion := range completions {
			fmt.Printf("%s  ", completion)
		}
		fmt.Print("\r\n")
		r.displayPrompt()
		fmt.Print(string(r.line))
		if r.cursor < len(r.line) {
			fmt.Printf("\x1b[%dD", len(r.line)-r.cursor)
		}
	}
}

func (r *Readline) getCompletions(prefix string) []string {
	var completions []string

	// Add command completions (simple implementation)
	commands := []string{"cd", "pwd", "echo", "exit", "help", "history", "alias", "unalias", "env", "export", "which", "type"}
	for _, cmd := range commands {
		if strings.HasPrefix(cmd, prefix) {
			completions = append(completions, cmd)
		}
	}

	// Add file completions (basic implementation)
	// This could be expanded to include proper file system traversal

	return completions
}

// Terminal control functions
func isTerminal() bool {
	var termios syscall.Termios
	_, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&termios)),
		0, 0, 0,
	)
	return errno == 0
}

func setRawMode() (*syscall.Termios, error) {
	var oldState syscall.Termios

	// Get current terminal state
	_, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&oldState)),
		0, 0, 0,
	)
	if errno != 0 {
		return nil, errno
	}

	// Set raw mode
	newState := oldState
	newState.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	newState.Cc[syscall.VMIN] = 1
	newState.Cc[syscall.VTIME] = 0

	_, _, errno = syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(&newState)),
		0, 0, 0,
	)
	if errno != 0 {
		return nil, errno
	}

	return &oldState, nil
}

func restoreTerminal(oldState *syscall.Termios) {
	syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(oldState)),
		0, 0, 0,
	)
}

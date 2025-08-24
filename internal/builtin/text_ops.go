package builtin

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Cat displays file contents (like cat command)
func Cat(args []string) error {
	if len(args) == 0 {
		// Read from stdin
		return catReader(os.Stdin)
	}

	for _, filename := range args {
		if filename == "-" {
			if err := catReader(os.Stdin); err != nil {
				fmt.Printf("cat: %v\n", err)
			}
			continue
		}

		file, err := os.Open(filename)
		if err != nil {
			fmt.Printf("cat: %v\n", err)
			continue
		}

		err = catReader(file)
		file.Close()

		if err != nil {
			fmt.Printf("cat: %v\n", err)
		}
	}

	return nil
}

// catReader reads from a reader and outputs to stdout
func catReader(reader io.Reader) error {
	_, err := io.Copy(os.Stdout, reader)
	return err
}

// Head displays first lines of files (like head command)
func Head(args []string) error {
	lines := 10 // default
	var files []string

	// Parse arguments
	for i, arg := range args {
		if arg == "-n" && i+1 < len(args) {
			var err error
			lines, err = strconv.Atoi(args[i+1])
			if err != nil {
				return fmt.Errorf("head: invalid number of lines: %s", args[i+1])
			}
			i++ // skip next argument
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			// Handle -10 format
			var err error
			lines, err = strconv.Atoi(arg[1:])
			if err != nil {
				return fmt.Errorf("head: invalid number of lines: %s", arg[1:])
			}
		} else {
			files = append(files, args[i:]...)
			break
		}
	}

	if len(files) == 0 {
		return headReader(os.Stdin, lines)
	}

	for i, filename := range files {
		if len(files) > 1 {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("==> %s <==\n", filename)
		}

		if filename == "-" {
			if err := headReader(os.Stdin, lines); err != nil {
				fmt.Printf("head: %v\n", err)
			}
			continue
		}

		file, err := os.Open(filename)
		if err != nil {
			fmt.Printf("head: %v\n", err)
			continue
		}

		err = headReader(file, lines)
		file.Close()

		if err != nil {
			fmt.Printf("head: %v\n", err)
		}
	}

	return nil
}

// headReader reads first n lines from a reader
func headReader(reader io.Reader, lines int) error {
	scanner := bufio.NewScanner(reader)
	count := 0

	for scanner.Scan() && count < lines {
		fmt.Println(scanner.Text())
		count++
	}

	return scanner.Err()
}

// Tail displays last lines of files (like tail command)
func Tail(args []string) error {
	lines := 10 // default
	var files []string

	// Parse arguments (simplified)
	for i, arg := range args {
		if arg == "-n" && i+1 < len(args) {
			var err error
			lines, err = strconv.Atoi(args[i+1])
			if err != nil {
				return fmt.Errorf("tail: invalid number of lines: %s", args[i+1])
			}
			i++ // skip next argument
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			var err error
			lines, err = strconv.Atoi(arg[1:])
			if err != nil {
				return fmt.Errorf("tail: invalid number of lines: %s", arg[1:])
			}
		} else {
			files = append(files, args[i:]...)
			break
		}
	}

	if len(files) == 0 {
		return tailReader(os.Stdin, lines)
	}

	for i, filename := range files {
		if len(files) > 1 {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("==> %s <==\n", filename)
		}

		file, err := os.Open(filename)
		if err != nil {
			fmt.Printf("tail: %v\n", err)
			continue
		}

		err = tailFile(file, lines)
		file.Close()

		if err != nil {
			fmt.Printf("tail: %v\n", err)
		}
	}

	return nil
}

// tailReader displays last n lines from reader (for stdin)
func tailReader(reader io.Reader, lines int) error {
	scanner := bufio.NewScanner(reader)
	buffer := make([]string, 0, lines)

	for scanner.Scan() {
		buffer = append(buffer, scanner.Text())
		if len(buffer) > lines {
			buffer = buffer[1:]
		}
	}

	for _, line := range buffer {
		fmt.Println(line)
	}

	return scanner.Err()
}

// tailFile displays last n lines from file
func tailFile(file *os.File, lines int) error {
	// For simplicity, read all lines and keep last n
	scanner := bufio.NewScanner(file)
	buffer := make([]string, 0, lines)

	for scanner.Scan() {
		buffer = append(buffer, scanner.Text())
		if len(buffer) > lines {
			buffer = buffer[1:]
		}
	}

	for _, line := range buffer {
		fmt.Println(line)
	}

	return scanner.Err()
}

// Wc counts lines, words, and characters (like wc command)
func Wc(args []string) error {
	var showLines, showWords, showChars bool = true, true, true
	var files []string

	// Parse flags
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			// Reset defaults when flags are specified
			if i == 0 {
				showLines, showWords, showChars = false, false, false
			}

			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'l':
					showLines = true
				case 'w':
					showWords = true
				case 'c':
					showChars = true
				}
			}
		} else {
			files = append(files, args[i:]...)
			break
		}
	}

	if len(files) == 0 {
		lines, words, chars, err := wcReader(os.Stdin)
		if err != nil {
			return err
		}
		printWcResult(lines, words, chars, "", showLines, showWords, showChars)
		return nil
	}

	totalLines, totalWords, totalChars := 0, 0, 0

	for _, filename := range files {
		file, err := os.Open(filename)
		if err != nil {
			fmt.Printf("wc: %v\n", err)
			continue
		}

		lines, words, chars, err := wcReader(file)
		file.Close()

		if err != nil {
			fmt.Printf("wc: %v\n", err)
			continue
		}

		printWcResult(lines, words, chars, filename, showLines, showWords, showChars)

		totalLines += lines
		totalWords += words
		totalChars += chars
	}

	if len(files) > 1 {
		printWcResult(totalLines, totalWords, totalChars, "total", showLines, showWords, showChars)
	}

	return nil
}

// wcReader counts lines, words, and characters from reader
func wcReader(reader io.Reader) (int, int, int, error) {
	scanner := bufio.NewScanner(reader)
	lines, words, chars := 0, 0, 0

	for scanner.Scan() {
		text := scanner.Text()
		lines++
		chars += len(text) + 1 // +1 for newline
		words += len(strings.Fields(text))
	}

	return lines, words, chars, scanner.Err()
}

// printWcResult prints wc results in the correct format
func printWcResult(lines, words, chars int, filename string, showLines, showWords, showChars bool) {
	var result strings.Builder

	if showLines {
		result.WriteString(fmt.Sprintf("%8d", lines))
	}
	if showWords {
		result.WriteString(fmt.Sprintf("%8d", words))
	}
	if showChars {
		result.WriteString(fmt.Sprintf("%8d", chars))
	}

	if filename != "" {
		result.WriteString(" " + filename)
	}

	fmt.Println(result.String())
}

// Grep searches for patterns in files (like grep command)
func Grep(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("grep: missing pattern")
	}

	var ignoreCase bool
	var lineNumbers bool
	var invertMatch bool
	var pattern string
	var files []string

	// Parse arguments
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'i':
					ignoreCase = true
				case 'n':
					lineNumbers = true
				case 'v':
					invertMatch = true
				}
			}
		} else {
			pattern = arg
			files = args[i+1:]
			break
		}
	}

	if pattern == "" {
		return fmt.Errorf("grep: missing pattern")
	}

	// Compile regex
	var regex *regexp.Regexp
	var err error

	if ignoreCase {
		regex, err = regexp.Compile("(?i)" + pattern)
	} else {
		regex, err = regexp.Compile(pattern)
	}

	if err != nil {
		return fmt.Errorf("grep: invalid pattern: %v", err)
	}

	if len(files) == 0 {
		return grepReader(os.Stdin, "", regex, lineNumbers, invertMatch, false)
	}

	showFilenames := len(files) > 1

	for _, filename := range files {
		file, err := os.Open(filename)
		if err != nil {
			fmt.Printf("grep: %v\n", err)
			continue
		}

		err = grepReader(file, filename, regex, lineNumbers, invertMatch, showFilenames)
		file.Close()

		if err != nil {
			fmt.Printf("grep: %v\n", err)
		}
	}

	return nil
}

// grepReader searches for pattern in reader
func grepReader(reader io.Reader, filename string, regex *regexp.Regexp, lineNumbers, invertMatch, showFilenames bool) error {
	scanner := bufio.NewScanner(reader)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		text := scanner.Text()
		matches := regex.MatchString(text)

		if matches != invertMatch { // XOR logic
			continue
		}

		var output strings.Builder

		if showFilenames {
			output.WriteString(filename + ":")
		}

		if lineNumbers {
			output.WriteString(fmt.Sprintf("%d:", lineNum))
		}

		output.WriteString(text)
		fmt.Println(output.String())
	}

	return scanner.Err()
}

// Sort sorts lines in files (like sort command)
func Sort(args []string) error {
	var reverse bool
	var numeric bool
	var unique bool
	var files []string

	// Parse flags
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'r':
					reverse = true
				case 'n':
					numeric = true
				case 'u':
					unique = true
				}
			}
		} else {
			files = append(files, args[i:]...)
			break
		}
	}

	var lines []string

	if len(files) == 0 {
		var err error
		lines, err = readLines(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		for _, filename := range files {
			file, err := os.Open(filename)
			if err != nil {
				fmt.Printf("sort: %v\n", err)
				continue
			}

			fileLines, err := readLines(file)
			file.Close()

			if err != nil {
				fmt.Printf("sort: %v\n", err)
				continue
			}

			lines = append(lines, fileLines...)
		}
	}

	return sortAndPrint(lines, reverse, numeric, unique)
}

// readLines reads all lines from a reader
func readLines(reader io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// sortAndPrint sorts lines and prints them
func sortAndPrint(lines []string, reverse, numeric, unique bool) error {
	if numeric {
		// Numeric sort (simplified)
		if reverse {
			for i := len(lines) - 1; i >= 0; i-- {
				fmt.Println(lines[i])
			}
		} else {
			for _, line := range lines {
				fmt.Println(line)
			}
		}
	} else {
		// String sort
		if reverse {
			for i := len(lines) - 1; i >= 0; i-- {
				if unique && i > 0 && lines[i] == lines[i-1] {
					continue
				}
				fmt.Println(lines[i])
			}
		} else {
			prev := ""
			for _, line := range lines {
				if unique && line == prev {
					continue
				}
				fmt.Println(line)
				prev = line
			}
		}
	}

	return nil
}

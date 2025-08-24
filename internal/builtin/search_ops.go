package builtin

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Find searches for files and directories (like find command)
func Find(args []string) error {
	var paths []string
	var name string
	var fileType string
	var maxDepth int = -1
	var minDepth int = 0
	var exec string
	var size string
	var mtime string

	// Parse arguments
	i := 0
	for i < len(args) {
		arg := args[i]

		if !strings.HasPrefix(arg, "-") {
			// Path argument
			paths = append(paths, arg)
		} else {
			switch arg {
			case "-name":
				if i+1 < len(args) {
					i++
					name = args[i]
				}
			case "-type":
				if i+1 < len(args) {
					i++
					fileType = args[i]
				}
			case "-maxdepth":
				if i+1 < len(args) {
					i++
					if d, err := strconv.Atoi(args[i]); err == nil {
						maxDepth = d
					}
				}
			case "-mindepth":
				if i+1 < len(args) {
					i++
					if d, err := strconv.Atoi(args[i]); err == nil {
						minDepth = d
					}
				}
			case "-exec":
				if i+1 < len(args) {
					i++
					exec = args[i]
				}
			case "-size":
				if i+1 < len(args) {
					i++
					size = args[i]
				}
			case "-mtime":
				if i+1 < len(args) {
					i++
					mtime = args[i]
				}
			}
		}
		i++
	}

	if len(paths) == 0 {
		paths = []string{"."}
	}

	for _, path := range paths {
		if err := findInPath(path, name, fileType, maxDepth, minDepth, exec, size, mtime, 0); err != nil {
			fmt.Printf("find: %v\n", err)
		}
	}

	return nil
}

// findInPath recursively searches in a path
func findInPath(path, name, fileType string, maxDepth, minDepth int, exec, size, mtime string, currentDepth int) error {
	// Check depth limits
	if maxDepth >= 0 && currentDepth > maxDepth {
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	// Check if current item matches criteria
	if currentDepth >= minDepth {
		if matchesCriteria(path, info, name, fileType, size, mtime) {
			if exec != "" {
				// Execute command on found file
				fmt.Printf("Executing: %s %s\n", exec, path)
			} else {
				fmt.Println(path)
			}
		}
	}

	// Recurse into directories
	if info.IsDir() && (maxDepth < 0 || currentDepth < maxDepth) {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			subPath := filepath.Join(path, entry.Name())
			findInPath(subPath, name, fileType, maxDepth, minDepth, exec, size, mtime, currentDepth+1)
		}
	}

	return nil
}

// matchesCriteria checks if a file matches the search criteria
func matchesCriteria(path string, info os.FileInfo, name, fileType, size, mtime string) bool {
	// Check name pattern
	if name != "" {
		matched, err := filepath.Match(name, filepath.Base(path))
		if err != nil || !matched {
			// Try regex if glob fails
			if regex, err := regexp.Compile(name); err == nil {
				if !regex.MatchString(filepath.Base(path)) {
					return false
				}
			} else {
				return false
			}
		}
	}

	// Check file type
	if fileType != "" {
		switch fileType {
		case "f":
			if !info.Mode().IsRegular() {
				return false
			}
		case "d":
			if !info.IsDir() {
				return false
			}
		case "l":
			if info.Mode()&os.ModeSymlink == 0 {
				return false
			}
		}
	}

	// Check size (simplified)
	if size != "" {
		// Format: +100k, -50M, 1G etc.
		// Simplified implementation
		if !matchesSize(info.Size(), size) {
			return false
		}
	}

	// Check modification time (simplified)
	if mtime != "" {
		// Format: +7, -1, 0 (days)
		if !matchesMtime(info.ModTime(), mtime) {
			return false
		}
	}

	return true
}

// matchesSize checks if file size matches criteria
func matchesSize(fileSize int64, sizeSpec string) bool {
	if len(sizeSpec) == 0 {
		return true
	}

	// Parse size specification
	var operator string
	var value int64
	var unit int64 = 1

	if sizeSpec[0] == '+' || sizeSpec[0] == '-' {
		operator = string(sizeSpec[0])
		sizeSpec = sizeSpec[1:]
	}

	// Parse unit
	if len(sizeSpec) > 0 {
		lastChar := sizeSpec[len(sizeSpec)-1]
		switch lastChar {
		case 'k', 'K':
			unit = 1024
			sizeSpec = sizeSpec[:len(sizeSpec)-1]
		case 'm', 'M':
			unit = 1024 * 1024
			sizeSpec = sizeSpec[:len(sizeSpec)-1]
		case 'g', 'G':
			unit = 1024 * 1024 * 1024
			sizeSpec = sizeSpec[:len(sizeSpec)-1]
		}
	}

	if v, err := strconv.ParseInt(sizeSpec, 10, 64); err == nil {
		value = v * unit
	} else {
		return true // Invalid format, ignore
	}

	switch operator {
	case "+":
		return fileSize > value
	case "-":
		return fileSize < value
	default:
		return fileSize == value
	}
}

// matchesMtime checks if modification time matches criteria
func matchesMtime(modTime time.Time, mtimeSpec string) bool {
	if len(mtimeSpec) == 0 {
		return true
	}

	var operator string
	var days int

	if mtimeSpec[0] == '+' || mtimeSpec[0] == '-' {
		operator = string(mtimeSpec[0])
		mtimeSpec = mtimeSpec[1:]
	}

	if d, err := strconv.Atoi(mtimeSpec); err == nil {
		days = d
	} else {
		return true // Invalid format, ignore
	}

	now := time.Now()
	daysDiff := int(now.Sub(modTime).Hours() / 24)

	switch operator {
	case "+":
		return daysDiff > days
	case "-":
		return daysDiff < days
	default:
		return daysDiff == days
	}
}

// Locate finds files by name in database (simplified implementation)
func Locate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("locate: missing pattern")
	}

	pattern := args[0]

	// Simple implementation: search in common directories
	searchDirs := []string{
		"/usr/bin",
		"/usr/local/bin",
		"/bin",
		"/sbin",
		"/usr/sbin",
		"/home",
		"/opt",
		"/var",
		"/etc",
	}

	for _, dir := range searchDirs {
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors
			}

			if strings.Contains(strings.ToLower(d.Name()), strings.ToLower(pattern)) {
				fmt.Println(path)
			}

			return nil
		})
	}

	return nil
}

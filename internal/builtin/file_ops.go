package builtin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gex/internal/ui"
)

// Ls lists directory contents (like ls command)
func Ls(args []string) error {
	var paths []string
	var showHidden bool
	var longFormat bool
	var humanReadable bool
	var sortByTime bool
	var reverse bool

	// Parse flags
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'a':
					showHidden = true
				case 'l':
					longFormat = true
				case 'h':
					humanReadable = true
				case 't':
					sortByTime = true
				case 'r':
					reverse = true
				}
			}
		} else {
			paths = append(paths, args[i:]...)
			break
		}
	}

	if len(paths) == 0 {
		paths = []string{"."}
	}

	for _, path := range paths {
		if err := listDirectory(path, showHidden, longFormat, humanReadable, sortByTime, reverse); err != nil {
			fmt.Printf("ls: %v\n", err)
		}
	}

	return nil
}

// listDirectory implements the directory listing logic
func listDirectory(path string, showHidden, longFormat, humanReadable, sortByTime, reverse bool) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	var files []os.DirEntry
	for _, entry := range entries {
		if !showHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		files = append(files, entry)
	}

	// Sort files
	if sortByTime {
		sort.Slice(files, func(i, j int) bool {
			info1, _ := files[i].Info()
			info2, _ := files[j].Info()
			if reverse {
				return info1.ModTime().Before(info2.ModTime())
			}
			return info1.ModTime().After(info2.ModTime())
		})
	} else {
		sort.Slice(files, func(i, j int) bool {
			if reverse {
				return files[i].Name() > files[j].Name()
			}
			return files[i].Name() < files[j].Name()
		})
	}

	if longFormat {
		return printLongFormat(files, path, humanReadable)
	}

	// Simple format with colors
	for _, file := range files {
		info, _ := file.Info()
		isDir := file.IsDir()
		isExecutable := info != nil && info.Mode()&0111 != 0

		coloredName := ui.ColorizeFilename(file.Name(), isDir, isExecutable)
		fmt.Printf("%s  ", coloredName)
	}
	if len(files) > 0 {
		fmt.Println()
	}

	return nil
}

// printLongFormat prints files in long format (-l flag)
func printLongFormat(files []os.DirEntry, basePath string, humanReadable bool) error {
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}

		// File mode (permissions)
		mode := info.Mode()
		modeStr := mode.String()

		// Number of links (simplified)
		links := "1"
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			links = strconv.FormatUint(uint64(stat.Nlink), 10)
		}

		// Owner and group (simplified)
		owner := "user"
		group := "group"

		// Size
		var sizeStr string
		if humanReadable {
			sizeStr = formatHumanReadable(info.Size())
		} else {
			sizeStr = strconv.FormatInt(info.Size(), 10)
		}

		// Modification time
		modTime := info.ModTime().Format("Jan 02 15:04")

		// Colorize filename
		isDir := info.IsDir()
		isExecutable := info.Mode()&0111 != 0
		coloredName := ui.ColorizeFilename(file.Name(), isDir, isExecutable)

		fmt.Printf("%s %s %s %s %8s %s %s\n",
			modeStr, links, owner, group, sizeStr, modTime, coloredName)
	}

	return nil
}

// formatHumanReadable formats file size in human readable format
func formatHumanReadable(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := "KMGTPE"
	return fmt.Sprintf("%.1f%c", float64(size)/float64(div), units[exp])
}

// Mkdir creates directories (like mkdir command)
func Mkdir(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("mkdir: missing operand")
	}

	var createParents bool
	var mode os.FileMode = 0755
	var paths []string

	// Parse flags
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'p':
					createParents = true
				}
			}
		} else {
			paths = append(paths, args[i:]...)
			break
		}
	}

	if len(paths) == 0 {
		return fmt.Errorf("mkdir: missing operand")
	}

	for _, path := range paths {
		var err error
		if createParents {
			err = os.MkdirAll(path, mode)
		} else {
			err = os.Mkdir(path, mode)
		}

		if err != nil {
			fmt.Printf("mkdir: %v\n", err)
		}
	}

	return nil
}

// Rmdir removes empty directories (like rmdir command)
func Rmdir(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("rmdir: missing operand")
	}

	for _, path := range args {
		if err := os.Remove(path); err != nil {
			fmt.Printf("rmdir: %v\n", err)
		}
	}

	return nil
}

// Rm removes files and directories (like rm command)
func Rm(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("rm: missing operand")
	}

	var recursive bool
	var force bool
	var paths []string

	// Parse flags
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'r', 'R':
					recursive = true
				case 'f':
					force = true
				}
			}
		} else {
			paths = append(paths, args[i:]...)
			break
		}
	}

	if len(paths) == 0 {
		return fmt.Errorf("rm: missing operand")
	}

	for _, path := range paths {
		var err error
		if recursive {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}

		if err != nil && !force {
			fmt.Printf("rm: %v\n", err)
		}
	}

	return nil
}

// Cp copies files and directories (like cp command)
func Cp(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("cp: missing operand")
	}

	var recursive bool
	var preserve bool
	var paths []string

	// Parse flags
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'r', 'R':
					recursive = true
				case 'p':
					preserve = true
				}
			}
		} else {
			paths = append(paths, args[i:]...)
			break
		}
	}

	if len(paths) < 2 {
		return fmt.Errorf("cp: missing operand")
	}

	dest := paths[len(paths)-1]
	sources := paths[:len(paths)-1]

	// Check if destination is a directory
	destInfo, err := os.Stat(dest)
	isDestDir := err == nil && destInfo.IsDir()

	for _, src := range sources {
		var destPath string
		if isDestDir {
			destPath = filepath.Join(dest, filepath.Base(src))
		} else {
			destPath = dest
		}

		if err := copyFile(src, destPath, recursive, preserve); err != nil {
			fmt.Printf("cp: %v\n", err)
		}
	}

	return nil
}

// copyFile implements file copying logic
func copyFile(src, dest string, recursive, preserve bool) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		if !recursive {
			return fmt.Errorf("omitting directory '%s'", src)
		}
		return copyDir(src, dest, preserve)
	}

	return copyRegularFile(src, dest, srcInfo, preserve)
}

// copyRegularFile copies a regular file
func copyRegularFile(src, dest string, srcInfo os.FileInfo, preserve bool) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}

	if preserve {
		if err := os.Chmod(dest, srcInfo.Mode()); err != nil {
			return err
		}
		if err := os.Chtimes(dest, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
			return err
		}
	}

	return nil
}

// copyDir copies a directory recursively
func copyDir(src, dest string, preserve bool) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dest, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if err := copyFile(srcPath, destPath, true, preserve); err != nil {
			return err
		}
	}

	if preserve {
		if err := os.Chmod(dest, srcInfo.Mode()); err != nil {
			return err
		}
		if err := os.Chtimes(dest, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
			return err
		}
	}

	return nil
}

// Mv moves/renames files and directories (like mv command)
func Mv(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("mv: missing operand")
	}

	if len(args) == 2 {
		// Simple move/rename
		return os.Rename(args[0], args[1])
	}

	// Multiple sources, destination must be a directory
	dest := args[len(args)-1]
	sources := args[:len(args)-1]

	destInfo, err := os.Stat(dest)
	if err != nil || !destInfo.IsDir() {
		return fmt.Errorf("mv: target '%s' is not a directory", dest)
	}

	for _, src := range sources {
		destPath := filepath.Join(dest, filepath.Base(src))
		if err := os.Rename(src, destPath); err != nil {
			fmt.Printf("mv: %v\n", err)
		}
	}

	return nil
}

// Touch creates empty files or updates timestamps (like touch command)
func Touch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("touch: missing operand")
	}

	now := time.Now()

	for _, path := range args {
		// Try to update timestamp if file exists
		if err := os.Chtimes(path, now, now); err != nil {
			// File doesn't exist, create it
			file, createErr := os.Create(path)
			if createErr != nil {
				fmt.Printf("touch: %v\n", createErr)
				continue
			}
			file.Close()
		}
	}

	return nil
}

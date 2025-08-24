package builtin

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
)

// Chmod changes file permissions (like chmod command)
func Chmod(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("chmod: missing operand")
	}

	var recursive bool
	var modeStr string
	var files []string

	// Parse arguments
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'R':
					recursive = true
				}
			}
		} else {
			if modeStr == "" {
				modeStr = arg
			} else {
				files = append(files, args[i:]...)
				break
			}
		}
	}

	if modeStr == "" || len(files) == 0 {
		return fmt.Errorf("chmod: missing operand")
	}

	// Parse mode
	mode, err := parseMode(modeStr)
	if err != nil {
		return fmt.Errorf("chmod: invalid mode: %s", modeStr)
	}

	// Apply to files
	for _, file := range files {
		if err := chmodFile(file, mode, recursive); err != nil {
			fmt.Printf("chmod: %v\n", err)
		}
	}

	return nil
}

// parseMode parses chmod mode string
func parseMode(modeStr string) (os.FileMode, error) {
	// Handle octal mode (e.g., 755, 644)
	if len(modeStr) >= 3 && len(modeStr) <= 4 {
		if octal, err := strconv.ParseUint(modeStr, 8, 32); err == nil {
			return os.FileMode(octal), nil
		}
	}

	// Handle symbolic mode (e.g., u+x, go-w, a=r)
	// Simplified implementation
	var mode os.FileMode = 0644 // Default

	parts := strings.Split(modeStr, ",")
	for _, part := range parts {
		if err := applySymbolicMode(&mode, part); err != nil {
			return 0, err
		}
	}

	return mode, nil
}

// applySymbolicMode applies symbolic mode changes
func applySymbolicMode(mode *os.FileMode, modeStr string) error {
	// Parse who (u/g/o/a)
	var user, group, other bool

	i := 0
	for i < len(modeStr) {
		switch modeStr[i] {
		case 'u':
			user = true
		case 'g':
			group = true
		case 'o':
			other = true
		case 'a':
			user, group, other = true, true, true
		default:
			goto parseOp
		}
		i++
	}

	if !user && !group && !other {
		user, group, other = true, true, true // Default to all
	}

parseOp:
	if i >= len(modeStr) {
		return fmt.Errorf("invalid mode")
	}

	// Parse operation (+/-/=)
	op := modeStr[i]
	i++

	// Parse permissions (r/w/x)
	var perms os.FileMode
	for i < len(modeStr) {
		switch modeStr[i] {
		case 'r':
			perms |= 0444
		case 'w':
			perms |= 0222
		case 'x':
			perms |= 0111
		}
		i++
	}

	// Apply changes
	switch op {
	case '+':
		if user {
			*mode |= (perms & 0700)
		}
		if group {
			*mode |= (perms & 0070)
		}
		if other {
			*mode |= (perms & 0007)
		}
	case '-':
		if user {
			*mode &^= (perms & 0700)
		}
		if group {
			*mode &^= (perms & 0070)
		}
		if other {
			*mode &^= (perms & 0007)
		}
	case '=':
		// Clear and set
		if user {
			*mode &^= 0700
			*mode |= (perms & 0700)
		}
		if group {
			*mode &^= 0070
			*mode |= (perms & 0070)
		}
		if other {
			*mode &^= 0007
			*mode |= (perms & 0007)
		}
	}

	return nil
}

// chmodFile changes permissions of a file
func chmodFile(path string, mode os.FileMode, recursive bool) error {
	if recursive {
		return chmodRecursive(path, mode)
	}
	return os.Chmod(path, mode)
}

// chmodRecursive changes permissions recursively
func chmodRecursive(path string, mode os.FileMode) error {
	err := os.Chmod(path, mode)
	if err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			subPath := path + "/" + entry.Name()
			if err := chmodRecursive(subPath, mode); err != nil {
				return err
			}
		}
	}

	return nil
}

// Chown changes file ownership (like chown command)
func Chown(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("chown: missing operand")
	}

	var recursive bool
	var owner string
	var files []string

	// Parse arguments
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'R':
					recursive = true
				}
			}
		} else {
			if owner == "" {
				owner = arg
			} else {
				files = append(files, args[i:]...)
				break
			}
		}
	}

	if owner == "" || len(files) == 0 {
		return fmt.Errorf("chown: missing operand")
	}

	// Parse owner:group
	var uid, gid int = -1, -1

	parts := strings.Split(owner, ":")
	if len(parts) >= 1 && parts[0] != "" {
		if u, err := user.Lookup(parts[0]); err == nil {
			if id, err := strconv.Atoi(u.Uid); err == nil {
				uid = id
			}
		} else if id, err := strconv.Atoi(parts[0]); err == nil {
			uid = id
		}
	}

	if len(parts) >= 2 && parts[1] != "" {
		if g, err := user.LookupGroup(parts[1]); err == nil {
			if id, err := strconv.Atoi(g.Gid); err == nil {
				gid = id
			}
		} else if id, err := strconv.Atoi(parts[1]); err == nil {
			gid = id
		}
	}

	// Apply to files
	for _, file := range files {
		if err := chownFile(file, uid, gid, recursive); err != nil {
			fmt.Printf("chown: %v\n", err)
		}
	}

	return nil
}

// chownFile changes ownership of a file
func chownFile(path string, uid, gid int, recursive bool) error {
	if recursive {
		return chownRecursive(path, uid, gid)
	}
	return os.Chown(path, uid, gid)
}

// chownRecursive changes ownership recursively
func chownRecursive(path string, uid, gid int) error {
	err := os.Chown(path, uid, gid)
	if err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			subPath := path + "/" + entry.Name()
			if err := chownRecursive(subPath, uid, gid); err != nil {
				return err
			}
		}
	}

	return nil
}

// Chgrp changes group ownership (like chgrp command)
func Chgrp(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("chgrp: missing operand")
	}

	var recursive bool
	var group string
	var files []string

	// Parse arguments
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'R':
					recursive = true
				}
			}
		} else {
			if group == "" {
				group = arg
			} else {
				files = append(files, args[i:]...)
				break
			}
		}
	}

	if group == "" || len(files) == 0 {
		return fmt.Errorf("chgrp: missing operand")
	}

	// Parse group
	var gid int = -1
	if g, err := user.LookupGroup(group); err == nil {
		if id, err := strconv.Atoi(g.Gid); err == nil {
			gid = id
		}
	} else if id, err := strconv.Atoi(group); err == nil {
		gid = id
	}

	if gid == -1 {
		return fmt.Errorf("chgrp: invalid group: %s", group)
	}

	// Apply to files
	for _, file := range files {
		if err := chownFile(file, -1, gid, recursive); err != nil {
			fmt.Printf("chgrp: %v\n", err)
		}
	}

	return nil
}

package builtin

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Ps shows running processes (simplified version)
func Ps(args []string) error {
	var showAll bool
	var showUser bool

	// Parse flags
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'a':
					showAll = true
				case 'u':
					showUser = true
				case 'x':
					showAll = true
				}
			}
		}
	}

	// Read from /proc directory
	procDir := "/proc"
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return fmt.Errorf("ps: cannot read /proc: %v", err)
	}

	if showUser {
		fmt.Printf("%-8s %-8s %-8s %-8s %-8s %s\n", "USER", "PID", "CPU%", "MEM%", "TIME", "COMMAND")
	} else {
		fmt.Printf("%-8s %-8s %s\n", "PID", "TTY", "CMD")
	}

	for _, entry := range entries {
		// Check if directory name is a number (PID)
		if pid, err := strconv.Atoi(entry.Name()); err == nil {
			if err := showProcess(pid, showAll, showUser); err == nil {
				// Process shown successfully
			}
		}
	}

	return nil
}

// showProcess displays information about a process
func showProcess(pid int, showAll, showUser bool) error {
	procPath := fmt.Sprintf("/proc/%d", pid)

	// Read command line
	cmdlineFile := filepath.Join(procPath, "cmdline")
	cmdlineData, err := os.ReadFile(cmdlineFile)
	if err != nil {
		return err
	}

	cmdline := strings.ReplaceAll(string(cmdlineData), "\x00", " ")
	if cmdline == "" {
		cmdline = fmt.Sprintf("[%d]", pid)
	}

	if showUser {
		// Simplified user format
		fmt.Printf("%-8s %-8d %-8s %-8s %-8s %s\n",
			"user", pid, "0.0", "0.0", "00:00:00", cmdline)
	} else {
		fmt.Printf("%-8d %-8s %s\n", pid, "?", cmdline)
	}

	return nil
}

// Kill sends signals to processes (like kill command)
func Kill(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("kill: missing operand")
	}

	signal := syscall.SIGTERM // default signal
	var pids []int

	// Parse arguments
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			// Parse signal
			sigStr := arg[1:]
			switch sigStr {
			case "9", "KILL":
				signal = syscall.SIGKILL
			case "15", "TERM":
				signal = syscall.SIGTERM
			case "1", "HUP":
				signal = syscall.SIGHUP
			case "2", "INT":
				signal = syscall.SIGINT
			default:
				return fmt.Errorf("kill: invalid signal: %s", sigStr)
			}
		} else {
			// Parse PIDs
			for _, pidStr := range args[i:] {
				pid, err := strconv.Atoi(pidStr)
				if err != nil {
					fmt.Printf("kill: invalid PID: %s\n", pidStr)
					continue
				}
				pids = append(pids, pid)
			}
			break
		}
	}

	if len(pids) == 0 {
		return fmt.Errorf("kill: missing PID")
	}

	for _, pid := range pids {
		if err := syscall.Kill(pid, signal); err != nil {
			fmt.Printf("kill: cannot kill %d: %v\n", pid, err)
		}
	}

	return nil
}

// Df shows filesystem disk space usage (like df command)
func Df(args []string) error {
	var humanReadable bool
	var paths []string

	// Parse flags
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'h':
					humanReadable = true
				}
			}
		} else {
			paths = append(paths, args[i:]...)
			break
		}
	}

	if len(paths) == 0 {
		paths = []string{"/"}
	}

	if humanReadable {
		fmt.Printf("%-20s %-8s %-8s %-8s %-5s %s\n",
			"Filesystem", "Size", "Used", "Avail", "Use%", "Mounted on")
	} else {
		fmt.Printf("%-20s %-12s %-12s %-12s %-5s %s\n",
			"Filesystem", "1K-blocks", "Used", "Available", "Use%", "Mounted on")
	}

	for _, path := range paths {
		if err := showDiskUsage(path, humanReadable); err != nil {
			fmt.Printf("df: %v\n", err)
		}
	}

	return nil
}

// showDiskUsage displays disk usage for a path
func showDiskUsage(path string, humanReadable bool) error {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return err
	}

	blockSize := uint64(stat.Bsize)
	totalBlocks := stat.Blocks
	freeBlocks := stat.Bavail
	usedBlocks := totalBlocks - stat.Bfree

	total := totalBlocks * blockSize
	used := usedBlocks * blockSize
	available := freeBlocks * blockSize

	var usePercent int
	if totalBlocks > 0 {
		usePercent = int((usedBlocks * 100) / totalBlocks)
	}

	if humanReadable {
		fmt.Printf("%-20s %-8s %-8s %-8s %4d%% %s\n",
			"filesystem",
			formatHumanReadable(int64(total)),
			formatHumanReadable(int64(used)),
			formatHumanReadable(int64(available)),
			usePercent,
			path)
	} else {
		fmt.Printf("%-20s %-12d %-12d %-12d %4d%% %s\n",
			"filesystem",
			total/1024,
			used/1024,
			available/1024,
			usePercent,
			path)
	}

	return nil
}

// Du shows directory disk usage (like du command)
func Du(args []string) error {
	var humanReadable bool
	var summarize bool
	var paths []string

	// Parse flags
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'h':
					humanReadable = true
				case 's':
					summarize = true
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
		if err := showDirectoryUsage(path, humanReadable, summarize); err != nil {
			fmt.Printf("du: %v\n", err)
		}
	}

	return nil
}

// showDirectoryUsage displays directory usage
func showDirectoryUsage(path string, humanReadable, summarize bool) error {
	totalSize, err := calculateDirSize(path)
	if err != nil {
		return err
	}

	if humanReadable {
		fmt.Printf("%s\t%s\n", formatHumanReadable(totalSize), path)
	} else {
		fmt.Printf("%d\t%s\n", totalSize/1024, path) // in KB
	}

	return nil
}

// calculateDirSize calculates total size of directory
func calculateDirSize(path string) (int64, error) {
	var totalSize int64

	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return nil // skip errors
			}
			totalSize += info.Size()
		}

		return nil
	})

	return totalSize, err
}

// Free shows memory usage (like free command)
func Free(args []string) error {
	var humanReadable bool

	// Parse flags
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'h':
					humanReadable = true
				}
			}
		}
	}

	return showMemoryUsage(humanReadable)
}

// showMemoryUsage displays memory usage information
func showMemoryUsage(humanReadable bool) error {
	// Read /proc/meminfo
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return fmt.Errorf("free: cannot read /proc/meminfo: %v", err)
	}
	defer file.Close()

	memInfo := make(map[string]int64)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			key := strings.TrimSuffix(parts[0], ":")
			if value, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				memInfo[key] = value * 1024 // Convert from KB to bytes
			}
		}
	}

	total := memInfo["MemTotal"]
	free := memInfo["MemFree"]
	available := memInfo["MemAvailable"]
	if available == 0 {
		available = free
	}
	used := total - free

	if humanReadable {
		fmt.Printf("%-12s %-8s %-8s %-8s\n", "", "total", "used", "free")
		fmt.Printf("%-12s %-8s %-8s %-8s\n", "Mem:",
			formatHumanReadable(total),
			formatHumanReadable(used),
			formatHumanReadable(free))
	} else {
		fmt.Printf("%-12s %-12s %-12s %-12s\n", "", "total", "used", "free")
		fmt.Printf("%-12s %-12d %-12d %-12d\n", "Mem:",
			total/1024, used/1024, free/1024) // in KB
	}

	return scanner.Err()
}

// Uptime shows system uptime (like uptime command)
func Uptime(args []string) error {
	// Read /proc/uptime
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return fmt.Errorf("uptime: cannot read /proc/uptime: %v", err)
	}

	parts := strings.Fields(string(data))
	if len(parts) == 0 {
		return fmt.Errorf("uptime: invalid format")
	}

	uptimeSeconds, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return fmt.Errorf("uptime: cannot parse uptime: %v", err)
	}

	duration := time.Duration(uptimeSeconds) * time.Second

	now := time.Now()
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	fmt.Printf(" %s up ", now.Format("15:04:05"))

	if days > 0 {
		fmt.Printf("%d day", days)
		if days > 1 {
			fmt.Print("s")
		}
		fmt.Print(", ")
	}

	if hours > 0 {
		fmt.Printf("%d:%02d, ", hours, minutes)
	} else {
		fmt.Printf("%d min, ", minutes)
	}

	// Get load average (simplified)
	loadData, err := os.ReadFile("/proc/loadavg")
	if err == nil {
		loadParts := strings.Fields(string(loadData))
		if len(loadParts) >= 3 {
			fmt.Printf("load average: %s, %s, %s", loadParts[0], loadParts[1], loadParts[2])
		}
	}

	fmt.Println()
	return nil
}

// Uname shows system information (like uname command)
func Uname(args []string) error {
	var showAll bool
	var showKernel bool
	var showNode bool
	var showRelease bool
	var showVersion bool
	var showMachine bool

	// Parse flags
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'a':
					showAll = true
				case 's':
					showKernel = true
				case 'n':
					showNode = true
				case 'r':
					showRelease = true
				case 'v':
					showVersion = true
				case 'm':
					showMachine = true
				}
			}
		}
	}

	// Default behavior
	if !showAll && !showKernel && !showNode && !showRelease && !showVersion && !showMachine {
		showKernel = true
	}

	if showAll {
		showKernel = true
		showNode = true
		showRelease = true
		showVersion = true
		showMachine = true
	}

	var parts []string

	if showKernel {
		parts = append(parts, runtime.GOOS)
	}

	if showNode {
		hostname, _ := os.Hostname()
		parts = append(parts, hostname)
	}

	if showRelease {
		// Try to read kernel release
		if data, err := os.ReadFile("/proc/version"); err == nil {
			version := strings.Fields(string(data))
			if len(version) >= 3 {
				parts = append(parts, version[2])
			}
		} else {
			parts = append(parts, "unknown")
		}
	}

	if showVersion {
		parts = append(parts, "#1 SMP")
	}

	if showMachine {
		parts = append(parts, runtime.GOARCH)
	}

	fmt.Println(strings.Join(parts, " "))
	return nil
}

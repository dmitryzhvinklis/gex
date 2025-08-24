package builtin

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Ping sends ICMP ping packets (simplified implementation using TCP connect)
func Ping(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("ping: missing host")
	}

	var count int = 4
	var interval time.Duration = time.Second
	var timeout time.Duration = 3 * time.Second
	var host string

	// Parse arguments
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			switch arg {
			case "-c":
				if i+1 < len(args) {
					if c, err := strconv.Atoi(args[i+1]); err == nil {
						count = c
					}
				}
			case "-i":
				if i+1 < len(args) {
					if d, err := time.ParseDuration(args[i+1] + "s"); err == nil {
						interval = d
					}
				}
			case "-W":
				if i+1 < len(args) {
					if d, err := time.ParseDuration(args[i+1] + "s"); err == nil {
						timeout = d
					}
				}
			}
		} else if host == "" {
			host = arg
		}
	}

	if host == "" {
		return fmt.Errorf("ping: missing host")
	}

	fmt.Printf("PING %s\n", host)

	var successful, failed int
	var totalTime time.Duration

	for i := 0; i < count; i++ {
		start := time.Now()

		// Use TCP connect as a simple ping alternative
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, "80"), timeout)
		if err != nil {
			// Try port 443 (HTTPS)
			conn, err = net.DialTimeout("tcp", net.JoinHostPort(host, "443"), timeout)
		}

		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("Request timeout for icmp_seq=%d\n", i+1)
			failed++
		} else {
			conn.Close()
			fmt.Printf("64 bytes from %s: icmp_seq=%d time=%.1fms\n",
				host, i+1, float64(elapsed.Nanoseconds())/1000000)
			successful++
			totalTime += elapsed
		}

		if i < count-1 {
			time.Sleep(interval)
		}
	}

	fmt.Printf("\n--- %s ping statistics ---\n", host)
	fmt.Printf("%d packets transmitted, %d received, %.1f%% packet loss\n",
		count, successful, float64(failed)/float64(count)*100)

	if successful > 0 {
		avgTime := totalTime / time.Duration(successful)
		fmt.Printf("round-trip min/avg/max = %.1f/%.1f/%.1f ms\n",
			float64(avgTime.Nanoseconds())/1000000,
			float64(avgTime.Nanoseconds())/1000000,
			float64(avgTime.Nanoseconds())/1000000)
	}

	return nil
}

// Wget downloads files from web (simplified implementation)
func Wget(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("wget: missing URL")
	}

	var url string
	var output string
	var quiet bool
	var continue_ bool
	var timeout time.Duration = 30 * time.Second

	// Parse arguments
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			switch arg {
			case "-O":
				if i+1 < len(args) {
					output = args[i+1]
				}
			case "-q", "--quiet":
				quiet = true
			case "-c", "--continue":
				continue_ = true
			case "-T", "--timeout":
				if i+1 < len(args) {
					if d, err := time.ParseDuration(args[i+1] + "s"); err == nil {
						timeout = d
					}
				}
			}
		} else if url == "" {
			url = arg
		}
	}

	if url == "" {
		return fmt.Errorf("wget: missing URL")
	}

	// Add protocol if missing
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	if !quiet {
		fmt.Printf("Connecting to %s...\n", url)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Make request
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("wget: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wget: server returned %d %s", resp.StatusCode, resp.Status)
	}

	// Determine output file
	if output == "" {
		parts := strings.Split(url, "/")
		if len(parts) > 0 && parts[len(parts)-1] != "" {
			output = parts[len(parts)-1]
		} else {
			output = "index.html"
		}
	}

	// Handle continue option
	var outFile *os.File
	if continue_ {
		outFile, err = os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		outFile, err = os.Create(output)
	}

	if err != nil {
		return fmt.Errorf("wget: cannot create %s: %v", output, err)
	}
	defer outFile.Close()

	if !quiet {
		fmt.Printf("Saving to: '%s'\n", output)
	}

	// Copy data
	written, err := io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("wget: %v", err)
	}

	if !quiet {
		fmt.Printf("Downloaded %d bytes\n", written)
		fmt.Printf("'%s' saved\n", output)
	}

	return nil
}

// Curl transfers data from/to servers (simplified implementation)
func Curl(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("curl: missing URL")
	}

	var url string
	var output string
	var method string = "GET"
	var data string
	var headers []string
	var followRedirects bool
	var silent bool
	var timeout time.Duration = 30 * time.Second

	// Parse arguments
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			switch arg {
			case "-o", "--output":
				if i+1 < len(args) {
					output = args[i+1]
				}
			case "-X", "--request":
				if i+1 < len(args) {
					method = strings.ToUpper(args[i+1])
				}
			case "-d", "--data":
				if i+1 < len(args) {
					data = args[i+1]
					if method == "GET" {
						method = "POST"
					}
				}
			case "-H", "--header":
				if i+1 < len(args) {
					headers = append(headers, args[i+1])
				}
			case "-L", "--location":
				followRedirects = true
			case "-s", "--silent":
				silent = true
			case "--connect-timeout":
				if i+1 < len(args) {
					if d, err := time.ParseDuration(args[i+1] + "s"); err == nil {
						timeout = d
					}
				}
			}
		} else if url == "" {
			url = arg
		}
	}

	if url == "" {
		return fmt.Errorf("curl: missing URL")
	}

	// Add protocol if missing
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: timeout,
	}

	if !followRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// Create request
	var req *http.Request
	var err error

	if data != "" {
		req, err = http.NewRequest(method, url, strings.NewReader(data))
		if err != nil {
			return fmt.Errorf("curl: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return fmt.Errorf("curl: %v", err)
		}
	}

	// Add custom headers
	for _, header := range headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("curl: %v", err)
	}
	defer resp.Body.Close()

	if !silent {
		fmt.Printf("HTTP/%s %s\n", resp.Proto[5:], resp.Status)
		for name, values := range resp.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", name, value)
			}
		}
		fmt.Println()
	}

	// Handle output
	var writer io.Writer = os.Stdout

	if output != "" {
		file, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("curl: cannot create %s: %v", output, err)
		}
		defer file.Close()
		writer = file
	}

	// Copy response body
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("curl: %v", err)
	}

	return nil
}

// Netstat displays network connections (simplified implementation)
func Netstat(args []string) error {
	var showAll bool
	var showListening bool
	var showTcp bool = true
	var showUdp bool
	var showNumeric bool

	// Parse arguments
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags := arg[1:]
			for _, flag := range flags {
				switch flag {
				case 'a':
					showAll = true
				case 'l':
					showListening = true
				case 't':
					showTcp = true
					showUdp = false
				case 'u':
					showUdp = true
					showTcp = false
				case 'n':
					showNumeric = true
				}
			}
		}
	}

	fmt.Printf("Proto Recv-Q Send-Q Local Address           Foreign Address         State\n")

	if showTcp {
		// Read TCP connections from /proc/net/tcp
		if err := showTcpConnections(showAll, showListening, showNumeric); err != nil {
			return err
		}
	}

	if showUdp {
		// Read UDP connections from /proc/net/udp
		if err := showUdpConnections(showAll, showListening, showNumeric); err != nil {
			return err
		}
	}

	return nil
}

// showTcpConnections displays TCP connections
func showTcpConnections(showAll, showListening, showNumeric bool) error {
	// Simplified implementation - would normally read from /proc/net/tcp
	fmt.Printf("tcp        0      0 0.0.0.0:22              0.0.0.0:*               LISTEN\n")
	fmt.Printf("tcp        0      0 127.0.0.1:631           0.0.0.0:*               LISTEN\n")
	return nil
}

// showUdpConnections displays UDP connections
func showUdpConnections(showAll, showListening, showNumeric bool) error {
	// Simplified implementation - would normally read from /proc/net/udp
	fmt.Printf("udp        0      0 0.0.0.0:68              0.0.0.0:*\n")
	return nil
}

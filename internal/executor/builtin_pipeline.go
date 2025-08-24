package executor

import (
	"fmt"
	"io"
	"os"

	"gex/internal/cli"
)

// executeBuiltinPipeline executes a pipeline with built-in commands
func (e *Executor) executeBuiltinPipeline(commands []*cli.Command) error {
	if len(commands) == 0 {
		return fmt.Errorf("empty pipeline")
	}

	// Create pipes between commands
	var pipes []*os.File
	var readers []*os.File

	for i := 0; i < len(commands)-1; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			return err
		}
		pipes = append(pipes, w)
		readers = append(readers, r)
	}

	// Defer closing all pipes
	defer func() {
		for _, p := range pipes {
			p.Close()
		}
		for _, r := range readers {
			r.Close()
		}
	}()

	// Execute each command in the pipeline
	for i, command := range commands {
		var stdin io.Reader = os.Stdin
		var stdout io.Writer = os.Stdout

		// Set up input
		if i > 0 {
			stdin = readers[i-1]
		}

		// Set up output
		if i < len(commands)-1 {
			stdout = pipes[i]
		}

		// Execute the command with redirected I/O
		if err := e.executeBuiltinWithIO(command, stdin, stdout, os.Stderr); err != nil {
			return err
		}
	}

	return nil
}

// executeBuiltinWithIO executes a built-in command with custom I/O
func (e *Executor) executeBuiltinWithIO(cmd *cli.Command, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	// Save original I/O
	origStdin := os.Stdin
	origStdout := os.Stdout
	origStderr := os.Stderr

	// Create temporary files for I/O redirection
	stdinFile, stdinCleanup := createTempReader(stdin)
	stdoutFile, stdoutCleanup := createTempWriter(stdout)
	stderrFile, stderrCleanup := createTempWriter(stderr)

	defer func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
		os.Stderr = origStderr
		stdinCleanup()
		stdoutCleanup()
		stderrCleanup()
	}()

	// Redirect I/O
	os.Stdin = stdinFile
	os.Stdout = stdoutFile
	os.Stderr = stderrFile

	// Execute the built-in command
	return e.executeBuiltin(cmd)
}

// createTempReader creates a temporary file for reading
func createTempReader(reader io.Reader) (*os.File, func()) {
	if file, ok := reader.(*os.File); ok {
		return file, func() {}
	}

	// Create pipe for non-file readers
	r, w, err := os.Pipe()
	if err != nil {
		return os.Stdin, func() {}
	}

	go func() {
		defer w.Close()
		io.Copy(w, reader)
	}()

	return r, func() { r.Close() }
}

// createTempWriter creates a temporary file for writing
func createTempWriter(writer io.Writer) (*os.File, func()) {
	if file, ok := writer.(*os.File); ok {
		return file, func() {}
	}

	// Create pipe for non-file writers
	r, w, err := os.Pipe()
	if err != nil {
		return os.Stdout, func() {}
	}

	go func() {
		defer r.Close()
		io.Copy(writer, r)
	}()

	return w, func() { w.Close() }
}

// hasBuiltinCommand checks if any command in the pipeline is built-in
func hasBuiltinCommand(commands []*cli.Command) bool {
	for _, cmd := range commands {
		if cli.IsBuiltin(cmd.Name) {
			return true
		}
	}
	return false
}

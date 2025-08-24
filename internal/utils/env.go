package utils

import (
	"os"
	"strings"
)

// ExpandVariables expands environment variables in a string
// Supports both $VAR and ${VAR} syntax
func ExpandVariables(input string) string {
	if input == "" {
		return input
	}

	var result strings.Builder
	result.Grow(len(input) * 2) // Pre-allocate for performance

	i := 0
	for i < len(input) {
		if input[i] == '$' && i+1 < len(input) {
			if input[i+1] == '{' {
				// Handle ${VAR} syntax
				end := strings.Index(input[i+2:], "}")
				if end != -1 {
					varName := input[i+2 : i+2+end]
					value := os.Getenv(varName)
					result.WriteString(value)
					i = i + 3 + end
					continue
				}
			} else if isVarChar(input[i+1]) {
				// Handle $VAR syntax
				start := i + 1
				end := start
				for end < len(input) && isVarChar(input[end]) {
					end++
				}
				varName := input[start:end]
				value := os.Getenv(varName)
				result.WriteString(value)
				i = end
				continue
			}
		}

		result.WriteByte(input[i])
		i++
	}

	return result.String()
}

// isVarChar checks if a character is valid for a variable name
func isVarChar(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_'
}

// SetEnvVar sets an environment variable
func SetEnvVar(name, value string) error {
	return os.Setenv(name, value)
}

// GetEnvVar gets an environment variable with default value
func GetEnvVar(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}

// UnsetEnvVar unsets an environment variable
func UnsetEnvVar(name string) error {
	return os.Unsetenv(name)
}

// GetAllEnvVars returns all environment variables as a map
func GetAllEnvVars() map[string]string {
	env := make(map[string]string)
	for _, pair := range os.Environ() {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	return env
}

// ExpandPath expands ~ to home directory in paths
func ExpandPath(path string) string {
	if path == "" || path[0] != '~' {
		return path
	}

	if path == "~" {
		return GetEnvVar("HOME", "")
	}

	if len(path) > 1 && path[1] == '/' {
		home := GetEnvVar("HOME", "")
		if home != "" {
			return home + path[1:]
		}
	}

	return path
}

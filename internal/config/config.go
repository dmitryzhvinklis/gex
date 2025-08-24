package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents shell configuration
type Config struct {
	HistoryLimit   int               `json:"history_limit"`
	Prompt         string            `json:"prompt"`
	Aliases        map[string]string `json:"aliases"`
	AutoComplete   bool              `json:"auto_complete"`
	ColorOutput    bool              `json:"color_output"`
	TabCompletion  bool              `json:"tab_completion"`
	HistorySearch  bool              `json:"history_search"`
	CaseSensitive  bool              `json:"case_sensitive"`
	MaxJobs        int               `json:"max_jobs"`
	TimeoutSeconds int               `json:"timeout_seconds"`
}

// Default configuration
var defaultConfig = Config{
	HistoryLimit:   1000,
	Prompt:         "gex> ",
	Aliases:        make(map[string]string),
	AutoComplete:   true,
	ColorOutput:    true,
	TabCompletion:  true,
	HistorySearch:  true,
	CaseSensitive:  false,
	MaxJobs:        10,
	TimeoutSeconds: 30,
}

// New creates a new configuration with defaults
func New() *Config {
	cfg := defaultConfig
	cfg.Aliases = make(map[string]string)

	// Copy default aliases
	for k, v := range defaultConfig.Aliases {
		cfg.Aliases[k] = v
	}

	return &cfg
}

// Load loads configuration from file
func Load(path string) (*Config, error) {
	// If file doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return New(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Ensure aliases map is initialized
	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]string)
	}

	// Set defaults for unspecified values
	if cfg.HistoryLimit == 0 {
		cfg.HistoryLimit = defaultConfig.HistoryLimit
	}
	if cfg.Prompt == "" {
		cfg.Prompt = defaultConfig.Prompt
	}
	if cfg.MaxJobs == 0 {
		cfg.MaxJobs = defaultConfig.MaxJobs
	}
	if cfg.TimeoutSeconds == 0 {
		cfg.TimeoutSeconds = defaultConfig.TimeoutSeconds
	}

	return &cfg, nil
}

// Save saves configuration to file
func (c *Config) Save(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetConfigPath returns the default configuration file path
func GetConfigPath() string {
	home := os.Getenv("HOME")
	if home == "" {
		return ".gexrc"
	}
	return filepath.Join(home, ".gexrc")
}

// LoadDefault loads configuration from the default location
func LoadDefault() (*Config, error) {
	return Load(GetConfigPath())
}

// SaveDefault saves configuration to the default location
func (c *Config) SaveDefault() error {
	return c.Save(GetConfigPath())
}

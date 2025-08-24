package shell

import (
	"os"
	"sync"
)

// Session manages shell state and history
type Session struct {
	workingDir   string
	previousDir  string
	history      []string
	aliases      map[string]string
	variables    map[string]string
	mutex        sync.RWMutex
	historyLimit int
}

// NewSession creates a new shell session
func NewSession(cfg interface{}) *Session {
	wd, _ := os.Getwd()

	return &Session{
		workingDir:   wd,
		previousDir:  "",
		history:      make([]string, 0),
		aliases:      make(map[string]string),
		variables:    make(map[string]string),
		historyLimit: 1000, // Default history limit
	}
}

// Working Directory Management
func (s *Session) GetWorkingDir() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.workingDir
}

func (s *Session) SetWorkingDir(dir string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.workingDir = dir
}

func (s *Session) GetPreviousDir() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.previousDir
}

func (s *Session) SetPreviousDir(dir string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.previousDir = dir
}

// History Management
func (s *Session) AddHistory(cmd string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Don't add empty commands or duplicates of the last command
	if cmd == "" || (len(s.history) > 0 && s.history[len(s.history)-1] == cmd) {
		return
	}

	s.history = append(s.history, cmd)

	// Limit history size for performance
	if len(s.history) > s.historyLimit {
		// Remove oldest entries
		copy(s.history, s.history[len(s.history)-s.historyLimit:])
		s.history = s.history[:s.historyLimit]
	}
}

func (s *Session) GetHistory() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy to prevent external modification
	result := make([]string, len(s.history))
	copy(result, s.history)
	return result
}

func (s *Session) GetHistoryEntry(index int) string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if index < 0 || index >= len(s.history) {
		return ""
	}
	return s.history[index]
}

func (s *Session) GetHistorySize() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.history)
}

// Alias Management
func (s *Session) SetAlias(name, value string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.aliases[name] = value
}

func (s *Session) GetAliases() map[string]string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]string)
	for k, v := range s.aliases {
		result[k] = v
	}
	return result
}

func (s *Session) RemoveAlias(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.aliases, name)
}

// Variable Management
func (s *Session) SetVariable(name, value string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.variables[name] = value
}

func (s *Session) GetVariable(name string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	value, exists := s.variables[name]
	return value, exists
}

func (s *Session) GetVariables() map[string]string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]string)
	for k, v := range s.variables {
		result[k] = v
	}
	return result
}

func (s *Session) RemoveVariable(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.variables, name)
}

// Configuration
func (s *Session) SetHistoryLimit(limit int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if limit > 0 {
		s.historyLimit = limit

		// Truncate current history if needed
		if len(s.history) > limit {
			copy(s.history, s.history[len(s.history)-limit:])
			s.history = s.history[:limit]
		}
	}
}

func (s *Session) GetHistoryLimit() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.historyLimit
}

package session

import (
	"encoding/json"
	"fmt"
	"hama-shell/pkg/types"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

// Persistence implements the SessionPersistence interface for file-based persistence
type Persistence struct {
	mu      sync.RWMutex
	dataDir string
}

// NewPersistence creates a new session persistence manager
func NewPersistence(dataDir string) (*Persistence, error) {
	if dataDir == "" {
		return nil, fmt.Errorf("data directory cannot be empty")
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &Persistence{
		dataDir: dataDir,
	}, nil
}

// Store stores a session to persistent storage
func (p *Persistence) Store(session *types.Session) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	if session.ID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	filePath := p.getSessionFilePath(session.ID)

	// Marshal session to JSON
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Write to file atomically using a temporary file
	tempFile := filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	// Atomic move
	if err := os.Rename(tempFile, filePath); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to move session file: %w", err)
	}

	return nil
}

// Retrieve retrieves a session from persistent storage
func (p *Persistence) Retrieve(sessionID string) (*types.Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	filePath := p.getSessionFilePath(sessionID)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	// Unmarshal JSON
	var session types.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// Remove removes a session from persistent storage
func (p *Persistence) Remove(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	filePath := p.getSessionFilePath(sessionID)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Remove file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove session file: %w", err)
	}

	return nil
}

// List returns all persisted session IDs
func (p *Persistence) List() ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	entries, err := os.ReadDir(p.dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	var sessionIDs []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) == ".json" {
			sessionID := name[:len(name)-5] // Remove .json extension
			sessionIDs = append(sessionIDs, sessionID)
		}
	}

	return sessionIDs, nil
}

// ListSessions returns all persisted sessions
func (p *Persistence) ListSessions() ([]*types.Session, error) {
	sessionIDs, err := p.List()
	if err != nil {
		return nil, err
	}

	var sessions []*types.Session
	for _, sessionID := range sessionIDs {
		session, err := p.Retrieve(sessionID)
		if err != nil {
			// Log error but continue with other sessions
			fmt.Printf("Warning: failed to retrieve session %s: %v\n", sessionID, err)
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Exists checks if a session exists in persistent storage
func (p *Persistence) Exists(sessionID string) bool {
	if sessionID == "" {
		return false
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	filePath := p.getSessionFilePath(sessionID)
	_, err := os.Stat(filePath)
	return err == nil
}

// Clean removes all session files from persistent storage
func (p *Persistence) Clean() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	entries, err := os.ReadDir(p.dataDir)
	if err != nil {
		return fmt.Errorf("failed to read data directory: %w", err)
	}

	var errors []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) == ".json" {
			filePath := filepath.Join(p.dataDir, name)
			if err := os.Remove(filePath); err != nil {
				errors = append(errors, fmt.Errorf("failed to remove %s: %w", name, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to clean some session files: %v", errors)
	}

	return nil
}

// GetDataDir returns the data directory path
func (p *Persistence) GetDataDir() string {
	return p.dataDir
}

// getSessionFilePath returns the file path for a session
func (p *Persistence) getSessionFilePath(sessionID string) string {
	return filepath.Join(p.dataDir, sessionID+".json")
}

// Stats returns statistics about persisted sessions
func (p *Persistence) Stats() (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	entries, err := os.ReadDir(p.dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	var totalSize int64
	var fileCount int

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) == ".json" {
			fileCount++

			info, err := entry.Info()
			if err == nil {
				totalSize += info.Size()
			}
		}
	}

	stats := map[string]interface{}{
		"data_dir":         p.dataDir,
		"session_count":    fileCount,
		"total_size_bytes": totalSize,
	}

	// Add disk usage if possible
	if stat, err := os.Stat(p.dataDir); err == nil {
		if statT, ok := stat.Sys().(*fs.FileInfo); ok {
			_ = statT // Use statT for platform-specific disk usage if needed
		}
	}

	return stats, nil
}

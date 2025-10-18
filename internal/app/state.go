// Package app provides core application functionality
package app

import (
	"sync"

	"theboys-launcher/internal/config"
	"theboys-launcher/internal/logging"
)

// State represents the application state
type State struct {
	// Configuration and logging
	Config *config.Config
	Logger *logging.Logger

	// UI state
	CurrentScreen     string
	IsLaunching       bool
	CurrentModpackID  string

	// Download state
	ActiveDownloads   map[string]*DownloadState
	DownloadMutex     sync.RWMutex

	// Application state mutex
	stateMutex        sync.RWMutex
}

// DownloadState represents the state of a download operation
type DownloadState struct {
	ID           string
	URL          string
	Progress     float64
	Speed        int64 // bytes per second
	TotalSize    int64
	Downloaded   int64
	Status       string // "downloading", "paused", "completed", "error"
	Error        error
	Cancel       chan struct{}
}

// NewState creates a new application state
func NewState(cfg *config.Config, logger *logging.Logger) *State {
	return &State{
		Config:           cfg,
		Logger:           logger,
		CurrentScreen:    "main",
		ActiveDownloads:  make(map[string]*DownloadState),
	}
}

// SetCurrentScreen sets the current screen
func (s *State) SetCurrentScreen(screen string) {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()
	s.CurrentScreen = screen
}

// GetCurrentScreen returns the current screen
func (s *State) GetCurrentScreen() string {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	return s.CurrentScreen
}

// SetLaunching sets the launching state
func (s *State) SetLaunching(launching bool) {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()
	s.IsLaunching = launching
}

// IsLaunching returns whether the application is currently launching
func (s *State) GetLaunching() bool {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	return s.IsLaunching
}

// SetCurrentModpack sets the current modpack ID
func (s *State) SetCurrentModpack(modpackID string) {
	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()
	s.CurrentModpackID = modpackID
}

// GetCurrentModpack returns the current modpack ID
func (s *State) GetCurrentModpack() string {
	s.stateMutex.RLock()
	defer s.stateMutex.RUnlock()
	return s.CurrentModpackID
}

// AddDownload adds a new download to the active downloads
func (s *State) AddDownload(state *DownloadState) {
	s.DownloadMutex.Lock()
	defer s.DownloadMutex.Unlock()
	s.ActiveDownloads[state.ID] = state
}

// RemoveDownload removes a download from the active downloads
func (s *State) RemoveDownload(id string) {
	s.DownloadMutex.Lock()
	defer s.DownloadMutex.Unlock()
	delete(s.ActiveDownloads, id)
}

// GetDownload returns a download state by ID
func (s *State) GetDownload(id string) (*DownloadState, bool) {
	s.DownloadMutex.RLock()
	defer s.DownloadMutex.RUnlock()
	state, exists := s.ActiveDownloads[id]
	return state, exists
}

// GetAllDownloads returns all active downloads
func (s *State) GetAllDownloads() map[string]*DownloadState {
	s.DownloadMutex.RLock()
	defer s.DownloadMutex.RUnlock()

	result := make(map[string]*DownloadState)
	for id, state := range s.ActiveDownloads {
		result[id] = state
	}
	return result
}

// UpdateDownloadProgress updates the progress of a download
func (s *State) UpdateDownloadProgress(id string, progress float64, downloaded, speed int64) {
	s.DownloadMutex.Lock()
	defer s.DownloadMutex.Unlock()

	if state, exists := s.ActiveDownloads[id]; exists {
		state.Progress = progress
		state.Downloaded = downloaded
		state.Speed = speed
	}
}

// UpdateDownloadStatus updates the status of a download
func (s *State) UpdateDownloadStatus(id string, status string, err error) {
	s.DownloadMutex.Lock()
	defer s.DownloadMutex.Unlock()

	if state, exists := s.ActiveDownloads[id]; exists {
		state.Status = status
		state.Error = err
	}
}
package launcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"theboys-launcher/pkg/types"
	"theboys-launcher/internal/config"
	"theboys-launcher/internal/platform"
	"theboys-launcher/internal/logging"
)

const (
	defaultRemoteModpacksURL = "https://raw.githubusercontent.com/dilllxd/theboys-launcher/refs/heads/main/modpacks.json"
	defaultTimeout          = 30 * time.Second
	userAgentComponent       = "Launcher"
)

// ModpackManager handles modpack operations including loading, validation, and updates
type ModpackManager struct {
	configManager    *config.Manager
	modpackManager   *config.ModpackManager
	platform         platform.Platform
	logger           *logging.Logger
	remoteModpacksURL string
	loadedModpacks   []types.Modpack
}

// NewModpackManager creates a new modpack manager
func NewModpackManager(configManager *config.Manager, modpackManager *config.ModpackManager, platform platform.Platform, logger *logging.Logger) *ModpackManager {
	return &ModpackManager{
		configManager:     configManager,
		modpackManager:    modpackManager,
		platform:          platform,
		logger:            logger,
		remoteModpacksURL: defaultRemoteModpacksURL,
		loadedModpacks:    []types.Modpack{},
	}
}

// LoadModpacks loads modpacks from local configuration and optionally fetches remote updates
func (m *ModpackManager) LoadModpacks(fetchRemote bool) ([]types.Modpack, error) {
	m.logger.Info("Loading modpacks...")

	// First, load local modpacks
	localModpacks, err := m.modpackManager.LoadModpacks()
	if err != nil {
		m.logger.Error("Failed to load local modpacks: %v", err)
		return nil, fmt.Errorf("failed to load local modpacks: %w", err)
	}

	m.loadedModpacks = localModpacks
	m.logger.Info("Loaded %d modpack(s) from local configuration", len(localModpacks))

	// Optionally fetch remote updates
	if fetchRemote {
		if err := m.fetchRemoteModpacks(); err != nil {
			m.logger.Warn("Failed to fetch remote modpacks: %v", err)
			// Continue with local modpacks if remote fetch fails
		}
	}

	// Update default modpack based on settings
	m.updateDefaultModpackID()

	return m.loadedModpacks, nil
}

// GetModpacks returns the currently loaded modpacks
func (m *ModpackManager) GetModpacks() []types.Modpack {
	// Return a copy to prevent external modification
	modpacksCopy := make([]types.Modpack, len(m.loadedModpacks))
	copy(modpacksCopy, m.loadedModpacks)
	return modpacksCopy
}

// SelectModpack selects a modpack by ID or returns the default
func (m *ModpackManager) SelectModpack(requestedID string) (*types.Modpack, error) {
	if len(m.loadedModpacks) == 0 {
		return nil, fmt.Errorf("no modpacks available")
	}

	// If no specific ID requested, use the last selected or default
	if strings.TrimSpace(requestedID) == "" {
		settings := m.configManager.GetSettings()

		// Try to use last selected modpack
		if settings.LastModpackID != "" {
			for _, mp := range m.loadedModpacks {
				if strings.EqualFold(mp.ID, settings.LastModpackID) {
					m.logger.Debug("Selected last used modpack: %s", mp.DisplayName)
					return &mp, nil
				}
			}
		}

		// Fall back to default modpack
		for _, mp := range m.loadedModpacks {
			if mp.Default {
				m.logger.Debug("Selected default modpack: %s", mp.DisplayName)
				return &mp, nil
			}
		}

		// If no default, return first modpack
		if len(m.loadedModpacks) > 0 {
			m.logger.Debug("Selected first available modpack: %s", m.loadedModpacks[0].DisplayName)
			return &m.loadedModpacks[0], nil
		}
	}

	// Find modpack by ID
	id := strings.ToLower(strings.TrimSpace(requestedID))
	for _, mp := range m.loadedModpacks {
		if strings.ToLower(mp.ID) == id {
			m.logger.Debug("Selected modpack by ID: %s", mp.DisplayName)

			// Update last selected modpack in settings
			m.configManager.SetLastModpackID(mp.ID)
			m.configManager.SaveSettings()

			return &mp, nil
		}
	}

	return nil, fmt.Errorf("unknown modpack %q. Available modpacks: %v",
		requestedID, m.getAvailableModpackIDs())
}

// GetModpackByID returns a modpack by its ID
func (m *ModpackManager) GetModpackByID(id string) (*types.Modpack, error) {
	for _, modpack := range m.loadedModpacks {
		if strings.EqualFold(modpack.ID, id) {
			modpackCopy := modpack
			return &modpackCopy, nil
		}
	}
	return nil, fmt.Errorf("modpack with ID '%s' not found", id)
}

// ValidateModpack checks if a modpack configuration is valid
func (m *ModpackManager) ValidateModpack(modpack types.Modpack) error {
	if modpack.ID == "" {
		return fmt.Errorf("modpack ID cannot be empty")
	}
	if modpack.DisplayName == "" {
		return fmt.Errorf("modpack display name cannot be empty")
	}
	if modpack.PackURL == "" {
		return fmt.Errorf("modpack pack URL cannot be empty")
	}
	if modpack.InstanceName == "" {
		return fmt.Errorf("modpack instance name cannot be empty")
	}

	// Optional: Validate URL format
	if !strings.HasPrefix(modpack.PackURL, "http://") && !strings.HasPrefix(modpack.PackURL, "https://") {
		return fmt.Errorf("modpack pack URL must be a valid HTTP/HTTPS URL")
	}

	return nil
}

// GetModpackLabel returns a formatted label for a modpack
func (m *ModpackManager) GetModpackLabel(modpack types.Modpack) string {
	if name := strings.TrimSpace(modpack.DisplayName); name != "" {
		return name
	}
	return modpack.ID
}

// fetchRemoteModpacks fetches modpacks from the remote URL
func (m *ModpackManager) fetchRemoteModpacks() error {
	m.logger.Info("Fetching remote modpacks from %s", m.remoteModpacksURL)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", m.remoteModpacksURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", m.getUserAgent(userAgentComponent))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch remote modpacks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var remoteModpacks []types.Modpack
	if err := json.Unmarshal(body, &remoteModpacks); err != nil {
		return fmt.Errorf("failed to parse remote modpacks: %w", err)
	}

	// Normalize and validate remote modpacks
	normalizedModpacks := m.normalizeModpacks(remoteModpacks)
	if len(normalizedModpacks) == 0 {
		return fmt.Errorf("remote modpacks did not contain any valid modpacks")
	}

	// Merge with local modpacks
	if err := m.mergeModpacks(normalizedModpacks); err != nil {
		return fmt.Errorf("failed to merge remote modpacks: %w", err)
	}

	m.logger.Info("Successfully fetched and merged %d remote modpack(s)", len(normalizedModpacks))
	return nil
}

// normalizeModpacks normalizes a list of modpacks by validating and deduplicating them
func (m *ModpackManager) normalizeModpacks(mods []types.Modpack) []types.Modpack {
	if len(mods) == 0 {
		return nil
	}

	normalized := make([]types.Modpack, 0, len(mods))
	seen := make(map[string]bool, len(mods))

	for _, mod := range mods {
		// Validate modpack
		if err := m.ValidateModpack(mod); err != nil {
			m.logger.Warn("Skipping invalid modpack '%s': %v", mod.ID, err)
			continue
		}

		// Check for duplicates (case-insensitive)
		lowerID := strings.ToLower(mod.ID)
		if seen[lowerID] {
			m.logger.Warn("Skipping duplicate modpack: %s", mod.ID)
			continue
		}

		seen[lowerID] = true
		normalized = append(normalized, mod)
	}

	return normalized
}

// mergeModpacks merges remote modpacks with local configuration
func (m *ModpackManager) mergeModpacks(remoteModpacks []types.Modpack) error {
	// Create a map of existing modpacks by ID
	existing := make(map[string]types.Modpack)
	for _, mp := range m.loadedModpacks {
		existing[strings.ToLower(mp.ID)] = mp
	}

	// Merge remote modpacks
	for _, remoteMod := range remoteModpacks {
		lowerID := strings.ToLower(remoteMod.ID)

		if _, exists := existing[lowerID]; exists {
			// Update existing modpack if remote version is newer
			// For now, we'll replace it completely
			// TODO: Add version comparison logic
			m.logger.Debug("Updating existing modpack: %s", remoteMod.DisplayName)
		} else {
			// Add new modpack
			m.logger.Debug("Adding new remote modpack: %s", remoteMod.DisplayName)
		}

		// Add or update in modpack manager
		if err := m.modpackManager.UpdateModpack(remoteMod); err != nil {
			// If update fails, try to add it
			if err := m.modpackManager.AddModpack(remoteMod); err != nil {
				m.logger.Warn("Failed to add/update modpack '%s': %v", remoteMod.ID, err)
			}
		}
	}

	// Reload modpacks after merging
	updated, err := m.modpackManager.LoadModpacks()
	if err != nil {
		return fmt.Errorf("failed to reload modpacks after merge: %w", err)
	}

	m.loadedModpacks = updated
	return nil
}

// updateDefaultModpackID updates the default modpack based on available modpacks
func (m *ModpackManager) updateDefaultModpackID() {
	if len(m.loadedModpacks) == 0 {
		return
	}

	// Check if we have a default set
	hasDefault := false
	for _, mp := range m.loadedModpacks {
		if mp.Default {
			hasDefault = true
			break
		}
	}

	// If no default is set, make the first modpack the default
	if !hasDefault {
		m.loadedModpacks[0].Default = true
		m.logger.Info("Set %s as default modpack", m.loadedModpacks[0].DisplayName)
	}
}

// getAvailableModpackIDs returns a list of available modpack IDs
func (m *ModpackManager) getAvailableModpackIDs() []string {
	ids := make([]string, len(m.loadedModpacks))
	for i, mp := range m.loadedModpacks {
		ids[i] = mp.ID
	}
	return ids
}

// getUserAgent returns a user agent string with the launcher version and component name
func (m *ModpackManager) getUserAgent(component string) string {
	// For now, use a simple user agent
	// TODO: Integrate with version system when implemented
	return fmt.Sprintf("TheBoys-%s/dev", component)
}

// SetRemoteModpacksURL sets the URL for fetching remote modpacks
func (m *ModpackManager) SetRemoteModpacksURL(url string) {
	m.remoteModpacksURL = url
	m.logger.Debug("Remote modpacks URL set to: %s", url)
}

// RefreshModpacks forces a refresh of modpacks from remote source
func (m *ModpackManager) RefreshModpacks() error {
	m.logger.Info("Refreshing modpacks...")

	// Clear current modpacks
	m.loadedModpacks = []types.Modpack{}

	// Reload with remote fetch enabled
	_, err := m.LoadModpacks(true)
	if err != nil {
		return fmt.Errorf("failed to refresh modpacks: %w", err)
	}

	m.logger.Info("Modpacks refreshed successfully")
	return nil
}
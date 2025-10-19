package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func loadModpacks(root string) []Modpack {
	remote, err := fetchRemoteModpacks(remoteModpacksURL, 30*time.Second)
	if err != nil {
		fail(fmt.Errorf("failed to fetch remote modpacks.json: %w", err))
	}

	if len(remote) == 0 {
		fail(errors.New("remote modpacks.json returned no modpacks"))
	}

	normalized := normalizeModpacks(remote)
	if len(normalized) == 0 {
		fail(errors.New("remote modpacks.json did not contain any valid modpacks"))
	}

	logf("Loaded %d modpack(s) from remote catalog", len(normalized))
	updateDefaultModpackID(normalized)
	return normalized
}

func updateDefaultModpackID(modpacks []Modpack) {
	if len(modpacks) == 0 {
		return
	}
	for _, mp := range modpacks {
		if mp.Default {
			defaultModpackID = mp.ID
			return
		}
	}
	defaultModpackID = modpacks[0].ID
}

func fetchRemoteModpacks(url string, timeout time.Duration) ([]Modpack, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", getUserAgent("Launcher"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var mods []Modpack
	if err := json.Unmarshal(body, &mods); err != nil {
		return nil, err
	}

	return normalizeModpacks(mods), nil
}

func normalizeModpacks(mods []Modpack) []Modpack {
	if len(mods) == 0 {
		return nil
	}

	normalized := make([]Modpack, 0, len(mods))
	index := make(map[string]int, len(mods))

	for _, raw := range mods {
		id := strings.TrimSpace(raw.ID)
		packURL := strings.TrimSpace(raw.PackURL)
		instance := strings.TrimSpace(raw.InstanceName)

		if id == "" || packURL == "" || instance == "" {
			continue
		}

		display := strings.TrimSpace(raw.DisplayName)
		if display == "" {
			display = id
		}

		// Set defaults for new fields if not present
		author := strings.TrimSpace(raw.Author)
		if author == "" {
			author = "Unknown"
		}

		if raw.Tags == nil {
			raw.Tags = []string{}
		}

		if raw.LastUpdated == "" {
			raw.LastUpdated = time.Now().Format(time.RFC3339)
		}

		if raw.Category == "" {
			raw.Category = "general"
		}

		if raw.MinRam <= 0 {
			raw.MinRam = 2048 // Default 2GB minimum
		}

		if raw.RecommendedRam <= 0 {
			raw.RecommendedRam = 4096 // Default 4GB recommended
		}

		if raw.Changelog == "" {
			raw.Changelog = "No changelog available"
		}

		entry := Modpack{
			ID:             id,
			DisplayName:    display,
			PackURL:        packURL,
			InstanceName:   instance,
			Description:    strings.TrimSpace(raw.Description),
			Author:         author,
			Tags:           raw.Tags,
			LastUpdated:    raw.LastUpdated,
			Category:       raw.Category,
			MinRam:         raw.MinRam,
			RecommendedRam: raw.RecommendedRam,
			Changelog:      raw.Changelog,
			Default:        raw.Default,
		}

		key := strings.ToLower(id)
		if idx, ok := index[key]; ok {
			normalized[idx] = entry
		} else {
			index[key] = len(normalized)
			normalized = append(normalized, entry)
		}
	}

	return normalized
}

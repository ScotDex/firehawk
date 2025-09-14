package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type ServerStatus struct {
	Players       int       `json:"players"`
	ServerVersion string    `json:"server_version"`
	StartTime     time.Time `json:"start_time"`
	VIP           bool      `json:"vip,omitempty"`
}

// getAPIStatus is now a method on ESIClient to ensure it uses the correct HTTP client.
func (c *ESIClient) getAPIStatus() (*ServerStatus, error) {
	resp, err := c.httpClient.Get("https://esi.evetech.net/latest/status/?datasource=tranquility&vip")
	if err != nil {
		return nil, fmt.Errorf("failed to make ESI status request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ESI status endpoint returned a non-200 status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading ESI status response body: %w", err)
	}

	var status ServerStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("error unmarshaling ESI status JSON: %w", err)
	}
	return &status, nil
}

type cacheData struct {
	CharacterNames   map[int]string `json:"characterNames"`
	CorporationNames map[int]string `json:"corporationNames"`
	ShipNames        map[int]string `json:"shipNames"`
	SystemNames      map[int]string `json:"systemNames"`

	SearchResults map[string]SearchResponse `json:"searchResults"`
}

// SaveCacheToFile saves the ESI client's in-memory cache to a JSON file.
func (c *ESIClient) SaveCacheToFile(filePath string) error {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	data := cacheData{
		CharacterNames:   c.characterNames,
		CorporationNames: c.corporationNames,
		ShipNames:        c.shipNames,
		SystemNames:      c.systemNames,

		SearchResults: c.searchResults,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data for saving: %w", err)
	}

	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write cache file to %s: %w", filePath, err)
	}

	log.Println("Successfully saved ESI cache to", filePath)
	return nil
}

// LoadCacheFromFile loads the ESI cache from a JSON file into the EsiClient's memory.
func (c *ESIClient) LoadCacheFromFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Println("Cache file not found, starting with an empty cache.")
		return nil
	}

	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read cache file from %s: %w", filePath, err)
	}

	var data cacheData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal cache data from %s: %w", filePath, err)
	}

	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.characterNames = data.CharacterNames
	c.corporationNames = data.CorporationNames
	c.shipNames = data.ShipNames
	c.systemNames = data.SystemNames
	// Corrected field name
	c.searchResults = data.SearchResults

	log.Println("Successfully loaded ESI cache from", filePath)
	return nil
}

// ... (Structs: SearchResponse, Hit, EntityCounts are unchanged) ...
type SearchResponse struct {
	Hits               []Hit        `json:"hits"`
	Query              string       `json:"query"`
	ProcessingTimeMs   int          `json:"processingTimeMs"`
	Limit              int          `json:"limit"`
	Offset             int          `json:"offset"`
	EstimatedTotalHits int          `json:"estimatedTotalHits"`
	EntityCounts       EntityCounts `json:"entityCounts"`
	EntityOrder        []string     `json:"entityOrder"`
	IsExactMatch       bool         `json:"isExactMatch"`
}
type Hit struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Rank       int       `json:"rank"`
	Lang       string    `json:"lang"`
	Deleted    bool      `json:"deleted,omitempty"`
	UpdatedAt  time.Time `json:"updatedAt,omitempty"`
	Ticker     string    `json:"ticker,omitempty"`
	LastActive time.Time `json:"last_active,omitempty"`
}
type EntityCounts struct {
	Items        int `json:"items"`
	Ships        int `json:"ships"`
	Alliances    int `json:"alliances"`
	Corporations int `json:"corporations"`
	Factions     int `json:"factions"`
	Systems      int `json:"systems"`
	Regions      int `json:"regions"`
	Characters   int `json:"characters"`
}

// performSearch is now a method on ESIClient.
func (c *ESIClient) performSearch(searchTerm string) (*SearchResponse, error) {
	baseURL := "https://eve-kill.com/api/search/"
	fullURL := baseURL + url.PathEscape(searchTerm)

	// It now uses the configured client, solving the gzip issue.
	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make search request to %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned non-200 status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read search response body: %w", err)
	}

	var apiResult SearchResponse
	if err := json.Unmarshal(body, &apiResult); err != nil {
		return nil, fmt.Errorf("error unmarshaling search JSON: %w", err)
	}
	return &apiResult, nil
}

// ... (find...InResults functions are unchanged) ...
// findHitByType searches for the first hit matching a specific type.
func findHitByType(response *SearchResponse, hitType string) (Hit, error) {
	log.Printf("Searching for type: '%s'", hitType) // Add this
	for _, hit := range response.Hits {
		log.Printf("...checking hit of type: '%s'", hit.Type) // And this
		// Now checks against the type passed into the function
		if hit.Type == hitType {
			return hit, nil
		}
	}

	// The error message is now dynamic as well
	return Hit{}, fmt.Errorf("no hit of type '%s' found in search results", hitType)
}

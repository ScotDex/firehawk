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
}

func getAPIStatus() (*ServerStatus, error) {
	resp, err := http.Get("https://esi.evetech.net/status?players&server_version&start_time?datasource=tranquility")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var status ServerStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}
	return &status, nil
}

type cacheData struct {
	CharacterNames   map[int]string `json:"characterNames"`
	CorporationNames map[int]string `json:"corporationNames"`
	ShipNames        map[int]string `json:"shipNames"`
	SystemNames      map[int]string `json:"systemNames"`
}

// SaveCacheToFile saves the ESI client's in-memory cache to a JSON file.
func (c *ESIClient) SaveCacheToFile(filePath string) error {
	// Lock the mutex to ensure no other part of the program can change the maps while we save.
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	data := cacheData{
		CharacterNames:   c.characterNames,
		CorporationNames: c.corporationNames,
		ShipNames:        c.shipNames,
		SystemNames:      c.systemNames,
	}

	// Marshal the data into a nicely formatted JSON byte slice.
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	// Write the JSON data to the specified file.
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	log.Println("Successfully saved ESI cache to", filePath)
	return nil
}

// LoadCacheFromFile loads the ESI cache from a JSON file into the EsiClient's memory.
func (c *ESIClient) LoadCacheFromFile(filePath string) error {
	// First, check if the file even exists. If not, it's the first run.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Println("Cache file not found, starting with an empty cache.")
		return nil
	}

	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	var data cacheData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	// Lock the mutex to safely write the loaded data into our maps.
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.characterNames = data.CharacterNames
	c.corporationNames = data.CorporationNames
	c.shipNames = data.ShipNames
	c.systemNames = data.SystemNames

	log.Println("Successfully loaded ESI cache from", filePath)
	return nil
}

// SearchResponse is the top-level object for the entire API response.
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

// Hit represents a single item within the search results.
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

// EntityCounts holds the breakdown of hits per category.
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

func performSearch(SearchTerm string) (*SearchResponse, error) {

	baseUrl := "https://eve-kill.com/api/search/"
	fullURL := baseUrl + url.PathEscape(SearchTerm)

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var searchResult SearchResponse
	if err := json.Unmarshal(body, &searchResult); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return &searchResult, nil
}

func findSystemInResults(response *SearchResponse) (Hit, error) {
	// Loop through each hit in the response's Hits slice.
	for _, hit := range response.Hits {
		// Check if the type is "system".
		if hit.Type == "system" {
			// If it is, we found our match. Return the hit and no error.
			return hit, nil
		}
	}

	// If the loop finishes without finding a system, return an empty Hit and an error.
	return Hit{}, fmt.Errorf("no system found in search results")
}

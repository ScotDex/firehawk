package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type ESIClient struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string

	characterNames   map[int]string
	corporationNames map[int]string
	shipNames        map[int]string
	systemNames      map[int]string
	cacheMutex       sync.RWMutex
	contactInfo      string
}

func NewESIClient() *ESIClient {
	return &ESIClient{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL:   "https://esi.evetech.net/latest",
		userAgent: fmt.Sprintf("Firehawk Discord Bot (%s)"),

		characterNames:   make(map[int]string),
		corporationNames: make(map[int]string),
		shipNames:        make(map[int]string),
		systemNames:      make(map[int]string),
	}
}

type ESINameResponse struct {
	Name string `json:"name"`
}

// getName is a generic, internal helper function to resolve any ID to a name with caching.
func (c *ESIClient) getName(id int, category string, cache map[int]string) string {
	if id == 0 {
		return "Unknown"
	}
	// 1. Check cache first (with a read lock)
	c.cacheMutex.RLock()
	name, found := cache[id]
	c.cacheMutex.RUnlock()
	if found {
		return name
	}

	// 2. Not in cache, make API call
	fullURL := fmt.Sprintf("%s/%s/%d/", c.baseURL, category, id)
	req, _ := http.NewRequest("GET", fullURL, nil)
	req.Header.Set("User-Agent", c.userAgent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "Unknown"
	}
	defer resp.Body.Close()

	var nameData ESINameResponse
	if json.NewDecoder(resp.Body).Decode(&nameData) != nil {
		return "Unknown"
	}

	// 3. Store result in cache (with a write lock)
	c.cacheMutex.Lock()
	cache[id] = nameData.Name
	c.cacheMutex.Unlock()

	return nameData.Name
}

// GetCharacterName resolves a character ID to a name.
func (c *ESIClient) GetCharacterName(id int) string {
	return c.getName(id, "characters", c.characterNames)
}

// GetCorporationName resolves a corporation ID to a name.
func (c *ESIClient) GetCorporationName(id int) string {
	return c.getName(id, "corporations", c.corporationNames)
}

// GetShipName resolves a ship type ID to a name.
func (c *ESIClient) GetShipName(id int) string {
	return c.getName(id, "universe/types", c.shipNames)
}

// GetSystemName resolves a solar system ID to a name.
func (c *ESIClient) GetSystemName(id int) string {
	return c.getName(id, "universe/systems", c.systemNames)
}

// Add this struct to parse the JSON from the /search endpoint
type EsiSearchCharacterResponse struct {
	Character []int `json:"character"`
}

// This is the only new method you need for the lookup command.
func (c *ESIClient) GetCharacterID(characterName string) (int, error) {
	// Use url.QueryEscape to safely handle spaces and special characters in names
	fullURL := fmt.Sprintf("%s/search/?categories=character&search=%s&strict=true", c.baseURL, url.QueryEscape(characterName))

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return 0, err
	}

	// Set the required User-Agent header
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var searchData EsiSearchCharacterResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchData); err != nil {
		return 0, fmt.Errorf("failed to decode character search response")
	}

	// Check if the search returned any results
	if len(searchData.Character) == 0 {
		return 0, fmt.Errorf("character not found: %s", characterName)
	}

	// Return the first ID found
	return searchData.Character[0], nil
}

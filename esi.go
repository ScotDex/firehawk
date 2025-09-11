package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type ESIClient struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string

	characterNames   map[int]string
	characterIDs     map[string]int
	corporationNames map[int]string
	shipNames        map[int]string
	systemNames      map[int]string
	cacheMutex       sync.RWMutex
	contactInfo      string
}

func NewESIClient(contactInfo string) *ESIClient {
	return &ESIClient{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL:   "https://esi.evetech.net/latest",
		userAgent: fmt.Sprintf("Firehawk Discord Bot (%s)", contactInfo),

		characterNames:   make(map[int]string),
		corporationNames: make(map[int]string),
		shipNames:        make(map[int]string),
		systemNames:      make(map[int]string),
		characterIDs:     make(map[string]int),
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

// In esi_api.go

// Your new, correct struct for the /universe/ids/ endpoint
type EsiIDResponse struct {
	Characters []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"characters"`
}

// GetCharacterID resolves a character name to its ID using the POST endpoint.
func (c *ESIClient) GetCharacterID(characterName string) (int, error) {
	log.Println("DEBUG: Looking up character ID for:", characterName)
	// First, check the cache
	c.cacheMutex.RLock()
	id, found := c.characterIDs[characterName]
	c.cacheMutex.RUnlock()
	if found {
		log.Println("DEBUG: Cache HIT for character:", characterName)
		return id, nil // If found, return the ID immediately.
	}
	// Cache miss; proceed to call the API
	log.Println("DEBUG: Cache MISS for character:", characterName)
	requestBody, _ := json.Marshal([]string{characterName})
	fullURL := fmt.Sprintf("%s/universe/ids", c.baseURL)

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	log.Printf("DEBUG: ESI responded with HTTP Status: %s", resp.Status)

	// Use your new, correct struct here
	var idData EsiIDResponse
	if err := json.NewDecoder(resp.Body).Decode(&idData); err != nil {
		return 0, fmt.Errorf("failed to decode ID response")
	}

	// Check if the response contains any character data
	if idData.Characters == nil || len(idData.Characters) == 0 {
		return 0, fmt.Errorf("character not found: %s", characterName)
	}
	// Assuming the first match is the desired character
	fetchedID := idData.Characters[0].ID
	c.cacheMutex.Lock()
	c.characterIDs[characterName] = fetchedID
	c.cacheMutex.Unlock()

	// Return the ID of the first character found
	return fetchedID, nil
}

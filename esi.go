package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// ESIClient holds the configuration for making ESI calls and the in-memory cache.
type ESIClient struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string

	// Caches for ID -> Name lookups
	characterNames   map[int]string
	corporationNames map[int]string
	shipNames        map[int]string
	systemNames      map[int]string
	// Cache for Name -> ID lookups
	characterIDs map[string]int
	// Cache for detailed system info
	systemInfoCache map[int]*ESISystemInfo
	cacheMutex      sync.RWMutex
}

// NewESIClient creates and configures a new ESI client.
func NewESIClient(contactInfo string) *ESIClient {
	return &ESIClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		baseURL:    "https://esi.evetech.net/latest",
		userAgent:  fmt.Sprintf("Firehawk Discord Bot (%s)", contactInfo),
		// Initialize all cache maps
		characterNames:   make(map[int]string),
		corporationNames: make(map[int]string),
		shipNames:        make(map[int]string),
		systemNames:      make(map[int]string),
		characterIDs:     make(map[string]int),
		systemInfoCache:  make(map[int]*ESISystemInfo),
	}
}

// --- Struct Definitions ---
type ESINameResponse struct {
	Name string `json:"name"`
}
type EsiIDResponse struct {
	Characters []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"characters"`
}
type ESISystemInfo struct {
	Name           string  `json:"name"`
	SecurityStatus float64 `json:"security_status"`
	Stargates      []int   `json:"stargates"`
	Stations       []int   `json:"stations"`
	SystemID       int     `json:"system_id"`
}
type EsiSearchSystemResponse struct {
	SolarSystem []int `json:"solar_system"`
}

// --- CORE HELPER FUNCTION ---
// makeRequest is a single, generic function to handle all ESI API calls.
func (c *ESIClient) makeRequest(method, url string, body io.Reader, target interface{}) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.userAgent)
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Printf("DEBUG: ESI responded to %s with HTTP Status: %s", url, resp.Status)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ESI returned non-200 status: %s", resp.Status)
	}

	// Decode the JSON response directly into the provided target interface.
	return json.NewDecoder(resp.Body).Decode(target)
}

// --- NAME -> ID METHODS ---
func (c *ESIClient) GetCharacterID(characterName string) (int, error) {
	// 1. Check cache
	c.cacheMutex.RLock()
	id, found := c.characterIDs[characterName]
	c.cacheMutex.RUnlock()
	if found {
		return id, nil
	}

	// 2. API call on cache miss
	requestBody, _ := json.Marshal([]string{characterName})
	fullURL := fmt.Sprintf("%s/universe/ids/", c.baseURL)
	var idData EsiIDResponse
	if err := c.makeRequest("POST", fullURL, bytes.NewBuffer(requestBody), &idData); err != nil {
		return 0, err
	}
	if idData.Characters == nil || len(idData.Characters) == 0 {
		return 0, fmt.Errorf("character not found: %s", characterName)
	}

	// 3. Store result in cache and return
	fetchedID := idData.Characters[0].ID
	c.cacheMutex.Lock()
	c.characterIDs[characterName] = fetchedID
	c.cacheMutex.Unlock()
	return fetchedID, nil
}

// --- ID -> NAME METHODS (Now much shorter) ---
func (c *ESIClient) getName(id int, category string, cache map[int]string) string {
	if id == 0 {
		return "Unknown"
	}
	c.cacheMutex.RLock()
	name, found := cache[id]
	c.cacheMutex.RUnlock()
	if found {
		return name
	}

	fullURL := fmt.Sprintf("%s/%s/%d/", c.baseURL, category, id)
	var nameData ESINameResponse
	if c.makeRequest("GET", fullURL, nil, &nameData) != nil {
		return "Unknown"
	}

	c.cacheMutex.Lock()
	cache[id] = nameData.Name
	c.cacheMutex.Unlock()
	return nameData.Name
}

func (c *ESIClient) GetCharacterName(id int) string {
	return c.getName(id, "characters", c.characterNames)
}
func (c *ESIClient) GetCorporationName(id int) string {
	return c.getName(id, "corporations", c.corporationNames)
}
func (c *ESIClient) GetShipName(id int) string {
	return c.getName(id, "universe/types", c.shipNames)
}
func (c *ESIClient) GetSystemName(id int) string {
	return c.getName(id, "universe/systems", c.systemNames)
}
func (c *ESIClient) GetSystemID(id int) string {
	return c.getName(id, "universe/systems", c.systemNames)
}
func (c *ESIClient) GetSystemInfo(id int) string {
	return c.getName(id, "universe/systems", c.systemNames)
}

// ... (Add GetSystemInfo and GetSystemID here, following the same pattern)

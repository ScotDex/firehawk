package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// ESIClient holds the configuration for making ESI calls and the in-memory cache.
type ESIClient struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string

	characterNames   map[int]string
	corporationNames map[int]string
	shipNames        map[int]string
	systemNames      map[int]string
	characterIDs     map[string]int
	systemInfoCache  map[int]*ESISystemInfo // Store pointers
	cacheMutex       sync.RWMutex
	searchResults    map[string]SearchResponse
}

// NewESIClient creates and configures a new ESI client.
func NewESIClient(contactInfo string) *ESIClient {
	transport := &http.Transport{
		DisableCompression: false,
	}
	return &ESIClient{
		httpClient: &http.Client{
			Timeout:   15 * time.Second,
			Transport: transport,
		},

		baseURL:          "https://esi.evetech.net/latest",
		userAgent:        fmt.Sprintf("Firehawk Discord Bot (%s)", contactInfo),
		characterNames:   make(map[int]string),
		corporationNames: make(map[int]string),
		shipNames:        make(map[int]string),
		systemNames:      make(map[int]string),
		characterIDs:     make(map[string]int),
		systemInfoCache:  make(map[int]*ESISystemInfo), // Initialize map of pointers
		searchResults:    make(map[string]SearchResponse),
	}
}

// --- Struct Definitions ---
type ESINameResponse struct {
	Name string `json:"name"`
}
type ESIIDResponse struct { // Renamed for Go convention
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

// --- CORE HELPER FUNCTION ---
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ESI returned non-200 status: %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// --- NAME -> ID METHODS ---
func (c *ESIClient) GetCharacterID(characterName string) (int, error) {
	c.cacheMutex.RLock()
	id, found := c.characterIDs[characterName]
	c.cacheMutex.RUnlock()
	if found {
		return id, nil
	}

	requestBody, _ := json.Marshal([]string{characterName})
	fullURL := fmt.Sprintf("%s/universe/ids/", c.baseURL)
	var idData ESIIDResponse
	if err := c.makeRequest("POST", fullURL, bytes.NewBuffer(requestBody), &idData); err != nil {
		return 0, err
	}
	if len(idData.Characters) == 0 {
		return 0, fmt.Errorf("character not found: %s", characterName)
	}

	fetchedID := idData.Characters[0].ID
	c.cacheMutex.Lock()
	c.characterIDs[characterName] = fetchedID
	c.cacheMutex.Unlock()
	return fetchedID, nil
}

// --- ID -> NAME METHODS ---
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
	if err := c.makeRequest("GET", fullURL, nil, &nameData); err != nil {
		log.Printf("Failed to get name for ID %d in category %s: %v", id, category, err)
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

// --- STATIC CACHE METHODS ---
func (c *ESIClient) LoadSystemCache(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open cache file: %w", err)
	}
	defer file.Close()

	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	if err := json.NewDecoder(file).Decode(&c.systemInfoCache); err != nil {
		return fmt.Errorf("failed to unmarshal JSON data: %w", err)
	}

	log.Printf("Successfully loaded %d systems from local cache.", len(c.systemInfoCache))
	return nil
}

func (c *ESIClient) GetSystemDetails(systemID int) (*ESISystemInfo, error) {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	cachedInfo, found := c.systemInfoCache[systemID]
	if found {
		return cachedInfo, nil
	}

	return nil, fmt.Errorf("system ID %d not found in local cache", systemID)
}

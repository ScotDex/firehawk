package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

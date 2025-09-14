//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// ESISystemInfo defines the data we want to save for each system.
type ESISystemInfo struct {
	Name            string  `json:"name"`
	SecurityStatus  float64 `json:"security_status"`
	ConstellationID int     `json:"constellation_id"`
	SystemID        int     `json:"system_id"`
}

// Global variable for the number of concurrent workers. Adjust as needed.
const numWorkers = 25 // safer than 50 for ESI rate limits

// Shared HTTP client with timeout and connection reuse
var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

func main() {
	log.Println("Starting to build the system cache using a worker pool...")

	// 1. Get a list of ALL system IDs from the ESI.
	resp, err := httpClient.Get("https://esi.evetech.net/latest/universe/systems/")
	if err != nil {
		log.Fatalf("Failed to get system ID list: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var systemIDs []int
	if err := json.Unmarshal(body, &systemIDs); err != nil {
		log.Fatalf("Failed to unmarshal system IDs: %v", err)
	}
	log.Printf("Found %d total systems to process.", len(systemIDs))

	// 2. Collect results and failed IDs
	systemCache := make(map[int]ESISystemInfo)
	failedIDs := runWorkerPool(systemIDs, systemCache)

	// 3. Retry failed systems once more
	if len(failedIDs) > 0 {
		log.Printf("Retrying %d failed systems...", len(failedIDs))
		retryFails := runWorkerPool(failedIDs, systemCache)
		if len(retryFails) > 0 {
			log.Printf("WARNING: %d systems still failed after retries", len(retryFails))
		}
	}

	// 4. Save the completed map to a JSON file.
	fileData, err := json.MarshalIndent(systemCache, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal cache data to JSON: %v", err)
	}

	err = os.WriteFile("systems.json", fileData, 0644)
	if err != nil {
		log.Fatalf("Failed to write cache file: %v", err)
	}

	log.Println("Successfully created systems.json!")
}

// runWorkerPool handles worker spawning, job distribution, and result collection
func runWorkerPool(systemIDs []int, cache map[int]ESISystemInfo) []int {
	jobs := make(chan int, len(systemIDs))
	results := make(chan ESISystemInfo, len(systemIDs))
	failed := make(chan int, len(systemIDs))
	var wg sync.WaitGroup

	// Start workers
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, failed, &wg)
	}

	// Feed jobs
	for _, id := range systemIDs {
		jobs <- id
	}
	close(jobs)

	// Wait for workers to finish
	wg.Wait()
	close(results)
	close(failed)

	// Collect results
	for systemInfo := range results {
		cache[systemInfo.SystemID] = systemInfo
	}

	var failedIDs []int
	for id := range failed {
		failedIDs = append(failedIDs, id)
	}

	return failedIDs
}

// worker is the function that each goroutine will run.
func worker(id int, jobs <-chan int, results chan<- ESISystemInfo, failed chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()

	for systemID := range jobs {
		systemInfo, err := fetchSystemWithRetry(systemID)
		if err != nil {
			log.Printf("Worker %d: Failed to fetch system %d after retries: %v", id, systemID, err)
			failed <- systemID
			continue
		}
		results <- *systemInfo
	}
}

// fetchSystemWithRetry tries to fetch system data with retry + backoff
func fetchSystemWithRetry(systemID int) (*ESISystemInfo, error) {
	var systemInfo ESISystemInfo

	for attempt := 0; attempt < 5; attempt++ {
		resp, err := httpClient.Get(fmt.Sprintf("https://esi.evetech.net/latest/universe/systems/%d/", systemID))
		if err != nil {
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		func() {
			defer resp.Body.Close()

			// Handle rate limiting
			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == 420 {
				reset := resp.Header.Get("X-Esi-Error-Limit-Reset")
				if reset != "" {
					if wait, err := strconv.Atoi(reset); err == nil {
						time.Sleep(time.Duration(wait+1) * time.Second)
					}
				} else {
					time.Sleep(time.Duration(attempt+1) * time.Second)
				}
				return
			}

			if resp.StatusCode != http.StatusOK {
				return
			}

			if err := json.NewDecoder(resp.Body).Decode(&systemInfo); err != nil {
				return
			}
			// Success
			systemInfo.SystemID = systemID
			err = nil
		}()

		// If we filled the struct, return success
		if systemInfo.SystemID == systemID {
			return &systemInfo, nil
		}
	}

	return nil, fmt.Errorf("max retries reached")
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

// ESISystemInfo defines the data we want to save for each system.
type ESISystemInfo struct {
	Name            string  `json:"name"`
	SecurityStatus  float64 `json:"security_status"`
	ConstellationID int     `json:"constellation_id"`
	SystemID        int     `json:"system_id"`
}

// Global variable for the number of concurrent workers. Adjust as needed.
const numWorkers = 50

func main() {
	log.Println("Starting to build the system cache using a worker pool...")

	// 1. Get a list of ALL system IDs from the ESI.
	resp, err := http.Get("https://esi.evetech.net/latest/universe/systems/")
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

	// 2. Create channels and a wait group for concurrency.
	jobs := make(chan int, len(systemIDs))
	results := make(chan ESISystemInfo, len(systemIDs))
	var wg sync.WaitGroup

	// 3. Start the worker goroutines.
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}

	// 4. Send all system IDs to the jobs channel.
	for _, id := range systemIDs {
		jobs <- id
	}
	close(jobs) // Close the jobs channel to signal no more work is coming.

	// 5. Wait for all workers to finish their jobs.
	wg.Wait()
	close(results) // Close the results channel to signal we are done.

	// 6. Collect all results from the results channel into a map.
	systemCache := make(map[int]ESISystemInfo)
	for systemInfo := range results {
		systemCache[systemInfo.SystemID] = systemInfo
	}

	// 7. Save the completed map to a JSON file.
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

// worker is the function that each goroutine will run.
func worker(id int, jobs <-chan int, results chan<- ESISystemInfo, wg *sync.WaitGroup) {
	defer wg.Done()
	for systemID := range jobs {
		log.Printf("Worker %d: Processing system %d", id, systemID)

		systemURL := fmt.Sprintf("https://esi.evetech.net/latest/universe/systems/%d/", systemID)
		resp, err := http.Get(systemURL)
		if err != nil {
			log.Printf("Worker %d: Error fetching system %d: %v", id, systemID, err)
			continue
		}
		defer resp.Body.Close()

		var systemInfo ESISystemInfo
		if err := json.NewDecoder(resp.Body).Decode(&systemInfo); err != nil {
			log.Printf("Worker %d: Error decoding system %d: %v", id, systemID, err)
			continue
		}

		results <- systemInfo
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func hasTopicMatch(killmailTopics, subscribedTopics []string) bool {
	for _, kTopic := range killmailTopics {
		for _, sTopic := range subscribedTopics {
			if kTopic == sTopic {
				return true
			}
		}
	}
	return false
}

func goSafely(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("CRITICAL: Panic recovered in goroutine: %v", err)
			}
		}()
		fn()
	}()
}

func formatISKHuman(value float64) string {
	if value >= 1_000_000_000 {
		return fmt.Sprintf("%.2fB ISK", value/1_000_000_000)
	}
	if value >= 1_000_000 {
		return fmt.Sprintf("%.2fM ISK", value/1_000_000)
	}
	if value >= 1_000 {
		return fmt.Sprintf("%.2fK ISK", value/1_000)
	}
	// For values less than 1,000
	return fmt.Sprintf("%.2f ISK", value)
}

// generateKillmailTopics checks a killmail and returns a slice of matching topics.
func generateKillmailTopics(data *KillmailData) []string {
	topics := []string{"all"}

	// --- Value-based topics ---
	if data.Killmail.TotalValue >= 10_000_000_000 {
		topics = append(topics, "10b")
	}
	if data.Killmail.TotalValue >= 5_000_000_000 {
		topics = append(topics, "5b")
	}
	if data.Killmail.TotalValue >= 1_000_000_000 {
		topics = append(topics, "bigkills")
	}

	// --- Attribute-based topics ---
	if data.Killmail.IsSolo {
		topics = append(topics, "solo")
	}
	if data.Killmail.IsNpc {
		topics = append(topics, "npc")
	}

	// --- Location-based topics ---
	if data.Killmail.RegionID >= 12000001 {
		topics = append(topics, "abyssal")
	}
	if data.Killmail.RegionID >= 11000000 && data.Killmail.RegionID < 12000000 {
		topics = append(topics, "wspace")
	}
	if data.Killmail.SystemSecurity >= 0.5 {
		topics = append(topics, "highsec")
	} else if data.Killmail.SystemSecurity > 0.0 {
		topics = append(topics, "lowsec")
	} else {
		topics = append(topics, "nullsec")
	}

	// --- Ship-based topics (using the victim's ship group ID) ---
	switch data.Killmail.Victim.ShipGroupID {
	// T1 Subcaps
	case 25: // Frigate
		topics = append(topics, "frigates", "t1")
	case 420: // Destroyer
		topics = append(topics, "destroyers", "t1")
	case 26: // Cruiser
		topics = append(topics, "cruisers", "t1")
	case 540: // Battlecruiser
		topics = append(topics, "battlecruisers", "t1")
	case 27: // Battleship
		topics = append(topics, "battleships", "t1")

	// T2 and Faction Ships
	case 324, 541, 830, 834, 893: // Assault Frigate, Interdictor, Covert Ops, Stealth Bomber, Interceptor
		topics = append(topics, "frigates", "t2")
	case 358, 831, 894: // Heavy Assault Cruiser, Combat Recon Ship, Heavy Interdiction Cruiser
		topics = append(topics, "cruisers", "t2")
	case 902: // Command Destroyer
		topics = append(topics, "destroyers", "t2")
	case 832, 833, 898, 900: // Logistics, Force Recon Ship, Black Ops, Marauder
		topics = append(topics, "battleships", "t2")

	// T3
	case 906: // Strategic Cruiser (T3 Cruiser)
		topics = append(topics, "cruisers", "t3")
	case 941: // Tactical Destroyer (T3 Destroyer)
		topics = append(topics, "destroyers", "t3")

	// Capitals & Structures
	case 513: // Freighter
		topics = append(topics, "freighters")
	case 30: // Titan
		topics = append(topics, "titans", "capitals")
	case 659: // Supercarrier
		topics = append(topics, "supercarriers", "capitals")
	case 485, 547, 1538: // Dreadnought, Carrier, FAX
		topics = append(topics, "capitals")
	case 1657: // Upwell Structures
		topics = append(topics, "citadel")
	}

	return topics
}

func loadSubscriptions() {
	mu.Lock()
	defer mu.Unlock()

	data, err := os.ReadFile("subscriptions.json") // Use your SubMapFile variable

	// Case 1: Handle errors during file reading.
	if err != nil {
		// If the file simply doesn't exist, that's okay. We'll just start fresh.
		if os.IsNotExist(err) {
			log.Println("subscriptions.json not found, starting with empty subscriptions.")
			return
		}
		// For any other reading error, log it as a more serious problem.
		log.Printf("Error reading subscriptions.json: %v", err)
		return
	}

	// Case 2: If the file exists but is empty, there's nothing to load.
	if len(data) == 0 {
		log.Println("subscriptions.json is empty, starting with empty subscriptions.")
		return
	}

	// Only if we have data, we try to unmarshal it.
	err = json.Unmarshal(data, &subscriptions)
	if err != nil {
		log.Printf("Error unmarshaling subscriptions from JSON: %v", err)
	} else {
		log.Printf("Successfully loaded %d channel(s) with subscriptions.", len(subscriptions))
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// redisqURL is the endpoint for the EVE-KILL real-time data feed.
const redisqURL = "https://eve-kill.com/redisq?queueID=firehawkv1"

// KillmailData is the struct for parsing the killmail JSON response from the data feed.
type KillmailData struct {
	// The 'Package' is now a pointer, which allows the JSON decoder to handle cases where it's null.
	Package *struct {
		KillID   int `json:"killID"`
		Killmail struct {
			KillmailID    int       `json:"killmail_id"`
			KillmailTime  time.Time `json:"killmail_time"`
			SolarSystemID int       `json:"solar_system_id"`
			Victim        struct {
				ShipTypeID    int `json:"ship_type_id"`
				CharacterID   int `json:"character_id"`
				CorporationID int `json:"corporation_id"`
				AllianceID    int `json:"alliance_id"`
			} `json:"victim"`
			Attackers []struct {
				CharacterID   int  `json:"character_id"`
				CorporationID int  `json:"corporation_id"`
				ShipTypeID    int  `json:"ship_type_id"`
				FinalBlow     bool `json:"final_blow"`
			} `json:"attackers"`
		} `json:"killmail"`
		Zkb struct {
			TotalValue float64 `json:"totalValue"`
			Solo       bool    `json:"solo"`
			Npc        bool    `json:"npc"`
		} `json:"zkb"`
	} `json:"package"`
}

// buildKillmailURL constructs the public URL for a given killmail ID.
func buildKillmailURL(killID int) string {
	return fmt.Sprintf("https://eve-kill.com/kill/%d", killID)
}

// formatISKHuman formats a float64 into a human-readable string (e.g., 1.23B, 45.6M).
func formatISKHuman(value float64) string {
	if value >= 1_000_000_000 {
		return fmt.Sprintf("%.2fB ISK", value/1_000_000_000)
	}
	if value >= 1_000_000 {
		return fmt.Sprintf("%.2fM ISK", value/1_000_000)
	}
	return fmt.Sprintf("%.2fK ISK", value/1_000)
}

// killmailPoller is the main background process that fetches, enriches, and posts killmails.
func killmailPoller(s *discordgo.Session, channelID string) {
	client := &http.Client{
		Timeout: 90 * time.Second,
	}
	log.Println("Killmail poller started. Waiting for killmails...")

	for {
		// --- 1. Fetch data from the stream ---
		resp, err := client.Get(redisqURL)
		if err != nil {
			log.Printf("Error fetching from redisq: %v", err)
			time.Sleep(10 * time.Second) // Wait longer on network errors.
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Redisq returned non-200 status: %s", resp.Status)
			resp.Body.Close()
			time.Sleep(10 * time.Second)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("Error reading redisq response body: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// --- 2. Parse the JSON response ---
		if string(body) == "{\"package\":null}" {
			time.Sleep(1 * time.Second) // Empty queue, short wait.
			continue
		}
		var killmailData KillmailData
		err = json.Unmarshal(body, &killmailData)
		if err != nil {
			log.Printf("Error unmarshaling JSON: %v", err)
			log.Printf("Failing JSON body: %s", string(body))
			time.Sleep(5 * time.Second)
			continue
		}

		// Add a nil check to prevent panics on malformed data.
		if killmailData.Package == nil {
			log.Println("Received killmail with null package, skipping.")
			continue
		}

		// --- 3. Filter out old killmails --- Commented out for testing
		//	killTime := killmailData.Package.Killmail.KillmailTime
		//	if time.Since(killTime) > 5*time.Minute {
		//		log.Printf("Stale killmail (ID %d) received, skipping...", killmailData.Package.KillID)
		//		continue
		//	}

		log.Printf("New killmail received: ID %d", killmailData.Package.KillID)

		pkg := killmailData.Package
		killID := pkg.KillID
		if killID > 0 {
			// --- 4. Extract key information ---
			victim := pkg.Killmail.Victim
			attackers := pkg.Killmail.Attackers
			var finalBlowAttacker struct {
				CharacterID   int
				CorporationID int
				ShipTypeID    int
			}
			for _, attacker := range attackers {
				if attacker.FinalBlow {
					finalBlowAttacker.CharacterID = attacker.CharacterID
					finalBlowAttacker.CorporationID = attacker.CorporationID
					finalBlowAttacker.ShipTypeID = attacker.ShipTypeID
					break
				}
			}

			// --- 5. Enrich the data by fetching names for IDs ---
			var victimName, victimCorp, victimShip, finalBlowName, finalBlowCorp, finalBlowShip, systemName string
			var wg sync.WaitGroup
			wg.Add(7)
			go func() { defer wg.Done(); victimName = esiClient.GetCharacterName(victim.CharacterID) }()
			go func() { defer wg.Done(); victimCorp = esiClient.GetCorporationName(victim.CorporationID) }()
			go func() { defer wg.Done(); victimShip = esiClient.GetShipName(victim.ShipTypeID) }()
			go func() { defer wg.Done(); finalBlowName = esiClient.GetCharacterName(finalBlowAttacker.CharacterID) }()
			go func() { defer wg.Done(); finalBlowCorp = esiClient.GetCorporationName(finalBlowAttacker.CorporationID) }()
			go func() { defer wg.Done(); finalBlowShip = esiClient.GetShipName(finalBlowAttacker.ShipTypeID) }()
			go func() { defer wg.Done(); systemName = esiClient.GetSystemName(pkg.Killmail.SolarSystemID) }()
			wg.Wait()

			// --- 6. Build and send the final Discord embed ---
			url := buildKillmailURL(killID)
			totalValueFormatted := formatISKHuman(pkg.Zkb.TotalValue)

			embed := &discordgo.MessageEmbed{
				Title:     fmt.Sprintf("%s destroyed in %s", victimShip, systemName),
				URL:       url,
				Color:     0xBF2A2A,
				Timestamp: pkg.Killmail.KillmailTime.Format(time.RFC3339),
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: fmt.Sprintf("https://images.evetech.net/types/%d/render?size=128", victim.ShipTypeID),
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Victim",
						Value:  fmt.Sprintf("**%s**\n%s", victimName, victimCorp),
						Inline: true,
					},
					{
						Name:   fmt.Sprintf("Final Blow (%d Attackers)", len(attackers)),
						Value:  fmt.Sprintf("**%s**\n%s (%s)", finalBlowName, finalBlowCorp, finalBlowShip),
						Inline: true,
					},
					{
						Name:  "Total Value",
						Value: totalValueFormatted,
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Powered by Firehawk",
				},
			}
			_, err := s.ChannelMessageSendEmbed(channelID, embed)
			if err != nil {
				log.Printf("Error sending enriched embed to Discord: %v", err)
			} else {
				log.Printf("Successfully posted enriched embed for kill %d.", killID)
			}
		}
		time.Sleep(1 * time.Second)
	}
}

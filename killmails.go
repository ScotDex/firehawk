package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

// HandleKillmailMessage serves as the main entry point for processing raw messages from the WebSocket.
// It determines the message type and directs the data to the appropriate handler.
func HandleKillmailMessage(s *discordgo.Session, message []byte) {
	// First, unmarshal the message into our generic SocketMessage to read its type.
	var msg SocketMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Error unmarshaling socket envelope: %s", string(message))
		return
	}

	// Use a switch statement to handle different kinds of server messages.
	switch msg.Type {
	case "killmail":
		// If it's a killmail, unmarshal its specific data payload.
		var killmailData KillmailData
		if err := json.Unmarshal(msg.Data, &killmailData); err != nil {
			log.Printf("Error unmarshaling killmail payload: %v", err)
			return
		}
		// Pass the fully parsed killmail to the main processing function.
		processAndSendKillmail(s, &killmailData)

	case "info", "subscribed", "ping":
		// Handle standard server status messages by logging them.
		log.Printf("Received server message: type=%s", msg.Type)

	default:
		// Log any message types we don't currently handle, for debugging purposes.
		log.Printf("Received unhandled type: '%s'", msg.Type)
	}
}

// processAndSendKillmail takes parsed killmail data, finds all subscribed channels that
// match the killmail's topics, and sends a formatted embed to them.
func processAndSendKillmail(s *discordgo.Session, data *KillmailData) {
	log.Printf("Processing new killmail: ID %d | Value: %.2f ISK", data.Killmail.KillmailID, data.Killmail.TotalValue)

	// Step 1: Generate a list of topics (or "tags") for this specific killmail
	// by calling the helper function from another file.
	killmailTopics := generateKillmailTopics(data)

	// Step 2: Build the rich Discord embed for the killmail. This is done once to avoid repeat work.
	embed := buildKillmailEmbed(data)

	// Step 3: Efficiently find matching channels and send the embed.
	// A read-lock allows multiple killmails to be processed at the same time without data corruption.
	mu.RLock()
	defer mu.RUnlock()

	// Iterate over each channel that has at least one subscription.
	for channelID, subscribedTopics := range subscriptions {
		// For each channel, check if any of its subscribed topics match the topics of this killmail.
		for _, kmTopic := range killmailTopics {
			// This is an efficient check to see if the key 'kmTopic' exists in the 'subscribedTopics' map.
			if _, isSubscribed := subscribedTopics[kmTopic]; isSubscribed {
				// A match is found! The channel is subscribed to a topic this killmail has.
				_, err := s.ChannelMessageSendEmbed(channelID, embed)
				if err != nil {
					log.Printf("Failed to send killmail embed to channel %s: %v", channelID, err)
				}

				// Since we've found a match and sent the message, we can stop checking topics for this channel
				// and move on to the next one. This 'break' makes the process much more efficient.
				break
			}
		}
	}
}

// buildKillmailEmbed is a factory function that constructs a rich Discord embed from killmail data.
func buildKillmailEmbed(data *KillmailData) *discordgo.MessageEmbed {
	// Extract key figures for clarity.
	victim := data.Killmail.Victim
	var finalBlowAttacker struct {
		CharacterName   string
		CorporationName string
	}

	// Safely iterate through attackers to find the one who got the final blow.
	for _, a := range data.Killmail.Attackers {
		if a.FinalBlow {
			finalBlowAttacker.CharacterName = a.CharacterName
			finalBlowAttacker.CorporationName = a.CorporationName
			break
		}
	}
	// Provide default values if names are not present (e.g., for NPCs).
	if finalBlowAttacker.CharacterName == "" {
		finalBlowAttacker.CharacterName = "Unknown"
	}
	if finalBlowAttacker.CorporationName == "" {
		finalBlowAttacker.CorporationName = "Unknown"
	}

	// Assemble and return the complete embed structure.
	return &discordgo.MessageEmbed{
		Title:     fmt.Sprintf("%s destroyed in %s", victim.ShipName.En, data.Killmail.SystemName),
		URL:       fmt.Sprintf("https://eve-kill.com/kill/%d", data.Killmail.KillmailID),
		Color:     0xBF2A2A, // A deep red color
		Timestamp: data.Killmail.KillmailTime.Format(time.RFC3339),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: fmt.Sprintf("https://images.evetech.net/types/%d/render?size=128", victim.ShipID),
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Victim", Value: victim.CharacterName, Inline: true},
			{Name: "Corporation", Value: victim.CorporationName, Inline: true},
			{Name: "Value", Value: formatISKHuman(data.Killmail.TotalValue), Inline: true},
			{Name: "Final Blow", Value: finalBlowAttacker.CharacterName, Inline: true},
			{Name: "Corporation", Value: finalBlowAttacker.CorporationName, Inline: true},
			{Name: "System", Value: fmt.Sprintf("%s (%s)", data.Killmail.SystemName, data.Killmail.RegionName.En), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{Text: "Powered by Firehawk"},
	}
}

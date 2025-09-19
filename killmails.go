package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

// HandleKillmailMessage is the entry point for processing messages from the WebSocket.
func HandleKillmailMessage(s *discordgo.Session, message []byte) {
	var msg SocketMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Error unmarshaling socket envelope: %s", string(message))
		return
	}

	switch msg.Type {
	case "killmail":
		var killmailData KillmailData
		if err := json.Unmarshal(msg.Data, &killmailData); err != nil {
			log.Printf("Error unmarshaling killmail payload: %v", err)
			return
		}
		// Pass the killmail to the processing function.
		processAndSendKillmail(s, &killmailData)

	case "info", "subscribed", "ping":
		log.Printf("Received server message: type=%s", msg.Type)

	default:
		log.Printf("Received unhandled type: '%s'", msg.Type)
	}
}

// --- Killmail Processing and Sending ---

// processAndSendKillmail creates the Discord embed and sends it to all matching subscribed channels.
func processAndSendKillmail(s *discordgo.Session, data *KillmailData) {
	killID := data.Killmail.KillmailID
	log.Printf("Processing new killmail: ID %d | Value: %.2f ISK", killID, data.Killmail.TotalValue)

	// 1. Generate a set of "tags" or "topics" for this specific killmail.
	killmailTopics := generateKillmailTopics(data)

	// 2. Build the Discord embed once.
	embed := buildKillmailEmbed(data)

	// 3. Efficiently find and send to subscribed channels.
	mu.RLock() // Lock for safe concurrent reading of the subscriptions map.
	defer mu.RUnlock()

	// Create a set of channels we've already posted to, to avoid duplicate messages.
	sentToChannels := make(map[string]bool)

	// Loop over every topic this killmail generated.
	for _, kmTopic := range killmailTopics {
		// Now, loop over every channel that is subscribed to anything.
		for channelID, subscribedTopics := range subscriptions {
			// If we've already sent to this channel, skip it.
			if sentToChannels[channelID] {
				continue
			}

			// This is the efficient check: does this channel's set of topics contain our killmail's topic?
			if _, ok := subscribedTopics[kmTopic]; ok {
				_, err := s.ChannelMessageSendEmbed(channelID, embed)
				if err != nil {
					log.Printf("Failed to send killmail embed to channel %s: %v", channelID, err)
				}
				// Mark this channel as "sent" and stop checking other topics for it.
				sentToChannels[channelID] = true
			}
		}
	}
}

// --- Helper Functions ---

// buildKillmailEmbed creates the discordgo.MessageEmbed from the killmail data.
func buildKillmailEmbed(data *KillmailData) *discordgo.MessageEmbed {
	victim := data.Killmail.Victim
	var finalBlowAttacker struct {
		CharacterName   string
		CorporationName string
	}

	// Safely find the final blow attacker.
	for _, a := range data.Killmail.Attackers {
		if a.FinalBlow {
			finalBlowAttacker.CharacterName = a.CharacterName
			finalBlowAttacker.CorporationName = a.CorporationName
			break
		}
	}
	if finalBlowAttacker.CharacterName == "" {
		finalBlowAttacker.CharacterName = "Unknown"
	}
	if finalBlowAttacker.CorporationName == "" {
		finalBlowAttacker.CorporationName = "Unknown"
	}

	return &discordgo.MessageEmbed{
		Title:     fmt.Sprintf("%s destroyed in %s", victim.ShipName.En, data.Killmail.SystemName),
		URL:       fmt.Sprintf("https://eve-kill.com/kill/%d", data.Killmail.KillmailID),
		Color:     0xBF2A2A, // Red
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

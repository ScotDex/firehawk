package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

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
		processAndSendKillmail(s, &killmailData)

	case "info", "subscribed", "ping":
		log.Printf("Received server message: type=%s", msg.Type)

	default:
		log.Printf("Received unhandled type: '%s'", msg.Type)
	}
}

func processAndSendKillmail(s *discordgo.Session, data *KillmailData) {
	killID := data.Killmail.KillmailID
	log.Printf("Processing new killmail: ID %d | Value: %.2f ISK", killID, data.Killmail.TotalValue)

	systemName := data.Killmail.SystemName
	victimName := data.Killmail.Victim.CharacterName
	victimCorp := data.Killmail.Victim.CorporationName
	victimShip := data.Killmail.Victim.ShipName.En

	// Safely find the final blow attacker
	var attackerCorp = "Unknown"
	var finalBlowName = "Unknown"
	for _, attacker := range data.Killmail.Attackers {
		if attacker.FinalBlow {
			finalBlowName = attacker.CharacterName
			attackerCorp = attacker.CorporationName
			break
		}
	}

	totalValueFormatted := formatISKHuman(data.Killmail.TotalValue)
	url := fmt.Sprintf("https://eve-kill.com/kill/%d", killID)

	embed := &discordgo.MessageEmbed{
		Title:     fmt.Sprintf("%s destroyed in %s", victimShip, systemName),
		URL:       url,
		Color:     0xBF2A2A, // Red color
		Timestamp: data.Killmail.KillmailTime.Format(time.RFC3339),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: fmt.Sprintf("https://images.evetech.net/types/%d/render?size=128", data.Killmail.Victim.ShipID),
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Victim", Value: victimName, Inline: true},
			{Name: "Corporation", Value: victimCorp, Inline: true},

			{Name: "Final Blow", Value: finalBlowName, Inline: true},
			{Name: "Corporation", Value: attackerCorp, Inline: true},

			{Name: "System", Value: systemName, Inline: true},
			{Name: "Region", Value: data.Killmail.RegionName.En, Inline: true},
			{Name: "Value", Value: totalValueFormatted, Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{Text: "Powered by Firehawk"},
	}

	_, err := s.ChannelMessageSendEmbed("1417920853894500352", embed) // Replace with your channel ID
	if err != nil {
		log.Printf("Failed to send killmail embed: %v", err)
		return
	}
	log.Printf("Successfully sent embed for kill %d.", killID)
}

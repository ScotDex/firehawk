package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

// You will also need a getAPIStatus() function and a global esiClient variable defined elsewhere.

var subscriptions = make(map[string]map[string][]string)

// Define the choices once as a separate variable to be reused.
var killmailTopicChoices = []*discordgo.ApplicationCommandOptionChoice{
	{Name: "All Kills", Value: "all"},
	{Name: "Big Kills", Value: "bigkills"},
	{Name: "Solo Kills", Value: "solo"},
	{Name: "NPC Kills", Value: "npc"},
	{Name: "Citadel Kills", Value: "citadel"},
	{Name: "10b+ ISK Kills", Value: "10b"},
	{Name: "5b+ ISK Kills", Value: "5b"},
	{Name: "Abyssal Kills", Value: "abyssal"},
	{Name: "W-Space Kills", Value: "wspace"},
	{Name: "High-sec Kills", Value: "highsec"},
	{Name: "Low-sec Kills", Value: "lowsec"},
	{Name: "Null-sec Kills", Value: "nullsec"},
	{Name: "T1 Ship Kills", Value: "t1"},
	{Name: "T2 Ship Kills", Value: "t2"},
	{Name: "T3 Ship Kills", Value: "t3"},
	{Name: "Frigate Kills", Value: "frigates"},
	{Name: "Destroyer Kills", Value: "destroyers"},
	{Name: "Cruiser Kills", Value: "cruisers"},
	{Name: "Battlecruiser Kills", Value: "battlecruisers"},
	{Name: "Battleship Kills", Value: "battleships"},
	{Name: "Capital Kills", Value: "capitals"},
	{Name: "Freighter Kills", Value: "freighters"},
	{Name: "Supercarrier Kills", Value: "supercarriers"},
	{Name: "Titan Kills", Value: "titans"},
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "status",
		Description: "Live status of EVE Online server",
	},
	{
		Name:        "lookup",
		Description: "Lookup an EVE Online character by name",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "character_name",
				Description: "Name of the character to look up",
				Required:    true,
			},
		},
	},
	{
		Name:        "subscribe",
		Description: "Sub this channel to a killmail feed",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "topic",
				Description: "The feed to subscribe to",
				Required:    true,
				Choices:     killmailTopicChoices,
			},
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel to subscribe to",
				Required:    false,
			},
		},
	},
	{
		Name:        "unsubscribe",
		Description: "Unsubscribe this channel from a killmail feed",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "topic",
				Description: "The feed to unsubscribe from",
				Required:    true,
				Choices:     killmailTopicChoices,
			},
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel to unsubscribe (defaults to current)",
				Required:    false,
			},
		},
	},
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"status": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		status, err := getAPIStatus()
		if err != nil {
			log.Printf("Error fetching API status: %v", err)
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "❌ Error fetching EVE Online server status.",
			})
			return
		}
		uptime := time.Since(status.StartTime)
		uptimeStr := fmt.Sprintf("%d hours, %d minutes", int(uptime.Hours()), int(uptime.Minutes())%60)
		embed := &discordgo.MessageEmbed{
			Title: "EVE Online Server Status",
			Color: 0x00ff00,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Players Online", Value: fmt.Sprintf("%d", status.Players), Inline: true},
				{Name: "Server Uptime", Value: uptimeStr, Inline: true},
				{Name: "Server Version", Value: status.ServerVersion},
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Embeds: []*discordgo.MessageEmbed{embed},
		})
		if err != nil {
			log.Printf("Failed to send followup message: %v", err)
		}
	},

	// In your commandHandlers map in commands.go

	"lookup": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Println("--- /lookup command initiated ---")

		// Defer the response immediately
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		if err != nil {
			log.Printf("DEBUG: Failed to defer interaction: %v", err)
			return
		}

		// Get the character name from the user's command
		charName := i.ApplicationCommandData().Options[0].StringValue()
		log.Printf("DEBUG: User is looking for character: '%s'", charName)

		// Use your ESI client to get the character's ID
		log.Println("DEBUG: Calling esiClient.GetCharacterID...")
		charID, err := esiClient.GetCharacterID(charName)

		// Check for an error from the ESI call
		if err != nil {
			log.Printf("DEBUG: esiClient returned an error: %v", err)
			errorMessage := fmt.Sprintf("❌ Could not find a character named `%s`.", charName)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &errorMessage,
			})
			return
		}

		// If successful, log the result
		log.Printf("DEBUG: Found Character ID: %d", charID)

		// Build the final URL using the ID
		finalURL := fmt.Sprintf("https://eve-kill.com/character/%d", charID)
		log.Printf("DEBUG: Constructed final URL: %s", finalURL)

		// Send the URL as the response
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &finalURL,
		})
		if err != nil {
			log.Printf("DEBUG: Failed to edit interaction response: %v", err)
		}

		log.Println("--- /lookup command finished ---")
	},

	"subscribe": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
		for _, opt := range options {
			optionMap[opt.Name] = opt
		}
		topic := optionMap["topic"].StringValue()
		var channelID string
		if channelOption, ok := optionMap["channel"]; ok {
			channelID = channelOption.ChannelValue(s).ID
		} else {
			channelID = i.ChannelID
		}
		guildID := i.GuildID
		if _, ok := subscriptions[guildID]; !ok {
			subscriptions[guildID] = make(map[string][]string)
		}
		if _, ok := subscriptions[guildID][channelID]; !ok {
			subscriptions[guildID][channelID] = []string{}
		}
		subscriptions[guildID][channelID] = append(subscriptions[guildID][channelID], topic)
		log.Printf("New subscription added: Guild %s, Channel %s, Topic %s", guildID, channelID, topic)
		responseMessage := fmt.Sprintf("✅ Subscribed channel <#%s> to the '%s' topic.", channelID, topic)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: responseMessage,
			},
		})
	},

	"unsubscribe": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := i.ApplicationCommandData().Options
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
		for _, opt := range options {
			optionMap[opt.Name] = opt
		}
		topicToRemove := optionMap["topic"].StringValue()
		var channelID string
		if channelOption, ok := optionMap["channel"]; ok {
			channelID = channelOption.ChannelValue(s).ID
		} else {
			channelID = i.ChannelID
		}
		guildID := i.GuildID
		if _, ok := subscriptions[guildID][channelID]; !ok {
			responseMessage := fmt.Sprintf("⚠️ This channel isn't subscribed to any topics.")
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: responseMessage, Flags: discordgo.MessageFlagsEphemeral},
			})
			return
		}
		originalTopics := subscriptions[guildID][channelID]
		newTopics := []string{}
		found := false
		for _, topic := range originalTopics {
			if topic != topicToRemove {
				newTopics = append(newTopics, topic)
			} else {
				found = true
			}
		}
		var responseMessage string
		if !found {
			responseMessage = fmt.Sprintf("⚠️ Channel <#%s> was not subscribed to the '%s' topic.", channelID, topicToRemove)
		} else {
			subscriptions[guildID][channelID] = newTopics
			responseMessage = fmt.Sprintf("✅ Unsubscribed channel <#%s> from the '%s' topic.", channelID, topicToRemove)
			log.Printf("Subscription removed: Guild %s, Channel %s, Topic %s", guildID, channelID, topicToRemove)
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: responseMessage,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	},
}

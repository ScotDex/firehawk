package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language" // <-- 1. Import language package
	"golang.org/x/text/message"
)

// You will also need getAPIStatus() and a global esiClient variable defined elsewhere.

// map[channelID][]topics
var subscriptions = make(map[string][]string)

var killmailTopicChoices = []*discordgo.ApplicationCommandOptionChoice{
	{Name: "All Kills", Value: "all"}, {Name: "Big Kills", Value: "bigkills"},
	{Name: "Solo Kills", Value: "solo"}, {Name: "NPC Kills", Value: "npc"},
	{Name: "Citadel Kills", Value: "citadel"}, {Name: "10b+ ISK Kills", Value: "10b"},
	{Name: "5b+ ISK Kills", Value: "5b"}, {Name: "Abyssal Kills", Value: "abyssal"},
	{Name: "W-Space Kills", Value: "wspace"}, {Name: "High-sec Kills", Value: "highsec"},
	{Name: "Low-sec Kills", Value: "lowsec"}, {Name: "Null-sec Kills", Value: "nullsec"},
	{Name: "T1 Ship Kills", Value: "t1"}, {Name: "T2 Ship Kills", Value: "t2"},
	{Name: "T3 Ship Kills", Value: "t3"}, {Name: "Frigate Kills", Value: "frigates"},
	{Name: "Destroyer Kills", Value: "destroyers"}, {Name: "Cruiser Kills", Value: "cruisers"},
	{Name: "Battlecruiser Kills", Value: "battlecruisers"}, {Name: "Battleship Kills", Value: "battleships"},
	{Name: "Capital Kills", Value: "capitals"}, {Name: "Freighter Kills", Value: "freighters"},
	{Name: "Supercarrier Kills", Value: "supercarriers"}, {Name: "Titan Kills", Value: "titans"},
}

var commands = []*discordgo.ApplicationCommand{
	{Name: "status", Description: "Live Tranquility Status"},
	{Name: "scout", Description: "Provides intel on a specific solar system.", Options: []*discordgo.ApplicationCommandOption{{Type: discordgo.ApplicationCommandOptionString, Name: "system_name", Description: "The name of the solar system to scout.", Required: true}}},
	{Name: "lookup", Description: "Lookup an EVE Online character by name", Options: []*discordgo.ApplicationCommandOption{{Type: discordgo.ApplicationCommandOptionString, Name: "character_name", Description: "Name of the character (exact spelling required).", Required: true}}},
	{Name: "subscribe", Description: "Subscribe this channel to a killmail feed", Options: []*discordgo.ApplicationCommandOption{{Type: discordgo.ApplicationCommandOptionString, Name: "topic", Description: "The feed to subscribe to", Required: true, Choices: killmailTopicChoices}, {Type: discordgo.ApplicationCommandOptionChannel, Name: "channel", Description: "The channel to subscribe to", Required: false}}},
	{Name: "unsubscribe", Description: "Unsubscribe this channel from a killmail feed", Options: []*discordgo.ApplicationCommandOption{{Type: discordgo.ApplicationCommandOptionString, Name: "topic", Description: "The feed to unsubscribe from", Required: true, Choices: killmailTopicChoices}, {Type: discordgo.ApplicationCommandOptionChannel, Name: "channel", Description: "The channel to unsubscribe", Required: false}}},
	{Name: "alliance", Description: "Provides intel on a specific alliance.", Options: []*discordgo.ApplicationCommandOption{{Type: discordgo.ApplicationCommandOptionString, Name: "alliances", Description: "The name of an alliance you want to scout.", Required: true}}},
	{Name: "group", Description: "Provides intel on a specific corporation.", Options: []*discordgo.ApplicationCommandOption{{Type: discordgo.ApplicationCommandOptionString, Name: "corporations", Description: "The name of a corporation you want to scout.", Required: true}}},
	{Name: "tools", Description: "An up to date list of third party tools for Eve Online"},
}
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"status": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		status, err := esiClient.getAPIStatus()
		if err != nil {
			log.Printf("Error fetching API status: %v", err)
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "❌ Error fetching EVE Online server status.",
			})
			return
		}
		p := message.NewPrinter(language.English)
		playersStr := p.Sprintf("%d", status.Players)

		corpLogoURL := esiClient.GetRandomCorporationLogoURL()
		uptime := time.Since(status.StartTime)
		uptimeStr := fmt.Sprintf("%d hours, %d minutes", int(uptime.Hours()), int(uptime.Minutes())%60)
		embed := &discordgo.MessageEmbed{
			Title: "EVE Online Server Status",
			Color: 0x00ff00,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: corpLogoURL,
			},
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Players Online", Value: playersStr, Inline: true},
				{Name: "Server Uptime", Value: uptimeStr, Inline: true},
				{Name: "Server Version", Value: status.ServerVersion},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Powered by Firehawk | Data from EVE ESI",
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

	"lookup": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		charName := i.ApplicationCommandData().Options[0].StringValue()
		charID, err := esiClient.GetCharacterID(charName)
		if err != nil {
			errorMessage := fmt.Sprintf("❌ Could not find a character named `%s`.", charName)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &errorMessage,
			})
			return
		}
		finalURL := fmt.Sprintf("https://eve-kill.com/character/%d", charID)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &finalURL,
		})
	},

	"subscribe": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
		for _, opt := range i.ApplicationCommandData().Options {
			optionMap[opt.Name] = opt
		}
		topic := optionMap["topic"].StringValue()
		channelID := i.ChannelID
		if channelOption, ok := optionMap["channel"]; ok {
			channelID = channelOption.ChannelValue(s).ID
		}

		// Check for duplicate subscriptions
		for _, existingTopic := range subscriptions[channelID] {
			if existingTopic == topic {
				responseMessage := fmt.Sprintf("⚠️ Channel <#%s> is already subscribed to the '%s' topic.", channelID, topic)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{Content: responseMessage, Flags: discordgo.MessageFlagsEphemeral},
				})
				return
			}
		}

		// Add the new subscription
		subscriptions[channelID] = append(subscriptions[channelID], topic)
		log.Printf("New subscription added: Channel %s, Topic %s", channelID, topic)
		responseMessage := fmt.Sprintf("✅ Subscribed channel <#%s> to the '%s' topic.", channelID, topic)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: responseMessage},
		})
	},

	"tools": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		toolLinks := []string{
			"[EVE-KILL](https://eve-kill.com/)",
			"[Z-Kill](https://zkillboard.com/)",
			"[SeAT](https://github.com/eveseat/seat/)",
			"[EVE Workbench](https://github.com/EVE-Workbench)",
			"[Alliance Auth](https://gitlab.com/allianceauth/allianceauth/)",
			"[DSCAN ICU](https://dscan.icu/)",
			"[Eve Buddy](https://github.com/ErikKalkoken/evebuddy)",
		}

		// You will need to import the "strings" package for this to work.
		var description strings.Builder
		for _, link := range toolLinks {
			description.WriteString(fmt.Sprintf("• %s\n", link))
		}

		// First, create the complete embed.
		corpLogoURL := esiClient.GetRandomCorporationLogoURL()
		embed := &discordgo.MessageEmbed{
			Title:       "EVE Community Tools",
			Description: description.String(),
			Color:       0x1a81ab,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: corpLogoURL,
			},
		}

		// Then, send the response containing the embed.
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			}, // <-- The missing '}' was here.
		})
	},

	"unsubscribe": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
		for _, opt := range i.ApplicationCommandData().Options {
			optionMap[opt.Name] = opt
		}
		topicToRemove := optionMap["topic"].StringValue()
		channelID := i.ChannelID
		if channelOption, ok := optionMap["channel"]; ok {
			channelID = channelOption.ChannelValue(s).ID
		}

		originalTopics, found := subscriptions[channelID]
		if !found || len(originalTopics) == 0 {
			responseMessage := "⚠️ This channel isn't subscribed to any topics."
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: responseMessage, Flags: discordgo.MessageFlagsEphemeral},
			})
			return
		}

		newTopics := []string{}
		topicWasFound := false
		for _, topic := range originalTopics {
			if topic != topicToRemove {
				newTopics = append(newTopics, topic)
			} else {
				topicWasFound = true
			}
		}

		var responseMessage string
		if !topicWasFound {
			responseMessage = fmt.Sprintf("⚠️ Channel <#%s> was not subscribed to the '%s' topic.", channelID, topicToRemove)
		} else {
			subscriptions[channelID] = newTopics
			responseMessage = fmt.Sprintf("✅ Unsubscribed channel <#%s> from the '%s' topic.", channelID, topicToRemove)
			log.Printf("Subscription removed: Channel %s, Topic %s", channelID, topicToRemove)
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: responseMessage, Flags: discordgo.MessageFlagsEphemeral},
		})
	},

	"scout": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// 1. Defer the response so Discord knows we're working on it.
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		// 2. Get the system name from the user's input.
		options := i.ApplicationCommandData().Options
		systemName := options[0].StringValue()

		// 3. Call your API search function.
		searchResult, err := esiClient.performSearch(systemName)
		if err != nil {
			log.Printf("Error performing search for '%s': %v", systemName, err)
			errorMessage := "❌ An error occurred while contacting the search API."
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &errorMessage,
			})
			return
		}

		// 4. Filter the results to find the system.
		// (This assumes you have the findSystemInResults function we discussed)
		systemHit, err := findHitByType(searchResult, "system")
		if err != nil {
			// This error means a system wasn't found in the results.
			errorMessage := fmt.Sprintf("❌ Could not find a system named `%s`.", systemName)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &errorMessage,
			})
			return
		}

		systemDetails, err := esiClient.GetSystemDetails(systemHit.ID)
		if err != nil {
			log.Printf("Error fetching system details: %v", err)
			errorMessage := "❌ Could not retrieve detailed intel for that system (missing from local cache)."
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &errorMessage,
			})
			return
		}

		secStatusColor := getSecStatusColor(systemDetails.SecurityStatus)
		regionName := esiClient.GetRegionName(systemDetails.RegionID)
		constellationName := esiClient.GetConstellationName(systemDetails.ConstellationID)
		finalURL := fmt.Sprintf("https://eve-kill.com/system/%d", systemHit.ID)
		embed := &discordgo.MessageEmbed{
			Title: fmt.Sprintf("Intel Report: %s", systemHit.Name),
			URL:   finalURL,
			Color: secStatusColor, // A nice blue color
			Fields: []*discordgo.MessageEmbedField{
				{Name: "System Details", Value: finalURL, Inline: false},
				{Name: "System Report Link", Value: fmt.Sprintf("%v", systemHit.Name), Inline: false},
				{Name: "Region", Value: regionName, Inline: false},
				{Name: "Constellation", Value: constellationName, Inline: false},
				{Name: "Security Status", Value: fmt.Sprintf("%.2f", systemDetails.SecurityStatus), Inline: false},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Powered by Firehawk | Data from EVE ESI",
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		// 6. Send the final message.
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Embeds: []*discordgo.MessageEmbed{embed},
		})
		if err != nil {
			log.Printf("Failed to send scout followup message: %v", err)
		}
	},
	"group": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// 1. Defer the response so Discord knows we're working on it.
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		// 2. Get the system name from the user's input.
		options := i.ApplicationCommandData().Options
		corporationNames := options[0].StringValue()

		// 3. Call your API search function.
		searchResult, err := esiClient.performSearch(corporationNames)
		if err != nil {
			log.Printf("Error performing search for '%s': %v", corporationNames, err)
			errorMessage := "❌ An error occurred while contacting the search API."
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &errorMessage,
			})
			return
		}

		// 4. Filter the results to find the system.
		// (This assumes you have the findSystemInResults function we discussed)
		corpHit, err := findHitByType(searchResult, "corporation")
		if err != nil {
			// This error means a system wasn't found in the results.
			errorMessage := fmt.Sprintf("❌ Could not find a corporation named `%s`.", corporationNames)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &errorMessage,
			})
			return
		}

		// 5. Format a nice reply with an embed.
		// (You can add more fields here later, like security status, region, etc.)
		finalURL := fmt.Sprintf("https://eve-kill.com/corporation/%d", corpHit.ID)
		embed := &discordgo.MessageEmbed{
			Title: fmt.Sprintf("Intel Report: %s", corporationNames),
			Color: 0x00bfff, // A nice blue color
			Fields: []*discordgo.MessageEmbedField{
				{Name: "System Details", Value: finalURL, Inline: true},
				{Name: "EVE-KILL Link", Value: fmt.Sprintf("[View Corporation](%s)", finalURL)},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Source: EVE-KILL Search API",
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		// 6. Send the final message.
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Embeds: []*discordgo.MessageEmbed{embed},
		})
		if err != nil {
			log.Printf("Failed to send scout followup message: %v", err)
		}
	},
	"alliance": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// 1. Defer the response so Discord knows we're working on it.
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		// 2. Get the system name from the user's input.
		options := i.ApplicationCommandData().Options
		allianceName := options[0].StringValue()

		// 3. Call your API search function.
		searchResult, err := esiClient.performSearch(allianceName)
		if err != nil {
			log.Printf("Error performing search for '%s': %v", allianceName, err)
			errorMessage := "❌ An error occurred while contacting the search API."
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &errorMessage,
			})
			return
		}

		// 4. Filter the results to find the system.
		// (This assumes you have the findSystemInResults function we discussed)
		allianceHit, err := findHitByType(searchResult, "alliance")
		if err != nil {
			// This error means a system wasn't found in the results.
			errorMessage := fmt.Sprintf("❌ Could not find an alliance named `%s`.", allianceHit.Name)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &errorMessage,
			})
			return
		}

		// 5. Format a nice reply with an embed.
		// (You can add more fields here later, like security status, region, etc.)
		finalURL := fmt.Sprintf("https://eve-kill.com/alliance/%d", allianceHit.ID)
		embed := &discordgo.MessageEmbed{
			Title: fmt.Sprintf("Intel Report: %s", allianceHit.Name),
			Color: 0x00bfff, // A nice blue color
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Alliance Details", Value: finalURL, Inline: true},
				{Name: "EVE-KILL Link", Value: fmt.Sprintf("[View Alliance](%s)", finalURL)},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Source: EVE-KILL Search API",
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		// 6. Send the final message.
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Embeds: []*discordgo.MessageEmbed{embed},
		})
		if err != nil {
			log.Printf("Failed to send scout followup message: %v", err)
		}
	},
}

package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
)

// You will also need a getAPIStatus() function and a global esiClient variable defined elsewhere.

var subscriptions = make(map[string]map[string][]string)

// Global variable to re-use for subscribe/unsubscribe process.
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
		Description: "Live Tranquility Status",
	},
	{
		Name:        "scout",
		Description: "Provides intel on a specific solar system.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "system_name",
				Description: "The name of the solar system to scout.",
				Required:    true,
			},
		},
	},
	{
		Name:        "lookup",
		Description: "Lookup an EVE Online character by name",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "character_name",
				Description: "The name of the charachter you want to look up (I need precise spelling)",
				Required:    true,
			},
		},
	},
	{
		Name:        "subscribe",
		Description: "Subscribe this channel to a killmail feed of choice",
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
				Description: "The channel to unsubscribe (defaults to the current channel)",
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
				Content: "Server possibly offline, or you are on a VPN...",
			})
			return
		}

		uptime := time.Since(status.StartTime)
		uptimeStr := fmt.Sprintf("%d hours, %d minutes", int(uptime.Hours()), int(uptime.Minutes())%60)

		embed := &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    "Tranquility",
				URL:     "https://esi.evetech.net/ui/",
				IconURL: "https://web.ccpgamescdn.com/forums/img/eve_favicon.ico",
			},
			Title: "Server Status - Online",
			Color: 0x00ff00,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: fmt.Sprintf("https://images.evetech.net/corporations/%d/logo?size=128", getRandomCachedCorpLogo()),
			},
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Players Online", Value: fmt.Sprintf("%d", status.Players), Inline: true},
				{Name: "Uptime", Value: uptimeStr, Inline: true},
				{Name: "Version", Value: status.ServerVersion},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Powered by Firehawk",
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
			log.Printf("Error looking up character ID for '%s': %v", charName, err)
			errormessage := fmt.Sprintf("❌ Error looking up character ID for '%s'.", charName)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &errormessage,
			})
			return
		}

		finalURL := fmt.Sprintf("https://eve-kill.com/character/%d", charID)
		embed := &discordgo.MessageEmbed{
			Title:       "View on EVE-KILL",
			Description: fmt.Sprintf("Killboard for %s", charName),
			URL:         finalURL,
			Color:       0x42b6f5,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: fmt.Sprintf("https://images.evetech.net/characters/%d/portrait", charID),
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{embed},
		})
		if err != nil {
			log.Printf("Failed to edit interaction for lookup: %v", err)
		}
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

	"scout": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		systemName := i.ApplicationCommandData().Options[0].StringValue()

		systemID, err := esiClient.GetSystemID(systemName)
		if err != nil {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &[]string{"❌ Could not find a solar system with that name."}[0],
			})
			return
		}

		systemInfo, err := esiClient.GetSystemInfo(systemID)
		if err != nil {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &[]string{"❌ Could not retrieve details for that system."}[0],
			})
			return
		}

		embed := &discordgo.MessageEmbed{
			Title: "System Intel: " + systemInfo.Name,
			Color: 0x42b6f5, // Blue
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Security Status",
					Value:  fmt.Sprintf("%.1f", systemInfo.SecurityStatus),
					Inline: true,
				},
				{
					Name:   "Stargates",
					Value:  fmt.Sprintf("%d", len(systemInfo.Stargates)),
					Inline: true,
				},
				{
					Name:   "Stations",
					Value:  fmt.Sprintf("%d", len(systemInfo.Stations)),
					Inline: true,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("System ID: %d", systemInfo.SystemID),
			},
		}

		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{embed},
		})
	},
}

func getRandomCachedCorpLogo() int {
	esiClient.cacheMutex.RLock()
	defer esiClient.cacheMutex.RUnlock()

	if len(esiClient.corporationNames) == 0 {
		return 1000001 // default to CONCORD as error handling
	}
	corpIDs := make([]int, 0, len(esiClient.corporationNames))
	for id := range esiClient.corporationNames {
		corpIDs = append(corpIDs, id)
	}

	randomIndex := rand.Intn(len(corpIDs))
	return corpIDs[randomIndex]
}

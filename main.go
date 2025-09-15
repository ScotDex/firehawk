package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var esiClient *ESIClient

const cacheFilePath = "esi_cache.json"
const systemCachePath = "systems.json"

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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("DISCORD_BOT_TOKEN environment variable not set")
	}

	log.Println("Bot token loaded successfully")
	esiClient = NewESIClient("themadlyscientific@gmail.com")
	if err := esiClient.LoadSystemCache(systemCachePath); err != nil {
		log.Printf("WARNING: could not load static system cache: %v", err)
	}
	if err := esiClient.LoadCacheFromFile(cacheFilePath); err != nil {
		log.Printf("Warning: could not load dynamic ESI cache: %v", err)
	}

	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	dg.AddHandler(interactionCreate) // Using the named handler function

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer dg.Close()

	// Start the killmail poller. It no longer needs a channel ID.
	goSafely(func() {
		killmailPoller(dg)
	})

	log.Println("Registering Commands")
	for _, cmd := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("Cannot create '%v' command: %v", cmd.Name, err)
		}
	}

	log.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Shutting down bot. Saving cache...")
	if err := esiClient.SaveCacheToFile(cacheFilePath); err != nil {
		log.Printf("Error saving ESI cache: %v", err)
	}
}

// interactionCreate is the handler for all slash command interactions.
func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			handler(s, i)
		}
	}
}

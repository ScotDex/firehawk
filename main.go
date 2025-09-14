package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// The only global needed is the client itself.
var esiClient *ESIClient

// Create one shared client for the entire application to use.
var sharedHttpClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		DisableCompression: false, // Enable Gzip
	},
}

const cacheFilePath = "esi_cache.json"
const systemCachePath = "systems.json" // Renamed for clarity

// goSafely launches a function in a new goroutine and recovers from panics.
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

	// Create the client and load all caches.
	esiClient = NewESIClient("themadlyscientific@gmail.com")

	// Load static system data using the dedicated client method.
	if err := esiClient.LoadSystemCache(systemCachePath); err != nil {
		log.Printf("WARNING: could not load static system cache: %v", err)

	}
	// Load the dynamic cache.
	if err := esiClient.LoadCacheFromFile(cacheFilePath); err != nil {
		log.Printf("Warning: could not load dynamic ESI cache: %v", err)
	}

	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				handler(s, i)
			}
		}
	})

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer dg.Close()

	// Use os.Getenv for the channel ID for better configuration
	killmailChannelID := os.Getenv("KILLMAIL_CHANNEL_ID")
	if killmailChannelID != "" {
		goSafely(func() {
			killmailPoller(dg, killmailChannelID)
		})
	}

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

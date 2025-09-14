package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var esiClient *ESIClient
var systemsData map[int]ESISystemInfo

// Create one shared client for the entire application to use.
var sharedHttpClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		DisableCompression: false, // Enable Gzip
	},
}

const targetChannelID = "1415431475368693823"
const cacheFilePath = "esi_cache.json"
const systemCache = "systems.json"

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
	esiClient = NewESIClient("themadlyscientific@gmail.com")
	if err := esiClient.LoadCacheFromFile(cacheFilePath); err != nil {
		log.Printf("Warning: could not load ESI cache: %v", err)
	}

	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// Load static system data, but don't crash if it fails.
	fileData, err := os.ReadFile(systemCache)
	if err != nil {
		log.Printf("WARNING: Could not read systems.json: %v. System lookups may fail.", err)
	} else {
		if err := json.Unmarshal(fileData, &systemsData); err != nil {
			log.Printf("WARNING: Could not parse systems.json: %v. System lookups may fail.", err)
		} else {
			log.Printf("Successfully loaded %d systems into cache.", len(systemsData))
		}
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

	// Start the killmail poller safely.
	goSafely(func() {
		killmailPoller(dg, targetChannelID)
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

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

const targetChannelID = "1415431475368693823"
const cacheFilePath = "esi_cache.json"

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
	// Start the killmail poller in a separate goroutine
	go killmailPoller(dg, targetChannelID)

	log.Println("Logging Commands")
	for _, cmd := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("Cannot create '%v' command: %v", cmd.Name, err)
		}
	}

	// Kill command for bot to respond to.

	log.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Shutting down bot. Saving cache...")
	if err := esiClient.SaveCacheToFile(cacheFilePath); err != nil {
		log.Printf("Error saving ESI cache: %v", err)
	}
}

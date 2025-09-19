package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// Global ESIClient variable
var esiClient *ESIClient

// --- Constants ---
const (
	cacheFilePath        = "esi_cache.json"
	systemCachePath      = "systems.json"
	killmailWebSocketURL = "wss://ws.eve-kill.com/killmails" // Correct WebSocket URL

)

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

	dg.AddHandler(interactionCreate)

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer dg.Close()

	// Start background services
	go startHealthCheckServer()
	go killmailStreamer(dg, esiClient) // Correctly start the streamer with dependencies

	// Register commands after the bot is running
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

func killmailStreamer(s *discordgo.Session, esi *ESIClient) {
	log.Println("Kicking off web socket connection")

	for { // Main reconnection loop
		conn, _, err := websocket.DefaultDialer.Dial(killmailWebSocketURL, nil)
		if err != nil {
			log.Printf("Error connecting to WebSocket: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}
		log.Println("Web socket connected - streaming messages.")

		// Handles low-level protocol pings to keep the connection alive.
		pongWait := 60 * time.Second
		conn.SetReadDeadline(time.Now().Add(pongWait))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})

		if err := conn.WriteMessage(websocket.TextMessage, []byte("all")); err != nil {
			log.Printf("Error subscribing to killmail feed: %v", err)
			conn.Close()
			continue
		}
		log.Println("Subscribed to 'all' killmails topic, filters will apply accordingly.")

		for { // Message reading loop
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error, disconnecting:", err)
				break
			}

			conn.SetReadDeadline(time.Now().Add(pongWait))

			// Handles application-level pings by echoing the timestamp in a JSON reply.
			var msg SocketMessage
			if err := json.Unmarshal(message, &msg); err == nil && msg.Type == "ping" {
				var pingMsg PingMessage
				if err := json.Unmarshal(message, &pingMsg); err == nil {
					log.Println("Recieved Ping, responding with Pong")

					pongReply := fmt.Sprintf(`{"type":"pong","timestamp":"%s"}`, pingMsg.Timestamp)

					if err := conn.WriteMessage(websocket.TextMessage, []byte(pongReply)); err != nil {
						log.Printf("Error sending application pong: %v", err)
					}
				}
				continue // Don't process this ping message any further.
			}

			// If it's not a ping, pass it to the killmail handler.
			msgCopy := make([]byte, len(message))
			copy(msgCopy, message)
			go HandleKillmailMessage(s, msgCopy)
		}

		conn.Close()
		log.Println("Disconnected. Attempting to reconnect...")
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

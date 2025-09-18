package main

import (
	"encoding/json"
	"time"
)

// This file is the ONLY place these structs should be defined.

type SocketMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type PingMessage struct {
	Timestamp string `json:"timestamp"`
}

type KillmailData struct {
	// This outer 'killmail' object was the missing layer in your previous struct.
	Killmail struct {
		KillmailID   int       `json:"killmail_id"`
		KillmailTime time.Time `json:"kill_time"`
		SystemID     int       `json:"system_id"`
		SystemName   string    `json:"system_name"`
		TotalValue   float64   `json:"total_value"`
		IsNpc        bool      `json:"is_npc"`
		IsSolo       bool      `json:"is_solo"`
		RegionName   struct {
			En string `json:"en"`
		} `json:"region_name"`
		Victim struct {
			CharacterID     int    `json:"character_id"`
			CharacterName   string `json:"character_name"`
			CorporationID   int    `json:"corporation_id"`
			CorporationName string `json:"corporation_name"`
			ShipID          int    `json:"ship_id"`
			ShipName        struct {
				En string `json:"en"`
			} `json:"ship_name"`
		} `json:"victim"`
		Attackers []struct {
			CharacterID     int    `json:"character_id"`
			CharacterName   string `json:"character_name"`
			CorporationID   int    `json:"corporation_id"`
			CorporationName string `json:"corporation_name"`
			ShipID          int    `json:"ship_id"`
			ShipName        struct {
				En string `json:"en"`
			} `json:"ship_name"`
			FinalBlow bool `json:"final_blow"`
		} `json:"attackers"`
	} `json:"killmail"` // This is the key for the nested object.

	// These fields are at the top level, outside the 'killmail' object.

}

// In models.go

type EnrichedKillmailData struct {
	KillmailID     int       `json:"killmail_id"`
	KillTime       time.Time `json:"kill_time"`
	SystemID       int       `json:"system_id"`
	SystemName     string    `json:"system_name"`
	SystemSecurity float64   `json:"system_security"` // <-- Needed for high/low/null
	RegionID       int       `json:"region_id"`       // <-- NEW: For w-space/abyssal
	TotalValue     float64   `json:"total_value"`
	IsNpc          bool      `json:"is_npc"`
	IsSolo         bool      `json:"is_solo"`
	Victim         struct {
		CharacterName   string `json:"character_name"`
		CorporationName string `json:"corporation_name"`
		ShipID          int    `json:"ship_id"`
		ShipName        struct {
			En string `json:"en"`
		} `json:"ship_name"`
		ShipGroupID int `json:"ship_group_id"` // <-- NEW: For ship classes
	} `json:"victim"`
	Attackers []struct {
		// ...
	} `json:"attackers"`
}

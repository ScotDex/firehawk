# Firehawk Bot 🔥 (DONT SELF HOST YET!)

[![Go Version](https://img.shields.io/badge/go-1.22-blue.svg)](https://golang.org)
[![Discord](https://img.shields.io/discord/YOUR_SERVER_ID?label=Support%20Server)](https://discord.gg/YOUR_INVITE_LINK)
[![Status](https://img.shields.io/badge/status-probably%20broken-red.svg)](https://github.com/YOUR_USERNAME/firehawk)



Another EVE Online Discord bot, because apparently there aren't enough already. This one is badly written in Golang (Cos I hate listening to tutorial videos, learn by building damn it) but the positive is that the language is designed to be fast, so if you see lag at any point its most likely due to discord unable to keep up. Inspired by the actually good [Firetail Bot](https://forums.eveonline.com/t/firetail-eve-discord-bot/45283).

Please dont try and self host just yet, as this is still in a testing phase.

*(A screenshot proving it occasionally works)*

<img width="480" height="204" alt="image" src="https://github.com/user-attachments/assets/a6d42196-813f-444e-9701-85b94d3d99cc" />

---

## ✨ Features (The Things It Does)

* **Killmail Spam**: Get real-time killmails from `eve-kill.com` piped directly into your channel of choice.
* **Endless Filtering**: Use the ridiculously long list of subscription topics to pretend you're only getting the "important" kills.
* **Mostly Correct Data**: Fetches and caches names for things so you don't have to remember what ID `30000142` is.
* **Server Status**: Tells you if Tranquility is online, so you know who to blame when you can't log in.
* **Character Lookup**: Look up a name and get their killboard.
* **Location Scout**: Get some information on the location you are about to visit for you to ignore and get killed by. (not finished yet)
---

## 🚀 Commands (The Buttons You Can Press)

It's a slash command bot. You know the drill.

| Command                             | Description                                            | Example                               |
| ----------------------------------- | ------------------------------------------------------ | ------------------------------------- |
| `/status`                           | Checks if the server hamster is still running.         | `/status`                             |
| `/lookup <character_name>`          | Finds a character's public record of shame.            | `/lookup The Mittani`                 |
| `/subscribe <topic> [channel]`      | Starts the spam of killmails in a channel.         | `/subscribe topic:Big Kills`          |
| `/unsubscribe <topic> [channel]`    | Mercifully stops the spam.                         | `/unsubscribe topic:All Kills`        |

---

## 🔗 How to Get It

It's a public bot. You don't build it, you just invite it and hope for the best.

---

## 🗺️ Development Roadmap

A list of things I'll probably get around to building eventually.

### In Progress
* `/price <item_name>`: So you can be disappointed by market prices without logging in.
* `/scout <system_name>`: Get intel on a solar system.
* `/group <name>`: Look up corps and alliances.

### Planned
* Group Lookup: Looks up information about EVE corporations and alliances.
* Price: Checks the market price for items in-game.
* Whatever else I feel like doing 
* Dad Joke API - because why not?
* Weather API because, global warming means its sunny in december these days...

---

## ❤️ Acknowledgements

* This bot stands on the shoulders of giants, mostly the **[Firetail Bot](https://forums.eveonline.com/t/firetail-eve-discord-bot/45283)**.
* Killmail data is graciously provided by the endless stream at **[Eve-Kill.com](https://eve-kill.com/)**.
* Game data comes from the **EVE Online ESI API**, when it's working.

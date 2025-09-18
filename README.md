# Firehawk Bot 🔥

A high-performance, self-hostable EVE Online Discord bot written in Go.

[![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org)
[![Status](https://img.shields.io/badge/status-active-green.svg)](https://github.com/ScotDex/firehawk)
[![Discord](https://img.shields.io/discord/790348216587255808?label=Support%20Server&logo=discord)](https://discord.gg/tas2ggVUr3)

---

## 📖 About Firehawk

Firehawk is a self-hosted Discord bot for EVE Online, providing real-time killmail alerts and useful in-game information lookups. Inspired by the excellent [Firetail Bot](https://forums.eveonline.com/t/firetail-eve-discord-bot/45283), this project was built as a learning exercise to explore high-performance, concurrent applications in Go.

It's designed to be lightweight, fast, and easily run by anyone using Docker.



---

## ✨ Features

* **📰 Real-time Killmail Subscriptions:** Subscribe channels to filtered killmail feeds from `eve-kill.com`. Get alerts for big kills, solo kills, specific regions, and more.
* **🛰️ Advanced Intel Lookups:** Get detailed, cached information on in-game entities like solar systems, corporations, and alliances.
* **⚡ High-Performance Caching:** Utilizes a pre-seeded static cache for system data and a dynamic cache for API results to make lookups incredibly fast.
* **🛠️ Utilities:** Includes commands for checking server status, looking up characters, and listing useful third-party tools.

---

## 🚀 Getting Started (Self-Hosting)

Firehawk is designed to be self-hosted using Docker. Follow these instructions to get your own instance running.

### Prerequisites

Before you begin, you will need:
* [Docker](https://www.docker.com/get-started) and Docker Compose installed on your system.
* A **Discord Bot Token**. You can get one by creating a new application in the [Discord Developer Portal](https://discord.com/developers/applications).
*test

### Installation

1.  **Clone the Repository**
    Open your terminal, navigate to where you want to store the bot, and run the following command:
    ```bash
    git clone [https://github.com/ScotDex/firehawk.git](https://github.com/ScotDex/firehawk.git)
    cd firehawk
    ```

2.  **Create Your Configuration File**
    Copy the example configuration file to create your own. This file will store your secret bot token.
    ```bash
    cp .env.example .env
    ```

3.  **Edit the `.env` File**
    Open the newly created `.env` file in a text editor and add your Discord bot token and a contact email (this is required by the ESI API).
    ```env
    DISCORD_BOT_TOKEN=YOUR_SECRET_BOT_TOKEN_HERE
    ESI_CONTACT_EMAIL=your-email@example.com
    ```

4.  **Build and Run the Bot**
    Use Docker Compose to build the container and run it in the background.
    ```bash
    docker-compose up --build -d
    ```

Your Firehawk bot should now be online and ready to be invited to your Discord server.

---
## 📋 Command Reference

| Command                  | Description                                | Example                            |
| ------------------------ | ------------------------------------------ | ---------------------------------- |
| `/status`                | Checks the live status of the EVE server.  | `/status`                          |
| `/scout [system]`        | Provides a detailed intel report on a system. | `/scout Jita`                      |
| `/group [corporation]`   | Looks up a corporation.                    | `/group Pandemic Horde`            |
| `/alliance [alliance]`   | Looks up an alliance.                      | `/alliance Goonswarm Federation`   |
| `/lookup [character]`    | Provides a killboard link for a character. | `/lookup The Mittani`              |
| `/tools`                 | Lists useful third-party websites.         | `/tools`                           |
| `/subscribe [topic]`     | Subscribes the channel to a killmail feed. | `/subscribe topic:Big Kills`       |
| `/unsubscribe [topic]`   | Unsubscribes the channel from a feed.      | `/unsubscribe topic:All Kills`     |

---

## ❤️ Data Sources & Acknowledgements

* **Game Data:** All core game data is sourced from the official [**EVE Online ESI API**](https://esi.evetech.net/).
* **Killmail Data:** Real-time killmail data is provided by the [**Eve-Kill.com**](https://eve-kill.com/) RedisQ feed.
* **Inspiration:** This project was heavily inspired by the legendary [**Firetail Bot**](https://forums.eveonline.com/t/firetail-eve-discord-bot/45283).

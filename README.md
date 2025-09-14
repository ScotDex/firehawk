# Firehawk Bot 🔥  
**(Don’t self-host yet, seriously!)**

[![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org)  
[![Discord](https://img.shields.io/discord/YOUR_SERVER_ID?label=Support%20Server)](https://discord.gg/tas2ggVUr3)  
[![Status](https://img.shields.io/badge/status-probably%20broken-red.svg)](https://github.com/ScotDex/firehawk)  

---

## 📖 What Is This?
Another EVE Online Discord bot—because apparently there aren’t enough already.  

Written badly in **Go** (because I refuse to watch tutorial videos and prefer to learn by building), but the upside is: Go is fast. So if you notice lag, it’s probably Discord, not me.  

Inspired by the *actually good* [Firetail Bot](https://forums.eveonline.com/t/firetail-eve-discord-bot/45283).  

⚠️ **Don’t try to self-host this yet**. It’s still in a testing phase.  

*(Proof it sometimes works)*  
<img width="480" height="204" alt="image" src="https://github.com/user-attachments/assets/a6d42196-813f-444e-9701-85b94d3d99cc" />  

---

## ✨ Features
- **Killmail Spam** → Real-time killmails from `eve-kill.com` piped into your channel.  
- **Endless Filtering** → Subscribe to specific killmail topics so you can pretend you only see the “important” ones.  
- **Mostly Correct Data** → Fetches & caches IDs so you don’t have to remember what `30000142` is.  
- **Server Status** → Tells you if Tranquility is alive.  
- **Character Lookup** → Pulls a character’s killboard info.  
- **Group & Alliance Lookup** → Groups and alliances are now supported!  
- **Tools Command** → Lists useful third-party EVE Online tools.  
- **Location Scout** → Info on systems before you blindly jump in (WIP).  

---

## 🚀 Commands
Slash commands only. You know the drill.  

| Command                          | Description                                 | Example                          |
| -------------------------------- | ------------------------------------------- | -------------------------------- |
| `/status`                        | Checks if the TQ hamster is alive.          | `/status`                        |
| `/lookup <character_name>`       | Shows a character’s public record of shame. | `/lookup The Mittani`            |
| `/group <group_name>`            | Looks up information about a group.         | `/group Pandemic Horde`          |
| `/alliance <alliance_name>`      | Looks up information about an alliance.     | `/alliance Goonswarm Federation` |
| `/tools`                         | Lists useful third-party tools.             | `/tools`                         |
| `/subscribe <topic> [channel]`   | Starts killmail spam in a channel.          | `/subscribe topic:Big Kills`     |
| `/unsubscribe <topic> [channel]` | Stops the spam (mercifully).                | `/unsubscribe topic:All Kills`   |

---

## 🔗 How to Get It
It’s a **public bot**. No building required—just invite it and pray.  

### Quick Invite Link
👉 [**Invite Firehawk to Your Server**](https://discord.com/oauth2/authorize?client_id=YOUR_CLIENT_ID&scope=bot%20applications.commands&permissions=8)  

*(Replace `YOUR_CLIENT_ID` with the bot’s actual client ID.)*  

---

## 🛠️ Roadmap / TODO
- Refine **embeds** for better presentation.  
- Add **AI connection?** 🤔  (Because I want to be trendy)
- Dad Joke API (because why not).  
- Weather API (because it’s somehow sunny in December).  
- Static data improvements.  
- Help function.  (When people dont want to read this)
- Broadcast function.  (Intrusive)
- Market price lookup command.  (Can I be fucked?)
- Integrate thera bot?

---

## ❤️ Acknowledgements
- Inspired by the legendary [**Firetail Bot**](https://forums.eveonline.com/t/firetail-eve-discord-bot/45283).  
- Killmail data from **[Eve-Kill.com](https://eve-kill.com/)**.  
- Game data from the **EVE Online ESI API** (when it feels like working).  

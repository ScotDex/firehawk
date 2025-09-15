# Firehawk Bot 🔥  

See the bot working here https://discord.gg/tas2ggVUr3

---

## 📖 What Is This?
Another EVE Online Discord bot—because apparently there aren’t enough already.  

Written in GO - which was relatively simple. Although this could have been done better the example and code is here for you to use as you wish.

Currently going through 24h testing to ensure the code is robust enough to be hosted.

Inspired by the *actually good* [Firetail Bot](https://forums.eveonline.com/t/firetail-eve-discord-bot/45283).  

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
- **Tools Command** → Lists useful third-party EVE Online tools.  (Feel free to add more)
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

## How do I install?

Visit the wiki for detailed information for your tech guy to look at.

---

## 🛠️ Roadmap / TODO
- Refine **embeds** for better presentation.
- Add more sub choices.
- Add **AI connection?** 🤔  (Because I want to be trendy)
- Dad Joke API (because why not).  
- Weather API (because it’s somehow sunny in December).  
- Static data improvements.  
- Help function.  (When people dont want to read this)
- Market price lookup command.  (Can I be fucked?)
- Integrate thera bot?

---

## ❤️ Acknowledgements
- Inspired by the legendary [**Firetail Bot**](https://forums.eveonline.com/t/firetail-eve-discord-bot/45283).  
- Killmail data from **[Eve-Kill.com](https://eve-kill.com/)**.  
- Game data from the **EVE Online ESI API** (when it feels like working).  

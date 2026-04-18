# RSS Bot for XMPP

A bot that delivers RSS feed updates directly to your XMPP inbox.

## Description

RSS Bot is an XMPP bot that polls RSS/Atom feeds at configurable intervals and sends new items to users via XMPP direct messages. Users manage their subscriptions through simple chat commands.

## Features

- Add, delete, and list RSS feed subscriptions
- Enable/disable individual feeds
- Automatic feed polling at configurable intervals
- Automatic feed disabling after 5 consecutive fetch errors
- SQLite database for persistent storage

---

## Configuration

Create a `config.yaml` file with the following options:

```yaml
xmpp:
  host: "xmpp.example.com"    # XMPP server hostname
  port: 5222                  # XMPP server port
  jid: "rssbot@example.com"   # Bot's JID
  password: "your-password"   # Bot's password

rss:
  poll_interval: 3600         # Polling interval in seconds (default: 3600 = 1 hour)
  user_agent: "RSSBot/1.0"    # User-Agent header for fetching feeds

database:
  type: "sqlite"              # Database type: sqlite or postgres
  path: "./data/rss.db"       # Path to SQLite database file (if type=sqlite)

  # For PostgreSQL use:
  # type: "postgres"
  # host: "localhost"
  # port: 5432
  # user: "rssbot"
  # password: "your-password"
  # dbname: "rssbot"
```

### Configuration Options

| Section | Option | Description | Default |
|---------|--------|-------------|---------|
| `xmpp` | `host` | XMPP server hostname | - |
| `xmpp` | `port` | XMPP server port | 5222 |
| `xmpp` | `jid` | Bot's JID (e.g., bot@example.com) | - |
| `xmpp` | `password` | Bot's XMPP password | - |
| `rss` | `poll_interval` | Seconds between feed checks | 3600 |
| `rss` | `user_agent` | HTTP User-Agent for fetching | "RSSBot/1.0" |
| `database` | `type` | Database type: sqlite or postgres | "sqlite" |
| `database` | `path` | SQLite database path | "./rss.db" |
| `database` | `host` | PostgreSQL host | "localhost" |
| `database` | `port` | PostgreSQL port | 5432 |
| `database` | `user` | PostgreSQL user | - |
| `database` | `password` | PostgreSQL password | - |
| `database` | `dbname` | PostgreSQL database name | "rssbot" |

---

## Environment Variables

All configuration options can be set via environment variables:

| Variable | Description |
|----------|-------------|
| `XMPP_HOST` | XMPP server hostname |
| `XMPP_PORT` | XMPP server port |
| `XMPP_JID` | Bot's JID |
| `XMPP_PASSWORD` | Bot's XMPP password |
| `RSS_POLL_INTERVAL` | Polling interval in seconds |
| `RSS_USER_AGENT` | HTTP User-Agent for fetching |
| `DB_TYPE` | Database type: sqlite or postgres |
| `DB_PATH` | SQLite database path |
| `DB_HOST` | PostgreSQL host |
| `DB_PORT` | PostgreSQL port |
| `DB_USER` | PostgreSQL user |
| `DB_PASSWORD` | PostgreSQL password |
| `DB_NAME` | PostgreSQL database name |

Environment variables override values from config.yaml.

---

## Commands

Send direct messages to the bot:

| Command | Description |
|---------|-------------|
| `add <url>` | Subscribe to an RSS feed |
| `delete <url>` | Unsubscribe from a feed |
| `disable <url>` | Temporarily disable a feed |
| `enable <url>` | Re-enable a disabled feed |
| `list` | List all your subscriptions |
| `help` | Show available commands |

---

## Running

### Using Docker Compose

```bash
docker-compose up --build
```

### Using Docker

```bash
# Build
docker build -t rssbot .

# Run
docker run --rm -v $(pwd)/config.yaml:/app/config.yaml -v $(pwd)/data:/app/data rssbot
```

### Local Development

```bash
cd src
go build -o rssbot .
./rssbot -config ../config.yaml
```

---

## Database Schema

The bot uses SQLite with the following schema:

```sql
CREATE TABLE feeds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_jid TEXT NOT NULL,
    url TEXT NOT NULL,
    title TEXT,
    last_guid TEXT,
    last_poll TIMESTAMP,
    disabled INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    UNIQUE(user_jid, url)
);
```

---

## Project Structure

```
.
├── src/
│   ├── main.go        # Entry point
│   ├── config.go      # Configuration handling
│   ├── db.go          # SQLite database layer
│   ├── fetcher.go     # RSS feed fetching
│   ├── bot.go         # XMPP bot logic
│   └── go.mod         # Go module dependencies
├── config.yaml        # Configuration file
├── Dockerfile         # Docker build
├── docker-compose.yml # Docker Compose
└── README.md          # This file
```

---

## License

MIT

---

## Author

Developed by [inCloudy](https://incloudy.com.tr)
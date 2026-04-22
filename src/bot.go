package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	xmpp "gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
)

type Bot struct {
	cfg     *Config
	db      DB
	fetcher *Fetcher
	client  *xmpp.Client
	mu      sync.Mutex
	running bool
}

func NewBot(cfg *Config, db DB, fetcher *Fetcher) *Bot {
	return &Bot{
		cfg:     cfg,
		db:      db,
		fetcher: fetcher,
	}
}

func (b *Bot) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		return nil
	}

	config := xmpp.Config{
		TransportConfiguration: xmpp.TransportConfiguration{
			Address: fmt.Sprintf("%s:%d", b.cfg.XMPP.Host, b.cfg.XMPP.Port),
		},
		Jid:          b.cfg.XMPP.JID,
		Credential:   xmpp.Password(b.cfg.XMPP.Password),
		StreamLogger: os.Stdout,
		Insecure:     true,
	}

	router := xmpp.NewRouter()
	router.HandleFunc("message", b.handleMessage)

	client, err := xmpp.NewClient(&config, router, b.errorHandler)
	if err != nil {
		return err
	}

	if err := client.Connect(); err != nil {
		return err
	}

	b.client = client
	b.running = true

	go b.pollingLoop()

	log.Printf("Bot started as %s", b.cfg.XMPP.JID)
	return nil
}

func (b *Bot) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return nil
	}

	b.running = false
	if b.client != nil {
		b.client.Disconnect()
	}
	return nil
}

func (b *Bot) errorHandler(err error) {
	log.Printf("XMPP error: %v", err)
}

func (b *Bot) handleMessage(s xmpp.Sender, p stanza.Packet) {
	msg, ok := p.(stanza.Message)
	if !ok {
		return
	}

	if msg.Body == "" {
		return
	}

	if msg.Type == "groupchat" {
		return
	}

	userJID := strings.Split(msg.From, "/")[0]
	cmd := strings.TrimSpace(msg.Body)

	log.Printf("Command from %s: %s", userJID, cmd)

	response := b.handleCommand(userJID, cmd)
	b.sendMessage(msg.From, response)
}

func (b *Bot) handleCommand(userJID, cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return usageText()
	}

	switch parts[0] {
	case "add":
		if len(parts) < 2 {
			return "Usage: add <feed-url>"
		}
		return b.addFeed(userJID, parts[1])
	case "delete", "del", "remove":
		if len(parts) < 2 {
			return "Usage: delete <feed-url>"
		}
		return b.deleteFeed(userJID, parts[1])
	case "disable":
		if len(parts) < 2 {
			return "Usage: disable <feed-url>"
		}
		return b.setFeedDisabled(userJID, parts[1], 1)
	case "enable":
		if len(parts) < 2 {
			return "Usage: enable <feed-url>"
		}
		return b.setFeedDisabled(userJID, parts[1], 0)
	case "list":
		return b.listFeeds(userJID)
	case "help":
		return usageText()
	default:
		return "Unknown command. Use 'help' for available commands."
	}
}

func (b *Bot) addFeed(userJID, url string) string {
	items, title, err := b.fetcher.Fetch(url)
	if err != nil {
		return fmt.Sprintf("Error fetching feed: %v", err)
	}

	if title == "" {
		title = url
	}

	var lastGUID string
	if len(items) > 0 {
		lastGUID = items[0].GUID
	}

	feedID, err := b.db.AddFeed(userJID, url, title)
	if err != nil {
		return fmt.Sprintf("Error adding feed: %v", err)
	}

	if lastGUID != "" {
		b.db.UpdateFeedLastGUID(feedID, lastGUID)
	}

	return fmt.Sprintf("Added feed: %s", title)
}

func (b *Bot) deleteFeed(userJID, url string) string {
	affected, err := b.db.RemoveFeed(userJID, url)
	if err != nil {
		return fmt.Sprintf("Error deleting feed: %v", err)
	}
	if affected == 0 {
		return "Feed not found"
	}
	return "Feed deleted"
}

func (b *Bot) setFeedDisabled(userJID, url string, disabled int) string {
	affected, err := b.db.SetFeedDisabled(userJID, url, disabled)
	if err != nil {
		return fmt.Sprintf("Error updating feed: %v", err)
	}
	if affected == 0 {
		return "Feed not found"
	}
	if disabled == 1 {
		return "Feed disabled"
	}
	return "Feed enabled"
}

func (b *Bot) listFeeds(userJID string) string {
	feeds, err := b.db.ListFeeds(userJID)
	if err != nil {
		return fmt.Sprintf("Error listing feeds: %v", err)
	}
	if len(feeds) == 0 {
		return "No feeds registered. Use 'add <url>' to add a feed."
	}

	var lines []string
	for _, f := range feeds {
		lines = append(lines, fmt.Sprintf("- %s", f.URL))
		if f.Title != "" && f.Title != f.URL {
			lines = append(lines, fmt.Sprintf("  Title: %s", f.Title))
		}
		if f.Disabled != 0 {
			lines = append(lines, "  [DISABLED]")
		}
	}
	return "Your feeds:\n" + strings.Join(lines, "\n")
}

func (b *Bot) sendMessage(to, body string) {
	msg := stanza.Message{
		Attrs: stanza.Attrs{To: to},
		Body:  body,
	}
	b.client.Send(msg)
}

func (b *Bot) pollingLoop() {
	ticker := time.NewTicker(time.Duration(b.cfg.RSS.PollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.checkFeeds()
		case <-context.Background().Done():
			return
		}
	}
}

func (b *Bot) checkFeeds() {
	feeds, err := b.db.GetAllFeeds()
	if err != nil {
		log.Printf("Error getting feeds: %v", err)
		return
	}

	for _, feed := range feeds {
		if feed.Disabled != 0 {
			continue
		}
		items, _, err := b.fetcher.Fetch(feed.URL)
		if err != nil {
			log.Printf("Error fetching %s: %v", feed.URL, err)
			count, err := b.db.IncrementErrorCount(feed.ID)
			if err == nil && count >= 5 {
				b.db.SetFeedDisabled(feed.UserJID, feed.URL, 1)
				b.sendMessage(feed.UserJID, fmt.Sprintf("Feed disabled due to 5 consecutive errors: %s\nError: %v", feed.URL, err))
			}
			continue
		}

		b.db.ResetErrorCount(feed.ID)

		var newItems []FeedItem
		for _, item := range items {
			if item.GUID != feed.LastGUID {
				newItems = append(newItems, item)
			} else {
				break
			}
		}

		if len(newItems) > 0 {
			b.db.UpdateFeedLastGUID(feed.ID, newItems[0].GUID)
			b.notifyNewItems(feed.UserJID, feed.Title, newItems)
		}
	}
}

func (b *Bot) notifyNewItems(userJID, feedTitle string, items []FeedItem) {
	var lines []string
	for _, item := range items {
		if item.Title != "" {
			lines = append(lines, fmt.Sprintf("%s\n%s", item.Title, item.Link))
		}
	}
	if len(lines) > 0 {
		msg := fmt.Sprintf("New from %s:\n%s", feedTitle, strings.Join(lines, "\n\n"))
		b.sendMessage(userJID, msg)
	}
}

func usageText() string {
	return `Available commands:
  add <url>      - Add an RSS feed
  delete <url>   - Remove an RSS feed
  disable <url>  - Disable a feed
  enable <url>   - Enable a disabled feed
  list           - List your feeds
  help           - Show this help`
}

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

type Fetcher struct {
	client *http.Client
	agent  string
}

type FeedItem struct {
	GUID   string
	Title  string
	Link   string
	Posted time.Time
}

func NewFetcher(userAgent string) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		agent: userAgent,
	}
}

func (f *Fetcher) Fetch(url string) ([]FeedItem, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", f.agent)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	parser := gofeed.NewParser()
	feed, err := parser.Parse(resp.Body)
	if err != nil {
		return nil, "", err
	}

	var items []FeedItem
	for _, item := range feed.Items {
		g := item.GUID
		if g == "" {
			g = item.Link
		}
		if g == "" {
			g = item.Title
		}
		var posted time.Time
		if item.PublishedParsed != nil {
			posted = *item.PublishedParsed
		}
		items = append(items, FeedItem{
			GUID:   g,
			Title:  item.Title,
			Link:   item.Link,
			Posted: posted,
		})
	}

	title := feed.Title
	return items, title, nil
}

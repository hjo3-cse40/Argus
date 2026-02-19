package main

import (
	"argus-backend/internal/config"
	"argus-backend/internal/events"
	"argus-backend/internal/mq"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const seenIDsFile = "seen_ids.txt"

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title string `xml:"title"`
	Items []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Author      string `xml:"author"`
	PubDate     string `xml:"pubDate"`
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	mqClient, err := mq.Connect(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer mqClient.Close()

	if err := mqClient.DeclareQueue("raw_events"); err != nil {
		log.Fatalf("failed to declare queue: %v", err)
	}

	feeds := cfg.RSSHub.Feeds
	if len(feeds) == 0 {
		log.Fatal("no feeds configured — set RSSHUB_FEEDS (e.g. \"youtube:youtube/channel/ABC123,reddit:reddit/subreddit/golang\")")
	}
	log.Printf("Polling %d feed(s)", len(feeds))

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	seenIDs := loadSeenIDs()
	// Poll every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	for {
		processFeeds(client, feeds, mqClient, seenIDs)
		saveSeenIDs(seenIDs)

		select {
		case <-stop:
			log.Println("Shutting down RSS poller...")
			return
		case <-ticker.C:
		}
	}
}

func processFeeds(client *http.Client, feeds []config.Feed, mqClient *mq.Client, seenIDs map[string]map[string]bool) {
	for _, feed := range feeds {
		log.Printf("[%s] Fetching: %s", feed.SourceType, feed.URL)

		resp, err := client.Get(feed.URL)
		if err != nil {
			log.Printf("[%s] request failed: %v", feed.SourceType, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("[%s] bad status: %s", feed.SourceType, resp.Status)
			resp.Body.Close()
			continue
		}

		var rss RSS
		if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
			log.Printf("[%s] xml decode error: %v", feed.SourceType, err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		log.Printf("[%s] %s — %d items", feed.SourceType, rss.Channel.Title, len(rss.Channel.Items))

		if len(rss.Channel.Items) == 0 {
			continue
		}

		if seenIDs[feed.URL] == nil {
			seenIDs[feed.URL] = make(map[string]bool)
		}

		items := rss.Channel.Items
		if len(items) > 5 {
			items = items[:5]
		}

		for _, item := range items {
			eventID := generateDeterministicID(item.Link)

			if seenIDs[feed.URL][eventID] {
				continue
			}

			event := events.NewEvent(eventID, rss.Channel.Title, item.Title, item.Link)
			event.Metadata["source_type"] = feed.SourceType
			if item.Description != "" {
				event.Metadata["description"] = item.Description
			}
			if item.Author != "" {
				event.Metadata["author"] = item.Author
			}
			if item.PubDate != "" {
				event.Metadata["pub_date"] = item.PubDate
			}

			if err := event.Validate(); err != nil {
				log.Printf("[%s] invalid event: %v", feed.SourceType, err)
				continue
			}

			body, err := event.ToJSON()
			if err != nil {
				log.Printf("[%s] marshal failed: %v", feed.SourceType, err)
				continue
			}

			if err := mqClient.Publish("raw_events", body); err != nil {
				log.Printf("[%s] publish failed: %v", feed.SourceType, err)
				continue
			}

			log.Printf("✓ [%s] %s — %s", feed.SourceType, rss.Channel.Title, item.Title)
			seenIDs[feed.URL][eventID] = true
		}
	}
}

func generateDeterministicID(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// loadSeenIDs reads the persisted seen IDs from disk.
// File format: one "feedURL\teventID" per line.
func loadSeenIDs() map[string]map[string]bool {
	seen := make(map[string]map[string]bool)
	f, err := os.Open(seenIDsFile)
	if err != nil {
		return seen
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "\t", 2)
		if len(parts) != 2 {
			continue
		}
		feedURL, eventID := parts[0], parts[1]
		if seen[feedURL] == nil {
			seen[feedURL] = make(map[string]bool)
		}
		seen[feedURL][eventID] = true
	}
	log.Printf("Loaded %d feed(s) from %s", len(seen), seenIDsFile)
	return seen
}

func saveSeenIDs(seenIDs map[string]map[string]bool) {
	f, err := os.Create(seenIDsFile)
	if err != nil {
		log.Printf("failed to save seen IDs: %v", err)
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for feedURL, ids := range seenIDs {
		for id := range ids {
			fmt.Fprintf(w, "%s\t%s\n", feedURL, id)
		}
	}
	w.Flush()
}

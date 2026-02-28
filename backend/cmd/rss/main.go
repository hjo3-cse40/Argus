package main

import (
	"argus-backend/internal/config"
	"argus-backend/internal/events"
	"argus-backend/internal/mq"
	"argus-backend/internal/store"
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

	// Connect to database to load subsources
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)
	st, err := store.NewPostgresStore(connStr, 1000)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer st.Close()

	mqClient, err := mq.Connect(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer mqClient.Close()

	if err := mqClient.DeclareQueue("raw_events"); err != nil {
		log.Fatalf("failed to declare queue: %v", err)
	}

	// Load subsources from database
	subsources, err := loadSubsources(st)
	if err != nil {
		log.Fatalf("failed to load subsources: %v", err)
	}

	log.Printf("Polling %d subsource(s)", len(subsources))

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
		processFeeds(client, subsources, cfg.RSSHub.BaseURL, mqClient, seenIDs)
		saveSeenIDs(seenIDs)

		select {
		case <-stop:
			log.Println("Shutting down RSS poller...")
			return
		case <-ticker.C:
		}
	}
}

// loadSubsources queries the database for all active subsources with platform information
func loadSubsources(st store.Store) ([]store.SubsourceWithPlatform, error) {
	subsources := st.ListAllSubsources()
	if len(subsources) == 0 {
		return nil, fmt.Errorf("no subsources configured in database")
	}
	
	for _, s := range subsources {
		log.Printf("Loaded subsource: %s - %s (identifier: %s)", s.PlatformName, s.Name, s.Identifier)
	}
	
	return subsources, nil
}

// constructRSSHubURL constructs the full RSSHub URL from platform name and identifier
func constructRSSHubURL(baseURL, platformName, identifier string) string {
	var path string
	switch platformName {
	case "youtube":
		path = fmt.Sprintf("youtube/channel/%s", identifier)
	case "reddit":
		path = fmt.Sprintf("reddit/subreddit/%s", identifier)
	case "x":
		path = fmt.Sprintf("twitter/user/%s", identifier)
	default:
		return ""
	}
	return fmt.Sprintf("%s/%s", baseURL, path)
}

func processFeeds(client *http.Client, subsources []store.SubsourceWithPlatform, baseURL string, mqClient *mq.Client, seenIDs map[string]map[string]bool) {
	for _, subsource := range subsources {
		feedURL := constructRSSHubURL(baseURL, subsource.PlatformName, subsource.Identifier)
		if feedURL == "" {
			log.Printf("[%s] unsupported platform: %s", subsource.Name, subsource.PlatformName)
			continue
		}

		log.Printf("[%s - %s] Fetching: %s", subsource.PlatformName, subsource.Name, feedURL)

		resp, err := client.Get(feedURL)
		if err != nil {
			log.Printf("[%s - %s] request failed: %v", subsource.PlatformName, subsource.Name, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("[%s - %s] bad status: %s", subsource.PlatformName, subsource.Name, resp.Status)
			resp.Body.Close()
			continue
		}

		var rss RSS
		if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
			log.Printf("[%s - %s] xml decode error: %v", subsource.PlatformName, subsource.Name, err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		log.Printf("[%s - %s] %s — %d items", subsource.PlatformName, subsource.Name, rss.Channel.Title, len(rss.Channel.Items))

		if len(rss.Channel.Items) == 0 {
			continue
		}

		// Initialize seen IDs map for this subsource
		if seenIDs[subsource.ID] == nil {
			seenIDs[subsource.ID] = make(map[string]bool)
		}

		items := rss.Channel.Items
		if len(items) > 5 {
			items = items[:5]
		}

		for _, item := range items {
			eventID := generateDeterministicID(item.Link)

			// Check if we've seen this event for this subsource
			if seenIDs[subsource.ID][eventID] {
				continue
			}

			// Create event with hierarchical metadata
			event := events.NewEvent(eventID, subsource.Name, item.Title, item.Link)
			event.Metadata["subsource_id"] = subsource.ID
			event.Metadata["platform_name"] = subsource.PlatformName
			event.Metadata["subsource_name"] = subsource.Name
			event.Metadata["source_type"] = subsource.PlatformName
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
				log.Printf("[%s - %s] invalid event: %v", subsource.PlatformName, subsource.Name, err)
				continue
			}

			body, err := event.ToJSON()
			if err != nil {
				log.Printf("[%s - %s] marshal failed: %v", subsource.PlatformName, subsource.Name, err)
				continue
			}

			if err := mqClient.Publish("raw_events", body); err != nil {
				log.Printf("[%s - %s] publish failed: %v", subsource.PlatformName, subsource.Name, err)
				continue
			}

			log.Printf("✓ [%s - %s] %s — %s", subsource.PlatformName, subsource.Name, rss.Channel.Title, item.Title)
			seenIDs[subsource.ID][eventID] = true
		}
	}
}

func processFeeds_DEPRECATED(client *http.Client, feeds []config.Feed, mqClient *mq.Client, seenIDs map[string]map[string]bool) {
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
// File format: one "subsourceID\teventID" per line.
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
		subsourceID, eventID := parts[0], parts[1]
		if seen[subsourceID] == nil {
			seen[subsourceID] = make(map[string]bool)
		}
		seen[subsourceID][eventID] = true
	}
	log.Printf("Loaded %d subsource(s) from %s", len(seen), seenIDsFile)
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
	for subsourceID, ids := range seenIDs {
		for id := range ids {
			fmt.Fprintf(w, "%s\t%s\n", subsourceID, id)
		}
	}
	w.Flush()
}

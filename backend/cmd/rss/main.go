package main

import (
	"argus-backend/internal/config"
	"argus-backend/internal/events"
	"argus-backend/internal/mq"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"log"
	"net/http"
	"time"
)

// Minimal structs to match RSSHub's XML output
type RSS struct {
	Channel Channel `xml:"channel"`
}

// channel element
type Channel struct {
	Title string `xml:"title"`
	Items []Item `xml:"item"`
}

// video entries
type Item struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

func main() { //swap localhost with serverip:<port>
	urls := []string{"http://localhost:1200/youtube/channel/UCUyeluBRhGPCW4rPe_UvBZQ",
		"http://localhost:1200/youtube/channel/UC_S45UpAYVuc0fYEcHN9BVQ",
		//"http://localhost:1200/youtube/search/golang", //google api id req
	}

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

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	// In-memory deduplication (resets on restart)
	lastProcessedID := make(map[string]string)
	// Poll ever 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		processFeeds(client, urls, mqClient, lastProcessedID)
		<-ticker.C
	}
}

func processFeeds(client *http.Client, urls []string, mqClient *mq.Client, lastProcessedID map[string]string) {
	for _, url := range urls {

		log.Printf("Fetching feed: %s", url)
		// FETCHING
		// Make HTTP GET request to RSSHub endpoint
		resp, err := client.Get(url)
		if err != nil {
			log.Printf("request failed: %v", err)
			continue
		}

		// Ensure response body gets closed
		if resp.StatusCode != http.StatusOK {
			log.Printf("bad status: %s", resp.Status)
			resp.Body.Close()
			continue
		}

		// PARSING
		var rss RSS
		if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
			log.Printf("xml decode error: %v", err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// PROCESSING
		// Access parsed channel + items
		log.Printf("Channel: %s", rss.Channel.Title)

		if len(rss.Channel.Items) == 0 {
			continue
		}

		item := rss.Channel.Items[0]
		log.Printf("Processing: %s (%s)", item.Title, item.Link)

		eventID := generateDeterministicID(item.Link)

		if lastProcessedID[url] == eventID {
			log.Printf("Skipping duplicate event: %s", item.Title)
			continue
		}

		event := events.NewEvent(eventID, rss.Channel.Title, item.Title, item.Link)
		if err := event.Validate(); err != nil {
			log.Printf("invalid event: %v", err)
			continue
		}

		body, err := event.ToJSON()
		if err != nil {
			log.Printf("failed to marshal event: %v", err)
			continue
		}

		if err := mqClient.Publish("raw_events", body); err != nil {
			log.Printf("failed to publish event: %v", err)
			continue
		}

		log.Printf("✓ Published event: event_id=%s source=%s title=%s\n", eventID, rss.Channel.Title, item.Title)

		lastProcessedID[url] = eventID
	}
}

func generateDeterministicID(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

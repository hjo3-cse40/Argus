package main

import (
	"context"
	"argus-backend/internal/config"
	"argus-backend/internal/events"
	"argus-backend/internal/mq"
	"argus-backend/internal/store"
	"argus-backend/internal/youtube"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

const defaultSeenIDsFile = "seen_ids.txt"

// seenIDsPath returns RSS_SEEN_IDS_FILE when set (e.g. /data/seen_ids.txt in Docker), else cwd default.
func seenIDsPath() string {
	if p := strings.TrimSpace(os.Getenv("RSS_SEEN_IDS_FILE")); p != "" {
		return p
	}
	return defaultSeenIDsFile
}

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

// Atom 1.0 (reddit.com *.rss returns this, not RSS 2.0).

type atomFeed struct {
	Title   string       `xml:"http://www.w3.org/2005/Atom title"`
	Entries []atomEntry  `xml:"http://www.w3.org/2005/Atom entry"`
}

type atomEntry struct {
	Title     string       `xml:"http://www.w3.org/2005/Atom title"`
	Links     []atomLink   `xml:"http://www.w3.org/2005/Atom link"`
	Published string       `xml:"http://www.w3.org/2005/Atom published"`
	Updated   string       `xml:"http://www.w3.org/2005/Atom updated"`
	Author    atomAuthor   `xml:"http://www.w3.org/2005/Atom author"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

type atomAuthor struct {
	Name string `xml:"http://www.w3.org/2005/Atom name"`
}

func atomEntryPermalink(e atomEntry) string {
	var fallback string
	for _, l := range e.Links {
		if l.Href == "" {
			continue
		}
		if l.Rel == "" || l.Rel == "alternate" {
			return l.Href
		}
		fallback = l.Href
	}
	return fallback
}

func atomFeedToItems(af atomFeed) []Item {
	out := make([]Item, 0, len(af.Entries))
	for _, e := range af.Entries {
		link := atomEntryPermalink(e)
		if link == "" {
			continue
		}
		pub := e.Published
		if pub == "" {
			pub = e.Updated
		}
		out = append(out, Item{
			Title:   e.Title,
			Link:    link,
			PubDate: pub,
			Author:  strings.TrimSpace(e.Author.Name),
		})
	}
	return out
}

// decodeFeedXML parses RSS 2.0 or Atom 1.0 (e.g. Reddit subreddit feeds).
func decodeFeedXML(data []byte) (feedTitle string, items []Item, err error) {
	var rss RSS
	rssErr := xml.Unmarshal(data, &rss)

	var af atomFeed
	atomErr := xml.Unmarshal(data, &af)

	if len(rss.Channel.Items) > 0 {
		return rss.Channel.Title, rss.Channel.Items, nil
	}
	if atomErr == nil && len(af.Entries) > 0 {
		return af.Title, atomFeedToItems(af), nil
	}
	if atomErr == nil {
		return af.Title, nil, nil
	}
	if rssErr == nil {
		return rss.Channel.Title, rss.Channel.Items, nil
	}
	return "", nil, fmt.Errorf("rss: %v; atom: %v", rssErr, atomErr)
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
	defer func() { _ = st.Close() }()

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
	seenIDs := loadSeenIDs(seenIDsPath())
	// Poll every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	for {
		subsources := loadSubsources(st)
		if len(subsources) == 0 {
			log.Printf("no subsources configured; idle until next check (every 5m)")
		} else {
			log.Printf("polling %d subsource(s)", len(subsources))
			processFeeds(client, subsources, cfg.RSSHub.BaseURL, mqClient, seenIDs, st)
		}
		saveSeenIDs(seenIDsPath(), seenIDs)

		select {
		case <-stop:
			log.Println("Shutting down RSS poller...")
			return
		case <-ticker.C:
		}
	}
}

// loadSubsources queries the database for all active subsources with platform information.
// An empty database is valid: the poller stays up and reloads on each tick.
func loadSubsources(st store.Store) []store.SubsourceWithPlatform {
	return st.ListAllSubsources()
}

// youTubeChannelIDRe matches a full YouTube channel id (e.g. UC…). Handles/usernames must use /youtube/user/.
var youTubeChannelIDRe = regexp.MustCompile(`^UC[-_0-9A-Za-z]{22}$`)

// normalizeRedditSubreddit strips optional r/ prefix and whitespace from stored identifiers.
func normalizeRedditSubreddit(identifier string) string {
	s := strings.TrimSpace(identifier)
	s = strings.TrimPrefix(s, "/")
	if len(s) >= 2 && strings.EqualFold(s[:2], "r/") {
		s = strings.TrimSpace(s[2:])
	}
	return strings.TrimSpace(s)
}

// feedURLForSubsource returns the RSS URL to fetch. Reddit uses reddit.com public feeds so we
// do not depend on RSSHub's reddit routes (often flaky or returning NotFound for valid subs).
// YouTube: channel ids (UC…) use Google's official Atom feed (no RSSHub). Handles / legacy
// names go through RSSHub as /youtube/user/@handle (RSSHub requires @ for channel handles).
func feedURLForSubsource(baseURL, platformName, identifier string) string {
	switch platformName {
	case "reddit":
		sub := normalizeRedditSubreddit(identifier)
		if sub == "" {
			return ""
		}
		return fmt.Sprintf("https://www.reddit.com/r/%s/.rss", url.PathEscape(sub))
	case "youtube":
		id := strings.TrimSpace(identifier)
		if id == "" {
			return ""
		}
		if youTubeChannelIDRe.MatchString(id) {
			return fmt.Sprintf(
				"https://www.youtube.com/feeds/videos.xml?channel_id=%s",
				url.QueryEscape(id),
			)
		}
		handle := strings.TrimPrefix(id, "@")
		if handle == "" {
			return ""
		}
		base := strings.TrimRight(baseURL, "/")
		return fmt.Sprintf("%s/youtube/user/@%s", base, handle)
	default:
		return constructRSSHubURL(baseURL, platformName, identifier)
	}
}

// constructRSSHubURL constructs the full RSSHub URL from platform name and identifier
func constructRSSHubURL(baseURL, platformName, identifier string) string {
	var path string
	switch platformName {
	case "youtube":
		id := strings.TrimSpace(identifier)
		if youTubeChannelIDRe.MatchString(id) {
			path = fmt.Sprintf("youtube/channel/%s", id)
		} else {
			// Handles must use @… for RSSHub; see feedURLForSubsource for direct youtube.com feeds.
			handle := strings.TrimPrefix(id, "@")
			if handle == "" {
				return ""
			}
			path = fmt.Sprintf("youtube/user/@%s", handle)
		}
	case "x":
		path = fmt.Sprintf("twitter/user/%s", identifier)
	default:
		return ""
	}
	return fmt.Sprintf("%s/%s", baseURL, path)
}

// fetchFeed performs GET with headers suited to the feed host (Reddit often rejects default Go user agents).
func fetchFeed(client *http.Client, feedURL string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	switch {
	case strings.Contains(feedURL, "reddit.com/r/"),
		strings.Contains(feedURL, "youtube.com/feeds/videos.xml"):
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ArgusRSS/1.0)")
	}
	return client.Do(req)
}

func processFeeds(client *http.Client, subsources []store.SubsourceWithPlatform, baseURL string, mqClient *mq.Client, seenIDs map[string]map[string]bool, st store.Store) {
	for _, subsource := range subsources {
		feedURL := feedURLForSubsource(baseURL, subsource.PlatformName, subsource.Identifier)
		if feedURL == "" {
			log.Printf("[%s] unsupported platform or empty identifier: %s", subsource.Name, subsource.PlatformName)
			continue
		}

		if subsource.PlatformName == "youtube" {
			yid := strings.TrimSpace(subsource.Identifier)
			if yid != "" && !youTubeChannelIDRe.MatchString(yid) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				uc, err := youtube.ResolveChannelID(ctx, yid)
				cancel()
				if err == nil && uc != "" {
					feedURL = youtube.FeedURL(uc)
					log.Printf("[%s - %s] resolved YouTube channel id %s (official feed)", subsource.PlatformName, subsource.Name, uc)
				}
			}
		}

		log.Printf("[%s - %s] Fetching: %s", subsource.PlatformName, subsource.Name, feedURL)

		resp, err := fetchFeed(client, feedURL)
		if err != nil {
			log.Printf("[%s - %s] request failed: %v", subsource.PlatformName, subsource.Name, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("[%s - %s] bad status: %s", subsource.PlatformName, subsource.Name, resp.Status)
			_ = resp.Body.Close()
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			log.Printf("[%s - %s] read body: %v", subsource.PlatformName, subsource.Name, err)
			continue
		}

		feedTitle, feedItems, err := decodeFeedXML(body)
		if err != nil {
			log.Printf("[%s - %s] xml decode error: %v", subsource.PlatformName, subsource.Name, err)
			continue
		}

		log.Printf("[%s - %s] %s — %d items", subsource.PlatformName, subsource.Name, feedTitle, len(feedItems))

		if len(feedItems) == 0 {
			continue
		}

		// Initialize seen IDs map for this subsource
		if seenIDs[subsource.ID] == nil {
			seenIDs[subsource.ID] = make(map[string]bool)
		}

		items := feedItems
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

			// Same contract as POST /api/ingest: queued row must exist before the worker can
			// mark the delivery delivered for the dashboard/API.
			sid := subsource.ID
			st.AddQueued(store.Delivery{
				EventID:     event.EventID,
				Source:      subsource.PlatformName,
				Title:       event.Title,
				URL:         event.URL,
				SubsourceID: &sid,
				UserID:      subsource.UserID,
			})

			log.Printf("✓ [%s - %s] %s — %s", subsource.PlatformName, subsource.Name, feedTitle, item.Title)
			seenIDs[subsource.ID][eventID] = true
		}
	}
}

func generateDeterministicID(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// loadSeenIDs reads the persisted seen IDs from disk.
// File format: one "subsourceID\teventID" per line.
func loadSeenIDs(path string) map[string]map[string]bool {
	seen := make(map[string]map[string]bool)
	f, err := os.Open(path)
	if err != nil {
		return seen
	}
	defer func() { _ = f.Close() }()

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
	log.Printf("Loaded %d subsource(s) from %s", len(seen), path)
	return seen
}

func saveSeenIDs(path string, seenIDs map[string]map[string]bool) {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Printf("failed to create seen-ID dir: %v", err)
			return
		}
	}
	f, err := os.Create(path)
	if err != nil {
		log.Printf("failed to save seen IDs: %v", err)
		return
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)
	for subsourceID, ids := range seenIDs {
		for id := range ids {
			_, _ = fmt.Fprintf(w, "%s\t%s\n", subsourceID, id)
		}
	}
	_ = w.Flush()
}

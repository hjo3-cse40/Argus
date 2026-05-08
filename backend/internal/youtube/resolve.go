package youtube

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// UC channel ids are UC + exactly 22 chars from YouTube's alphabet (includes '-' '_').
var ucChannelIDRe = regexp.MustCompile(`^UC[-_0-9A-Za-z]{22}$`)

// Embedded in ytInitialData / response HTML.
var channelIDPatterns = []*regexp.Regexp{
	regexp.MustCompile(`"channelId":"(UC[-_0-9A-Za-z]{22})"`),
	regexp.MustCompile(`"externalId":"(UC[-_0-9A-Za-z]{22})"`),
	regexp.MustCompile(`href="https://www\.youtube\.com/channel/(UC[-_0-9A-Za-z]{22})"`),
	regexp.MustCompile(`rel="canonical" href="https://www\.youtube\.com/channel/(UC[-_0-9A-Za-z]{22})"`),
	regexp.MustCompile(`channel_id=(UC[-_0-9A-Za-z]{22})`),
	regexp.MustCompile(`\/channel\/(UC[-_0-9A-Za-z]{22})`),
}

// ErrNotResolved is returned when no UC… id could be extracted from tried pages.
var ErrNotResolved = errors.New("youtube: channel id not found")

var resolveCache sync.Map // string -> string (normalized key -> UC id)

const maxHTMLBody = 4 << 20 // 4 MiB

var defaultClient = &http.Client{
	Timeout: 18 * time.Second,
}

// FeedURL returns https://www.youtube.com/feeds/videos.xml?channel_id=… for the given UC id.
func FeedURL(channelID string) string {
	return "https://www.youtube.com/feeds/videos.xml?channel_id=" + url.QueryEscape(channelID)
}

// ChannelIDFromHTML extracts the first UC… channel id found in a YouTube HTML/JSON response.
func ChannelIDFromHTML(html []byte) string {
	s := string(html)
	for _, re := range channelIDPatterns {
		m := re.FindStringSubmatch(s)
		if len(m) >= 2 && ucChannelIDRe.MatchString(m[1]) {
			return m[1]
		}
	}
	return ""
}

func normalizeKey(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "@")
	return strings.ToLower(s)
}

func candidateURLs(handle string) []string {
	h := url.PathEscape(handle)
	return []string{
		"https://www.youtube.com/@" + h,
		"https://www.youtube.com/c/" + h,
		"https://www.youtube.com/user/" + h,
	}
}

// ResolveChannelID maps a handle, legacy username, or custom-path slug to a channel UC… id
// by fetching public channel pages (no API key). If raw is already a UC… id, it is returned.
// Results are cached in-process for the lifetime of the process.
func ResolveChannelID(ctx context.Context, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrNotResolved
	}
	if ucChannelIDRe.MatchString(raw) {
		return raw, nil
	}

	key := normalizeKey(raw)
	if key == "" {
		return "", ErrNotResolved
	}
	if v, ok := resolveCache.Load(key); ok {
		return v.(string), nil
	}

	handle := strings.TrimPrefix(raw, "@")
	if handle == "" {
		return "", ErrNotResolved
	}

	var lastErr error
	for _, pageURL := range candidateURLs(handle) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		resp, err := defaultClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxHTMLBody))
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if resp.StatusCode != http.StatusOK {
			lastErr = errors.New("youtube: non-200 status")
			continue
		}

		if id := ChannelIDFromHTML(body); id != "" {
			resolveCache.Store(key, id)
			return id, nil
		}
	}

	if lastErr != nil {
		return "", lastErr
	}
	return "", ErrNotResolved
}

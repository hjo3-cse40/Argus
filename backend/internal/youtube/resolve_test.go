package youtube

import (
	"strings"
	"testing"
)

func TestChannelIDFromHTML(t *testing.T) {
	const embedded = `"responseContext":{},"contents":{},"channelId":"UCBaBbCcDdEeFfGgHhIiJjKk"`

	if got := ChannelIDFromHTML([]byte(embedded)); got != "UCBaBbCcDdEeFfGgHhIiJjKk" {
		t.Fatalf("got %q", got)
	}
}

func TestChannelIDFromHTML_Canonical(t *testing.T) {
	html := `<head><link rel="canonical" href="https://www.youtube.com/channel/UCxxxxxxxxxxxxxxxxxxxxxx"/>`
	if got := ChannelIDFromHTML([]byte(html)); got != "UCxxxxxxxxxxxxxxxxxxxxxx" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveChannelID_AlreadyUC(t *testing.T) {
	id := "UC" + strings.Repeat("a", 22)
	got, err := ResolveChannelID(t.Context(), id)
	if err != nil || got != id {
		t.Fatalf("got %q err %v", got, err)
	}
}

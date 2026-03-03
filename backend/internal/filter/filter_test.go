package filter

import (
	"testing"

	"argus-backend/internal/events"
	"argus-backend/internal/store"
)

func makeEvent(title, description string) *events.Event {
	ev := events.NewEvent("test-id", "test-source", title, "https://example.com")
	if description != "" {
		ev.Metadata["description"] = description
	}
	return ev
}

func TestEvaluate_NoFilters_PassesAll(t *testing.T) {
	ev := makeEvent("NBA Highlights", "Best plays of the week")
	if !Evaluate(ev, nil) {
		t.Error("expected event to pass with nil filters")
	}
	if !Evaluate(ev, []store.DestinationFilter{}) {
		t.Error("expected event to pass with empty filters")
	}
}

func TestEvaluate_KeywordExclude_BlocksMatch(t *testing.T) {
	ev := makeEvent("NBA Shorts Compilation", "Watch short clips")
	filters := []store.DestinationFilter{
		{FilterType: "keyword_exclude", Pattern: "shorts"},
	}
	if Evaluate(ev, filters) {
		t.Error("expected event to be blocked by exclude filter")
	}
}

func TestEvaluate_KeywordExclude_PassesNonMatch(t *testing.T) {
	ev := makeEvent("NBA Full Game Highlights", "Full game recap")
	filters := []store.DestinationFilter{
		{FilterType: "keyword_exclude", Pattern: "shorts"},
	}
	if !Evaluate(ev, filters) {
		t.Error("expected event to pass when exclude pattern doesn't match")
	}
}

func TestEvaluate_KeywordExclude_CaseInsensitive(t *testing.T) {
	ev := makeEvent("NBA SHORTS Compilation", "")
	filters := []store.DestinationFilter{
		{FilterType: "keyword_exclude", Pattern: "shorts"},
	}
	if Evaluate(ev, filters) {
		t.Error("expected case-insensitive exclude match")
	}
}

func TestEvaluate_KeywordInclude_PassesMatch(t *testing.T) {
	ev := makeEvent("NBA Playoffs Game 7", "Intense playoff action")
	filters := []store.DestinationFilter{
		{FilterType: "keyword_include", Pattern: "playoffs"},
	}
	if !Evaluate(ev, filters) {
		t.Error("expected event to pass with matching include filter")
	}
}

func TestEvaluate_KeywordInclude_BlocksNonMatch(t *testing.T) {
	ev := makeEvent("NBA Regular Season Recap", "Weekly recap")
	filters := []store.DestinationFilter{
		{FilterType: "keyword_include", Pattern: "playoffs"},
	}
	if Evaluate(ev, filters) {
		t.Error("expected event to be blocked when no include pattern matches")
	}
}

func TestEvaluate_MultipleIncludes_ORLogic(t *testing.T) {
	ev := makeEvent("NBA Finals Highlights", "Championship game")
	filters := []store.DestinationFilter{
		{FilterType: "keyword_include", Pattern: "playoffs"},
		{FilterType: "keyword_include", Pattern: "finals"},
	}
	if !Evaluate(ev, filters) {
		t.Error("expected event to pass when at least one include matches")
	}
}

func TestEvaluate_ExcludeTakesPriority(t *testing.T) {
	ev := makeEvent("NBA Playoffs Shorts", "Short clips from playoffs")
	filters := []store.DestinationFilter{
		{FilterType: "keyword_include", Pattern: "playoffs"},
		{FilterType: "keyword_exclude", Pattern: "shorts"},
	}
	if Evaluate(ev, filters) {
		t.Error("expected exclude to take priority over include")
	}
}

func TestEvaluate_MatchesDescription(t *testing.T) {
	ev := makeEvent("Game Highlights", "Watch the best dunks from the playoffs")
	filters := []store.DestinationFilter{
		{FilterType: "keyword_include", Pattern: "playoffs"},
	}
	if !Evaluate(ev, filters) {
		t.Error("expected include to match against description")
	}
}

func TestEvaluate_ExcludeMatchesDescription(t *testing.T) {
	ev := makeEvent("Game Highlights", "Quick shorts compilation of the week")
	filters := []store.DestinationFilter{
		{FilterType: "keyword_exclude", Pattern: "shorts"},
	}
	if Evaluate(ev, filters) {
		t.Error("expected exclude to match against description")
	}
}

func TestEvaluate_OnlyExcludes_NoInclude_PassesNonMatch(t *testing.T) {
	ev := makeEvent("Full Game Replay", "Complete game footage")
	filters := []store.DestinationFilter{
		{FilterType: "keyword_exclude", Pattern: "shorts"},
		{FilterType: "keyword_exclude", Pattern: "ad"},
	}
	if !Evaluate(ev, filters) {
		t.Error("expected event to pass when only excludes exist and none match")
	}
}

func TestEvaluate_EmptyDescription(t *testing.T) {
	ev := makeEvent("NBA Playoffs Recap", "")
	filters := []store.DestinationFilter{
		{FilterType: "keyword_include", Pattern: "playoffs"},
	}
	if !Evaluate(ev, filters) {
		t.Error("expected include to match title even without description")
	}
}

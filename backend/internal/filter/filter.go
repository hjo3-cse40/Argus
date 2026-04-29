package filter

import (
	"strings"

	"argus-backend/internal/events"
	"argus-backend/internal/store"
)

// Evaluate checks whether an event should be delivered given a set of destination filters.
//
// Rules:
//   - No filters: event passes (backwards compatible)
//   - keyword_exclude: event is dropped if title or description matches ANY exclude pattern
//   - keyword_include: event must match at least ONE include pattern to pass
//   - Excludes are evaluated first and take priority over includes
func Evaluate(ev *events.Event, filters []store.DestinationFilter) bool {
	pass, _ := EvaluateWithReason(ev, filters)
	return pass
}

// EvaluateWithReason returns whether an event passes filters and, when blocked,
// includes a human-readable reason.
func EvaluateWithReason(ev *events.Event, filters []store.DestinationFilter) (bool, string) {
	if len(filters) == 0 {
		return true, ""
	}

	text := buildSearchText(ev)

	var hasIncludes bool

	for _, f := range filters {
		pattern := strings.ToLower(f.Pattern)

		switch f.FilterType {
		case "keyword_exclude":
			if strings.Contains(text, pattern) {
				return false, "blocked because exclude keyword: " + f.Pattern
			}
		case "keyword_include":
			hasIncludes = true
		}
	}

	if !hasIncludes {
		return true, ""
	}

	for _, f := range filters {
		if f.FilterType == "keyword_include" {
			if strings.Contains(text, strings.ToLower(f.Pattern)) {
				return true, ""
			}
		}
	}

	return false, "blocked because no include keyword matched"
}

func buildSearchText(ev *events.Event) string {
	parts := []string{strings.ToLower(ev.Title)}

	if desc, ok := ev.Metadata["description"].(string); ok && desc != "" {
		parts = append(parts, strings.ToLower(desc))
	}

	return strings.Join(parts, " ")
}

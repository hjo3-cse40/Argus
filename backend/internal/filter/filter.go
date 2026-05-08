package filter

import (
	"strings"

	"argus-backend/internal/events"
	"argus-backend/internal/store"
)

// FilterCombineOpts configures how multiple keyword_include / keyword_exclude rows combine.
// IncludeCombine / ExcludeCombine use "any" or "all" (case-insensitive); empty defaults to "any".
//
// "any" — include: at least one pattern matches; exclude: block if any pattern matches.
// "all" — include: every pattern must match; exclude: block only if every pattern matches.
type FilterCombineOpts struct {
	IncludeCombine string
	ExcludeCombine string
}

// NormalizeCombine returns "any" or "all". Unknown values default to "any".
func NormalizeCombine(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "all":
		return "all"
	default:
		return "any"
	}
}

// Evaluate checks whether an event should be delivered given destination filters.
// Combine modes default to "any" (legacy OR behavior).
//
// Rules:
//   - No filters: event passes (backwards compatible)
//   - keyword_exclude: depends on ExcludeCombine (see FilterCombineOpts)
//   - keyword_include: depends on IncludeCombine
//   - Excludes are evaluated first and take priority over includes
func Evaluate(ev *events.Event, filters []store.DestinationFilter) bool {
	pass, _ := EvaluateWithReason(ev, filters, FilterCombineOpts{})
	return pass
}

// EvaluateWithOpts is like Evaluate but uses combine modes from the destination platform.
func EvaluateWithOpts(ev *events.Event, filters []store.DestinationFilter, opts FilterCombineOpts) bool {
	pass, _ := EvaluateWithReason(ev, filters, opts)
	return pass
}

// EvaluateWithReason returns whether an event passes filters and, when blocked,
// includes a human-readable reason.
func EvaluateWithReason(ev *events.Event, filters []store.DestinationFilter, opts FilterCombineOpts) (bool, string) {
	if len(filters) == 0 {
		return true, ""
	}

	text := buildSearchText(ev)
	incMode := NormalizeCombine(opts.IncludeCombine)
	excMode := NormalizeCombine(opts.ExcludeCombine)

	var includes, excludes []string
	for _, f := range filters {
		switch f.FilterType {
		case "keyword_exclude":
			excludes = append(excludes, f.Pattern)
		case "keyword_include":
			includes = append(includes, f.Pattern)
		}
	}

	if len(excludes) > 0 {
		if excMode == "all" {
			allExcludeMatch := true
			for _, pat := range excludes {
				if !strings.Contains(text, strings.ToLower(pat)) {
					allExcludeMatch = false
					break
				}
			}
			if allExcludeMatch {
				return false, "blocked because all exclude keywords matched"
			}
		} else {
			for _, pat := range excludes {
				if strings.Contains(text, strings.ToLower(pat)) {
					return false, "blocked because exclude keyword: " + pat
				}
			}
		}
	}

	if len(includes) == 0 {
		return true, ""
	}

	if incMode == "all" {
		for _, pat := range includes {
			if !strings.Contains(text, strings.ToLower(pat)) {
				return false, "blocked because not all include keywords matched"
			}
		}
		return true, ""
	}

	for _, pat := range includes {
		if strings.Contains(text, strings.ToLower(pat)) {
			return true, ""
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

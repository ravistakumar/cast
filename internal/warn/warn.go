// Package warn defines structured compile warnings and their report rendering.
package warn

import (
	"fmt"
	"sort"
	"strings"
)

// Severity classifies a warning.
type Severity string

const (
	Info Severity = "info"
	Warn Severity = "warn"
)

// Warning is a structured, machine-coded compile diagnostic.
type Warning struct {
	Harness    string
	Severity   Severity
	Location   string
	Code       string
	Message    string
	Suggestion string
}

// Render groups warnings by harness into a stable, human-readable report.
// Empty input renders an empty string.
func Render(ws []Warning) string {
	if len(ws) == 0 {
		return ""
	}
	byHarness := map[string][]Warning{}
	for _, w := range ws {
		byHarness[w.Harness] = append(byHarness[w.Harness], w)
	}
	harnesses := make([]string, 0, len(byHarness))
	for h := range byHarness {
		harnesses = append(harnesses, h)
	}
	sort.Strings(harnesses)

	var b strings.Builder
	for _, h := range harnesses {
		fmt.Fprintf(&b, "%s:\n", h)
		group := byHarness[h]
		sort.SliceStable(group, func(i, j int) bool { return group[i].Code < group[j].Code })
		for _, w := range group {
			fmt.Fprintf(&b, "  [%s] %s (%s): %s\n", w.Severity, w.Code, w.Location, w.Message)
			if w.Suggestion != "" {
				fmt.Fprintf(&b, "        ↳ %s\n", w.Suggestion)
			}
		}
	}
	return b.String()
}

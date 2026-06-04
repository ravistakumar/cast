// Package optimize optionally adapts a skill body per harness via an agent CLI.
package optimize

import (
	"fmt"
	"strings"

	"github.com/ravistakumar/cast/internal/agent"
	"github.com/ravistakumar/cast/internal/profile"
	"github.com/ravistakumar/cast/internal/skill"
	"github.com/ravistakumar/cast/internal/warn"
)

// Optimize asks the agent to rewrite the body for the target harness, resolving
// the supplied warnings. On any failure it returns the original body plus an
// optimize-failed warning — it never blocks.
func Optimize(s *skill.Skill, p profile.Profile, ws []warn.Warning, r agent.Runner) (string, []warn.Warning) {
	prompt := buildPrompt(s, p, ws)
	out, err := r.Run(prompt)
	if err != nil || strings.TrimSpace(out) == "" {
		return s.Body, []warn.Warning{{
			Harness: p.Name(), Severity: warn.Warn, Location: "body",
			Code: "optimize-failed", Message: fmt.Sprintf("optimization pass failed; using deterministic output: %v", err),
		}}
	}
	return out, nil
}

func buildPrompt(s *skill.Skill, p profile.Profile, ws []warn.Warning) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Rewrite the following Agent Skill body so it is idiomatic for the %s coding agent.\n", p.Name())
	b.WriteString("Resolve these portability issues; output ONLY the rewritten markdown body, no commentary:\n")
	for _, w := range ws {
		fmt.Fprintf(&b, "- %s\n", w.Message)
	}
	b.WriteString("\n--- BODY ---\n")
	b.WriteString(s.Body)
	return b.String()
}

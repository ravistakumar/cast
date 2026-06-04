// Package emit walks a Skill IR through a Profile into output files + warnings.
package emit

import (
	"regexp"

	"github.com/ravistakumar/cast/internal/profile"
	"github.com/ravistakumar/cast/internal/skill"
	"github.com/ravistakumar/cast/internal/warn"
)

// canonicalTools is the set of Claude tool names scanned for in the body.
var canonicalTools = []string{
	"Read", "Write", "Edit", "Bash", "Glob", "Grep",
	"Task", "WebFetch", "WebSearch", "NotebookEdit", "TodoWrite",
}

// Emit produces a profile's output files for a skill plus structured warnings.
// The body is copied verbatim; only deterministic detection happens here.
func Emit(s *skill.Skill, p profile.Profile) ([]profile.File, []warn.Warning) {
	dir := p.SkillDir(s.Name)
	var files []profile.File

	// SKILL.md = harness frontmatter + canonical body.
	content := "---\n" + p.Frontmatter(s) + "---\n\n" + s.Body
	files = append(files, profile.File{RelPath: dir + "/SKILL.md", Bytes: []byte(content)})

	// Copy assets verbatim under the skill dir.
	for _, a := range s.Assets {
		files = append(files, profile.File{RelPath: dir + "/" + a.RelPath, Bytes: a.Bytes})
	}

	// Harness enhancement files (placed under the skill dir).
	for _, f := range p.Enhancements(s) {
		files = append(files, profile.File{RelPath: dir + "/" + f.RelPath, Bytes: f.Bytes})
	}

	warns := scanTools(s, p)
	return files, warns
}

func scanTools(s *skill.Skill, p profile.Profile) []warn.Warning {
	tm := p.ToolMap()
	var warns []warn.Warning
	for _, tool := range canonicalTools {
		if !mentions(s.Body, tool) {
			continue
		}
		repl, known := tm[tool]
		if known && repl == "" {
			warns = append(warns, warn.Warning{
				Harness: p.Name(), Severity: warn.Warn, Location: "body",
				Code: "tool-no-equivalent",
				Message: "body references the " + tool + " tool, which has no " + p.Name() + " equivalent",
				Suggestion: "run with --optimize to let the agent CLI adapt this phrasing",
			})
		}
	}
	return warns
}

func mentions(body, tool string) bool {
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(tool) + `\b`)
	return re.MatchString(body)
}

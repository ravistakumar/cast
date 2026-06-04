package profile

import (
	"fmt"
	"strings"

	"github.com/ravistakumar/cast/internal/skill"
)

// geminiInstallSubdir is the per-skill output subdirectory. Gemini CLI loads
// skills from <root>/skills/<name>/SKILL.md (geminicli.com/docs/cli/skills).
const geminiInstallSubdir = "skills"

// Gemini emits a skill for the Gemini CLI.
type Gemini struct{}

func (Gemini) Name() string { return "gemini" }

func (Gemini) SkillDir(name string) string { return geminiInstallSubdir + "/" + name }

// ToolMap maps canonical Claude tool names to Gemini CLI equivalents. Tools with
// uncertain Gemini support (Task, NotebookEdit) are omitted so emit stays quiet
// about them rather than warning incorrectly.
func (Gemini) ToolMap() map[string]string {
	return map[string]string{
		"Read": "read_file", "Write": "write_file", "Edit": "replace",
		"Bash": "run_shell_command", "Glob": "glob", "Grep": "grep_search",
		"WebFetch": "web_fetch", "WebSearch": "google_web_search",
		"TodoWrite": "write_todos",
	}
}

func (Gemini) Frontmatter(s *skill.Skill) string {
	var b strings.Builder
	fmt.Fprintf(&b, "name: %s\n", s.Name)
	fmt.Fprintf(&b, "description: %s\n", s.Description)
	return b.String()
}

// Enhancements: Gemini CLI needs no extra files.
func (Gemini) Enhancements(s *skill.Skill) []File { return nil }

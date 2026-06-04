package profile

import (
	"fmt"
	"strings"

	"github.com/ravistakumar/cast/internal/skill"
)

// cursorInstallSubdir is the per-skill output subdirectory. Cursor loads skills
// from <root>/skills/<name>/SKILL.md (cursor.com/docs/context/skills).
const cursorInstallSubdir = "skills"

// Cursor emits a skill for the Cursor agent.
type Cursor struct{}

func (Cursor) Name() string { return "cursor" }

func (Cursor) SkillDir(name string) string { return cursorInstallSubdir + "/" + name }

func (Cursor) ToolMap() map[string]string {
	return map[string]string{
		"Read": "read_file", "Write": "edit_file", "Edit": "edit_file",
		"Bash": "run_terminal_cmd", "Glob": "", "Grep": "grep",
		"Task": "", "WebFetch": "", "WebSearch": "web", "NotebookEdit": "", "TodoWrite": "",
	}
}

func (Cursor) Frontmatter(s *skill.Skill) string {
	var b strings.Builder
	fmt.Fprintf(&b, "name: %s\n", s.Name)
	fmt.Fprintf(&b, "description: %s\n", s.Description)
	return b.String()
}

// Enhancements: Cursor needs no extra files in v1.
func (Cursor) Enhancements(s *skill.Skill) []File { return nil }

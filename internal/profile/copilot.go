package profile

import (
	"fmt"
	"strings"

	"github.com/ravistakumar/cast/internal/skill"
)

// copilotInstallSubdir is the per-skill output subdirectory. Copilot loads
// skills from <root>/skills/<name>/SKILL.md
// (code.visualstudio.com/docs/copilot/customization/agent-skills).
const copilotInstallSubdir = "skills"

// Copilot emits a skill for GitHub Copilot (VS Code).
type Copilot struct{}

func (Copilot) Name() string { return "copilot" }

func (Copilot) SkillDir(name string) string { return copilotInstallSubdir + "/" + name }

// ToolMap maps canonical Claude tool names to Copilot equivalents. Tools with
// uncertain Copilot support (WebSearch, Task, TodoWrite) are omitted so emit
// stays quiet about them rather than warning incorrectly.
func (Copilot) ToolMap() map[string]string {
	return map[string]string{
		"Read": "read_file", "Write": "create_file", "Edit": "replace_string_in_file",
		"Bash": "run_in_terminal", "Glob": "list_dir", "Grep": "grep_search",
		"WebFetch": "fetch_webpage", "NotebookEdit": "edit_notebook_file",
	}
}

func (Copilot) Frontmatter(s *skill.Skill) string {
	var b strings.Builder
	fmt.Fprintf(&b, "name: %s\n", s.Name)
	fmt.Fprintf(&b, "description: %s\n", s.Description)
	return b.String()
}

// Enhancements: Copilot needs no extra files.
func (Copilot) Enhancements(s *skill.Skill) []File { return nil }

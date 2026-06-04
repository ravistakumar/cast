package profile

import (
	"fmt"
	"strings"

	"github.com/ravistakumar/cast/internal/skill"
)

// opencodeInstallSubdir is the per-skill output subdirectory. OpenCode loads
// skills from <root>/skills/<name>/SKILL.md (opencode.ai/docs/skills).
const opencodeInstallSubdir = "skills"

// OpenCode emits a skill for the OpenCode agent.
type OpenCode struct{}

func (OpenCode) Name() string { return "opencode" }

func (OpenCode) SkillDir(name string) string { return opencodeInstallSubdir + "/" + name }

// ToolMap maps canonical Claude tool names to OpenCode equivalents. OpenCode has
// no sub-agent delegation or notebook tool, so Task and NotebookEdit map to ""
// (emit warns they have no equivalent).
func (OpenCode) ToolMap() map[string]string {
	return map[string]string{
		"Read": "read", "Write": "write", "Edit": "edit", "Bash": "bash",
		"Glob": "glob", "Grep": "grep", "WebFetch": "webfetch",
		"WebSearch": "websearch", "TodoWrite": "todowrite",
		"Task": "", "NotebookEdit": "",
	}
}

func (OpenCode) Frontmatter(s *skill.Skill) string {
	var b strings.Builder
	fmt.Fprintf(&b, "name: %s\n", s.Name)
	fmt.Fprintf(&b, "description: %s\n", s.Description)
	return b.String()
}

// Enhancements: OpenCode needs no extra files.
func (OpenCode) Enhancements(s *skill.Skill) []File { return nil }

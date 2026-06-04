package profile

import (
	"fmt"
	"strings"

	"github.com/ravistakumar/cast/internal/skill"
)

// codexInstallSubdir is the per-skill output subdirectory. Codex loads skills
// from <root>/skills/<name>/SKILL.md (developers.openai.com/codex/skills).
const codexInstallSubdir = "skills"

// Codex emits a skill for the OpenAI Codex CLI.
type Codex struct{}

func (Codex) Name() string { return "codex" }

func (Codex) SkillDir(name string) string { return codexInstallSubdir + "/" + name }

// ToolMap maps canonical Claude tool names to Codex equivalents.
// "" means no known equivalent (emit warns).
func (Codex) ToolMap() map[string]string {
	return map[string]string{
		"Read": "read_file", "Write": "write_file", "Edit": "apply_patch",
		"Bash": "shell", "Glob": "", "Grep": "", "Task": "",
		"WebFetch": "", "WebSearch": "", "NotebookEdit": "", "TodoWrite": "",
	}
}

func (Codex) Frontmatter(s *skill.Skill) string {
	var b strings.Builder
	fmt.Fprintf(&b, "name: %s\n", s.Name)
	fmt.Fprintf(&b, "description: %s\n", s.Description)
	if s.License != "" {
		fmt.Fprintf(&b, "license: %s\n", s.License)
	}
	return b.String()
}

// Enhancements emits Codex's agents/openai.yaml extension file, using Codex's
// documented schema (an `interface` block). See
// developers.openai.com/codex/skills.
func (Codex) Enhancements(s *skill.Skill) []File {
	var b strings.Builder
	b.WriteString("interface:\n")
	fmt.Fprintf(&b, "  display_name: %s\n", s.Name)
	fmt.Fprintf(&b, "  short_description: %s\n", s.Description)
	return []File{{RelPath: "agents/openai.yaml", Bytes: []byte(b.String())}}
}

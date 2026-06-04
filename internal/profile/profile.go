// Package profile describes how to emit a Skill for one target harness.
package profile

import (
	"sort"

	"github.com/ravistakumar/cast/internal/skill"
)

// File is one emitted output file, path relative to the harness output root.
type File struct {
	RelPath string
	Bytes   []byte
}

// Profile renders a Skill into one harness's native form.
type Profile interface {
	Name() string
	SkillDir(skillName string) string
	ToolMap() map[string]string
	Frontmatter(s *skill.Skill) string
	Enhancements(s *skill.Skill) []File
}

var registry = map[string]Profile{
	"codex":    Codex{},
	"copilot":  Copilot{},
	"cursor":   Cursor{},
	"gemini":   Gemini{},
	"opencode": OpenCode{},
}

// Get returns the profile registered under name.
func Get(name string) (Profile, bool) {
	p, ok := registry[name]
	return p, ok
}

// Names returns the sorted list of registered profile names.
func Names() []string {
	out := make([]string, 0, len(registry))
	for n := range registry {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

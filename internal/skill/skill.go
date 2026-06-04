// Package skill parses a canonical SKILL.md into a harness-neutral IR.
package skill

import (
	"fmt"
	"regexp"
)

// Asset is a non-SKILL.md file packaged with the skill.
type Asset struct {
	RelPath string
	Bytes   []byte
}

// Skill is the harness-neutral intermediate representation.
type Skill struct {
	Name          string
	Description   string
	License       string
	Compatibility string
	AllowedTools  []string
	Body          string
	Dir           string
	Assets        []Asset
}

var nameRe = regexp.MustCompile(`^[a-z0-9-]+$`)

// Validate enforces the agentskills.io required-field rules.
func (s *Skill) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("skill: name is required")
	}
	if len(s.Name) > 64 {
		return fmt.Errorf("skill: name %q exceeds 64 chars", s.Name)
	}
	if !nameRe.MatchString(s.Name) {
		return fmt.Errorf("skill: name %q must be lowercase letters, numbers, hyphens", s.Name)
	}
	if s.Description == "" {
		return fmt.Errorf("skill: description is required")
	}
	if len(s.Description) > 1024 {
		return fmt.Errorf("skill: description exceeds 1024 chars")
	}
	return nil
}

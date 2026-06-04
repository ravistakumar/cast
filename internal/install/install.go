// Package install writes emitted files into a target directory tree.
package install

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ravistakumar/cast/internal/profile"
)

// Install writes each file under root, creating parent directories.
func Install(files []profile.File, root string) error {
	for _, f := range files {
		dest := filepath.Join(root, filepath.FromSlash(f.RelPath))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return fmt.Errorf("install: creating directory for %s: %w", dest, err)
		}
		if err := os.WriteFile(dest, f.Bytes, 0o644); err != nil {
			return fmt.Errorf("install: writing %s: %w", dest, err)
		}
	}
	return nil
}

// liveRoots maps a harness to its live skills root under the home dir. Combined
// with a profile's SkillDir ("skills/<name>"), these resolve to the documented
// user-level skill locations, e.g. ~/.codex/skills/<name>/ and
// ~/.config/opencode/skills/<name>/ (developers.openai.com/codex/skills,
// cursor.com/docs/context/skills, geminicli.com/docs/cli/skills,
// code.visualstudio.com/docs/copilot/customization/agent-skills,
// opencode.ai/docs/skills).
var liveRoots = map[string]string{
	"codex":    ".codex",
	"copilot":  ".copilot",
	"cursor":   ".cursor",
	"gemini":   ".gemini",
	"opencode": ".config/opencode",
}

// DefaultRoot returns the live install root for a harness.
func DefaultRoot(harness string) (string, error) {
	sub, ok := liveRoots[harness]
	if !ok {
		return "", fmt.Errorf("install: no known live directory for harness %q", harness)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, sub), nil
}

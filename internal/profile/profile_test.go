package profile

import (
	"strings"
	"testing"

	"github.com/ravistakumar/cast/internal/skill"
)

func TestRegistry(t *testing.T) {
	names := Names()
	if len(names) != 5 {
		t.Fatalf("expected 5 profiles, got %v", names)
	}
	for _, n := range []string{"codex", "copilot", "cursor", "gemini", "opencode"} {
		if _, ok := Get(n); !ok {
			t.Fatalf("missing profile %q", n)
		}
	}
	if _, ok := Get("nope"); ok {
		t.Fatal("unknown profile should not resolve")
	}
}

func TestNewProfilesShape(t *testing.T) {
	s := &skill.Skill{Name: "pdf", Description: "PDFs"}
	// (canonical tool, expected mapped name) confirmed from each harness's docs.
	cases := []struct {
		profile  string
		tool     string
		mapped   string
	}{
		{"gemini", "Bash", "run_shell_command"},
		{"copilot", "Write", "create_file"},
		{"opencode", "Read", "read"},
	}
	for _, c := range cases {
		t.Run(c.profile, func(t *testing.T) {
			p, _ := Get(c.profile)
			if got := p.ToolMap()[c.tool]; got != c.mapped {
				t.Fatalf("%s ToolMap[%s] = %q, want %q", c.profile, c.tool, got, c.mapped)
			}
			if files := p.Enhancements(s); files != nil {
				t.Fatalf("%s should emit no enhancement files, got %+v", c.profile, files)
			}
			fm := p.Frontmatter(s)
			if !strings.Contains(fm, "name: pdf") || !strings.Contains(fm, "description: PDFs") {
				t.Fatalf("%s frontmatter missing fields:\n%s", c.profile, fm)
			}
		})
	}
}

func TestOpenCodeWarnsOnAbsentTools(t *testing.T) {
	p, _ := Get("opencode")
	tm := p.ToolMap()
	for _, tool := range []string{"Task", "NotebookEdit"} {
		if v, ok := tm[tool]; !ok || v != "" {
			t.Fatalf("opencode should map %s to no-equivalent, got %q ok=%v", tool, v, ok)
		}
	}
}

func TestCodexFrontmatterAndEnhancements(t *testing.T) {
	s := &skill.Skill{Name: "pdf", Description: "PDFs", AllowedTools: []string{"Read"}}
	p, _ := Get("codex")
	fm := p.Frontmatter(s)
	if !strings.Contains(fm, "name: pdf") || !strings.Contains(fm, "description: PDFs") {
		t.Fatalf("codex frontmatter missing fields:\n%s", fm)
	}
	files := p.Enhancements(s)
	if len(files) != 1 || files[0].RelPath != "agents/openai.yaml" {
		t.Fatalf("codex should emit agents/openai.yaml, got %+v", files)
	}
	yaml := string(files[0].Bytes)
	if !strings.Contains(yaml, "interface:") ||
		!strings.Contains(yaml, "display_name: pdf") ||
		!strings.Contains(yaml, "short_description: PDFs") {
		t.Fatalf("openai.yaml missing documented interface schema:\n%s", yaml)
	}
}

func TestCursorToolMap(t *testing.T) {
	p, _ := Get("cursor")
	if v, ok := p.ToolMap()["Task"]; !ok || v != "" {
		t.Fatalf("cursor should map Task to no-equivalent, got %q ok=%v", v, ok)
	}
}

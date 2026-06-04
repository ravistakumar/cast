package profile

import (
	"strings"
	"testing"

	"github.com/ravistakumar/cast/internal/skill"
)

func TestRegistry(t *testing.T) {
	names := Names()
	if len(names) != 2 {
		t.Fatalf("expected 2 profiles, got %v", names)
	}
	for _, n := range []string{"codex", "cursor"} {
		if _, ok := Get(n); !ok {
			t.Fatalf("missing profile %q", n)
		}
	}
	if _, ok := Get("nope"); ok {
		t.Fatal("unknown profile should not resolve")
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

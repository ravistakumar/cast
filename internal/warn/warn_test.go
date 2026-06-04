package warn

import (
	"strings"
	"testing"
)

func TestRenderGroupsByHarness(t *testing.T) {
	ws := []Warning{
		{Harness: "codex", Severity: Warn, Location: "body", Code: "tool-no-equivalent", Message: "Task has no Codex equivalent"},
		{Harness: "cursor", Severity: Info, Location: "frontmatter", Code: "dropped-field", Message: "allowed-tools dropped"},
	}
	out := Render(ws)
	if !strings.Contains(out, "codex") || !strings.Contains(out, "cursor") {
		t.Fatalf("expected both harnesses in report:\n%s", out)
	}
	if !strings.Contains(out, "tool-no-equivalent") {
		t.Fatalf("expected code in report:\n%s", out)
	}
}

func TestRenderEmpty(t *testing.T) {
	if Render(nil) != "" {
		t.Fatal("empty warnings should render empty string")
	}
}

package optimize

import (
	"errors"
	"testing"

	"github.com/ravistakumar/cast/internal/profile"
	"github.com/ravistakumar/cast/internal/skill"
	"github.com/ravistakumar/cast/internal/warn"
)

type stubRunner struct {
	out string
	err error
}

func (s stubRunner) Run(prompt string) (string, error) { return s.out, s.err }

func TestOptimizeUsesAgentOutput(t *testing.T) {
	s := &skill.Skill{Name: "pdf", Description: "PDFs", Body: "use Task"}
	p, _ := profile.Get("codex")
	ws := []warn.Warning{{Harness: "codex", Code: "tool-no-equivalent", Message: "Task"}}
	body, out := Optimize(s, p, ws, stubRunner{out: "adapted body"})
	if body != "adapted body" {
		t.Fatalf("expected adapted body, got %q", body)
	}
	if len(out) != 0 {
		t.Fatalf("successful optimize should clear warnings, got %+v", out)
	}
}

func TestOptimizeFallsBackOnError(t *testing.T) {
	s := &skill.Skill{Name: "pdf", Description: "PDFs", Body: "use Task"}
	p, _ := profile.Get("codex")
	ws := []warn.Warning{{Harness: "codex", Code: "tool-no-equivalent", Message: "Task"}}
	body, out := Optimize(s, p, ws, stubRunner{err: errors.New("cli down")})
	if body != "use Task" {
		t.Fatalf("expected original body on failure, got %q", body)
	}
	if !hasCode(out, "optimize-failed") {
		t.Fatalf("expected optimize-failed warning, got %+v", out)
	}
}

func TestOptimizeFallsBackOnEmptyOutput(t *testing.T) {
	s := &skill.Skill{Name: "pdf", Description: "PDFs", Body: "use Task"}
	p, _ := profile.Get("codex")
	ws := []warn.Warning{{Harness: "codex", Code: "tool-no-equivalent", Message: "Task"}}
	body, out := Optimize(s, p, ws, stubRunner{out: "   \n  "})
	if body != "use Task" {
		t.Fatalf("expected original body on empty output, got %q", body)
	}
	if !hasCode(out, "optimize-failed") {
		t.Fatalf("expected optimize-failed warning, got %+v", out)
	}
}

func hasCode(ws []warn.Warning, code string) bool {
	for _, w := range ws {
		if w.Code == code {
			return true
		}
	}
	return false
}

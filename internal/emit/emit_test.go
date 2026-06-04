package emit

import (
	"strings"
	"testing"

	"github.com/ravistakumar/cast/internal/profile"
	"github.com/ravistakumar/cast/internal/skill"
	"github.com/ravistakumar/cast/internal/warn"
)

func find(files []profile.File, rel string) (profile.File, bool) {
	for _, f := range files {
		if f.RelPath == rel {
			return f, true
		}
	}
	return profile.File{}, false
}

func TestEmitWritesSkillAndAssets(t *testing.T) {
	s := &skill.Skill{
		Name: "pdf", Description: "PDFs",
		Body:   "Use the Task tool then Read the file.",
		Assets: []skill.Asset{{RelPath: "scripts/run.sh", Bytes: []byte("echo hi")}},
	}
	p, _ := profile.Get("codex")
	files, warns := Emit(s, p)

	sk, ok := find(files, "skills/pdf/SKILL.md")
	if !ok {
		t.Fatalf("expected skills/pdf/SKILL.md, got %v", paths(files))
	}
	if !strings.Contains(string(sk.Bytes), "name: pdf") || !strings.Contains(string(sk.Bytes), "Use the Task tool") {
		t.Fatalf("SKILL.md content wrong:\n%s", sk.Bytes)
	}
	if _, ok := find(files, "skills/pdf/scripts/run.sh"); !ok {
		t.Fatalf("expected asset copied, got %v", paths(files))
	}
	if _, ok := find(files, "skills/pdf/agents/openai.yaml"); !ok {
		t.Fatalf("expected enhancement file, got %v", paths(files))
	}
	// "Task" has no Codex equivalent -> a warning; "Read" maps -> no warning.
	if !hasCode(warns, "tool-no-equivalent") {
		t.Fatalf("expected tool-no-equivalent warning, got %+v", warns)
	}
	for _, w := range warns {
		if w.Code == "tool-no-equivalent" && !strings.Contains(w.Message, "Task") {
			t.Fatalf("warning should name Task: %+v", w)
		}
	}
}

func paths(files []profile.File) []string {
	var out []string
	for _, f := range files {
		out = append(out, f.RelPath)
	}
	return out
}

func hasCode(ws []warn.Warning, code string) bool {
	for _, w := range ws {
		if w.Code == code {
			return true
		}
	}
	return false
}

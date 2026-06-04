package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSkill(t *testing.T, dir string) {
	t.Helper()
	sk := filepath.Join(dir, "demo")
	if err := os.MkdirAll(sk, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nname: demo\ndescription: A demo skill\n---\n\nUse the Task tool.\n"
	if err := os.WriteFile(filepath.Join(sk, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestTargetsCommand(t *testing.T) {
	var out bytes.Buffer
	root := New()
	root.SetOut(&out)
	root.SetArgs([]string{"targets"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "codex") || !strings.Contains(out.String(), "cursor") {
		t.Fatalf("targets output missing harnesses: %s", out.String())
	}
}

func TestBuildCommandWritesDist(t *testing.T) {
	tmp := t.TempDir()
	writeSkill(t, tmp)
	out := filepath.Join(tmp, "dist")
	var buf bytes.Buffer
	root := New()
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"build", filepath.Join(tmp, "demo"), "--target", "codex", "--outdir", out})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "codex", "skills", "demo", "SKILL.md")); err != nil {
		t.Fatalf("expected compiled SKILL.md: %v", err)
	}
	if !strings.Contains(buf.String(), "tool-no-equivalent") {
		t.Fatalf("expected warning report in output: %s", buf.String())
	}
}

func TestCheckCommandRejectsInvalid(t *testing.T) {
	tmp := t.TempDir()
	bad := filepath.Join(tmp, "bad")
	_ = os.MkdirAll(bad, 0o755)
	_ = os.WriteFile(filepath.Join(bad, "SKILL.md"), []byte("no frontmatter"), 0o644)
	root := New()
	root.SetArgs([]string{"check", bad})
	if err := root.Execute(); err == nil {
		t.Fatal("check should fail on invalid skill")
	}
}

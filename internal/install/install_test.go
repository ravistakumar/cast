package install

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ravistakumar/cast/internal/profile"
)

func TestInstallWritesNestedFiles(t *testing.T) {
	root := t.TempDir()
	files := []profile.File{
		{RelPath: "skills/pdf/SKILL.md", Bytes: []byte("hi")},
		{RelPath: "skills/pdf/agents/openai.yaml", Bytes: []byte("x")},
	}
	if err := Install(files, root); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(root, "skills/pdf/SKILL.md"))
	if err != nil || string(got) != "hi" {
		t.Fatalf("file not written correctly: %v %q", err, got)
	}
	if _, err := os.Stat(filepath.Join(root, "skills/pdf/agents/openai.yaml")); err != nil {
		t.Fatalf("nested file missing: %v", err)
	}
}

func TestDefaultRootKnownHarness(t *testing.T) {
	r, err := DefaultRoot("codex")
	if err != nil || r == "" {
		t.Fatalf("expected a root for codex, got %q err=%v", r, err)
	}
	if _, err := DefaultRoot("nope"); err == nil {
		t.Fatal("unknown harness should error")
	}
}

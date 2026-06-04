package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// End-to-end: a multi-asset skill compiles to both targets and writes a full tree.
func TestE2EBuildBothTargets(t *testing.T) {
	tmp := t.TempDir()
	sk := filepath.Join(tmp, "demo")
	if err := os.MkdirAll(filepath.Join(sk, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nname: demo\ndescription: A demo skill\n---\n\nRead then Edit the file.\n"
	if err := os.WriteFile(filepath.Join(sk, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sk, "scripts", "run.sh"), []byte("echo hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(tmp, "dist")
	var buf bytes.Buffer
	root := New()
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"build", sk, "--target", "codex,cursor", "--outdir", out})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"codex/skills/demo/SKILL.md",
		"codex/skills/demo/scripts/run.sh",
		"codex/skills/demo/agents/openai.yaml",
		"cursor/skills/demo/SKILL.md",
		"cursor/skills/demo/scripts/run.sh",
	} {
		if _, err := os.Stat(filepath.Join(out, filepath.FromSlash(want))); err != nil {
			t.Fatalf("missing expected output %s: %v", want, err)
		}
	}
}

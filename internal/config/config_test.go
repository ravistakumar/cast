package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // no config file present
	clearEnv(t)
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Targets) != 2 || c.Outdir != "dist" || c.Optimize || c.Agent != "auto" {
		t.Fatalf("unexpected defaults: %+v", c)
	}
}

func TestFileAndEnvOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	clearEnv(t)
	cfgDir := filepath.Join(dir, "cast")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "targets = [\"codex\"]\noutdir = \"out\"\noptimize = true\nagent = \"codex\"\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CAST_OUTDIR", "envout")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Targets) != 1 || c.Targets[0] != "codex" {
		t.Fatalf("file targets not loaded: %+v", c)
	}
	if c.Outdir != "envout" {
		t.Fatalf("env should override file outdir: %+v", c)
	}
	if !c.Optimize || c.Agent != "codex" {
		t.Fatalf("file values not loaded: %+v", c)
	}
}

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"CAST_TARGETS", "CAST_OUTDIR", "CAST_OPTIMIZE", "CAST_AGENT"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}
}

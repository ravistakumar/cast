# cast Skill Compiler Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `cast`, a Go CLI that compiles one canonical spec-compliant `SKILL.md` into harness-native skills for Codex and Cursor, with a deterministic structural core and an optional LLM adaptation pass.

**Architecture:** Functional core (`parse → normalize to IR → emit through per-harness profiles + structured warnings`) wrapped by an imperative shell (file I/O, config, install, optional agent-CLI subprocess). Adding a harness = adding one `Profile`. Mirrors the `prr`/`fan` sibling conventions.

**Tech Stack:** Go 1.23, `github.com/spf13/cobra` (CLI), `gopkg.in/yaml.v3` (SKILL.md frontmatter), `github.com/BurntSushi/toml` (config). Single static binary.

---

## Shared Type Reference (defined across tasks — do not redefine)

These are the canonical signatures every task must match exactly.

```go
// package skill
type Asset struct {
    RelPath string // path relative to the skill dir, e.g. "scripts/run.py"
    Bytes   []byte
}
type Skill struct {
    Name          string   // required
    Description   string   // required
    License       string   // optional frontmatter
    Compatibility string   // optional frontmatter (e.g. "Designed for Claude Code")
    AllowedTools  []string // optional, experimental ("allowed-tools")
    Body          string   // markdown body after frontmatter
    Dir           string   // source directory path
    Assets        []Asset  // all files under scripts/ references/ assets/
}
func Parse(dir string) (*Skill, error)
func (s *Skill) Validate() error

// package warn
type Severity string
const ( Info Severity = "info"; Warn Severity = "warn" )
type Warning struct {
    Harness, Code, Message, Suggestion, Location string
    Severity Severity
}
func Render(ws []Warning) string

// package profile
type File struct {
    RelPath string // relative to the harness output root
    Bytes   []byte
}
type Profile interface {
    Name() string
    SkillDir(skillName string) string                 // output subdir for this skill
    ToolMap() map[string]string                       // canonical tool -> harness tool ("" = no equivalent)
    Frontmatter(s *skill.Skill) string                // rendered SKILL.md frontmatter block (no fences)
    Enhancements(s *skill.Skill) []File               // extra harness-specific files
}
func Get(name string) (Profile, bool)
func Names() []string

// package emit
func Emit(s *skill.Skill, p profile.Profile) ([]profile.File, []warn.Warning)

// package config
type Config struct {
    Targets  []string
    Outdir   string
    Optimize bool
    Agent    string
}
func Load() (*Config, error)

// package agent
type Runner interface { Run(prompt string) (string, error) }
func Detect() (string, error)
func NewRunner(name string) (Runner, error)

// package install
func Install(files []profile.File, root string) error
func DefaultRoot(harness string) (string, error)

// package optimize
func Optimize(s *skill.Skill, p profile.Profile, ws []warn.Warning, r agent.Runner) (body string, out []warn.Warning)
```

Canonical Claude tool-name set used for body scanning (lowercase-compared, whole word):
`Read, Write, Edit, Bash, Glob, Grep, Task, WebFetch, WebSearch, NotebookEdit, TodoWrite`.

---

## Task 1: Project scaffold

**Files:**
- Create: `go.mod`
- Create: `cmd/cast/main.go`
- Create: `internal/version/version_test.go`
- Create: `internal/version/version.go`
- Create: `Makefile`

- [ ] **Step 1: Initialize the module**

Run:
```bash
cd /Users/ravisubedi/dev/opensource/cast
go mod init github.com/ravistakumar/cast
go get github.com/spf13/cobra@v1.8.1 gopkg.in/yaml.v3@v3.0.1 github.com/BurntSushi/toml@v1.6.0
```
Expected: `go.mod` created with `go 1.23` and the three requires.

- [ ] **Step 2: Write the failing test**

`internal/version/version_test.go`:
```go
package version

import "testing"

func TestStringIsNonEmpty(t *testing.T) {
	if String() == "" {
		t.Fatal("version.String() must not be empty")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./internal/version/`
Expected: build failure — `undefined: String`.

- [ ] **Step 4: Write minimal implementation**

`internal/version/version.go`:
```go
// Package version exposes the build version string.
package version

// Version is overridden at build time via -ldflags.
var Version = "dev"

// String returns the current version.
func String() string { return Version }
```

- [ ] **Step 5: Write minimal main**

`cmd/cast/main.go`:
```go
package main

import (
	"fmt"
	"os"

	"github.com/ravistakumar/cast/internal/version"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("cast", version.String())
		return
	}
	fmt.Fprintln(os.Stderr, "cast: no command (cli wired in a later task)")
	os.Exit(1)
}
```

`Makefile`:
```makefile
build:
	go build -o cast ./cmd/cast

test:
	go test ./...
```

- [ ] **Step 6: Run tests and build**

Run: `go test ./... && go build ./cmd/cast`
Expected: PASS, binary builds.

- [ ] **Step 7: Commit**

```bash
git add go.mod go.sum cmd internal/version Makefile
git commit -m "Scaffold cast module, version package, and entry point"
```

---

## Task 2: Parse and validate the canonical SKILL.md into the IR

**Files:**
- Test: `internal/skill/skill_test.go`
- Create: `internal/skill/skill.go`
- Create: `internal/skill/parse.go`

- [ ] **Step 1: Write the failing validation tests**

`internal/skill/skill_test.go`:
```go
package skill

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		s       Skill
		wantErr bool
	}{
		{"ok", Skill{Name: "pdf-tools", Description: "Work with PDFs"}, false},
		{"missing name", Skill{Description: "x"}, true},
		{"missing desc", Skill{Name: "pdf-tools"}, true},
		{"bad chars", Skill{Name: "PDF_Tools", Description: "x"}, true},
		{"too long name", Skill{Name: string(make([]byte, 65)), Description: "x"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.s.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/skill/`
Expected: build failure — `undefined: Skill`.

- [ ] **Step 3: Implement the IR type and Validate**

`internal/skill/skill.go`:
```go
// Package skill parses a canonical SKILL.md into a harness-neutral IR.
package skill

import (
	"fmt"
	"regexp"
)

// Asset is a non-SKILL.md file packaged with the skill.
type Asset struct {
	RelPath string
	Bytes   []byte
}

// Skill is the harness-neutral intermediate representation.
type Skill struct {
	Name          string
	Description   string
	License       string
	Compatibility string
	AllowedTools  []string
	Body          string
	Dir           string
	Assets        []Asset
}

var nameRe = regexp.MustCompile(`^[a-z0-9-]+$`)

// Validate enforces the agentskills.io required-field rules.
func (s *Skill) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("skill: name is required")
	}
	if len(s.Name) > 64 {
		return fmt.Errorf("skill: name %q exceeds 64 chars", s.Name)
	}
	if !nameRe.MatchString(s.Name) {
		return fmt.Errorf("skill: name %q must be lowercase letters, numbers, hyphens", s.Name)
	}
	if s.Description == "" {
		return fmt.Errorf("skill: description is required")
	}
	if len(s.Description) > 1024 {
		return fmt.Errorf("skill: description exceeds 1024 chars")
	}
	return nil
}
```

- [ ] **Step 4: Run validation tests**

Run: `go test ./internal/skill/ -run TestValidate -v`
Expected: PASS.

- [ ] **Step 5: Write the failing parse test**

Add a fixture and test. Create `internal/skill/testdata/good/SKILL.md`:
```markdown
---
name: good
description: A good skill
license: MIT
compatibility: Designed for Claude Code
allowed-tools:
  - Read
  - Bash
---
# Body heading

Do the thing with the Task tool.
```
Create `internal/skill/testdata/good/scripts/run.sh` with content `echo hi`.

Add `"strings"` to the imports of `internal/skill/skill_test.go`, then add:
```go
func TestParse(t *testing.T) {
	s, err := Parse("testdata/good")
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "good" || s.Description != "A good skill" {
		t.Fatalf("frontmatter wrong: %+v", s)
	}
	if s.License != "MIT" || s.Compatibility != "Designed for Claude Code" {
		t.Fatalf("optional fields wrong: %+v", s)
	}
	if len(s.AllowedTools) != 2 || s.AllowedTools[0] != "Read" {
		t.Fatalf("allowed-tools wrong: %v", s.AllowedTools)
	}
	if !strings.Contains(s.Body, "Body heading") {
		t.Fatalf("body not captured: %q", s.Body)
	}
	if len(s.Assets) != 1 || s.Assets[0].RelPath != "scripts/run.sh" {
		t.Fatalf("assets wrong: %+v", s.Assets)
	}
}
```

The full import block for `internal/skill/skill_test.go` is therefore:
```go
import (
	"strings"
	"testing"
)
```

- [ ] **Step 6: Run parse test to verify it fails**

Run: `go test ./internal/skill/ -run TestParse`
Expected: build failure — `undefined: Parse`.

- [ ] **Step 7: Implement Parse**

`internal/skill/parse.go`:
```go
package skill

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type frontmatter struct {
	Name          string   `yaml:"name"`
	Description   string   `yaml:"description"`
	License       string   `yaml:"license"`
	Compatibility string   `yaml:"compatibility"`
	AllowedTools  []string `yaml:"allowed-tools"`
}

// Parse reads dir/SKILL.md, parses YAML frontmatter and markdown body,
// loads assets from scripts/ references/ assets/, and validates.
func Parse(dir string) (*Skill, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		return nil, fmt.Errorf("skill: reading SKILL.md: %w", err)
	}
	fmText, body, err := splitFrontmatter(raw)
	if err != nil {
		return nil, err
	}
	var fm frontmatter
	if err := yaml.Unmarshal(fmText, &fm); err != nil {
		return nil, fmt.Errorf("skill: parsing frontmatter: %w", err)
	}
	s := &Skill{
		Name:          fm.Name,
		Description:   fm.Description,
		License:       fm.License,
		Compatibility: fm.Compatibility,
		AllowedTools:  fm.AllowedTools,
		Body:          body,
		Dir:           dir,
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	// name must match parent dir per spec.
	if base := filepath.Base(dir); base != s.Name {
		return nil, fmt.Errorf("skill: name %q must match directory %q", s.Name, base)
	}
	assets, err := loadAssets(dir)
	if err != nil {
		return nil, err
	}
	s.Assets = assets
	return s, nil
}

func splitFrontmatter(raw []byte) (fm []byte, body string, err error) {
	r := bytes.TrimLeft(raw, "\xef\xbb\xbf \t\r\n")
	if !bytes.HasPrefix(r, []byte("---")) {
		return nil, "", fmt.Errorf("skill: SKILL.md must start with YAML frontmatter delimited by ---")
	}
	rest := r[3:]
	idx := bytes.Index(rest, []byte("\n---"))
	if idx < 0 {
		return nil, "", fmt.Errorf("skill: unterminated frontmatter (missing closing ---)")
	}
	fm = rest[:idx]
	after := rest[idx+len("\n---"):]
	if nl := bytes.IndexByte(after, '\n'); nl >= 0 {
		after = after[nl+1:]
	} else {
		after = nil
	}
	return fm, string(after), nil
}

func loadAssets(dir string) ([]Asset, error) {
	var assets []Asset
	for _, sub := range []string{"scripts", "references", "assets"} {
		root := filepath.Join(dir, sub)
		if _, err := os.Stat(root); err != nil {
			continue
		}
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(dir, p)
			if err != nil {
				return err
			}
			assets = append(assets, Asset{RelPath: filepath.ToSlash(rel), Bytes: b})
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("skill: loading %s: %w", sub, err)
		}
	}
	return assets, nil
}
```

- [ ] **Step 8: Run all skill tests**

Run: `go test ./internal/skill/ -v`
Expected: PASS (TestValidate, TestParse).

- [ ] **Step 9: Commit**

```bash
git add internal/skill
git commit -m "Add SKILL.md parser and IR with spec validation"
```

---

## Task 3: Warning type and report rendering

**Files:**
- Test: `internal/warn/warn_test.go`
- Create: `internal/warn/warn.go`

- [ ] **Step 1: Write the failing test**

`internal/warn/warn_test.go`:
```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/warn/`
Expected: build failure — `undefined: Warning`.

- [ ] **Step 3: Implement warn**

`internal/warn/warn.go`:
```go
// Package warn defines structured compile warnings and their report rendering.
package warn

import (
	"fmt"
	"sort"
	"strings"
)

// Severity classifies a warning.
type Severity string

const (
	Info Severity = "info"
	Warn Severity = "warn"
)

// Warning is a structured, machine-coded compile diagnostic.
type Warning struct {
	Harness    string
	Severity   Severity
	Location   string
	Code       string
	Message    string
	Suggestion string
}

// Render groups warnings by harness into a stable, human-readable report.
// Empty input renders an empty string.
func Render(ws []Warning) string {
	if len(ws) == 0 {
		return ""
	}
	byHarness := map[string][]Warning{}
	for _, w := range ws {
		byHarness[w.Harness] = append(byHarness[w.Harness], w)
	}
	harnesses := make([]string, 0, len(byHarness))
	for h := range byHarness {
		harnesses = append(harnesses, h)
	}
	sort.Strings(harnesses)

	var b strings.Builder
	for _, h := range harnesses {
		fmt.Fprintf(&b, "%s:\n", h)
		group := byHarness[h]
		sort.SliceStable(group, func(i, j int) bool { return group[i].Code < group[j].Code })
		for _, w := range group {
			fmt.Fprintf(&b, "  [%s] %s (%s): %s\n", w.Severity, w.Code, w.Location, w.Message)
			if w.Suggestion != "" {
				fmt.Fprintf(&b, "        ↳ %s\n", w.Suggestion)
			}
		}
	}
	return b.String()
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/warn/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/warn
git commit -m "Add structured warning type and report rendering"
```

---

## Task 4: Profile interface, registry, and the Codex + Cursor profiles

**Files:**
- Test: `internal/profile/profile_test.go`
- Create: `internal/profile/profile.go`
- Create: `internal/profile/codex.go`
- Create: `internal/profile/cursor.go`

> **Verify during implementation:** the live install directory names and the exact `agents/openai.yaml` schema for Codex, and Cursor's skill directory convention, are version-sensitive. The constants below are centralized so a single edit corrects them; tests do not depend on real paths.

- [ ] **Step 1: Write the failing test**

`internal/profile/profile_test.go`:
```go
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
}

func TestCursorToolMap(t *testing.T) {
	p, _ := Get("cursor")
	if v, ok := p.ToolMap()["Task"]; !ok || v != "" {
		t.Fatalf("cursor should map Task to no-equivalent, got %q ok=%v", v, ok)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/profile/`
Expected: build failure — `undefined: Names`.

- [ ] **Step 3: Implement the interface and registry**

`internal/profile/profile.go`:
```go
// Package profile describes how to emit a Skill for one target harness.
package profile

import (
	"sort"

	"github.com/ravistakumar/cast/internal/skill"
)

// File is one emitted output file, path relative to the harness output root.
type File struct {
	RelPath string
	Bytes   []byte
}

// Profile renders a Skill into one harness's native form.
type Profile interface {
	Name() string
	SkillDir(skillName string) string
	ToolMap() map[string]string
	Frontmatter(s *skill.Skill) string
	Enhancements(s *skill.Skill) []File
}

var registry = map[string]Profile{
	"codex":  Codex{},
	"cursor": Cursor{},
}

// Get returns the profile registered under name.
func Get(name string) (Profile, bool) {
	p, ok := registry[name]
	return p, ok
}

// Names returns the sorted list of registered profile names.
func Names() []string {
	out := make([]string, 0, len(registry))
	for n := range registry {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
```

- [ ] **Step 4: Implement the Codex profile**

`internal/profile/codex.go`:
```go
package profile

import (
	"fmt"
	"strings"

	"github.com/ravistakumar/cast/internal/skill"
)

// codexInstallSubdir is the per-skill output subdirectory. VERIFY against
// current Codex docs before relying on --install.
const codexInstallSubdir = "skills"

// Codex emits a skill for the OpenAI Codex CLI.
type Codex struct{}

func (Codex) Name() string { return "codex" }

func (Codex) SkillDir(name string) string { return codexInstallSubdir + "/" + name }

// ToolMap maps canonical Claude tool names to Codex equivalents.
// "" means no known equivalent (emit warns).
func (Codex) ToolMap() map[string]string {
	return map[string]string{
		"Read": "read_file", "Write": "write_file", "Edit": "apply_patch",
		"Bash": "shell", "Glob": "", "Grep": "", "Task": "",
		"WebFetch": "", "WebSearch": "", "NotebookEdit": "", "TodoWrite": "",
	}
}

func (Codex) Frontmatter(s *skill.Skill) string {
	var b strings.Builder
	fmt.Fprintf(&b, "name: %s\n", s.Name)
	fmt.Fprintf(&b, "description: %s\n", s.Description)
	if s.License != "" {
		fmt.Fprintf(&b, "license: %s\n", s.License)
	}
	return b.String()
}

// Enhancements emits Codex's agents/openai.yaml extension file.
func (Codex) Enhancements(s *skill.Skill) []File {
	var b strings.Builder
	fmt.Fprintf(&b, "name: %s\n", s.Name)
	fmt.Fprintf(&b, "description: %s\n", s.Description)
	return []File{{RelPath: "agents/openai.yaml", Bytes: []byte(b.String())}}
}
```

- [ ] **Step 5: Implement the Cursor profile**

`internal/profile/cursor.go`:
```go
package profile

import (
	"fmt"
	"strings"

	"github.com/ravistakumar/cast/internal/skill"
)

// cursorInstallSubdir is the per-skill output subdirectory. VERIFY against
// current Cursor docs before relying on --install.
const cursorInstallSubdir = "skills"

// Cursor emits a skill for the Cursor agent.
type Cursor struct{}

func (Cursor) Name() string { return "cursor" }

func (Cursor) SkillDir(name string) string { return cursorInstallSubdir + "/" + name }

func (Cursor) ToolMap() map[string]string {
	return map[string]string{
		"Read": "read_file", "Write": "edit_file", "Edit": "edit_file",
		"Bash": "run_terminal_cmd", "Glob": "", "Grep": "grep",
		"Task": "", "WebFetch": "", "WebSearch": "web", "NotebookEdit": "", "TodoWrite": "",
	}
}

func (Cursor) Frontmatter(s *skill.Skill) string {
	var b strings.Builder
	fmt.Fprintf(&b, "name: %s\n", s.Name)
	fmt.Fprintf(&b, "description: %s\n", s.Description)
	return b.String()
}

// Enhancements: Cursor needs no extra files in v1.
func (Cursor) Enhancements(s *skill.Skill) []File { return nil }
```

- [ ] **Step 6: Run all profile tests**

Run: `go test ./internal/profile/ -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/profile
git commit -m "Add profile interface, registry, and Codex + Cursor profiles"
```

---

## Task 5: Emit — walk the IR through a profile into files + warnings

**Files:**
- Test: `internal/emit/emit_test.go`
- Create: `internal/emit/emit.go`

- [ ] **Step 1: Write the failing test**

`internal/emit/emit_test.go`:
```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/emit/`
Expected: build failure — `undefined: Emit`.

- [ ] **Step 3: Implement Emit**

`internal/emit/emit.go`:
```go
// Package emit walks a Skill IR through a Profile into output files + warnings.
package emit

import (
	"regexp"

	"github.com/ravistakumar/cast/internal/profile"
	"github.com/ravistakumar/cast/internal/skill"
	"github.com/ravistakumar/cast/internal/warn"
)

// canonicalTools is the set of Claude tool names scanned for in the body.
var canonicalTools = []string{
	"Read", "Write", "Edit", "Bash", "Glob", "Grep",
	"Task", "WebFetch", "WebSearch", "NotebookEdit", "TodoWrite",
}

// Emit produces a profile's output files for a skill plus structured warnings.
// The body is copied verbatim; only deterministic detection happens here.
func Emit(s *skill.Skill, p profile.Profile) ([]profile.File, []warn.Warning) {
	dir := p.SkillDir(s.Name)
	var files []profile.File

	// SKILL.md = harness frontmatter + canonical body.
	content := "---\n" + p.Frontmatter(s) + "---\n\n" + s.Body
	files = append(files, profile.File{RelPath: dir + "/SKILL.md", Bytes: []byte(content)})

	// Copy assets verbatim under the skill dir.
	for _, a := range s.Assets {
		files = append(files, profile.File{RelPath: dir + "/" + a.RelPath, Bytes: a.Bytes})
	}

	// Harness enhancement files (placed under the skill dir).
	for _, f := range p.Enhancements(s) {
		files = append(files, profile.File{RelPath: dir + "/" + f.RelPath, Bytes: f.Bytes})
	}

	warns := scanTools(s, p)
	return files, warns
}

func scanTools(s *skill.Skill, p profile.Profile) []warn.Warning {
	tm := p.ToolMap()
	var warns []warn.Warning
	for _, tool := range canonicalTools {
		if !mentions(s.Body, tool) {
			continue
		}
		repl, known := tm[tool]
		if known && repl == "" {
			warns = append(warns, warn.Warning{
				Harness: p.Name(), Severity: warn.Warn, Location: "body",
				Code: "tool-no-equivalent",
				Message: "body references the " + tool + " tool, which has no " + p.Name() + " equivalent",
				Suggestion: "run with --optimize to let the agent CLI adapt this phrasing",
			})
		}
	}
	return warns
}

func mentions(body, tool string) bool {
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(tool) + `\b`)
	return re.MatchString(body)
}
```

- [ ] **Step 4: Run emit tests**

Run: `go test ./internal/emit/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/emit
git commit -m "Add emit: IR + profile -> files and tool-divergence warnings"
```

---

## Task 6: Config (TOML + env + defaults)

**Files:**
- Test: `internal/config/config_test.go`
- Create: `internal/config/config.go`

- [ ] **Step 1: Write the failing test**

`internal/config/config_test.go`:
```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/`
Expected: build failure — `undefined: Load`.

- [ ] **Step 3: Implement config**

`internal/config/config.go`:
```go
// Package config loads cast settings from TOML, then CAST_* env overrides.
package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds resolved cast settings.
type Config struct {
	Targets  []string `toml:"targets"`
	Outdir   string   `toml:"outdir"`
	Optimize bool     `toml:"optimize"`
	Agent    string   `toml:"agent"`
}

func defaults() *Config {
	return &Config{Targets: []string{"codex", "cursor"}, Outdir: "dist", Optimize: false, Agent: "auto"}
}

// Load reads ~/.config/cast/config.toml (honoring XDG_CONFIG_HOME), then
// applies CAST_* environment overrides. Missing file is not an error.
func Load() (*Config, error) {
	c := defaults()
	if path := configPath(); path != "" {
		if _, err := os.Stat(path); err == nil {
			if _, err := toml.DecodeFile(path, c); err != nil {
				return nil, err
			}
		}
	}
	if v := os.Getenv("CAST_TARGETS"); v != "" {
		c.Targets = strings.Split(v, ",")
	}
	if v := os.Getenv("CAST_OUTDIR"); v != "" {
		c.Outdir = v
	}
	if v := os.Getenv("CAST_OPTIMIZE"); v == "1" || v == "true" {
		c.Optimize = true
	}
	if v := os.Getenv("CAST_AGENT"); v != "" {
		c.Agent = v
	}
	return c, nil
}

func configPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "cast", "config.toml")
}
```

- [ ] **Step 4: Run config tests**

Run: `go test ./internal/config/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/config
git commit -m "Add config loading: TOML defaults with CAST_* env overrides"
```

---

## Task 7: Install — write files to a root directory

**Files:**
- Test: `internal/install/install_test.go`
- Create: `internal/install/install.go`

- [ ] **Step 1: Write the failing test**

`internal/install/install_test.go`:
```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/install/`
Expected: build failure — `undefined: Install`.

- [ ] **Step 3: Implement install**

`internal/install/install.go`:
```go
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
			return err
		}
		if err := os.WriteFile(dest, f.Bytes, 0o644); err != nil {
			return fmt.Errorf("install: writing %s: %w", dest, err)
		}
	}
	return nil
}

// liveRoots maps a harness to its live skills root under the home dir.
// VERIFY these against current tool docs before relying on --install.
var liveRoots = map[string]string{
	"codex":  ".codex",
	"cursor": ".cursor",
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
```

- [ ] **Step 4: Run install tests**

Run: `go test ./internal/install/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/install
git commit -m "Add install: write emitted files to a root tree"
```

---

## Task 8: Agent — invoke an authenticated CLI (with a stubbable Runner)

**Files:**
- Test: `internal/agent/agent_test.go`
- Create: `internal/agent/agent.go`

- [ ] **Step 1: Write the failing test**

`internal/agent/agent_test.go`:
```go
package agent

import "testing"

func TestNewRunnerUnknown(t *testing.T) {
	if _, err := NewRunner("nope"); err == nil {
		t.Fatal("unknown agent should error")
	}
}

func TestCommandArgsPutPromptLast(t *testing.T) {
	args, err := commandFor("claude", "do the thing")
	if err != nil {
		t.Fatal(err)
	}
	if len(args) == 0 || args[len(args)-1] != "do the thing" {
		t.Fatalf("prompt must be the final argument: %v", args)
	}
	if args[0] != "claude" {
		t.Fatalf("expected claude binary first: %v", args)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/agent/`
Expected: build failure — `undefined: NewRunner`.

- [ ] **Step 3: Implement agent**

`internal/agent/agent.go`:
```go
// Package agent invokes an already-authenticated coding-agent CLI headlessly.
package agent

import (
	"fmt"
	"os/exec"
	"strings"
)

// Runner runs a prompt through an agent and returns its stdout.
type Runner interface {
	Run(prompt string) (string, error)
}

// commandFor builds the headless argv for a known agent; prompt is always last.
func commandFor(name, prompt string) ([]string, error) {
	switch name {
	case "claude":
		return []string{"claude", "-p", prompt}, nil
	case "codex":
		return []string{"codex", "exec", prompt}, nil
	default:
		return nil, fmt.Errorf("agent: unknown agent %q", name)
	}
}

type execRunner struct{ name string }

func (r execRunner) Run(prompt string) (string, error) {
	args, err := commandFor(r.name, prompt)
	if err != nil {
		return "", err
	}
	cmd := exec.Command(args[0], args[1:]...) // #nosec G204 — fixed binary, prompt is a single arg
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("agent: %s failed: %w", r.name, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// NewRunner returns a Runner for the named agent.
func NewRunner(name string) (Runner, error) {
	if _, err := commandFor(name, ""); err != nil {
		return nil, err
	}
	return execRunner{name: name}, nil
}

// Detect returns the first agent CLI found on PATH (claude, then codex).
func Detect() (string, error) {
	for _, name := range []string{"claude", "codex"} {
		if _, err := exec.LookPath(name); err == nil {
			return name, nil
		}
	}
	return "", fmt.Errorf("agent: no supported agent CLI found on PATH")
}
```

- [ ] **Step 4: Run agent tests**

Run: `go test ./internal/agent/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/agent
git commit -m "Add agent: headless authenticated-CLI runner with detection"
```

---

## Task 9: Optimize — optional LLM body adaptation with deterministic fallback

**Files:**
- Test: `internal/optimize/optimize_test.go`
- Create: `internal/optimize/optimize.go`

- [ ] **Step 1: Write the failing test (stub Runner: success + failure)**

`internal/optimize/optimize_test.go`:
```go
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

func hasCode(ws []warn.Warning, code string) bool {
	for _, w := range ws {
		if w.Code == code {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/optimize/`
Expected: build failure — `undefined: Optimize`.

- [ ] **Step 3: Implement optimize**

`internal/optimize/optimize.go`:
```go
// Package optimize optionally adapts a skill body per harness via an agent CLI.
package optimize

import (
	"fmt"
	"strings"

	"github.com/ravistakumar/cast/internal/agent"
	"github.com/ravistakumar/cast/internal/profile"
	"github.com/ravistakumar/cast/internal/skill"
	"github.com/ravistakumar/cast/internal/warn"
)

// Optimize asks the agent to rewrite the body for the target harness, resolving
// the supplied warnings. On any failure it returns the original body plus an
// optimize-failed warning — it never blocks.
func Optimize(s *skill.Skill, p profile.Profile, ws []warn.Warning, r agent.Runner) (string, []warn.Warning) {
	prompt := buildPrompt(s, p, ws)
	out, err := r.Run(prompt)
	if err != nil || strings.TrimSpace(out) == "" {
		return s.Body, []warn.Warning{{
			Harness: p.Name(), Severity: warn.Warn, Location: "body",
			Code: "optimize-failed", Message: fmt.Sprintf("optimization pass failed; using deterministic output: %v", err),
		}}
	}
	return out, nil
}

func buildPrompt(s *skill.Skill, p profile.Profile, ws []warn.Warning) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Rewrite the following Agent Skill body so it is idiomatic for the %s coding agent.\n", p.Name())
	b.WriteString("Resolve these portability issues; output ONLY the rewritten markdown body, no commentary:\n")
	for _, w := range ws {
		fmt.Fprintf(&b, "- %s\n", w.Message)
	}
	b.WriteString("\n--- BODY ---\n")
	b.WriteString(s.Body)
	return b.String()
}
```

- [ ] **Step 4: Run optimize tests**

Run: `go test ./internal/optimize/ -v`
Expected: PASS (both success and fallback).

- [ ] **Step 5: Commit**

```bash
git add internal/optimize
git commit -m "Add optimize: optional LLM body adaptation with safe fallback"
```

---

## Task 10: CLI wiring (build, check, targets)

**Files:**
- Test: `internal/cli/cli_test.go`
- Create: `internal/cli/cli.go`
- Modify: `cmd/cast/main.go`

- [ ] **Step 1: Write the failing test**

`internal/cli/cli_test.go`:
```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cli/`
Expected: build failure — `undefined: New`.

- [ ] **Step 3: Implement the CLI**

`internal/cli/cli.go`:
```go
// Package cli wires cast's cobra commands.
package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ravistakumar/cast/internal/agent"
	"github.com/ravistakumar/cast/internal/config"
	"github.com/ravistakumar/cast/internal/emit"
	"github.com/ravistakumar/cast/internal/install"
	"github.com/ravistakumar/cast/internal/optimize"
	"github.com/ravistakumar/cast/internal/profile"
	"github.com/ravistakumar/cast/internal/skill"
	"github.com/ravistakumar/cast/internal/warn"
)

// New builds the root cast command.
func New() *cobra.Command {
	root := &cobra.Command{Use: "cast", Short: "Compile one SKILL.md into harness-native skills"}
	root.AddCommand(buildCmd(), checkCmd(), targetsCmd())
	return root
}

func targetsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "targets",
		Short: "List supported target harnesses",
		RunE: func(cmd *cobra.Command, _ []string) error {
			for _, n := range profile.Names() {
				fmt.Fprintln(cmd.OutOrStdout(), n)
			}
			return nil
		},
	}
}

func checkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check <skill-dir>",
		Short: "Validate a canonical SKILL.md against the spec",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := skill.Parse(args[0]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "ok")
			return nil
		},
	}
}

func buildCmd() *cobra.Command {
	var targets []string
	var outdir string
	var doInstall, doOptimize bool
	var agentName string

	cmd := &cobra.Command{
		Use:   "build <skill-dir>",
		Short: "Compile a skill to harness-native output",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if len(targets) == 0 {
				targets = cfg.Targets
			}
			if outdir == "" {
				outdir = cfg.Outdir
			}
			if !cmd.Flags().Changed("optimize") {
				doOptimize = cfg.Optimize
			}
			if agentName == "" {
				agentName = cfg.Agent
			}

			s, err := skill.Parse(args[0]) // strict gate
			if err != nil {
				return err
			}

			var allWarns []warn.Warning
			for _, t := range targets {
				p, ok := profile.Get(t)
				if !ok {
					return fmt.Errorf("unknown target %q (see `cast targets`)", t)
				}
				files, warns := emit.Emit(s, p)

				if doOptimize && len(warns) > 0 {
					files, warns = applyOptimize(s, p, files, warns, agentName)
				}

				root := filepath.Join(outdir, t)
				if err := install.Install(files, root); err != nil {
					return err
				}
				if doInstall {
					live, err := install.DefaultRoot(t)
					if err != nil {
						allWarns = append(allWarns, warn.Warning{Harness: t, Severity: warn.Warn, Code: "install-skipped", Message: err.Error()})
					} else if err := install.Install(files, live); err != nil {
						return err
					}
				}
				allWarns = append(allWarns, warns...)
			}

			if report := warn.Render(allWarns); report != "" {
				fmt.Fprint(cmd.ErrOrStderr(), report)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "cast: compiled %s to %d target(s)\n", s.Name, len(targets))
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&targets, "target", nil, "target harness(es); default from config")
	cmd.Flags().StringVar(&outdir, "outdir", "", "output directory (default from config)")
	cmd.Flags().BoolVar(&doInstall, "install", false, "also install into live harness dirs")
	cmd.Flags().BoolVar(&doOptimize, "optimize", false, "run the optional LLM adaptation pass")
	cmd.Flags().StringVar(&agentName, "agent", "", "agent CLI for --optimize: auto|claude|codex")
	return cmd
}

// applyOptimize rewrites the body via the agent CLI and re-emits the SKILL.md.
func applyOptimize(s *skill.Skill, p profile.Profile, files []profile.File, warns []warn.Warning, agentName string) ([]profile.File, []warn.Warning) {
	name := agentName
	if name == "" || name == "auto" {
		detected, err := agent.Detect()
		if err != nil {
			return files, append(warns, warn.Warning{Harness: p.Name(), Severity: warn.Warn, Code: "optimize-failed", Message: err.Error()})
		}
		name = detected
	}
	r, err := agent.NewRunner(name)
	if err != nil {
		return files, append(warns, warn.Warning{Harness: p.Name(), Severity: warn.Warn, Code: "optimize-failed", Message: err.Error()})
	}
	newBody, optWarns := optimize.Optimize(s, p, warns, r)
	adapted := *s
	adapted.Body = newBody
	files, _ = emit.Emit(&adapted, p) // re-emit with adapted body; structural warns already known
	return files, optWarns
}
```

- [ ] **Step 4: Wire main to the CLI**

Replace `cmd/cast/main.go`:
```go
package main

import (
	"fmt"
	"os"

	"github.com/ravistakumar/cast/internal/cli"
	"github.com/ravistakumar/cast/internal/version"
)

func main() {
	root := cli.New()
	root.Version = version.String()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "cast:", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 5: Run all tests and build**

Run: `go test ./... && go build ./cmd/cast`
Expected: PASS; binary builds.

- [ ] **Step 6: Commit**

```bash
git add internal/cli cmd/cast
git commit -m "Wire cast CLI: build, check, targets commands"
```

---

## Task 11: End-to-end test with a stub agent

**Files:**
- Test: `internal/cli/e2e_test.go`

- [ ] **Step 1: Write the E2E test**

`internal/cli/e2e_test.go`:
```go
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
```

- [ ] **Step 2: Run the E2E test**

Run: `go test ./internal/cli/ -run TestE2E -v`
Expected: PASS.

- [ ] **Step 3: Run the full suite with race detector**

Run: `go test -race ./...`
Expected: PASS (matches fan's CI convention).

- [ ] **Step 4: Commit**

```bash
git add internal/cli/e2e_test.go
git commit -m "Add end-to-end build test across both targets"
```

---

## Final verification

- [ ] Run `go vet ./...` — expect no findings.
- [ ] Run `go test -race ./...` — expect all PASS.
- [ ] Run `go build -o cast ./cmd/cast && ./cast targets` — expect `codex` and `cursor`.
- [ ] Manually compile a sample skill: `./cast build ./internal/skill/testdata/good --target codex --outdir /tmp/cast-out` and inspect `/tmp/cast-out/codex/skills/good/`.

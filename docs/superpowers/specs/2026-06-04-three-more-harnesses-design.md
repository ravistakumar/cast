# cast — Three More Target Harnesses: Design

**Status:** Approved design (pre-implementation)
**Date:** 2026-06-04
**Author:** Ravi S

## Summary

Add three new target harnesses — **Gemini CLI**, **GitHub Copilot**, and
**OpenCode** — each as one new `Profile` in `internal/profile`. This takes
`cast` from 2 to 5 supported targets. No new commands or architecture; this
exercises the existing one-profile-per-harness extension point.

All three support the open Agent Skills `SKILL.md` format with the same
`<dir>/<skill-name>/SKILL.md` layout, require no agent-specific sidecar file
(unlike Codex's `agents/openai.yaml`), and accept `name`+`description`
frontmatter. (Copilot defines optional extra frontmatter fields; we skip them —
they are optional and ignored by other agents. YAGNI.)

Conventions confirmed against official docs: geminicli.com/docs/cli/skills,
code.visualstudio.com/docs/copilot/customization/agent-skills, and
opencode.ai/docs/skills.

## Components

### New profiles

Each mirrors `cursor.go` (value-receiver struct, `Enhancements` returns `nil`):
`internal/profile/gemini.go`, `internal/profile/copilot.go`,
`internal/profile/opencode.go`.

- `Name()`: `"gemini"`, `"copilot"`, `"opencode"`.
- `SkillDir(name)`: `"skills/" + name` (all three use a `skills/` subdir).
- `Frontmatter(s)`: `name: <name>\ndescription: <desc>\n`.
- `Enhancements(s)`: `nil` (no sidecar file).
- `ToolMap()`: see below.

### Tool-name maps

Principle: map **confirmed equivalents** to the harness's tool name; map a tool
to `""` (which makes `emit` warn) **only when confirmed absent**; **omit**
tools whose support is uncertain (no warning — under-claim rather than mis-warn).

`gemini`:
```
Read→read_file, Write→write_file, Edit→replace, Bash→run_shell_command,
Glob→glob, Grep→grep_search, WebFetch→web_fetch, WebSearch→google_web_search,
TodoWrite→write_todos
(omit Task, NotebookEdit — uncertain)
```

`copilot`:
```
Read→read_file, Write→create_file, Edit→replace_string_in_file,
Bash→run_in_terminal, Glob→list_dir, Grep→grep_search, WebFetch→fetch_webpage,
NotebookEdit→edit_notebook_file
(omit WebSearch, Task, TodoWrite — uncertain)
```

`opencode`:
```
Read→read, Write→write, Edit→edit, Bash→bash, Glob→glob, Grep→grep,
WebFetch→webfetch, WebSearch→websearch, TodoWrite→todowrite,
Task→"" , NotebookEdit→""   (both confirmed absent → warn)
```

### Registry

Register all three in `internal/profile/profile.go`'s `registry` map. `Names()`
then returns `["codex","copilot","cursor","gemini","opencode"]` (sorted).

### Install live roots

Add to `internal/install/install.go`'s `liveRoots`:
```
gemini   → ".gemini"            → ~/.gemini/skills/<name>/
copilot  → ".copilot"           → ~/.copilot/skills/<name>/
opencode → ".config/opencode"   → ~/.config/opencode/skills/<name>/
```
All match the documented user-level skill directories.

### Default targets — unchanged

The config default `targets` stays `["codex", "cursor"]`. The three new
harnesses are opt-in via `--target` or config. (Least surprise; smallest default
output.)

### README

Expand the **Supported targets** table from 2 to 5 rows. Each new row: target
name, `skills/<name>/SKILL.md`, "no enhancement file".

## Testing

- `TestRegistry` updated to expect **5** profiles and to resolve each name.
- One focused test per new profile asserting: a representative confirmed tool
  mapping, that `Enhancements` is empty, and that `Frontmatter` contains
  `name:`/`description:`. For `opencode`, also assert `ToolMap()["Task"] == ""`
  (confirmed-absent → warns).
- Existing `emit`, `cli`, and `e2e` tests are unaffected (they pin codex/cursor).

## Verification

- `cast targets` lists all five.
- `cast build examples/changelog --target gemini,copilot,opencode --outdir <tmp>`
  produces `dist/<harness>/skills/changelog/SKILL.md` (+ `scripts/...`) for each,
  with no sidecar files.
- `go vet ./...`, `golangci-lint run`, and `go test -race ./...` all clean.

## Scope

**In:** three profiles, their tool maps, registry + liveRoots entries, per-profile
tests, README table update.

**Out:** project-scope `--install` (still global-only, consistent with
codex/cursor); Copilot-specific optional frontmatter fields; changing the default
target set; any new commands or flags.

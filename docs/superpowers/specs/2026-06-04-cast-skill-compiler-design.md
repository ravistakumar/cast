# cast — Agent Skill Compiler: Design

**Status:** Approved design (pre-implementation)
**Date:** 2026-06-04
**Author:** Ravi S

## Summary

`cast` compiles one canonical Agent Skill (a spec-compliant `SKILL.md`) into
harness-native skills for multiple coding agents. You author once in the open
[agentskills.io](https://agentskills.io/specification) standard — which is
already runnable in Claude Code, so Claude Code is the **source format**, not a
compile target. `cast` then emits ready-to-install skills for the **v1 targets,
Codex and Cursor** — each with the correct directory layout, frontmatter
dialect, tool-name remapping, and harness-specific enhancement files.

It is **structural and deterministic by default**. An **optional pass** rides
your already-authenticated `claude`/`codex` CLI to semantically adapt and tune
each skill per harness — no API keys. Like its siblings
[`prr`](https://github.com/ravistakumar/prr) (refine one prompt) and
[`fan`](https://github.com/ravistakumar/fan) (run many in parallel), `cast` is
agent-agnostic and **never blocks**: it compiles best-effort and warns about
what it cannot confidently translate.

## Motivation

By mid-2026 the Agent Skills `SKILL.md` format is an open, cross-vendor standard
adopted across ~30 agents, and MCP + AGENTS.md sit under the Linux Foundation
Agentic AI Foundation. But **portability is not guaranteed**: the spec itself
annotates `allowed-tools` as *Experimental* with support that "may vary between
agent implementations," and script-language support "depends on the agent
implementation." Tools also extend the core spec (Codex `agents/openai.yaml`,
Claude context-forking) and use different skill directory conventions.

Two gaps motivate `cast`:

1. **Structural divergence.** A skill that parses everywhere still needs
   per-harness directory layout, frontmatter dialect, tool-name remapping, and
   enhancement files. Existing validators/linters
   (`claudelint`, `skill-validator`, `agent-skill-linter`, `SkillPort`) check
   single-standard conformance, security, and quality — **none compile across
   harnesses**. The closest tool, `wshobson/agents`, generates harness-native
   artifacts but only embedded inside a marketplace, not as standalone tooling.
2. **Performance divergence.** The SkCC paper (arXiv 2605.03353) reports a single
   static Markdown skill performs up to **40% differently across agents** from
   formatting alone, with per-harness compilation yielding large pass-rate gains
   (Claude +21–33%, Kimi +35–49%). A pure portability check is a *diagnosis*;
   per-harness adaptation is the *cure*.

`cast` addresses both: deterministic structural compilation, plus an optional
LLM pass for semantic adaptation and SkCC-style tuning.

## Design decisions (locked)

| Decision | Choice | Rationale |
|---|---|---|
| Core transform | Structural core + **optional** LLM pass | Deterministic & testable by default; optional pass rides authenticated CLI (no API keys), true to the prr/fan family |
| Canonical input | Spec-compliant `SKILL.md` as-is | Lowest authoring friction; no new format to learn; one-directional (canonical → targets) |
| v1 targets | Claude Code (source) → **Codex + Cursor** | Smallest credible MVP at high-divergence, high-traffic targets; grows via one-profile-per-harness |
| Output model | Build to `dist/<harness>/` + optional `--install` | Committable artifacts for publishing/CI; install for local convenience |
| Divergence handling | **Warn + best-effort, never block** | Matches prr's "helper, never gatekeeper"; warnings become the LLM pass work-list |
| Engine architecture | **IR + lightweight harness profiles** | Pure, testable functional core; IR feeds the LLM pass; scales via one-profile-per-harness |

## Architecture

```
SKILL.md (+scripts/, references/, assets/)
        │
   ┌────▼─────┐   parse + validate against agentskills.io spec
   │  parse   │   → frontmatter struct, body blocks, asset/script manifest
   └────┬─────┘
   ┌────▼─────┐   normalize into internal representation (IR)
   │normalize │   harness-neutral; single in-memory source of truth
   └────┬─────┘
        │           for each target harness:
   ┌────▼─────┐   walk IR through a Profile
   │  emit    │   (dir layout · frontmatter map · tool-name table ·
   └────┬─────┘    enhancement-file templates) → files + []Warning
        │
   ┌────▼──────────────────────────────┐
   │ dist/codex/…   dist/cursor/…       │  ← committable artifacts
   │ + warnings report                  │
   └────────────────────────────────────┘
        │  (optional, --optimize)
   ┌────▼─────┐  feed IR body + warnings to authenticated claude/codex CLI
   │ llm pass │  → semantic body remap + SkCC-style per-harness tuning
   └──────────┘  (frontmatter/paths/enhancement files stay deterministic)
```

- **Functional core** (`parse → normalize → emit`): pure, deterministic, table-
  tested. No I/O, no subprocess.
- **Imperative shell**: file I/O, the optional CLI subprocess, and `--install`.
- A **Profile** is mostly declarative data (paths, frontmatter field map, tool-
  name table, enhancement-file templates) with a small Go hook for rare special
  cases. Adding a harness mirrors the `internal/agent` one-case pattern in
  `prr`/`fan`.

## CLI surface

```bash
cast build ./my-skill                      # compile all configured targets → ./dist/<harness>/
cast build ./my-skill --target codex       # one target
cast build ./my-skill --install            # also copy into live harness skill dirs
cast build ./my-skill --optimize           # run the optional LLM pass
cast build ./my-skill --optimize --agent codex   # force which CLI drives optimization
cast check ./my-skill                      # validate canonical against the spec; emit no files
cast targets                               # list supported harnesses
```

Config at `~/.config/cast/config.toml` (overridden by `CAST_*` env, then flags):

```toml
targets  = ["codex", "cursor"]   # default targets for `build`
outdir   = "dist"
optimize = false                 # off by default (deterministic)
agent    = "auto"                # which CLI drives --optimize: auto | claude | codex
```

## Components (internal packages)

| Package | Purpose | Layer |
|---|---|---|
| `internal/skill` | Parse `SKILL.md` + assets, validate against spec, build the IR | core (pure) |
| `internal/profile` | Per-harness profiles (`codex`, `cursor`) + registry: dir layout, frontmatter map, tool-name table, enhancement templates | core (data) |
| `internal/emit` | Walk IR through a profile → files + `[]Warning` | core (pure) |
| `internal/warn` | `Warning{Harness, Severity, Location, Code, Message, Suggestion}` + report rendering | core (pure) |
| `internal/optimize` | Optional LLM pass: feed IR + warnings to the agent CLI, splice adapted body back | shell |
| `internal/agent` | Invoke authenticated `claude`/`codex` headless (reuse prr/fan pattern) | shell |
| `internal/install` | Place compiled artifacts into live harness dirs | shell |
| `internal/config` | TOML + env + flags | shell |
| `internal/cli` | cobra command wiring | shell |

`cmd/cast/main.go` is the entry point. Module path `github.com/ravistakumar/cast`.

## Data flow & the warning model

A `Warning` is structured, not free text:

```go
type Warning struct {
    Harness   string // "codex" | "cursor"
    Severity  string // "info" | "warn"
    Location  string // "frontmatter" | "body" | "script:foo.py"
    Code      string // stable machine code, e.g. "tool-no-equivalent"
    Message   string // human-readable
    Suggestion string // optional remediation hint
}
```

Warnings are produced deterministically by `emit` when the tool-name table or
profile rules cannot confidently translate something (e.g. a body reference to
Claude's `Task` tool with no Codex equivalent). They render as a stdout report
and, when `--optimize` is set, become the **exact work-list** handed to the LLM
pass for that target.

## The optional LLM pass (`--optimize`)

Deterministic structural scaffolding is **always** produced first. With
`--optimize`, for each target `cast` hands the agent CLI the IR body + that
target's warnings and asks it to (a) resolve flagged body divergences and
(b) apply SkCC-style per-harness formatting tuning. **Only the body is LLM-
touched**; frontmatter, paths, and enhancement files stay deterministic. If the
CLI errors or returns unusable output, `cast` falls back to the deterministic
output and warns — never blocks.

## Error handling

- **One strict gate:** if the *canonical* `SKILL.md` violates the spec, fail
  fast with precise errors. `cast` will not compile invalid input.
- Everything downstream is **warn-and-continue**: untranslatable body content, a
  missing `--install` target dir, an LLM-pass failure — all warn, never abort.
- **Exit codes:** `0` on success *even with warnings*; non-zero only on hard
  failure (invalid input / I/O error). Matches `fan`'s convention.

## Testing strategy

- **Golden-file tests** for the core: fixture canonical skills → expected
  `dist/<harness>/` trees. Table-driven.
- Unit tests for IR normalization, each profile's emit, and **warning
  generation** (assert specific warnings fire on known divergences — e.g. a
  `Task`-tool reference compiling to Codex).
- **Stub-agent test** for `--optimize`: inject a fake CLI returning canned
  output; assert both the wiring and the fallback path (as `fan` tests with a
  stub agent).
- One **E2E**: compile a sample skill to both targets via the stub.

## Scope

**In v1:** parse + validate, IR, **Codex + Cursor** profiles, `build` to
`dist/`, warnings report, `--target`, `--install`, `--optimize` (claude/codex +
tested fallback), `check`, `targets`, config.

**Deliberately out (later):** more harnesses (Gemini CLI, Copilot, OpenCode);
reverse / any-to-any import; author override syntax; marketplace publishing;
watch mode; a standalone portability score/badge.

## Open items to confirm during implementation

- **Live install paths** for Codex and Cursor skill directories (`--install`)
  are version-sensitive — the research saw Codex referencing `.agents/plugins`
  and `.claude-plugin` dirs. Confirm current paths against each tool's docs
  before hard-coding. The `build`-to-`dist/` core does not depend on this.
- **Codex `agents/openai.yaml`** exact schema for the enhancement-file template.
- **Cursor** skill/rules directory conventions and frontmatter expectations.

## Project conventions (inherited from prr/fan family)

- Go, single static binary, cobra CLI, `BurntSushi/toml` config.
- Agent-agnostic; drives already-authenticated CLIs; no API keys.
- Functional-core / imperative-shell.
- `internal/` split by responsibility; one-case-per-harness extensibility.
- Homebrew tap `ravistakumar/tap`; `go install`; `install.sh`.
- MIT license. GitHub org `ravistakumar`.

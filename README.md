<div align="center">

# cast — Cast one skill to every coding agent

**Write an Agent Skill once. Compile it into harness-native skills for each agent.**
`cast` takes a single spec-compliant `SKILL.md` and emits ready-to-install skills
for Codex and Cursor — correct directory layout, frontmatter dialect, tool-name
remapping, and harness-specific files — with an optional pass that adapts each
skill to its target using your own authenticated agent CLI.

[![CI](https://github.com/ravistakumar/cast/actions/workflows/ci.yml/badge.svg)](https://github.com/ravistakumar/cast/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/ravistakumar/cast)](https://github.com/ravistakumar/cast/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/ravistakumar/cast)](https://goreportcard.com/report/github.com/ravistakumar/cast)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

</div>

## Why cast?

The Agent Skills [`SKILL.md`](https://agentskills.io/specification) format is an
open standard, but a skill that *parses* everywhere does not *run well*
everywhere. Each agent has its own skill directory layout, frontmatter dialect,
tool names, and extension files (Codex's `agents/openai.yaml`, Claude's
experimental `allowed-tools`). `cast` takes the portable core you already wrote
and produces the harness-native form for each target:

- **Author once** — your canonical input is a plain spec-compliant `SKILL.md`.
  No new format to learn.
- **Deterministic by default** — structural compilation (layout, frontmatter,
  tool-name remapping, enhancement files) is pure and reproducible. No model,
  no network.
- **Optional adaptation** — `--optimize` rides your already-authenticated
  `claude` / `codex` CLI to adapt each skill's body to its target. No API keys.
- **Never blocks** — `cast` compiles best-effort and warns about what it can't
  cleanly translate; the only hard error is an invalid canonical `SKILL.md`.

`cast` is agent-agnostic and is the third in a family of single-purpose tools:
[`prr`](https://github.com/ravistakumar/prr) refines one prompt before handoff,
[`fan`](https://github.com/ravistakumar/fan) runs many prompts across agents in
parallel, and `cast` compiles one skill for every agent.

## Install

```bash
# Homebrew (macOS/Linux)
brew install ravistakumar/tap/cast

# Go
go install github.com/ravistakumar/cast/cmd/cast@latest

# Script
curl -fsSL https://raw.githubusercontent.com/ravistakumar/cast/main/install.sh | sh
```

## Usage

```bash
cast build ./my-skill                      # compile to ./dist/<harness>/ for all configured targets
cast build ./my-skill --target codex       # one target
cast build ./my-skill --install            # also copy into live harness skill dirs
cast build ./my-skill --optimize           # run the optional adaptation pass
cast build ./my-skill --optimize --agent codex   # force which CLI drives optimization
cast check ./my-skill                      # validate the canonical SKILL.md; emit nothing
cast targets                               # list supported harnesses
```

A canonical skill is just a directory with a `SKILL.md` (and optional `scripts/`,
`references/`, `assets/`). `cast build` writes a per-harness tree under `dist/`:

```
dist/
  codex/skills/my-skill/SKILL.md
  codex/skills/my-skill/agents/openai.yaml
  codex/skills/my-skill/scripts/...
  cursor/skills/my-skill/SKILL.md
  cursor/skills/my-skill/scripts/...
```

The committed `dist/` artifacts are what others install; `--install` is a
convenience that also drops them into your live harness directories.

## How it works

```
SKILL.md (+ scripts/ references/ assets/)
  → parse + validate against the agentskills.io spec   (strict: invalid input is the only hard error)
  → normalize into a harness-neutral representation
  → for each target: emit through a Profile
        (dir layout · frontmatter dialect · tool-name remap · enhancement files)
        → files + structured warnings
  → (optional) --optimize: hand the body + warnings to your authenticated
        claude/codex CLI to adapt it; fall back to deterministic output on any failure
```

When the body references a tool with no equivalent in a target (for example
Claude's `Task` tool in Codex), `cast` emits a warning rather than guessing.
Those warnings are exactly the work-list the `--optimize` pass resolves.

## Supported targets

The canonical source is a spec-compliant `SKILL.md` (already runnable in Claude
Code). `cast` compiles it to:

| Target | Output | Notes |
| --- | --- | --- |
| Codex  | `skills/<name>/SKILL.md` + `agents/openai.yaml` | OpenAI Codex CLI extension file |
| Cursor | `skills/<name>/SKILL.md`                        | Cursor skill directory |

To add a target, implement one `Profile` in `internal/profile` — directory
layout, frontmatter, a tool-name table, and any enhancement files.

> **Note:** `--install` places skills in the documented user-level locations —
> `~/.codex/skills/<name>/` and `~/.cursor/skills/<name>/`
> ([Codex](https://developers.openai.com/codex/skills),
> [Cursor](https://cursor.com/docs/context/skills)). These can drift across agent
> releases; the `dist/` output does not depend on them.

## Configuration

`~/.config/cast/config.toml` (overridden by `CAST_*` env vars, then flags):

```toml
targets  = ["codex", "cursor"]   # default targets for `build`
outdir   = "dist"
optimize = false                 # off by default (deterministic)
agent    = "auto"                # which CLI drives --optimize: auto | claude | codex
```

## Contributing

Contributions welcome — see [CONTRIBUTING.md](CONTRIBUTING.md). `cast` is written
in Go with a functional-core / imperative-shell design that keeps logic easy to
test.

## License

[MIT](LICENSE)

# cast — Social Preview Image: Design

**Status:** Approved design (pre-implementation)
**Date:** 2026-06-04
**Author:** Ravi S

## Summary

Add a GitHub social-preview (OpenGraph) image for `cast`, matching the `prr`
family convention. Pure content/asset work — no Go code.

## Components

- `docs/demo/social-card.html` — a 1280×640 HTML card (Catppuccin Mocha,
  JetBrains Mono), mirroring `prr`'s `social-card.html`. Content:
  - Wordmark `cast▌` (mauve `#cba6f7` with a green `#a6e3a1` cursor block).
  - Subtitle `ONE SKILL · EVERY AGENT` (uppercase, mauve).
  - Tagline: "Compile one Agent Skill into harness-native skills for every coding agent."
  - Flow line: `one SKILL.md → cast → per-harness skills + warnings`.
  - Footer pills: `Codex`, `Cursor`, `Gemini`, `Copilot`, `OpenCode`.
  - Repo URL `github.com/ravistakumar/cast`.
  - Accent: mauve (chosen over a teal variant for family consistency with `prr`).
- `docs/demo/render-social.sh` — renders the card to `docs/social-preview.png`
  via headless Google Chrome (1280×640), mirroring `prr`'s script.
- `docs/social-preview.png` — the committed rendered image (1280×640 PNG).

## Verification

- `docs/social-preview.png` exists, is a 1280×640 PNG, and renders the wordmark,
  tagline, flow line, all five target pills, and the repo URL legibly.
- `sh docs/demo/render-social.sh` reproducibly regenerates it.

## Manual step (out of automation)

GitHub's repository social-preview image is set via **Settings → Social preview →
Upload an image** in the web UI. There is no public API/`gh` command for it, so
after the PNG is committed, it must be uploaded once through that UI. The
committed file is the source of truth.

## Scope

**In:** the card HTML, the render script, and the committed PNG.

**Out:** changing the README (the demo GIF already heads it); animated/social
variants; CI rendering.

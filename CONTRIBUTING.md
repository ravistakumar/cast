# Contributing to cast

Thanks for your interest! `cast` is written in Go and uses standard tooling.

## Development

```bash
make build   # build the binary
make test    # run all tests
make lint    # golangci-lint
```

## Conventions

- **Functional core, imperative shell**: keep logic pure and unit-tested in
  `skill`, `profile`, `emit`, and `warn`; keep I/O and side effects in `config`,
  `install`, `agent`, `optimize`, and `cli`.
- **New target harnesses**: add one `Profile` in `internal/profile` — directory
  layout, frontmatter dialect, a tool-name remap table, and any harness-specific
  enhancement files. Register it in the profile registry; `emit` and the CLI
  pick it up generically.
- **TDD**: write the failing test first. Every PR runs `go test ./...` and lint.
- **No API keys in tests**: the optional optimize pass is tested with a stub
  `agent.Runner`, never a real CLI. No network, no agent cost.

## Submitting changes

1. Fork the repository.
2. Create a branch: `git checkout -b my-change`.
3. Make your changes and add tests.
4. Run `make test` and `make lint` locally.
5. Open a pull request with a clear description of the change and why.

All contributions are subject to the [Code of Conduct](CODE_OF_CONDUCT.md).

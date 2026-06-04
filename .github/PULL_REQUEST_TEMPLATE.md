## Summary

## Changes

## Testing
- [ ] `go test -race ./...` passes
- [ ] `go vet ./...` clean
- [ ] `golangci-lint run` clean

## Checklist
- [ ] Follows the functional-core / imperative-shell layout
- [ ] New behavior is covered by tests
- [ ] A new target harness (if any) is one `Profile` in `internal/profile`, registered in the registry, with its frontmatter/tool-map/enhancements documented

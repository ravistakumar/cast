#!/bin/sh
# Build cast and render the demo GIF with VHS (https://github.com/charmbracelet/vhs).
# `cast build` is deterministic and agent-free, so this never calls a real agent.
set -e

cd "$(git rev-parse --show-toplevel)"

TMPBIN=$(mktemp -d)
trap 'rm -rf "$TMPBIN" dist' EXIT

go build -o "$TMPBIN/cast" ./cmd/cast
PATH="$TMPBIN:$PATH" vhs docs/demo.tape

echo "Rendered docs/demo.gif"

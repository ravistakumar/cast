#!/bin/sh
# List the commits made since the most recent git tag (or all commits if the
# repository has no tags yet), one per line as "<short-hash> <subject>".
set -e

last_tag=$(git describe --tags --abbrev=0 2>/dev/null || true)
if [ -n "$last_tag" ]; then
  range="$last_tag..HEAD"
else
  range="HEAD"
fi

git log "$range" --no-merges --pretty=format:'%h %s'

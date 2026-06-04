---
name: changelog
description: Generate grouped release notes from git commit history since the last tag.
---

# Changelog

Generate a clean, grouped changelog (release notes) from the commits made since
the most recent git tag.

## Steps

1. Run `scripts/collect-commits.sh` to list the commits since the last tag.
2. Use Grep to scan the commit subjects for conventional-commit prefixes
   (`feat:`, `fix:`, `deps:`, `docs:`) and group them into sections.
3. For any entry that needs more context, use Bash to read the full message with
   `git show --no-patch <hash>`.
4. Write the result as Markdown under a `## <version>` heading, in this order:
   **Features**, **Fixes**, then everything else. Keep entries terse and
   user-facing, and drop noise like merge commits.

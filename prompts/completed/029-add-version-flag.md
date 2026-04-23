---
status: executed
container: code-reviewer-029-add-version-flag
dark-factory-version: v0.17.9
created: "2026-03-05T21:53:46Z"
queued: "2026-03-05T21:53:46Z"
started: "2026-03-05T21:53:46Z"
completed: "2026-03-05T22:00:46Z"
---
<objective>
Add a --version flag that prints the build-time version and exits.
</objective>

<context>
Read CLAUDE.md for project conventions.
Read main.go for the current flag parsing.
Read pkg/version/version.go for the existing Version variable.
</context>

<requirements>
1. Add `--version` flag to main.go.

2. When `--version` is passed, print `code-reviewer <version>` to stdout and exit 0.

3. Check `--version` before any other logic (before arg validation, config loading, etc.).

4. Update usage string to include `--version`.
</requirements>

<constraints>
- Use the existing `version.Version` variable (set at build time via ldflags)
- No other behavior changes
</constraints>

<verification>
Run `make precommit` — must pass.
</verification>

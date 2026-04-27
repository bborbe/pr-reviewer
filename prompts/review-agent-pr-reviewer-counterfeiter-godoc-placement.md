---
status: draft
created: "2026-04-28T12:00:00Z"
---

<summary>
- Seven interfaces have their `//counterfeiter:generate` directive placed inside the GoDoc comment block rather than above it
- Go treats consecutive `//`-prefixed lines as a single documentation block, so the directive text appears verbatim in `go doc` output
- The fix is purely mechanical: move each directive to the line immediately before the GoDoc comment and add a blank line between the directive and the comment
- The directive must NOT be separated from its interface by any non-comment lines or the counterfeiter tool will not associate it with the correct type
- No functional change to code generation or runtime behavior — only documentation correctness
</summary>

<objective>
Move all `//counterfeiter:generate` directives out of GoDoc comment blocks so they appear above the GoDoc comment with a blank line separating them. This prevents the directive text from leaking into `go doc` output. After this fix `go doc` shows clean documentation for all seven affected interfaces.
</objective>

<context>
Read `CLAUDE.md` for project conventions.

Files to read before making changes (read ALL first):
- `agent/pr-reviewer/pkg/review/review.go` — `Reviewer` interface (~line 16)
- `agent/pr-reviewer/pkg/git/git.go` — `WorktreeManager` interface (~line 19)
- `agent/pr-reviewer/pkg/config/config.go` — `Loader` interface (~line 116)
- `agent/pr-reviewer/pkg/github/client.go` — `Client` interface (~line 25)
- `agent/pr-reviewer/pkg/bitbucket/client.go` — `Client` interface (~line 23)
- `agent/pr-reviewer/pkg/plugins/plugins.go` — `Commander` interface (~line 25) and `Installer` interface (~line 32)
</context>

<requirements>
For each interface, apply this transformation:

**Before (directive inside GoDoc — WRONG):**
```go
// SomeInterface does something useful.
//
//counterfeiter:generate -o ../../mocks/some.go --fake-name Some . SomeInterface
type SomeInterface interface { ... }
```

**After (directive above GoDoc — CORRECT):**
```go
//counterfeiter:generate -o ../../mocks/some.go --fake-name Some . SomeInterface

// SomeInterface does something useful.
type SomeInterface interface { ... }
```

The blank line between the directive and the GoDoc comment is required so Go's doc parser treats them as separate blocks.

1. **`pkg/review/review.go`** — Move `//counterfeiter:generate` for `Reviewer` above the GoDoc comment. Add blank line between directive and GoDoc.

2. **`pkg/git/git.go`** — Move `//counterfeiter:generate` for `WorktreeManager` above the GoDoc comment. Add blank line.

3. **`pkg/config/config.go`** — Move `//counterfeiter:generate` for `Loader` above the GoDoc comment. Add blank line.

4. **`pkg/github/client.go`** — Move `//counterfeiter:generate` for `Client` above the GoDoc comment. Add blank line.

5. **`pkg/bitbucket/client.go`** — Move `//counterfeiter:generate` for `Client` above the GoDoc comment. Add blank line.

6. **`pkg/plugins/plugins.go`** — Move `//counterfeiter:generate` directives for both `Commander` and `Installer` above their respective GoDoc comments. Add blank line after each directive.

7. **After editing, run `go generate ./...`** from `agent/pr-reviewer/` to confirm counterfeiter still picks up the directives and regenerates mocks identically:
   ```bash
   cd agent/pr-reviewer && go generate ./...
   ```
   If generated mocks change (other than header comments), investigate before proceeding.

8. **Run `cd agent/pr-reviewer && make test`** — must pass.
</requirements>

<constraints>
- Only change files in `agent/pr-reviewer/`
- Do NOT commit — dark-factory handles git
- Do NOT change any interface definition, method signatures, or mock output paths
- The blank line between the directive and GoDoc comment is mandatory — without it the directive still falls inside the doc block
- Generated mock files must remain byte-for-byte identical after `go generate` (or differ only in tool-version comments)
- Existing tests must still pass
</constraints>

<verification>
cd agent/pr-reviewer && make precommit
</verification>

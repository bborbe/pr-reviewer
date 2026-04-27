---
status: draft
created: "2026-04-28T12:00:00Z"
---

<summary>
- Four more test suite files use the two-argument `RunSpecs` form that silently ignores timeout and reporter configuration
- The CLI test suite also lacks `time.Local = time.UTC` and `format.TruncatedDiff = false` initialization that all other suites set
- A test variable that holds a Counterfeiter mock uses a `fake` prefix in its name, which the project convention forbids
- This is the second of two sibling prompts (apply after the first ginkgo-suite-setup prompt to avoid conflicts)
</summary>

<objective>
Update four more Ginkgo suite files to use `GinkgoConfiguration()` with a 60-second timeout and the four-argument `RunSpecs` form, fix the `cmd/cli/main_test.go` initialization gap, and rename the `fakeCommander` variable in `plugins_test.go` to remove the `fake` prefix. After this fix all suite files in the service use the same consistent pattern.
</objective>

<context>
Read `CLAUDE.md` for project conventions.

Files to read before making changes (read ALL first):
- `agent/pr-reviewer/main_test.go` — reference pattern (4-arg RunSpecs, GinkgoConfiguration, suiteConfig.Timeout, time.Local, format.TruncatedDiff)
- `agent/pr-reviewer/pkg/bitbucket/client_suite_test.go` — needs update
- `agent/pr-reviewer/pkg/git/git_suite_test.go` — needs update
- `agent/pr-reviewer/pkg/steps/steps_suite_test.go` — needs update (also missing `//go:generate` directive)
- `agent/pr-reviewer/cmd/cli/main_test.go` — needs GinkgoConfiguration + initialization fixes
- `agent/pr-reviewer/pkg/plugins/plugins_test.go` — `fakeCommander` rename (~line 61)
</context>

<requirements>
For each of the four suite files, apply the same transformation as in the first ginkgo-suite-setup prompt:

```go
//go:generate go run -mod=mod github.com/maxbrunsfeld/counterfeiter/v6 -generate
func TestSuite(t *testing.T) {
    time.Local = time.UTC
    format.TruncatedDiff = false
    RegisterFailHandler(Fail)
    suiteConfig, reporterConfig := GinkgoConfiguration()
    suiteConfig.Timeout = 60 * time.Second
    RunSpecs(t, "<Suite Name>", suiteConfig, reporterConfig)
}
```

1. **`pkg/bitbucket/client_suite_test.go`** — Apply transformation. Keep existing suite name. Add `//go:generate` if missing.

2. **`pkg/git/git_suite_test.go`** — Apply transformation. Keep existing suite name. Add `//go:generate` if missing.

3. **`pkg/steps/steps_suite_test.go`** — Apply transformation. Keep existing suite name. Add `//go:generate` directive (currently missing from this file).

4. **`cmd/cli/main_test.go`** — This file does NOT have a standard `TestSuite` function with the project pattern. Read the file carefully. Apply:
   - Add `time.Local = time.UTC` and `format.TruncatedDiff = false` to the `TestSuite` function if missing.
   - Add `suiteConfig, reporterConfig := GinkgoConfiguration()` and `suiteConfig.Timeout = 60 * time.Second`.
   - Change `RunSpecs(t, "...")` to `RunSpecs(t, "...", suiteConfig, reporterConfig)`.
   - Add `//go:generate` directive if missing.
   - Ensure `"time"` and `"github.com/onsi/gomega/format"` are imported.

5. **`pkg/plugins/plugins_test.go` — Rename `fakeCommander`** (~line 61):
   - Find `var fakeCommander *mocks.Commander` and rename to `commander *mocks.Commander` (or another name that avoids collision with existing `commander plugins.Commander` if that variable is in scope — check the `Describe` block carefully).
   - If a collision exists with `var commander plugins.Commander`, use `mockCommander *mocks.Commander` (also acceptable — the key constraint is NO `fake` or `mock` prefix in the project style; but `mockCommander` is clearer than a bare `commander` when there's a collision).
   - Update all usages of `fakeCommander` in the test file to the new name.

6. **Run `cd agent/pr-reviewer && make test`** — must pass.
</requirements>

<constraints>
- Only change files in `agent/pr-reviewer/`
- Do NOT commit — dark-factory handles git
- Existing tests must still pass — this is a non-functional change except for the variable rename
- Apply this prompt AFTER the first ginkgo-suite-setup prompt to avoid merge conflicts (both prompts only touch different files, so they are actually independent — but conceptually this is part 2 of that work)
- Do NOT alter the suite name strings or test logic
- The `//go:generate` directive must be placed on the line immediately above `func TestSuite`
</constraints>

<verification>
cd agent/pr-reviewer && make precommit
</verification>

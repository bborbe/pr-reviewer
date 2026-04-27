---
status: draft
created: "2026-04-28T12:00:00Z"
---

<summary>
- Five test suite files use the two-argument `RunSpecs(t, "name")` form, which silently ignores Ginkgo configuration flags including timeout
- Without passing `GinkgoConfiguration()` to `RunSpecs`, tests can hang indefinitely when a subprocess or network call stalls
- The project already has the correct four-argument pattern in three other suite files — this prompt brings the first five packages into alignment
- Each file needs `GinkgoConfiguration()`, a 60-second suite timeout, and the four-argument `RunSpecs` call
- The `//go:generate` directive for counterfeiter is also missing from some of these files and should be added for consistency
</summary>

<objective>
Update five Ginkgo suite files to use `GinkgoConfiguration()` with a 60-second timeout and the four-argument `RunSpecs` form, matching the pattern already established in `main_test.go`, `cmd/run-task/main_test.go`, and `pkg/plugins/plugins_suite_test.go`. After this fix all test suites respect Ginkgo's timeout and reporter configuration flags.
</objective>

<context>
Read `CLAUDE.md` for project conventions.

Files to read before making changes (read ALL first):
- `agent/pr-reviewer/main_test.go` — the reference/correct pattern for suite setup (4-arg RunSpecs, GinkgoConfiguration, suiteConfig.Timeout)
- `agent/pr-reviewer/pkg/config/config_suite_test.go` — needs update
- `agent/pr-reviewer/pkg/verdict/verdict_suite_test.go` — needs update
- `agent/pr-reviewer/pkg/prurl/prurl_suite_test.go` — needs update
- `agent/pr-reviewer/pkg/github/github_suite_test.go` — needs update
- `agent/pr-reviewer/pkg/review/review_suite_test.go` — needs update
</context>

<requirements>
For each of the five suite files listed in the context, apply this transformation:

**Before (current 2-arg form):**
```go
func TestSuite(t *testing.T) {
    time.Local = time.UTC
    format.TruncatedDiff = false
    RegisterFailHandler(Fail)
    RunSpecs(t, "<Suite Name>")
}
```

**After (correct 4-arg form):**
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

1. **`pkg/config/config_suite_test.go`** — Apply the transformation. Keep the existing suite name string unchanged. Add `//go:generate` directive if missing.

2. **`pkg/verdict/verdict_suite_test.go`** — Apply the transformation. Keep the existing suite name string unchanged. Add `//go:generate` directive if missing.

3. **`pkg/prurl/prurl_suite_test.go`** — Apply the transformation. Keep the existing suite name string unchanged. Add `//go:generate` directive if missing.

4. **`pkg/github/github_suite_test.go`** — Apply the transformation. Keep the existing suite name string unchanged. Add `//go:generate` directive if missing.

5. **`pkg/review/review_suite_test.go`** — Apply the transformation. Keep the existing suite name string unchanged. Add `//go:generate` directive if missing.

6. **Verify imports** in each file: `"time"` must be imported (for `60 * time.Second`). The `GinkgoConfiguration` function is from `github.com/onsi/ginkgo/v2` — it is already imported via the dot import (`. "github.com/onsi/ginkgo/v2"`). No new import needed for `GinkgoConfiguration` itself.

7. **Run `cd agent/pr-reviewer && make test`** — must pass.
</requirements>

<constraints>
- Only change files in `agent/pr-reviewer/`
- Do NOT commit — dark-factory handles git
- Existing tests must still pass — this is a non-functional change
- Do NOT alter the suite name strings or any other existing content
- The `//go:generate` directive must be placed on the line immediately above `func TestSuite`
</constraints>

<verification>
cd agent/pr-reviewer && make precommit
</verification>

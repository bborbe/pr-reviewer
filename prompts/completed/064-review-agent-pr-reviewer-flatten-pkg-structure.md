---
status: completed
summary: Flattened agent/pr-reviewer/pkg/ by collapsing config, prurl, review, verdict, version, steps into flat pkg/ files and merging prompts/{execution,planning,review}/ into a single prompts/ package; regenerated mocks and all tests pass with make precommit exit 0.
container: code-reviewer-064-review-agent-pr-reviewer-flatten-pkg-structure
dark-factory-version: v0.135.19-1-gc08c946
created: "2026-04-28T20:00:00Z"
queued: "2026-04-28T19:58:35Z"
started: "2026-04-28T19:58:37Z"
completed: "2026-04-28T20:20:43Z"
---

<summary>
- The pr-reviewer module has 11 subpackages plus 3 nested under `prompts/`, most with only a single production caller
- The project's package-extraction rule says: default to one flat `pkg/` and only extract a subpackage with ≥2 distinct external callers, a clear subdomain, or cross-repo reuse
- Several packages fail this rule: `config`, `prurl`, `review`, `version` (single CLI consumer); `steps`, `prompts/execution`, `prompts/planning`, `prompts/review` (factory-only consumer — exact symptom of premature extraction)
- The 3-level nested `pkg/prompts/<phase>/` layout is also a "don't nest unnecessarily" violation
- External-API wrappers (`bitbucket`, `git`, `github`) and the wiring layer (`factory`) keep their own packages — they earn the boundary
- After this fix the module has 5 subpackages (`bitbucket`, `git`, `github`, `factory`, `prompts`) and a flat top-level `pkg/` containing the rest as files
</summary>

<objective>
Flatten `agent/pr-reviewer/pkg/` to comply with the project's package-extraction rules. Collapse single-consumer subpackages into files in a flat top-level `pkg/`, eliminate the 3-level nested `pkg/prompts/<phase>/` layout into a single `pkg/prompts/` package, and keep only `bitbucket/`, `git/`, `github/`, `factory/`, and `prompts/` as subpackages. `make precommit` passes at the end.
</objective>

<context>
Read `CLAUDE.md` for project conventions.

**Package-extraction rule being enforced** (from the project's coding guide; do NOT try to read host-absolute paths — the rule is summarised here):

- Default: one flat `pkg/` (plus `pkg/handler/`, `pkg/factory/`).
- Extract a subpackage only when one is true: ≥2 distinct external callers consume it (factory wiring counts as ZERO callers); a clear subdomain has formed with a stable external boundary; the code is reusable across repos.
- Symptoms of premature extraction (collapse if any apply): single consumer; local-interface duplication ("minimal" copy of an interface to dodge a cycle); adapter shims that only translate between two of your own packages; constructor sprawl (every extraction adds N parameters to the parent constructor); phantom-cycle workarounds via `SetDep(d)` after construction.
- Earned boundaries explicitly cited by the guide: `pkg/storage/`, `pkg/auth/`, `pkg/cache/`, `pkg/git/`, `pkg/handler/`, `pkg/factory/`, plus external API wrappers.
- The deletion test: "if we deleted this package, would the reorganisation be ambiguous?" If the answer is "obvious — it just becomes files in `pkg/`," it didn't earn its boundary.

Files to read before making changes (read ALL first):
- `agent/pr-reviewer/pkg/config/config.go` and `config_test.go`
- `agent/pr-reviewer/pkg/prurl/prurl.go` and `prurl_test.go`
- `agent/pr-reviewer/pkg/review/review.go` and `review_test.go`
- `agent/pr-reviewer/pkg/verdict/verdict.go` and `verdict_test.go`
- `agent/pr-reviewer/pkg/version/version.go`
- `agent/pr-reviewer/pkg/steps/review.go`, `gh_token.go`, `mocks.go`, `export_test.go`, `steps_suite_test.go`, plus their `*_test.go` files
- `agent/pr-reviewer/pkg/prompts/execution/prompts.go` plus its `workflow.md` and `output-format.md` (`//go:embed` assets)
- `agent/pr-reviewer/pkg/prompts/planning/prompts.go` plus its `workflow.md` and `output-format.md`
- `agent/pr-reviewer/pkg/prompts/review/prompts.go` plus its `workflow.md` and `output-format.md`
- `agent/pr-reviewer/pkg/factory/factory.go` — the only consumer of `steps` and `prompts/*`
- `agent/pr-reviewer/main.go`, `cmd/cli/main.go`, `cmd/run-task/main.go` — all production importers of `pkg/*`
- `agent/pr-reviewer/pkg/github/client.go` — currently imports `pkg/verdict`; check why
- `agent/pr-reviewer/mocks/` — counterfeiter-generated; will need regeneration. Check existing `//counterfeiter:generate` directives in the source interfaces (likely use relative `-o ../../mocks/...` paths that change after the move)
- All `*_suite_test.go` files in the affected packages — there must be exactly ONE `pkg_suite_test.go` after the move
</context>

<requirements>
**Execute steps in this order. Do not run `go generate` or `make precommit` until step 5.**

1. **Move single-consumer packages into flat `pkg/`**. Create `agent/pr-reviewer/pkg/pkg.go` if needed declaring `package pkg`. Move every `.go` file from these subdirs into `agent/pr-reviewer/pkg/`, change `package <name>` to `package pkg`, and rename per the table below. Keep the public API shape stable: callers must still be able to refer to a `Config`, `Verdict`, `PRURL`, etc. (now as `pkg.Config`, `pkg.Verdict`, `pkg.PRURL`, etc.).

   | Source | New file | Notes |
   |---|---|---|
   | `pkg/config/config.go` | `pkg/config.go` | keep exported names |
   | `pkg/prurl/prurl.go` | `pkg/prurl.go` | keep `PRURL` etc. |
   | `pkg/review/review.go` | `pkg/review.go` | keep names; if `Reviewer` clashes with anything in `steps_review.go`, prefix the steps version (`StepReviewer`) since the public Reviewer is the more-used type |
   | `pkg/verdict/verdict.go` | `pkg/verdict.go` | keep `Verdict` etc. |
   | `pkg/version/version.go` | `pkg/version.go` | keep `Version` constant |
   | `pkg/steps/review.go` | `pkg/steps_review.go` | resolve clashes per above |
   | `pkg/steps/gh_token.go` | `pkg/steps_gh_token.go` | keep names |
   | `pkg/steps/mocks.go` | inspect first | if it's only `//go:generate` directive aggregator, move to `pkg/steps_mocks.go`; if hand-written mocks, move into `agent/pr-reviewer/mocks/` instead |

   Move every corresponding `*_test.go` (including `*_suite_test.go` and `export_test.go`) alongside its source. Preserve the original `package` line of each test file (`package <name>` → `package pkg`; `package <name>_test` → `package pkg_test`). After all moves there must be exactly ONE `pkg_suite_test.go` — merge the seven existing suite files into one `RegisterFailHandler(Fail); RunSpecs(t, "pkg suite")` entrypoint.

   Delete the now-empty subdirectories (`pkg/config`, `pkg/prurl`, `pkg/review`, `pkg/verdict`, `pkg/version`, `pkg/steps`).

2. **Collapse `pkg/prompts/{execution,planning,review}/` into a single `pkg/prompts/` package**:
   - Move `pkg/prompts/execution/prompts.go` → `pkg/prompts/execution.go` (`package execution` → `package prompts`); rename `BuildInstructions` → `BuildExecutionInstructions`.
   - Move `pkg/prompts/planning/prompts.go` → `pkg/prompts/planning.go` (`package planning` → `package prompts`); rename `BuildInstructions` → `BuildPlanningInstructions`.
   - Move `pkg/prompts/review/prompts.go` → `pkg/prompts/review.go` (`package review` → `package prompts`); rename `BuildInstructions` → `BuildReviewInstructions`.
   - Move and rename the `//go:embed` assets to disambiguate (they share filenames across the three subdirs):
     - `execution/workflow.md` → `pkg/prompts/execution_workflow.md`
     - `execution/output-format.md` → `pkg/prompts/execution_output-format.md`
     - `planning/workflow.md` → `pkg/prompts/planning_workflow.md`
     - `planning/output-format.md` → `pkg/prompts/planning_output-format.md`
     - `review/workflow.md` → `pkg/prompts/review_workflow.md`
     - `review/output-format.md` → `pkg/prompts/review_output-format.md`
   - Update each `//go:embed <name>.md` directive to point at the renamed file.
   - Delete the now-empty `pkg/prompts/execution/`, `pkg/prompts/planning/`, `pkg/prompts/review/`.
   - Inside each moved file, also rename the package-level vars `workflow` and `outputFormat` to `executionWorkflow` / `executionOutputFormat` etc., since all three files now share a package and would otherwise collide.

3. **Update all importers**. Replace every `github.com/bborbe/code-reviewer/agent/pr-reviewer/pkg/<moved-name>` import (for `config`, `prurl`, `review`, `verdict`, `version`, `steps`) with `github.com/bborbe/code-reviewer/agent/pr-reviewer/pkg`. Update call sites: `config.Config` → `pkg.Config`, `verdict.Verdict` → `pkg.Verdict`, etc. For prompts, replace each `pkg/prompts/<phase>` import with the single `pkg/prompts`, and update `<phase>.BuildInstructions()` to `prompts.Build<Phase>Instructions()`. Files to edit:
   - `agent/pr-reviewer/main.go`
   - `agent/pr-reviewer/cmd/cli/main.go`
   - `agent/pr-reviewer/cmd/run-task/main.go`
   - `agent/pr-reviewer/pkg/factory/factory.go`
   - `agent/pr-reviewer/pkg/github/client.go` (currently imports `pkg/verdict`)
   - any `agent/pr-reviewer/mocks/*.go` that reference the moved packages
   - any remaining file inside `pkg/bitbucket/`, `pkg/git/`, `pkg/github/`, `pkg/factory/`, or `pkg/prompts/` that imports another moved subpackage

4. **Keep these subpackages — DO NOT flatten**:
   - `pkg/bitbucket/` (external API wrapper — earned)
   - `pkg/git/` (external dep wrapper — explicitly cited as earned in the guide)
   - `pkg/github/` (external API wrapper — earned)
   - `pkg/factory/` (wiring layer — explicitly allowed)
   - `pkg/prompts/` (single subdomain package after step 2; consolidates the three previously-nested phase packages)

5. **Regenerate counterfeiter mocks AFTER all moves and import updates**. For every `//counterfeiter:generate` directive on a moved interface:
   - Update the `-o` relative path: a directive previously at depth 3 (`pkg/<x>/<file>.go`) at `-o ../../mocks/...` is now at depth 2 (`pkg/<file>.go`), so the path becomes `-o ../mocks/...`.
   - Update the `--fake-name` and target if the package name in the directive changed (e.g., a directive using the explicit package qualifier).
   - Run `cd agent/pr-reviewer && go generate ./...`.
   - The set of generated files in `agent/pr-reviewer/mocks/` must be the same before and after (same filenames, same fake names) — if a file was renamed by `go generate`, you missed updating the directive.

6. **Verify the package-boundary test** for each remaining subpackage (`bitbucket`, `git`, `github`, `factory`, `prompts`): the deletion-test answer must be "ambiguous — multiple files would need re-homing." If any of the five would obviously become a single file in `pkg/`, log the concern in the prompt log and stop — that means the analysis was wrong.

7. **`make precommit` must pass** at `agent/pr-reviewer/`. All existing tests must still pass; test count (number of `It(...)` and `DescribeTable` entries) must be unchanged before and after the move — verify by counting pre-move and post-move.
</requirements>

<constraints>
- Only change files in `agent/pr-reviewer/`
- Do NOT commit — dark-factory handles git
- Do NOT delete or weaken existing tests; if a test moves, keep all its cases intact (test count preserved exactly)
- Do NOT introduce new dependencies
- If two moved files have a top-level identifier collision in `package pkg`, rename the less-public one (keep the originally-exported API stable for external callers — `pkg.Verdict`, `pkg.Config`, `pkg.PRURL`, `pkg.Reviewer`, `pkg.Version` must still exist with the same shape)
- Preserve all `//go:embed` assets — every `.md` file under the old `pkg/prompts/<phase>/` must end up referenced by a working `//go:embed` directive
- The `cmd/cli/main.go` ↔ `cmd/run-task/main.go` divergence is OUT OF SCOPE — do not refactor wiring; only update imports
- The `pkg/github` → `pkg/verdict` cross-coupling is also out of scope; after the move it becomes `pkg/github` → `pkg`, which is fine
</constraints>

<verification>
cd agent/pr-reviewer && make precommit
</verification>

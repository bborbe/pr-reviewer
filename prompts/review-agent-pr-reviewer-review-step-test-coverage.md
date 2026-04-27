---
status: draft
created: "2026-04-28T12:00:00Z"
---

<summary>
- The custom agent step that runs the AI review has zero test coverage for its three exported methods
- The `ShouldRun` method can be tested today without any mock — it only reads the markdown content
- The `Run` method requires an injectable Claude runner, which must be verified in the `bborbe/agent/lib` module or generated before tests can be written
- The `Name` method is trivially testable (pure string return)
- Covering these methods closes the most significant test gap in the service: the step that runs the AI review is currently completely unobservable in tests
</summary>

<objective>
Add Ginkgo/Gomega tests for `reviewStep.Name()`, `reviewStep.ShouldRun()`, and `reviewStep.Run()` in `pkg/steps/`. The `ShouldRun` and `Name` tests require no mocks. The `Run` tests require a Counterfeiter mock for `claudelib.ClaudeRunner` — generate it if it does not already exist. After this fix the core step of the review pipeline has ≥80% statement coverage.
</objective>

<context>
Read `CLAUDE.md` for project conventions.

Files to read before making changes (read ALL first):
- `agent/pr-reviewer/pkg/steps/review.go` — `reviewStep` struct (~line 35), `NewReviewStep` (~line 41), `Name` (~line 50), `ShouldRun` (~line 55), `Run` (~line 65); `extractVerdict` and `lastJSONBlock` helpers
- `agent/pr-reviewer/pkg/steps/review_test.go` — existing test coverage (verify what is already tested)
- `agent/pr-reviewer/pkg/steps/export_test.go` — exported symbols for white-box testing; check if `ClaudeRunner` is already exported for tests
- `agent/pr-reviewer/pkg/steps/steps_suite_test.go` — suite setup (counterfeiter generate directive)
- `agent/pr-reviewer/mocks/` — check if a `ClaudeRunner` mock already exists
</context>

<requirements>
1. **Check if `claudelib.ClaudeRunner` mock already exists**:
   ```bash
   ls agent/pr-reviewer/mocks/ | grep -i runner
   grep -rn "ClaudeRunner" agent/pr-reviewer/mocks/ 2>/dev/null | head
   ```
   Also check the `bborbe/agent/lib` module for an existing fake:
   ```bash
   find $(go env GOPATH)/pkg/mod/github.com/bborbe/agent/lib@*/ -name "*runner*" -o -name "*claude*" 2>/dev/null | grep -i fake
   ```

2. **If no mock exists, add a `//counterfeiter:generate` directive** for `claudelib.ClaudeRunner` to `export_test.go` or a new file in `pkg/steps/`:
   ```go
   //counterfeiter:generate -o ../../mocks/claude-runner.go --fake-name ClaudeRunner github.com/bborbe/agent/lib/claude ClaudeRunner
   ```
   First grep-verify the exact interface name and import path:
   ```bash
   grep -rn "type ClaudeRunner interface\|type Runner interface" \
     $(go env GOPATH)/pkg/mod/github.com/bborbe/agent/lib@*/claude/ 2>/dev/null
   ```
   Then run:
   ```bash
   cd agent/pr-reviewer && go generate ./pkg/steps/...
   ```

3. **Add test cases to `pkg/steps/review_test.go`** (external package `steps_test`, Ginkgo/Gomega):

   **`Name()` test:**
   ```go
   Describe("Name", func() {
       It("returns the step name", func() {
           step := steps.NewReviewStep(/* inject deps */)
           Expect(step.Name()).To(Equal("pr-ai-review"))
       })
   })
   ```

   **`ShouldRun()` tests — no mock needed (reads markdown only):**
   ```go
   Describe("ShouldRun", func() {
       DescribeTable("returns false when verdict section already exists",
           func(content string, expected bool) {
               step := steps.NewReviewStep(/* deps */)
               result, err := step.ShouldRun(ctx, content)
               Expect(err).NotTo(HaveOccurred())
               Expect(result).To(Equal(expected))
           },
           Entry("no verdict section", "# PR Review\n\nsome text", true),
           Entry("verdict section present", "# PR Review\n\n## Verdict\n\npass", false),
           Entry("empty content", "", true),
       )
   })
   ```

   **`Run()` tests — require `ClaudeRunner` mock:**
   - **Claude runner returns error** → `Run` should return `agentlib.AgentStatusFailed` result, NOT propagate error
   - **Claude runner returns unparseable output** → `Run` returns `AgentStatusDone` with `NextPhase = "human_review"`
   - **Claude runner returns `verdict: pass`** → `Run` returns `AgentStatusDone` with `NextPhase = "done"` (or appropriate terminal phase)
   - **Claude runner returns `verdict: fail`** → `Run` returns `AgentStatusDone` with `NextPhase = "human_review"`
   - Use the Counterfeiter mock for `ClaudeRunner` — call `.RunReturns(...)` or `.RunStub = ...` per the counterfeiter API

4. **Grep-verify `AgentStatusFailed`, `AgentStatusDone`, and `NextPhase` field names** from the agent lib before writing assertions:
   ```bash
   grep -rn "AgentStatusFailed\|AgentStatusDone\|type Result struct\|NextPhase" \
     $(go env GOPATH)/pkg/mod/github.com/bborbe/agent/lib@*/... 2>/dev/null | head -20
   ```

5. **Check coverage after adding tests**:
   ```bash
   cd agent/pr-reviewer && go test -coverprofile=/tmp/cover.out ./pkg/steps/... \
     && go tool cover -func=/tmp/cover.out | grep steps
   ```
   Coverage for `pkg/steps` should reach ≥80%.

6. **Run `cd agent/pr-reviewer && make test`** — must pass.
</requirements>

<constraints>
- Only change files in `agent/pr-reviewer/`
- Do NOT commit — dark-factory handles git
- Existing tests must still pass
- Use Ginkgo v2 / Gomega conventions — `DescribeTable`, `Entry`, `BeforeEach`, `context.Background()` in test setup
- Counterfeiter mocks generated via `go generate`, not hand-written
- Grep-verify all `agentlib.*` symbols before writing struct literals or calling methods
- If `claudelib.ClaudeRunner` interface does not exist (grep finds nothing), report `status: failed` with a note about the missing interface — do NOT invent the interface
</constraints>

<verification>
cd agent/pr-reviewer && make precommit
</verification>

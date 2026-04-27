---
status: draft
created: "2026-04-28T12:00:00Z"
---

<summary>
- A deprecated URL parser file has no callers in production code and duplicates functionality already provided by the canonical PR URL package
- A configuration method for looking up a repository path is dead — it is only referenced in its own test, not by any production code path
- Both items add maintenance surface without providing value: future contributors may update one path and miss the other
- Removing them reduces the codebase to a single authoritative implementation for each concern
</summary>

<objective>
Delete `pkg/github/url.go` (deprecated, no production callers) and its companion test file, and remove the `FindRepoPath` method from `pkg/config/config.go` along with its test cases. After this cleanup there is one canonical PR URL parser (`pkg/prurl`) and one canonical repo-lookup method (`Config.FindRepo`).
</objective>

<context>
Read `CLAUDE.md` for project conventions.

Files to read before making changes (read ALL first):
- `agent/pr-reviewer/pkg/github/url.go` — deprecated `ParsePRURL` function and its `PRInfo` type; confirm the `Deprecated` doc comment states "Use prurl.Parse instead"
- `agent/pr-reviewer/pkg/github/url_test.go` — tests for the deprecated function (to be deleted)
- `agent/pr-reviewer/pkg/config/config.go` — `FindRepoPath` method (~line 190) and its doc comment; also confirm `FindRepo` (~line 205) is the canonical replacement
- `agent/pr-reviewer/pkg/config/config_test.go` — any test cases for `FindRepoPath` (to be deleted)
- `agent/pr-reviewer/pkg/github/client.go` — confirm no production file imports `ParsePRURL` from `url.go`
</context>

<requirements>
1. **Confirm zero production callers** before deleting anything:
   ```bash
   grep -rn "ParsePRURL\|github\.PRInfo" agent/pr-reviewer/ --include="*.go" \
     | grep -v "_test.go" | grep -v "url.go"
   ```
   If any match is found (outside test files and the deprecated file itself), STOP and report `status: failed` — the caller must be migrated to `prurl.Parse` first.

2. **Delete `agent/pr-reviewer/pkg/github/url.go`** and **`agent/pr-reviewer/pkg/github/url_test.go`**. Use `git rm` or plain `os.Remove` equivalent in the shell:
   ```bash
   cd agent/pr-reviewer && rm pkg/github/url.go pkg/github/url_test.go
   ```

3. **Confirm zero production callers of `FindRepoPath`** before removing it:
   ```bash
   grep -rn "FindRepoPath" agent/pr-reviewer/ --include="*.go" \
     | grep -v "_test.go" | grep -v "config.go"
   ```
   If any match, STOP and report `status: failed`.

4. **Remove `FindRepoPath` from `pkg/config/config.go`** (~line 190): delete the entire method body.

5. **Remove the `FindRepoPath` test cases** from `pkg/config/config_test.go`: delete any `It`/`Entry`/`Describe` block that tests `FindRepoPath`. Keep all `FindRepo` tests intact.

6. **Run `cd agent/pr-reviewer && make test`** — must pass. The compilation error on `url_test.go` (referencing the now-deleted `ParsePRURL`) will disappear once the test file is deleted. Ensure `pkg/github` still compiles cleanly (the `Client` interface and implementation remain).

7. **Verify** the `pkg/github` package still has the test suite (`github_suite_test.go`) and that `client_test.go` tests still pass after removing `url_test.go`.
</requirements>

<constraints>
- Only change files in `agent/pr-reviewer/`
- Do NOT commit — dark-factory handles git
- Existing tests for `pkg/github/client.go`, `pkg/config/config.go`, and `pkg/prurl` must still pass
- If any production caller is found for `ParsePRURL` or `FindRepoPath`, STOP and report `status: failed` — do not silently migrate callers in this prompt
- Do NOT touch `pkg/prurl/prurl.go` — this prompt only removes dead code, not migrates callers
</constraints>

<verification>
cd agent/pr-reviewer && make precommit
</verification>

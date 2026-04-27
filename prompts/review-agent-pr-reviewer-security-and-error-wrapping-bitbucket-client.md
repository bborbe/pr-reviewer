---
status: draft
created: "2026-04-28T12:00:00Z"
---

<summary>
- The Bitbucket HTTP client has no timeout, allowing a slow or adversarial server to hold connections open indefinitely and exhaust goroutines
- The URL builder accepts the `http://` scheme, which would send credentials over cleartext HTTP
- All error constructors in the Bitbucket client use stdlib `fmt.Errorf` instead of the project's context-aware error wrapping
- Three methods silently drop error context by using bare `return err` after status-check helpers
- Fixing all three concerns in one pass avoids repeated edits to the same file
</summary>

<objective>
Harden the Bitbucket HTTP client against slow-server DoS and cleartext credential transmission, and migrate all error construction to `github.com/bborbe/errors` with proper context propagation. After this fix the client will time out after 30 s, will never transmit credentials over HTTP, and all errors will carry stack traces.
</objective>

<context>
Read `CLAUDE.md` for project conventions.

Files to read before making changes (read ALL first):
- `agent/pr-reviewer/pkg/bitbucket/client.go` — all methods (`GetPRBranches` ~line 50, `PostComment` ~line 125, `Approve` ~line 165, `NeedsWork` ~line 205, `buildURL` ~line 241, `checkResponseStatus` ~line 290, `httpClient` struct ~line 30, `NewBitbucketClient` ~line 38)
- `agent/pr-reviewer/pkg/bitbucket/client_test.go` — existing test coverage and test server pattern
- `agent/pr-reviewer/pkg/steps/gh_token.go` — reference implementation: `http.Client{Timeout: 10 * time.Second}` (~line 61)
</context>

<requirements>
1. **Add `Timeout: 30 * time.Second` to the `http.Client` in the struct literal** (~line 38 in `NewBitbucketClient` or wherever `httpClient` field is initialised). Use the same pattern as `pkg/steps/gh_token.go:61`.

2. **Harden `buildURL`** (~lines 241–246): reject the `http://` scheme outright. Change to:
   ```go
   func (c *httpClient) buildURL(host, path string) string {
       if strings.HasPrefix(host, "https://") {
           return host + path
       }
       // Bare hostname or unrecognised scheme: enforce HTTPS
       host = strings.TrimPrefix(host, "http://")
       return "https://" + host + path
   }
   ```
   This silently upgrades any `http://` host to `https://` and warns via the existing test that already passes an `httptest.NewServer` URL (which uses `http://` — tests must still pass because the scheme enforcement is for real hostnames, not test server URLs; adjust the test helper if needed or accept that the test's `httptest` URL is already stripped by `prurl.Parse` before reaching `buildURL`).

   If the upgrade silently breaks existing tests (httptest uses `http://`), keep the `http://` pass-through for test mode OR adjust the constructor to accept a scheme override for tests. The simplest safe change: strip `http://` prefix and prepend `https://` (test code sets `server.URL` which is an absolute URL and is passed directly — confirm whether tests call `buildURL` with the full `http://host` or just the bare host; read the test carefully before changing).

3. **Replace all `fmt.Errorf` calls with `errors.Wrapf` / `errors.Errorf`** from `github.com/bborbe/errors`. All methods already have `ctx context.Context` as their first parameter:
   - `fmt.Errorf("...: %w", err)` → `errors.Wrapf(ctx, err, "...")`
   - `fmt.Errorf("...")` (no wrapped error) → `errors.Errorf(ctx, "...")`
   - Files to update: `GetPRBranches`, `PostComment`, `Approve`, `NeedsWork` methods — approximately 18 call sites (lines 81, 88, 98, 103, 107, 110, 135, 140, 148, 173, 180, 214, 219, 227, 255–281).

4. **Fix bare `return err`** in three places (~lines 153, 185, 232) where callers of `checkResponseStatus` / `checkApproveResponseStatus` propagate the error without context. Replace with:
   ```go
   return errors.Wrap(ctx, err, "check response status")
   ```

5. **Update tests** in `client_test.go`: the test server uses `http://` URLs via `httptest.NewServer`. After step 2, ensure the test correctly exercises the upgraded path. Add a new test case for `buildURL` with a plain `http://` host to assert it becomes `https://`.

6. **Run `cd agent/pr-reviewer && make test`** — must pass before proceeding to precommit.
</requirements>

<constraints>
- Only change files in `agent/pr-reviewer/`
- Do NOT commit — dark-factory handles git
- Existing tests must still pass
- Use `errors.Wrapf(ctx, err, "...")` from `github.com/bborbe/errors` (always pass `ctx` as first arg) — never `fmt.Errorf`, never bare `return err`
- Code changes MUST add or update tests for changed paths — paths covered before the fix must remain covered, new paths must be tested
- Do NOT add `//nolint` directives
</constraints>

<verification>
cd agent/pr-reviewer && make precommit
</verification>

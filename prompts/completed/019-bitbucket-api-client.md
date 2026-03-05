---
status: completed
summary: Implemented Bitbucket Server API client with GetPRBranch and PostComment methods using REST API v1.0
container: pr-reviewer-019-bitbucket-api-client
dark-factory-version: v0.17.0
created: "2026-03-05T13:06:57Z"
queued: "2026-03-05T13:06:57Z"
started: "2026-03-05T13:06:57Z"
completed: "2026-03-05T13:13:49Z"
---
<objective>
Add Bitbucket Server API client for fetching PR metadata and posting comments via REST API. Mirrors the GitHub Client interface pattern.
</objective>

<context>
Read CLAUDE.md for project conventions.
Read ALL markdown files in ~/Documents/workspaces/coding-guidelines/ for Go patterns.
Read pkg/github/client.go for the existing Client interface pattern (GetPRBranch, PostComment, SubmitReview).
Read pkg/github/client_test.go for test patterns.
Precondition: Config has ResolvedBitbucketToken() method (from prompt 018).
Precondition: prurl.PRInfo has Host, Project, Repo, Number fields (from prompt 017).
</context>

<requirements>
1. Create `pkg/bitbucket/client.go` with:
   ```go
   //counterfeiter:generate -o ../../mocks/bitbucket-client.go --fake-name BitbucketClient . Client
   type Client interface {
       GetPRBranch(ctx context.Context, host, project, repo string, number int) (string, error)
       PostComment(ctx context.Context, host, project, repo string, number int, body string) error
   }
   ```

2. Implement `NewClient(token string) Client` returning `*httpClient`:
   - `GetPRBranch`: GET `https://{host}/rest/api/1.0/projects/{project}/repos/{repo}/pull-requests/{number}`
     - Parse JSON response: `.fromRef.displayId` is the source branch name
     - Set `Authorization: Bearer {token}` header
     - Handle 401 → "authentication failed for {host}"
     - Handle 403 → "insufficient permissions for {host}"
     - Handle 404 → "PR not found: {host}/projects/{project}/repos/{repo}/pull-requests/{number}"
   - `PostComment`: POST `https://{host}/rest/api/1.0/projects/{project}/repos/{repo}/pull-requests/{number}/comments`
     - Body: `{"text": "<body>"}`
     - Content-Type: `application/json`
     - Same auth header and error handling as GetPRBranch

3. Use `net/http` from stdlib — no external HTTP client library.

4. Use `encoding/json` for JSON parsing — no external JSON library.

5. Add tests in `pkg/bitbucket/client_test.go`:
   - Test error for empty token
   - Test that GetPRBranch constructs correct URL (use httptest.NewServer for mock)
   - Test that PostComment constructs correct URL and body
   - Test 401/403/404 error handling
   - Test context cancellation

6. Add `pkg/bitbucket/client_suite_test.go` for Ginkgo test suite.

7. Run `go generate ./...` to generate Counterfeiter mock.

8. Token must NEVER appear in error messages or log output. When logging errors, include the host but not the token.
</requirements>

<constraints>
- Existing GitHub client must continue to work unchanged
- Must use Bitbucket Server REST API v1.0 (not Bitbucket Cloud API)
- HTTPS required for all API calls
- Token must never appear in log output or error messages
- Use stdlib net/http and encoding/json only — no external deps
- Use Ginkgo v2 + Gomega for tests
- Do NOT modify main.go — wiring comes in a later prompt
- Do NOT modify pkg/github/ — keep GitHub and Bitbucket clients separate
</constraints>

<verification>
Run `make test` -- must pass.
Run `make precommit` -- must pass.
</verification>

<success_criteria>
- Bitbucket Client interface with GetPRBranch and PostComment
- REST API calls to correct Bitbucket Server endpoints
- Bearer token auth in headers
- Proper error handling for 401/403/404
- Token never in error messages
- Mock generated via counterfeiter
- make precommit passes
</success_criteria>

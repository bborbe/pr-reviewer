---
status: completed
summary: Wired Bitbucket Server support into main.go with platform routing and shared review workflow
container: pr-reviewer-020-wire-bitbucket-into-main
dark-factory-version: v0.17.0
created: "2026-03-05T13:13:51Z"
queued: "2026-03-05T13:13:51Z"
started: "2026-03-05T13:13:51Z"
completed: "2026-03-05T13:19:06Z"
---
<objective>
Wire Bitbucket Server support into main.go. Route GitHub and Bitbucket URLs to their respective clients while sharing the worktree/review/verdict workflow.
</objective>

<context>
Read CLAUDE.md for project conventions.
Read ALL markdown files in ~/Documents/workspaces/coding-guidelines/ for Go patterns.
Read main.go for the current flow: parse URL → load config → get branch → create worktree → review → post/submit.
Read pkg/prurl/prurl.go for Platform type and PRInfo struct (from prompt 017).
Read pkg/bitbucket/client.go for the Bitbucket Client interface (from prompt 019).
Read pkg/github/client.go for the GitHub Client interface.
Read pkg/config/config.go for ResolvedBitbucketToken and ResolvedGitHubToken.
Precondition: prurl.Parse returns PRInfo with Platform field (PlatformGitHub or PlatformBitbucket).
Precondition: Bitbucket Client has GetPRBranch and PostComment methods.
Precondition: Config has ResolvedBitbucketToken() method.
</context>

<requirements>
1. Update `run()` in main.go to route based on `prInfo.Platform`:
   - `PlatformGitHub`: use existing GitHub flow (unchanged)
   - `PlatformBitbucket`: use Bitbucket client for GetPRBranch and PostComment
   - Both share: worktree creation, Claude review, stdout output, verdict logging

2. For Bitbucket path in `run()`:
   - Resolve Bitbucket token: `cfg.ResolvedBitbucketToken()`
   - If token is empty, exit with error: "BITBUCKET_TOKEN not set"
   - Create Bitbucket client: `bitbucket.NewClient(resolvedBBToken)`
   - Get PR branch: `bbClient.GetPRBranch(ctx, prInfo.Host, prInfo.Project, prInfo.Repo, prInfo.Number)`
   - Worktree creation and review: identical to GitHub path
   - Post comment: `bbClient.PostComment(ctx, prInfo.Host, prInfo.Project, prInfo.Repo, prInfo.Number, reviewText)`
   - Log token status using same logTokenStatus pattern (but with "bitbucket token:" prefix)

3. For Bitbucket path, verdict handling:
   - Parse verdict (same as GitHub)
   - Log verdict to stderr (same as GitHub)
   - `--comment-only` → PostComment only (same as GitHub)
   - `VerdictApprove` / `VerdictRequestChanges` → PostComment only for now (Bitbucket SubmitReview is spec 005)
   - `VerdictComment` → PostComment (same as GitHub)

4. Extract shared review logic to avoid duplicating the worktree/review/verdict code path. Both GitHub and Bitbucket should share the same review execution, only differing in how they get the branch and post results.

5. Update logTokenStatus to accept a label parameter (e.g., "github token" vs "bitbucket token").

6. Ensure funlen lint compliance — extract functions if run() exceeds 80 lines.
</requirements>

<constraints>
- Existing GitHub flow must continue to work unchanged
- Bitbucket token missing → exit 1 with helpful message (not silent fallback)
- Token must never appear in log output
- Verdict approve/request-changes for Bitbucket → plain comment for now (spec 005 adds Bitbucket review verdicts)
- Use Ginkgo v2 + Gomega for any new tests
- Do NOT modify pkg/ packages — only main.go changes
</constraints>

<verification>
Run `make test` -- must pass.
Run `make precommit` -- must pass.
</verification>

<success_criteria>
- Bitbucket Server URLs routed to Bitbucket client
- GitHub URLs still routed to GitHub client (unchanged)
- Missing BITBUCKET_TOKEN exits with helpful error
- Worktree/review/verdict flow shared between platforms
- Review printed to stdout for both platforms
- Verdict logged to stderr for both platforms
- make precommit passes
</success_criteria>

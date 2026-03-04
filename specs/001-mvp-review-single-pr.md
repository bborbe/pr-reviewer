---
tags:
  - dark-factory
  - spec
---
Tags: [[Dark Factory - Prompt Strategy Guide]]

---

## Problem

No tool exists to review pull requests using the local Claude Code CLI with full project context (CLAUDE.md, coding guidelines, subagents, `/code-review` skill). The existing GitHub Actions approach (`claude-code-action`) runs generic reviews without project-specific rules. Bitbucket Server has no review automation at all.

## Goal

After completion, a Go CLI tool `pr-reviewer` accepts a single PR URL (GitHub or Bitbucket Server), resolves the repo to a local checkout via config, creates a temporary git worktree for the PR branch, runs Claude Code `/code-review` in that worktree, and posts the review as a comment on the PR.

## Non-goals

- Batch/polling mode (review all open PRs) — future work
- Creating or merging PRs
- Reviewing without a local checkout (no cloning)
- GitLab support
- Interactive mode / TUI

## Desired Behavior

1. User runs `pr-reviewer <URL>` where URL is a GitHub PR or Bitbucket Server PR
2. Tool parses the URL to determine provider (GitHub vs Bitbucket Server), repo identifier, and PR number
3. Tool looks up the local repo path from `~/.pr-reviewer.yaml` config
4. Tool fetches the PR metadata (source branch, target branch) via API
5. Tool runs `git fetch` in the local repo to ensure branches are up to date
6. Tool creates a git worktree at a temp path for the PR's source branch
7. Tool runs `claude --print "/code-review"` in the worktree directory — this picks up the repo's CLAUDE.md and all project-specific context
8. Tool posts the Claude output as a PR comment via API (GitHub: `gh pr comment`, Bitbucket Server: REST API)
9. Tool removes the git worktree and cleans up
10. Tool exits 0 on success, non-zero on failure with clear error message

## Constraints

- Config file location: `~/.pr-reviewer.yaml`
- Must use `gh` CLI for GitHub operations (already authenticated)
- Must use Bitbucket Server REST API with token auth for Bitbucket operations
- Must use `git worktree add/remove` — never checkout in the main working tree
- Must use `claude` CLI — not the SDK, not the API directly
- Worktree path: `<repo-path>/.worktrees/pr-review-<pr-number>` (inside repo, not /tmp)
- Always clean up worktree on exit (including on error/signal)

## Config Format

```yaml
# ~/.pr-reviewer.yaml
repos:
  - url: https://github.com/bborbe/teamvault-docker
    path: /Users/bborbe/Documents/workspaces/teamvault-docker
  - url: https://bitbucket.seibert.tools/projects/OC/repos/root
    path: /Users/bborbe/Documents/workspaces/oc-root

bitbucket:
  token: ${BITBUCKET_TOKEN}  # env var reference
```

GitHub auth is handled by `gh` CLI — no token needed in config.

## URL Patterns

### GitHub
```
https://github.com/{owner}/{repo}/pull/{number}
```
Extract: owner, repo, PR number.
Repo URL for config lookup: `https://github.com/{owner}/{repo}`

### Bitbucket Server
```
https://{host}/projects/{project}/repos/{repo}/pull-requests/{number}/overview
https://{host}/projects/{project}/repos/{repo}/pull-requests/{number}
```
Extract: host, project, repo, PR number.
Repo URL for config lookup: `https://{host}/projects/{project}/repos/{repo}`

## Failure Modes

| Trigger | Expected behavior | Recovery |
|---------|-------------------|----------|
| Unknown URL format | Exit 1 with "unsupported URL format: <url>" | User fixes URL |
| Repo not in config | Exit 1 with "repo not found in config, add to ~/.pr-reviewer.yaml: <url>" | User adds mapping |
| Local path doesn't exist | Exit 1 with "local path not found: <path>" | User clones repo |
| Local path not a git repo | Exit 1 with "not a git repo: <path>" | User checks path |
| git fetch fails | Exit 1 with git error | User checks network/auth |
| PR branch not found after fetch | Exit 1 with "branch not found: <branch>" | User checks PR is open |
| Worktree creation fails | Exit 1 with git error | User checks for conflicts |
| claude CLI not found | Exit 1 with "claude not found in PATH" | User installs Claude Code |
| claude exits non-zero | Exit 1 with claude's stderr | User investigates |
| Posting comment fails | Exit 1 with API error, but print review to stdout as fallback | User posts manually |
| Signal (SIGINT/SIGTERM) | Clean up worktree, then exit | Automatic |

## Security / Abuse Cases

- Bitbucket token in config: support `${ENV_VAR}` references, never hardcode tokens
- URL parsing: validate against known patterns only, reject anything else
- Worktree path: derived deterministically from repo path + PR number, no user-controlled path components
- Claude output: posted as-is — trusted internal tool, no sanitization needed

## Acceptance Criteria

- [ ] `pr-reviewer https://github.com/bborbe/teamvault-docker/pull/4` produces a review comment on the PR
- [ ] `pr-reviewer https://bitbucket.seibert.tools/projects/OC/repos/root/pull-requests/254/overview` produces a review comment on the PR
- [ ] Unknown URL exits 1 with helpful message
- [ ] Missing repo in config exits 1 with helpful message
- [ ] Worktree is cleaned up after success
- [ ] Worktree is cleaned up after failure
- [ ] Worktree is cleaned up after SIGINT
- [ ] Review is printed to stdout even if posting fails

## Verification

```
make precommit
```

## Do-Nothing Option

Continue using GitHub Actions `claude-code-action` for GitHub. No Bitbucket review automation. Reviews remain generic without project-specific rules. Acceptable for repos with simple conventions, but inadequate for repos with detailed CLAUDE.md and custom `/code-review` skills.

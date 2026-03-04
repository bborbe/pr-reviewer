---
tags:
  - dark-factory
  - spec
status: completed
---
Tags: [[Dark Factory - Spec Writing Guide]]

---

## Problem

When dark-factory creates a PR under user A's GitHub account, and pr-reviewer also runs as user A (same `gh` auth), GitHub rejects the review because a user cannot approve their own pull request. This blocks the autonomous pipeline: dark-factory creates PR → pr-reviewer reviews → but review is rejected as self-review.

## Goal

After completion, pr-reviewer supports using a separate GitHub identity for posting reviews. A dedicated `GH_TOKEN` (from a different GitHub user or GitHub App) is passed to all `gh` CLI calls, so reviews come from a different identity than the PR author.

## Non-goals

- Managing or creating GitHub accounts/apps
- Token rotation or refresh
- Per-repo reviewer identity (one global reviewer identity is sufficient)
- Changing how dark-factory authenticates (that stays as user A)

## Desired Behavior

1. Config adds optional `github.token` field referencing an env var
2. When set, all `gh` CLI calls in pr-reviewer pass `GH_TOKEN=<value>` in the subprocess environment
3. When not set, `gh` uses default auth (current user) — backward compatible
4. The token must be from a different GitHub identity than the PR author
5. Token value is resolved from the referenced env var at runtime, never stored in config

### Config

```yaml
# ~/.pr-reviewer.yaml
github:
  token: ${PR_REVIEWER_TOKEN}  # env var with PAT from reviewer bot account

repos:
  - url: https://github.com/bborbe/dark-factory
    path: /Users/bborbe/Documents/workspaces/dark-factory
```

When `github.token` is absent or empty, behavior is unchanged (uses default `gh` auth).

### Env Var Resolution

The `${VAR}` syntax in config is resolved at load time:
- `${PR_REVIEWER_TOKEN}` → looks up `os.Getenv("PR_REVIEWER_TOKEN")`
- Empty or unset → treated as "not configured" (falls back to default `gh` auth)
- Literal string without `${}` → used directly (for testing, not recommended)

### gh CLI Token Override

`gh` respects the `GH_TOKEN` environment variable. When a token is configured, every `exec.CommandContext` call for `gh` sets `cmd.Env` to include `GH_TOKEN=<resolved-token>`. The rest of the environment is inherited from the parent process.

## Constraints

- Backward compatible — existing configs without `github.token` keep working
- Token never appears in logs or error messages
- Token never written to disk (only read from env var)
- Must work with both GitHub PATs (classic and fine-grained) and GitHub App installation tokens
- `gh` CLI version must support `GH_TOKEN` env var (v2.0+)

## Failure Modes

| Trigger | Expected behavior | Recovery |
|---------|-------------------|----------|
| `github.token` references unset env var | Fall back to default `gh` auth, log warning | User sets env var or removes config |
| Token is invalid (401) | `gh` exits non-zero, pr-reviewer reports auth error | User checks token |
| Token lacks permissions (403) | `gh` exits non-zero, pr-reviewer reports permission error | User checks token scopes |
| Token is from same user as PR author | GitHub rejects review as self-review | User uses a different account's token |
| Config has `github.token` without `${}` | Used as literal token (works but not recommended) | User wraps in `${}` |

## Security / Abuse Cases

- Token in config file: only env var references, never raw tokens
- Token in logs: never — mask if it appears in error output
- Token in subprocess: only via `cmd.Env`, not as CLI argument (would appear in `ps`)
- Config file permissions: warn if `~/.pr-reviewer.yaml` is world-readable

## Acceptance Criteria

- [ ] `github.token: ${PR_REVIEWER_TOKEN}` in config resolves env var
- [ ] `gh` commands use `GH_TOKEN` from config when set
- [ ] Reviews posted under the bot account identity (not the PR author)
- [ ] Missing env var falls back to default `gh` auth with warning
- [ ] Existing configs without `github.token` work unchanged
- [ ] Token never appears in log output
- [ ] `make precommit` passes

## Verification

```
make precommit
```

## Do-Nothing Option

Run pr-reviewer under a different OS user or in a separate shell with `GH_TOKEN` exported manually. Works but requires manual setup per invocation and breaks the autonomous pipeline.

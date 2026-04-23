---
tags:
  - dark-factory
  - idea
status: idea
---

# Clone PR Repo into Workdir

## Idea

The agent clones the PR's repository into `/tmp/pr-<N>`, checks out the PR head branch, and cd's there before running review commands — instead of inspecting the PR only through `gh pr diff` output.

## Why

`/coding:pr-review` (and most real review tooling) expects a real git checkout: it needs `go.mod`/`pyproject.toml` to detect project type, runs `make precommit`, dispatches sub-agents on actual source files, and diffs `origin/<target>...HEAD` against working trees. A `gh pr diff` string is not enough.

## Sketch

- Agent workflow prepends:
  ```sh
  gh pr view <N> --repo <owner/repo> --json headRefName,baseRefName,headRepositoryOwner,headRepository
  git clone --depth=200 --branch <base> <clone-url> /tmp/pr-<N>
  cd /tmp/pr-<N>
  git fetch origin <head>:<head>
  git checkout <head>
  ```
- Ephemeral — cloned per task, wiped when pod terminates
- Fits current `ephemeral-storage: 2Gi` request for typical repos
- Requires AllowedTools: `Bash(git clone:*)`, `Bash(git fetch:*)`, `Bash(git checkout:*)`, `Bash(cd:*)` (already via `Bash(git:*)`)

## Risks / Open questions

- Large monorepos may exceed 2Gi ephemeral — bump to 5Gi? Or unshallow only on need?
- Fork PRs: head is on a different repo owner — URL construction differs. `gh pr checkout` handles this natively; consider using it instead of manual clone
- Private repos: `GH_TOKEN` must cover clone access, not just API reads
- `--depth=200` heuristic breaks for PRs older than 200 commits of base — add fallback to `git fetch --unshallow`

## Related

- Prerequisite for: `use-coding-pr-review`
- Superseded partially by: `pvc-backed-repo-cache` (avoid cloning same repo every time)
- Touches: `pkg/prompts/workflow.md`, `pkg/factory/factory.go` (AllowedTools)

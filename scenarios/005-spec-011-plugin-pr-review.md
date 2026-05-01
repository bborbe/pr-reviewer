---
status: draft
spec: 011-swap-execution-to-coding-pr-review
---

# Scenario 005: execution phase delegates to /coding:pr-review plugin

Validates that the execution phase invokes the `/coding:pr-review` slash command
inside the prepared worktree, receives a structured findings report from specialist
sub-agents, translates it into the verdict JSON schema, and cleans up the worktree.
This is the required subprocess + slash-command interface seam test for spec-011.

## Setup
- [ ] Scenario 001 prerequisites met (GitHub token in `GH_TOKEN`, `claude` CLI available)
- [ ] `bborbe/coding` plugin is installed: `claude plugin list | grep coding`
- [ ] PR #2 exists in `bborbe/code-reviewer` with a non-empty diff against `master`
- [ ] Task file generated for PR #2:
  ```bash
  cd ~/Documents/workspaces/code-reviewer/agent/pr-reviewer/cmd/run-task
  make generate-dummy-task TASK_FILE=test5-plugin.md
  ```
  (uses default PR #2 task configuration from the Makefile)

## Action — standard mode (default)
- [ ] Run the agent in `in_progress` phase:
  ```bash
  cd ~/Documents/workspaces/code-reviewer/agent/pr-reviewer/cmd/run-task
  make run-dummy-task TASK_FILE=test5-plugin.md
  ```

## Expected — standard mode
- [ ] Agent log contains `/coding:pr-review` (confirms slash command was invoked)
- [ ] Agent log contains at least one sub-agent name: `go-quality-assistant`,
      `go-security-specialist`, `go-factory-pattern-assistant`, or similar
- [ ] `## Review` section in `test5-plugin.md` contains a valid JSON object:
  ```bash
  grep -A 5 "## Review" test5-plugin.md
  ```
- [ ] Verdict JSON parses without error:
  ```bash
  grep -A 50 "## Review" test5-plugin.md | python3 -c "import sys,json; json.load(sys.stdin)"
  ```
- [ ] `verdict` field is one of `approve`, `request_changes`, `comment`
- [ ] `summary` is a non-empty string
- [ ] `concerns_addressed` is a non-empty list (reflects sub-agent or Must Fix/Should Fix findings, not generic boilerplate from the old hand-rolled prompt)
- [ ] Workdir cleaned up (no stale worktree after run):
  ```bash
  ls ~/.cache/code-reviewer/work/
  ```
  Worktree directory for the task should be absent (or pruned on next startup)
- [ ] Process exits 0

## Action — short mode
- [ ] Re-generate a fresh task file (or copy test5-plugin.md before the Review section is written):
  ```bash
  make generate-dummy-task TASK_FILE=test5-short.md
  ```
- [ ] Run with short mode:
  ```bash
  REVIEW_MODE=short make run-dummy-task TASK_FILE=test5-short.md
  ```

## Expected — short mode
- [ ] Agent log shows `/coding:pr-review master short`
- [ ] No specialist sub-agent names appear in the log (short mode skips agents)
- [ ] Verdict JSON is still valid and parses

## Cleanup
- [ ] `rm test5-plugin.md test5-short.md`
- [ ] Worktrees: removed automatically on next pod/process startup via `PruneAllWorktrees`

## Notes
Last run: (not yet run — scenario created for spec-011)

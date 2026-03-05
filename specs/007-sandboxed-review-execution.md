---
tags:
  - dark-factory
  - spec
status: completed
---
Tags: [[Dark Factory - Spec Writing Guide]]

---

## Problem

pr-reviewer runs `claude --print "/pr-review <target>"` directly on the host machine inside a git worktree of untrusted code. A malicious PR could craft a Makefile, pre-commit hook, or build script that executes destructive commands (e.g., `rm -rf ~/`) when Claude runs `make precommit` or other build tools during the review.

Additionally, pr-reviewer currently uses `~/.claude/` config which contains the user's personal commands and agents. It should use `~/.claude-yolo/` — the same sandboxed Claude config that dark-factory uses.

## Goal

After completion, pr-reviewer executes Claude reviews using `~/.claude-yolo/` config with a dedicated `/pr-review` command. The review output includes a structured JSON verdict block that the Go verdict parser extracts reliably, replacing heuristic markdown section scanning.

Future work (not this spec): run the review inside the claude-yolo Docker container for full sandboxing.

## Non-goals

- Running reviews inside Docker container (future spec)
- Changing dark-factory's existing `/code-review` command
- Bitbucket Cloud support
- Per-repo agent configuration
- Changing the host `~/.claude/commands/pr-review.md` (stays for human use)

## Desired Behavior

1. `~/.claude-yolo/commands/pr-review.md` exists with diff-scoped review logic, adapted for yolo container paths
2. All agents referenced by `/pr-review` are present in `~/.claude-yolo/agents/`
3. The `/pr-review` command instructs Claude to append a JSON verdict block at the end of the review: `{"verdict": "approve|request-changes|comment", "reason": "..."}`
4. pr-reviewer invokes `claude --print --model <model> --claude-config-dir ~/.claude-yolo "/pr-review <target>"` (or equivalent flag to use yolo config)
5. The Go verdict parser first tries to extract JSON verdict from the review text; falls back to heuristic section scanning if no JSON found
6. Existing `--comment-only` flag still works
7. Existing verdict behavior (approve/request-changes/comment) is unchanged

## Constraints

- `~/.claude-yolo/` is shared with dark-factory — do not break existing dark-factory commands or agents
- Agents in yolo must be self-contained (no references to host-specific paths like `~/Documents/workspaces/`)
- The JSON verdict block must be on its own line and parseable even if surrounded by markdown code fences
- Review output posted to PR comments must NOT include the JSON verdict block (strip it before posting)
- `make precommit` must pass

## Failure Modes

| Trigger | Expected behavior | Recovery |
|---------|-------------------|----------|
| JSON verdict missing from output | Fall back to heuristic section scanning | Automatic |
| JSON verdict malformed | Fall back to heuristic section scanning | Automatic |
| `--claude-config-dir` flag not supported | Check Claude CLI version, document minimum version | Manual |
| Agent file missing in yolo | Review runs with fewer agents (degraded, not broken) | Manual — copy missing agent |
| Dark-factory `/code-review` breaks | Not affected — separate command, no changes | N/A |

## Acceptance Criteria

- [ ] `~/.claude-yolo/commands/pr-review.md` exists and includes JSON verdict instruction
- [ ] All agents needed by `/pr-review` are in `~/.claude-yolo/agents/`
- [ ] pr-reviewer uses `~/.claude-yolo` config dir when invoking Claude
- [ ] Verdict parser extracts JSON verdict from review output
- [ ] Verdict parser falls back to heuristic if no JSON found
- [ ] JSON verdict block is stripped from review text before posting as comment
- [ ] `--comment-only` flag still works
- [ ] `make precommit` passes

## Verification

```
make precommit
```

## Resolved Questions

- **Config dir flag**: No `--claude-config-dir` exists. Dark-factory uses Docker volume mount: `-v ~/.claude-yolo:/home/node/.claude`. pr-reviewer should do the same — run Claude inside the `claude-yolo` container.
- **Executor pattern**: Reuse dark-factory's Docker executor approach (`docker run --rm -v ... claude-yolo`).
- **Agents**: Copy from host `~/.claude/agents/` to `~/.claude-yolo/agents/`, adapting any host-specific paths to container paths (`/home/node/.claude/docs/`).

## Do-Nothing Option

Keep running reviews on the host with `~/.claude/` config. Reviews work but are vulnerable to malicious code in PRs. The host `/pr-review.md` with mandatory section headers provides a partial fix for verdict parsing reliability but remains heuristic-based.

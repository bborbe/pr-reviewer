---
tags:
  - dark-factory
  - idea
status: idea
---

# Use /coding:pr-review Instead of Free-Form Workflow

## Idea

Replace the hand-rolled `pkg/prompts/workflow.md` flow with the plugin-provided `/coding:pr-review <target-branch> standard` slash command. The agent's prompt becomes: fetch PR metadata → clone → checkout → `/coding:pr-review` → extract verdict.

## Why

`workflow.md` is reinventing what the coding plugin already does better:
- Worktree-based diff against target branch (excludes vendor/node_modules)
- Project-type detection (Go / Python)
- Parallel dispatch of specialist sub-agents (go-quality, go-security, ...)
- Severity-bucketed report (Must Fix / Should Fix / Nice to Have)

We'd inherit improvements upstream for free and shed maintenance of our own prompt.

## Sketch

- Rewrite `pkg/prompts/workflow.md`:
  1. `gh pr view <N> --json headRefName,baseRefName,headRepository`
  2. `git clone --depth=200 --branch <base> <url> /tmp/pr-<N>`
  3. `cd /tmp/pr-<N> && git fetch origin <head>:<head> && git checkout <head>`
  4. Invoke `/coding:pr-review <base> standard`
  5. Parse the plugin's consolidated report → emit our verdict JSON
- `factory.go` AllowedTools: add `Task` (sub-agent dispatch), `Bash(make:*)`, `Bash(rm:*)`, `Bash(pwd:*)`; drop `WebFetch`
- `agent/.claude/CLAUDE.md`: loosen "no package installation" wording → "no runtime package installs" (plugin install is build-time-ish)
- Output-format contract unchanged — still `{"status":"...","message":"..."}`

## Risks / Open questions

- Plugin's report format might drift → our verdict parser needs to cope; fallback to heuristic scan
- `Task` tool broadens trust surface; sub-agents inherit restricted allowlist but can still Read/Grep/Glob broadly
- `/coding:pr-review` uses `git worktree` — requires full clone, not shallow. Our `--depth=200` may need to be unshallow
- Keep current `workflow.md` as fallback when plugin absent? (answered separately)

## Related

- Depends on: `install-coding-plugin-on-start`
- Depends on: `clone-pr-repo-into-workdir`
- Touches: `pkg/prompts/workflow.md`, `pkg/factory/factory.go`, `agent/.claude/CLAUDE.md`

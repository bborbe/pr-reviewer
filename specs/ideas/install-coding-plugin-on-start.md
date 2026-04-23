---
tags:
  - dark-factory
  - idea
status: idea
---

# Install bborbe/coding Plugin on Start

## Idea

Container entrypoint script installs (or updates) the [bborbe/coding](https://github.com/bborbe/coding) Claude Code plugin before exec'ing the agent binary. On first pod boot the plugin is installed; subsequent boots refresh to latest.

## Why

The plugin ships `/coding:pr-review` and a suite of review sub-agents (go-quality, go-security, python-quality, etc.) that are far richer than our free-form `workflow.md`. Having them present inside every pr-reviewer Job unlocks the full review stack with zero bespoke prompt engineering.

## Sketch

- Add `docker-entrypoint.sh`:
  ```sh
  #!/bin/sh
  set -e
  if claude plugin list 2>/dev/null | grep -q "^coding"; then
    claude plugin marketplace update coding || true
    claude plugin update coding@coding || true
  else
    claude plugin marketplace add bborbe/coding
    claude plugin install coding@coding
  fi
  exec /main -v=2 "$@"
  ```
- Dockerfile: `COPY docker-entrypoint.sh /` + `chmod +x` + change `ENTRYPOINT`
- PVC-backed `~/.claude/plugins/` makes install persist across pod restarts
- Plugin install runs *before* agent subprocess — outside the headless `--allowedTools` sandbox, so no guardrail change for the agent
- Always track latest (master) — no version pinning

## Risks / Open questions

- Pod start-up hits github.com; graceful fallback if offline? (use cached install, skip update)
- Marketplace add is idempotent? Need to check behavior on re-add
- Plugin updates could break the agent mid-run if a command is renamed — but plugin is pulled once per pod, so a running task is stable
- `claude plugin install` exit code on "already installed"?

## Related

- Prerequisite for: `use-coding-pr-review` idea
- Touches: `agent/pr-reviewer/Dockerfile`, new `docker-entrypoint.sh`

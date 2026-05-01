---
status: draft
spec: 010-git-checkout-review-workdir
---

# Scenario 001: pr-reviewer reviews GitHub PR via HTTPS clone URL

Validates that the pr-reviewer agent clones, checks out, and reviews a real PR using a standard `https://github.com/...` URL — the most common form.

## Setup
- [ ] code-reviewer at v0.20.0+ (`cd ~/Documents/workspaces/code-reviewer && git describe --tags`)
- [ ] Local Claude config dir authenticated: `CLAUDE_CONFIG_DIR=~/.claude-agent claude /login`
- [ ] `gh auth status` shows authenticated GitHub session
- [ ] Test PR available: `gh pr view 2 --repo bborbe/code-reviewer` returns OPEN state

## Action
- [ ] Generate task file with HTTPS clone URL:
  ```bash
  cd ~/Documents/workspaces/code-reviewer/agent/pr-reviewer/cmd/run-task
  make generate-dummy-task TASK_FILE=test1-https.md \
    CLONE_URL=https://github.com/bborbe/code-reviewer.git
  ```
- [ ] Run agent against the task file:
  ```bash
  make run-dummy-task TASK_FILE=test1-https.md
  ```

## Expected
- [ ] Agent log shows successful `git clone` of HTTPS URL (no auth errors)
- [ ] Agent log shows successful `git checkout` of PR head SHA
- [ ] Final stdout contains a JSON verdict block with `"verdict":"approve"|"request_changes"|"comment"`
- [ ] Result line shows `{"Status":"done","NextPhase":"ai_review",...}`
- [ ] Process exits 0
- [ ] Workdir under `~/.cache/code-reviewer/work/<task-id>/` contains the cloned repo

## Cleanup
- [ ] `rm test1-https.md`
- [ ] `rm -rf ~/.cache/code-reviewer/work/22222222-aaaa-bbbb-cccc-333333333333`

## Notes
Last run: 2026-05-01 — passed in ~17s. Verdict: approve, 1 nit comment on the test PR's marker HTML comment.

---
status: draft
spec: 010-git-checkout-review-workdir
---

# Scenario 002: pr-reviewer reviews GitHub PR via SCP-style SSH clone URL

Validates that `ParseCloneURL` correctly handles SCP-form URLs (`git@host:owner/repo.git`) and that the agent can clone+review using the SSH transport. Regression coverage for the SCP support shipped in v0.20.0.

## Setup
- [ ] Scenario 001 prerequisites met
- [ ] SSH key registered with GitHub: `ssh -T git@github.com` returns "successfully authenticated"

## Action
- [ ] Generate task file with SCP/SSH clone URL:
  ```bash
  cd ~/Documents/workspaces/code-reviewer/agent/pr-reviewer/cmd/run-task
  make generate-dummy-task TASK_FILE=test2-scp.md \
    CLONE_URL=git@github.com:bborbe/code-reviewer.git
  ```
- [ ] Run agent against the task file:
  ```bash
  make run-dummy-task TASK_FILE=test2-scp.md
  ```

## Expected
- [ ] Agent log shows successful `git clone` of SCP URL via SSH transport
- [ ] No regex/parser errors in `ParseCloneURL`
- [ ] Final stdout contains a JSON verdict block
- [ ] Result line shows `{"Status":"done",...}`
- [ ] Process exits 0

## Cleanup
- [ ] `rm test2-scp.md`
- [ ] `rm -rf ~/.cache/code-reviewer/work/22222222-aaaa-bbbb-cccc-333333333333`

## Notes
Last run: 2026-05-01 — passed in ~14s. Verdict: approve, no comments. Confirms the `splitCloneURL` SCP-detection regex from `pkg/git/clone_url.go` works end-to-end.

---
status: draft
spec: 010-git-checkout-review-workdir
---

# Scenario 003: pr-reviewer reviews arbitrary ref (non-PR commit/branch/tag)

Validates that the agent can review any git ref — not just PR heads. Confirms the spec's "any ref on any git server" capability: a SHA, branch, or tag without a corresponding pull request.

## Setup
- [ ] Scenario 001 prerequisites met
- [ ] Two commits exist on `master` of `bborbe/code-reviewer`:
  ```bash
  cd ~/Documents/workspaces/code-reviewer
  git log master --oneline -2
  ```

## Action
- [ ] Generate task file with arbitrary ref pair (master HEAD vs master~1):
  ```bash
  cd ~/Documents/workspaces/code-reviewer
  BASE=$(git rev-parse master~1)
  HEAD=$(git rev-parse master)
  cd agent/pr-reviewer/cmd/run-task
  make generate-dummy-task TASK_FILE=test3-branch.md \
    REF=$HEAD BASE_REF=$BASE \
    TASK_ID="44444444-aaaa-bbbb-cccc-555555555555" \
    PR_URL="(none — arbitrary ref review)"
  ```
- [ ] Run agent:
  ```bash
  make run-dummy-task TASK_FILE=test3-branch.md
  ```

## Expected
- [ ] Agent log shows successful clone + checkout of the SHA (`REF`)
- [ ] Agent log shows `git diff $BASE_REF...HEAD` invoked
- [ ] Final stdout contains a JSON verdict block — non-trivial review (real diff, not single-line marker)
- [ ] Result line shows `{"Status":"done",...}`
- [ ] No errors related to "PR not found" or similar (proves PR semantics not required)
- [ ] Process exits 0

## Cleanup
- [ ] `rm test3-branch.md`
- [ ] `rm -rf ~/.cache/code-reviewer/work/44444444-aaaa-bbbb-cccc-555555555555`

## Notes
Last run: 2026-05-01 — passed in ~80s reviewing v0.20.0 release commit. Verdict: approve, 3 nuanced comments on tools.env migration details. Demonstrates that the agent reviews real-content diffs at scale.

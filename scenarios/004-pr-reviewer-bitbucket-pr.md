---
status: blocked
spec: 010-git-checkout-review-workdir
blocker: needs accessible Bitbucket repo + PR target
---

# Scenario 004: pr-reviewer reviews Bitbucket PR via HTTPS clone URL

Validates provider-agnostic clone — confirms the agent works against a non-GitHub remote. The spec's "Provider-agnostic verified" DoD criterion (Behavioral: Bitbucket PR review completes end-to-end).

## Setup
- [ ] Scenario 001 prerequisites met
- [ ] Bitbucket auth available — `git clone` of target repo succeeds standalone
- [ ] Target Bitbucket repo + open PR identified (TBD)

## Action
- [ ] Generate task file with Bitbucket clone URL + PR head SHA:
  ```bash
  cd ~/Documents/workspaces/code-reviewer/agent/pr-reviewer/cmd/run-task
  make generate-dummy-task TASK_FILE=test4-bitbucket.md \
    CLONE_URL=https://bitbucket.example.com/scm/<project>/<repo>.git \
    REF=<bitbucket-pr-head-sha> \
    BASE_REF=<bitbucket-base-branch> \
    PR_URL=https://bitbucket.example.com/projects/<project>/repos/<repo>/pull-requests/<N>
  ```
- [ ] Run agent:
  ```bash
  make run-dummy-task TASK_FILE=test4-bitbucket.md
  ```

## Expected
- [ ] Agent log shows successful clone of Bitbucket URL (no provider-specific failures)
- [ ] Final stdout contains a JSON verdict block
- [ ] Result line shows `{"Status":"done",...}`
- [ ] Process exits 0

## Cleanup
- [ ] `rm test4-bitbucket.md`
- [ ] Remove workdir under `~/.cache/code-reviewer/work/<task-id>/`

## Notes
Currently blocked: no accessible Bitbucket repo + PR identified. Defer until target available.

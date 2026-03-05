---
tags:
  - dark-factory
  - spec
status: draft
---
Tags: [[Dark Factory - Spec Writing Guide]]

---

## Problem

Spec 004 posts Bitbucket reviews as plain comments. Like GitHub (spec 003), Bitbucket Server supports structured review verdicts (approve, needs work) that integrate with merge checks. Without this, the reviewer must manually approve/reject after reading the comment.

## Goal

After completion, `pr-reviewer` parses the Claude review output verdict (same as spec 003) and submits the appropriate Bitbucket Server review action (approve or needs-work) alongside the comment, instead of a plain comment only.

## Non-goals

- Changing the verdict parser (spec 003 handles that)
- Bitbucket Cloud approve/reject (different API)
- Configurable thresholds or scoring weights
- Revoking previous approvals

## Desired Behavior

1. After posting the review comment (spec 004), submit a review verdict via Bitbucket Server REST API
2. Verdict mapping uses the same parser from spec 003:
   - `VerdictApprove` → Bitbucket "APPROVED"
   - `VerdictRequestChanges` → Bitbucket "NEEDS_WORK"
   - `VerdictComment` → no verdict action (comment only, spec 004 fallback)
3. `--comment-only` flag skips verdict (same as GitHub path)
4. Verdict logged to stderr (already done by spec 003 wiring)

## Constraints

- Spec 004 plain-comment behavior must remain the fallback when verdict is unclear
- Must use Bitbucket Server REST API: `PUT /rest/api/1.0/projects/{project}/repos/{repo}/pull-requests/{number}/participants/{userSlug}`
- Same conservative parsing as GitHub — when in doubt, fall back to comment
- `--comment-only` flag applies to both GitHub and Bitbucket

## Failure Modes

| Trigger | Expected behavior | Recovery |
|---------|-------------------|----------|
| Verdict submit fails (API error) | Exit 1 with error, review comment already posted | User approves manually |
| User lacks approve permission | Exit 1 with "insufficient permissions" | User checks Bitbucket permissions |
| PR already approved by this user | Update existing verdict | Automatic |
| `--comment-only` flag set | Skip verdict, post comment only | By design |

## Acceptance Criteria

- [ ] Must Fix items → Bitbucket "NEEDS_WORK" via REST API
- [ ] No Must Fix items → Bitbucket "APPROVED" via REST API
- [ ] Unparseable review → comment only, no verdict action
- [ ] `--comment-only` → comment only, no verdict action
- [ ] Existing GitHub verdict flow unchanged
- [ ] `make precommit` passes

## Verification

```
make precommit
```

## Do-Nothing Option

Keep posting plain comments on Bitbucket (spec 004). Manual approve/reject after reading. Acceptable if merge checks don't require automated approval.

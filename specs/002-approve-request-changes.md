---
tags:
  - dark-factory
  - spec
---
Tags: [[Dark Factory - Prompt Strategy Guide]]

---

## Problem

Spec 001 posts reviews as plain comments. GitHub supports structured review verdicts (approve, request changes) that integrate with branch protection. Currently the reviewer must manually approve/reject after reading the comment.

## Goal

After completion, `pr-reviewer` parses the Claude review output for a pass/fail verdict and submits the review as a GitHub review with the appropriate action (`--approve` or `--request-changes`) instead of a plain comment.

## Non-goals

- Closing/merging PRs based on verdict
- Custom verdict thresholds or scoring
- Multi-reviewer workflows

## Desired Behavior

1. Claude `/code-review` output is analyzed for verdict (pass/fail/needs-work)
2. On pass: `gh pr review --approve --body "<review>"`
3. On fail/needs-work: `gh pr review --request-changes --body "<review>"`
4. Verdict detection must be robust — fallback to plain comment if verdict unclear

## Open Questions

- How does `/code-review` signal pass/fail? Structured output? Keywords? Exit code?
- Should the user be able to override the verdict via CLI flag (`--approve`, `--comment-only`)?

## Acceptance Criteria

- [ ] Clean review → PR approved via GitHub review API
- [ ] Review with issues → PR gets request-changes via GitHub review API
- [ ] Ambiguous verdict → falls back to plain comment (spec 001 behavior)
- [ ] `--comment-only` flag skips verdict and posts as plain comment

## Verification

```
make precommit
```

## Do-Nothing Option

Keep posting plain comments (spec 001). Manual approve/reject. Works fine for single-reviewer workflows.

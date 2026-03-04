---
tags:
  - dark-factory
  - spec
---
Tags: [[Dark Factory - Prompt Strategy Guide]]

---

## Problem

Spec 001 only supports GitHub. Bitbucket Server PRs at `bitbucket.seibert.tools` have no automated review tooling at all.

## Goal

After completion, `pr-reviewer` also accepts Bitbucket Server PR URLs, resolves the repo to a local checkout, runs the same Claude review workflow, and posts the review as a Bitbucket PR comment via REST API.

## Non-goals

- Bitbucket Cloud (different API)
- Approve/reject via Bitbucket API (separate spec if needed)

## Desired Behavior

1. URL parsing detects Bitbucket Server pattern: `https://{host}/projects/{project}/repos/{repo}/pull-requests/{number}`
2. Config lookup uses normalized repo URL
3. PR metadata fetched via Bitbucket Server REST API (`/rest/api/1.0/projects/{project}/repos/{repo}/pull-requests/{number}`)
4. Review posted via Bitbucket Server REST API (`/rest/api/1.0/.../comments`)
5. Auth via token from env var (`BITBUCKET_TOKEN`)

## Config Addition

```yaml
# ~/.pr-reviewer.yaml
repos:
  - url: https://bitbucket.seibert.tools/projects/OC/repos/root
    path: /Users/bborbe/Documents/workspaces/oc-root

bitbucket:
  token: ${BITBUCKET_TOKEN}  # env var reference
```

## URL Pattern

```
https://{host}/projects/{project}/repos/{repo}/pull-requests/{number}/overview
https://{host}/projects/{project}/repos/{repo}/pull-requests/{number}
```

## Open Questions

- Token per host or single global token?
- Bitbucket Server version minimum (API compatibility)?

## Acceptance Criteria

- [ ] `pr-reviewer https://bitbucket.seibert.tools/projects/OC/repos/root/pull-requests/254/overview` produces a review comment
- [ ] Auth failure exits 1 with helpful message
- [ ] Missing BITBUCKET_TOKEN exits 1 with helpful message

## Verification

```
make precommit
```

## Do-Nothing Option

No Bitbucket review automation. Manual reviews only. Acceptable if Bitbucket usage is low.

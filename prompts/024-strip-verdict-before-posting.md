---
status: draft
---
<objective>
Use StripJSONVerdict to remove the JSON verdict block from review text before posting it as a PR comment.
</objective>

<context>
Read CLAUDE.md for project conventions.
Read main.go — specifically `runReview()`, `submitGitHubReview()`, and `submitBitbucketReview()`.
Read pkg/verdict/verdict.go for the `StripJSONVerdict` function (added by prompt 024).
</context>

<requirements>
1. In `runReview()` in main.go, after parsing the verdict, strip the JSON verdict block from the review text:
   - Call `verdict.StripJSONVerdict(reviewText)` to get cleaned text
   - Use the cleaned text for posting comments (passed to submit functions)
   - Keep the original text for stdout printing (user sees full output including verdict JSON)

2. The verdict parsing must happen BEFORE stripping:
   - `result := verdict.Parse(reviewText)` — uses original text
   - `cleanedText := verdict.StripJSONVerdict(reviewText)` — for posting
   - Return `cleanedText` and `result` from `runReview()`

3. Both GitHub and Bitbucket submit functions receive the cleaned text (no changes needed in submit functions themselves).
</requirements>

<constraints>
- stdout output must still show the full review including JSON verdict (for debugging)
- Only the text posted to PR comments is stripped
- If no JSON verdict exists, text is unchanged (StripJSONVerdict is a no-op)
- `--comment-only` flag behavior unchanged
</constraints>

<verification>
Run `make precommit` — must pass.
</verification>

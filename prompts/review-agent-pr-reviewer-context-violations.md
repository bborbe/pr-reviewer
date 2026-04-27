---
status: draft
created: "2026-04-28T12:00:00Z"
---

<summary>
- A cleanup closure in the legacy CLI uses a bare `context.Background()` instead of deriving from the caller's context, discarding trace and value propagation for cleanup operations
- A loop that installs plugins executes network I/O per iteration but has no context-cancellation check, making it impossible to abort early if the context is cancelled between plugin installations
- Three error propagations in the plugin installer discard error context by using bare `return err` without wrapping
- Both issues are in different files but share the same root cause: incomplete context-propagation discipline
</summary>

<objective>
Replace `context.Background()` with `context.WithoutCancel(ctx)` in the CLI cleanup closure so cleanup operations still propagate context values, and add a non-blocking `ctx.Done()` check at the top of the plugin-install loop so cancellation is honored between iterations. Also fix three bare `return err` calls in the plugin package with proper `errors.Wrap`. After this fix, context cancellation reaches every loop iteration and the cleanup retains context values while remaining immune to parent cancellation.
</objective>

<context>
Read `CLAUDE.md` for project conventions.

Files to read before making changes (read ALL first):
- `agent/pr-reviewer/cmd/cli/main.go` — `createCloneAndFetch` function (~line 258), cleanup closure at ~line 280 where `cleanupCtx := context.Background()` is assigned; the parent function receives `ctx context.Context`
- `agent/pr-reviewer/pkg/plugins/plugins.go` — `EnsureInstalled` method (~line 75), `for _, spec := range specs` loop at ~line 81; `ensureOne` call at ~line 83 (bare `return err`); `runHard` calls at ~lines 108, 111 (bare `return err`)
- `agent/pr-reviewer/pkg/plugins/plugins_test.go` — existing tests for `EnsureInstalled`
</context>

<requirements>
1. **`cmd/cli/main.go` — Replace `context.Background()` in the cleanup closure** (~line 280):
   ```go
   // BEFORE
   cleanupCtx := context.Background()
   
   // AFTER — inherits context values (tracing, etc.) but is immune to parent cancellation
   cleanupCtx := context.WithoutCancel(ctx)
   ```
   `context.WithoutCancel` is available since Go 1.21. The module uses Go 1.26.2, so it is available. Add the `context` import if not already present (it likely is).

2. **`pkg/plugins/plugins.go` — Add non-blocking `ctx.Done()` check** at the top of the `for _, spec := range specs` loop body in `EnsureInstalled` (~line 81):
   ```go
   for _, spec := range specs {
       select {
       case <-ctx.Done():
           return errors.Wrap(ctx, ctx.Err(), "context cancelled during plugin install")
       default:
       }
       if err := i.ensureOne(ctx, spec); err != nil {
           return errors.Wrapf(ctx, err, "ensure plugin %s", spec.Name)
       }
   }
   ```

3. **`pkg/plugins/plugins.go` — Fix bare `return err`** in `ensureOne` (~lines 108, 111):
   ```go
   // BEFORE
   return err
   
   // AFTER
   return errors.Wrap(ctx, err, "run plugin install")
   ```
   Use a descriptive message that identifies the operation (e.g. `"list plugins"` for the list call, `"install plugin"` for the install call).

4. **`pkg/plugins/plugins_test.go` — Add a test for context cancellation in `EnsureInstalled`**: create a pre-cancelled context and verify that `EnsureInstalled` returns a non-nil error containing context-cancellation information when the spec list is non-empty.

5. **Run `cd agent/pr-reviewer && make test`** — must pass.
</requirements>

<constraints>
- Only change files in `agent/pr-reviewer/`
- Do NOT commit — dark-factory handles git
- Existing tests must still pass
- Use `errors.Wrapf(ctx, err, "...")` from `github.com/bborbe/errors` (always pass `ctx` as first arg) — never bare `return err`
- `context.WithoutCancel(ctx)` is available in Go 1.21+ — no version gate needed (module uses 1.26.2)
- Code changes MUST add or update tests for changed paths — paths covered before the fix must remain covered, new paths must be tested
</constraints>

<verification>
cd agent/pr-reviewer && make precommit
</verification>

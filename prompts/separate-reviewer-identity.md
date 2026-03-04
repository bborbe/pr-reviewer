# Add separate reviewer identity via GH_TOKEN

Implements spec 004.

## Config Changes

### `pkg/config/config.go`

Add `GitHubConfig` struct and field to `Config`:

```go
type GitHubConfig struct {
    Token string `yaml:"token"`
}

type Config struct {
    GitHub GitHubConfig `yaml:"github"`
    Repos  []RepoConfig `yaml:"repos"`
}
```

### Env var resolution

Add `resolveEnvVar(value string) string` to `pkg/config/config.go`:
- If value matches `${VAR_NAME}` pattern → return `os.Getenv("VAR_NAME")`
- Otherwise → return value as-is
- Called during `Load()` after YAML parsing, before validation

Add `ResolvedGitHubToken() string` method on `Config` that returns the resolved token (empty string if not configured).

## GitHub Client Changes

### `pkg/github/client.go`

Change `NewGHClient` to accept an optional token:

```go
func NewGHClient(token string) Client {
    return &ghClient{token: token}
}

type ghClient struct {
    token string
}
```

In both `GetPRBranch` and `PostComment`, before `cmd.Run()`, inject the token into the subprocess environment if set:

```go
if c.token != "" {
    cmd.Env = append(os.Environ(), "GH_TOKEN="+c.token)
}
```

This overrides `gh` CLI's default auth with the configured token.

## Wiring Changes

### `main.go`

After loading config, pass `cfg.ResolvedGitHubToken()` to `NewGHClient`:

```go
ghClient := github.NewGHClient(cfg.ResolvedGitHubToken())
```

If token is configured but env var is empty, log a warning:
```go
if cfg.GitHub.Token != "" && cfg.ResolvedGitHubToken() == "" {
    slog.Warn("github.token configured but env var is empty, using default gh auth")
}
```

## Tests

### `pkg/config/config_test.go`
- Config without `github` section → `ResolvedGitHubToken()` returns empty string
- Config with `github.token: ${TEST_VAR}` → resolves env var
- Config with `github.token: ${TEST_VAR}` and unset env var → returns empty string
- Config with `github.token: literal-value` → returns literal
- Existing configs without `github` field parse successfully (backward compat)

### `pkg/github/client_test.go`
- `NewGHClient("")` → `cmd.Env` is nil (inherits parent env)
- `NewGHClient("test-token")` → `cmd.Env` contains `GH_TOKEN=test-token`

## Constraints

- Run `make precommit` for validation only
- Do NOT commit, tag, or push (dark-factory handles git)
- Backward compatible — existing configs must keep working
- Token must never appear in log output or error messages
- Follow existing patterns exactly

# CLAUDE.md

CLI tool to review pull requests using local Claude Code with full project context.

## Project Overview

`pr-reviewer` takes a PR URL (GitHub or Bitbucket Server), resolves it to a local repo checkout, creates a git worktree, runs `claude --print "/code-review"`, and posts the review back as a PR comment.

## Development Standards

This project follows Benjamin Borbe's coding guidelines. For comprehensive guidance, see the [coding-guidelines repository](https://github.com/bborbe/coding-guidelines).

**For Claude Code**: Before working on this project, you MUST read ALL markdown files in `~/Documents/workspaces/coding-guidelines/` to understand the complete development patterns and standards.

### Key Reference Guides

- **[go-architecture-patterns.md](~/Documents/workspaces/coding-guidelines/go-architecture-patterns.md)** - Interface -> Constructor -> Struct -> Method pattern
- **[go-testing-guide.md](~/Documents/workspaces/coding-guidelines/go-testing-guide.md)** - Ginkgo v2/Gomega testing patterns
- **[go-makefile-commands.md](~/Documents/workspaces/coding-guidelines/go-makefile-commands.md)** - Standard build commands and workflow
- **[git-commit-workflow.md](~/Documents/workspaces/coding-guidelines/git-commit-workflow.md)** - Mandatory commit process with precommit checks
- **[go-mocking-guide.md](~/Documents/workspaces/coding-guidelines/go-mocking-guide.md)** - Mock generation and usage with Counterfeiter

## Build Commands

```bash
make precommit   # full check: format, generate, test, lint, vet, security
make test        # run tests only
make run ARGS="<pr-url>"  # run the tool
```

// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"
	"fmt"

	agentlib "github.com/bborbe/agent/lib"
	claudelib "github.com/bborbe/agent/lib/claude"
	"github.com/bborbe/errors"

	"github.com/bborbe/code-reviewer/agent/pr-reviewer/pkg/git"
)

// checkoutExecutionStep is the execution phase step that checks out the
// target ref as an on-disk worktree and runs Claude against the real files.
type checkoutExecutionStep struct {
	repoManager     git.RepoManager
	claudeConfigDir claudelib.ClaudeConfigDir
	agentDir        claudelib.AgentDir
	model           claudelib.ClaudeModel
	env             map[string]string
	allowedTools    claudelib.AllowedTools
	instructions    claudelib.Instructions
}

// NewCheckoutExecutionStep constructs the execution-phase step that wires
// RepoManager checkout into the Claude runner working directory.
func NewCheckoutExecutionStep(
	repoManager git.RepoManager,
	claudeConfigDir claudelib.ClaudeConfigDir,
	agentDir claudelib.AgentDir,
	model claudelib.ClaudeModel,
	env map[string]string,
	allowedTools claudelib.AllowedTools,
	instructions claudelib.Instructions,
) agentlib.Step {
	return &checkoutExecutionStep{
		repoManager:     repoManager,
		claudeConfigDir: claudeConfigDir,
		agentDir:        agentDir,
		model:           model,
		env:             env,
		allowedTools:    allowedTools,
		instructions:    instructions,
	}
}

// Name implements agentlib.Step.
func (s *checkoutExecutionStep) Name() string { return "pr-execute" }

// ShouldRun returns false if ## Review already exists (idempotent).
func (s *checkoutExecutionStep) ShouldRun(_ context.Context, md *agentlib.Markdown) (bool, error) {
	_, exists := md.FindSection("## Review")
	return !exists, nil
}

// Run checks out the target ref as a worktree, then runs Claude in that
// directory to produce the ## Review section.
func (s *checkoutExecutionStep) Run(
	ctx context.Context,
	md *agentlib.Markdown,
) (*agentlib.Result, error) {
	cloneURL, _ := md.Frontmatter.String("clone_url")
	ref, _ := md.Frontmatter.String("ref")
	taskID, _ := md.Frontmatter.String("task_identifier")

	if cloneURL == "" {
		return &agentlib.Result{
			Status:  agentlib.AgentStatusFailed,
			Message: "execution step: clone_url is missing from task frontmatter",
		}, nil
	}
	if ref == "" {
		return &agentlib.Result{
			Status:  agentlib.AgentStatusFailed,
			Message: "execution step: ref is missing from task frontmatter",
		}, nil
	}

	worktreePath, err := s.repoManager.EnsureWorktree(ctx, cloneURL, ref, taskID)
	if err != nil {
		return nil, errors.Wrapf(
			ctx,
			err,
			"ensure worktree clone_url=%s ref=%s task_id=%s",
			cloneURL,
			ref,
			taskID,
		)
	}

	runner := claudelib.NewClaudeRunner(claudelib.ClaudeRunnerConfig{
		ClaudeConfigDir:  s.claudeConfigDir,
		AllowedTools:     s.allowedTools,
		Model:            s.model,
		WorkingDirectory: claudelib.AgentDir(worktreePath),
		Env:              s.env,
	})

	taskContent, err := md.Marshal(ctx)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "execution marshal task")
	}

	prompt := claudelib.BuildPrompt(s.instructions.String(), nil, taskContent)

	runResult, runErr := runner.Run(ctx, prompt)
	if runErr != nil {
		return &agentlib.Result{
			Status:  agentlib.AgentStatusFailed,
			Message: fmt.Sprintf("execution claude run failed: %v", runErr),
		}, nil
	}

	md.ReplaceSection(agentlib.Section{
		Heading: "## Review",
		Body:    runResult.Result,
	})

	return &agentlib.Result{
		Status:    agentlib.AgentStatusDone,
		NextPhase: "ai_review",
	}, nil
}

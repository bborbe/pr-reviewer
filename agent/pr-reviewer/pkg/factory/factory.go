// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package factory

import (
	"context"
	"os"

	agentlib "github.com/bborbe/agent/lib"
	claudelib "github.com/bborbe/agent/lib/claude"
	delivery "github.com/bborbe/agent/lib/delivery"
	"github.com/bborbe/cqrs/base"
	"github.com/bborbe/errors"
	libkafka "github.com/bborbe/kafka"
	libtime "github.com/bborbe/time"

	"github.com/bborbe/code-reviewer/agent/pr-reviewer/pkg/prompts"
)

const serviceName = "agent-pr-reviewer"

// allowedTools pins the Claude tools pr-reviewer needs: read files, search,
// invoke git/gh for PR inspection, and fetch web content.
var allowedTools = claudelib.AllowedTools{
	"Read", "Grep", "Glob", "Bash(git:*)", "Bash(gh:*)", "WebFetch",
}

// passthroughEnvKeys lists process env vars forwarded into the Claude CLI
// subprocess. ClaudeRunnerConfig.Env bypasses the allowlist in
// lib/claude/claude-runner.go, so entries here cross the trust boundary —
// keep the list minimal. GH_TOKEN authenticates the gh CLI used by the
// agent to read PRs.
var passthroughEnvKeys = []string{"GH_TOKEN"}

// CreateTaskRunner wires a complete TaskRunner with ClaudeRunner,
// prompt assembly, and result delivery.
func CreateTaskRunner(
	claudeConfigDir claudelib.ClaudeConfigDir,
	agentDir claudelib.AgentDir,
	model claudelib.ClaudeModel,
	deliverer claudelib.ResultDeliverer[claudelib.AgentResult],
) claudelib.TaskRunner[claudelib.AgentResult] {
	return claudelib.NewTaskRunner[claudelib.AgentResult](
		claudelib.NewClaudeRunner(claudelib.ClaudeRunnerConfig{
			ClaudeConfigDir:  claudeConfigDir,
			AllowedTools:     allowedTools,
			Model:            model,
			WorkingDirectory: agentDir,
			Env:              passthroughEnv(passthroughEnvKeys),
		}),
		prompts.BuildInstructions(),
		nil,
		deliverer,
	)
}

// passthroughEnv reads the given keys from the current process environment
// and returns a map of those that are set (non-empty). Pure over os.Getenv
// — testable by manipulating env vars in the test.
func passthroughEnv(keys []string) map[string]string {
	out := map[string]string{}
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			out[k] = v
		}
	}
	return out
}

// CreateSyncProducer creates a Kafka sync producer.
func CreateSyncProducer(
	ctx context.Context,
	brokers libkafka.Brokers,
) (libkafka.SyncProducer, error) {
	producer, err := libkafka.NewSyncProducerWithName(ctx, brokers, serviceName)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "create sync producer failed")
	}
	return producer, nil
}

// CreateKafkaResultDeliverer creates a ResultDeliverer that publishes task updates to Kafka.
func CreateKafkaResultDeliverer(
	syncProducer libkafka.SyncProducer,
	branch base.Branch,
	taskID agentlib.TaskIdentifier,
	taskContent string,
	currentDateTime libtime.CurrentDateTimeGetter,
) claudelib.ResultDeliverer[claudelib.AgentResult] {
	return claudelib.NewResultDelivererAdapter[claudelib.AgentResult](
		delivery.NewKafkaResultDeliverer(
			syncProducer,
			branch,
			taskID,
			taskContent,
			delivery.NewFallbackContentGenerator(),
			currentDateTime,
		),
	)
}

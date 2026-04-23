// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package factory

import (
	"context"

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

// CreateTaskRunner wires a complete TaskRunner with ClaudeRunner,
// prompt assembly, and result delivery. ghToken is forwarded as GH_TOKEN
// into the Claude CLI subprocess env (bypasses the lib/claude allowlist)
// so the gh CLI can authenticate for PR reads.
func CreateTaskRunner(
	claudeConfigDir claudelib.ClaudeConfigDir,
	agentDir claudelib.AgentDir,
	model claudelib.ClaudeModel,
	ghToken string,
	deliverer claudelib.ResultDeliverer[claudelib.AgentResult],
) claudelib.TaskRunner[claudelib.AgentResult] {
	env := map[string]string{}
	if ghToken != "" {
		env["GH_TOKEN"] = ghToken
	}
	return claudelib.NewTaskRunner[claudelib.AgentResult](
		claudelib.NewClaudeRunner(claudelib.ClaudeRunnerConfig{
			ClaudeConfigDir:  claudeConfigDir,
			AllowedTools:     allowedTools,
			Model:            model,
			WorkingDirectory: agentDir,
			Env:              env,
		}),
		prompts.BuildInstructions(),
		nil,
		deliverer,
	)
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

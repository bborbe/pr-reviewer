// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"os"

	agentlib "github.com/bborbe/agent/lib"
	claudelib "github.com/bborbe/agent/lib/claude"
	"github.com/bborbe/cqrs/base"
	"github.com/bborbe/errors"
	libkafka "github.com/bborbe/kafka"
	libsentry "github.com/bborbe/sentry"
	"github.com/bborbe/service"
	libtime "github.com/bborbe/time"
	"github.com/golang/glog"

	"github.com/bborbe/code-reviewer/agent/pr-reviewer/pkg/factory"
)

func main() {
	app := &application{}
	os.Exit(service.Main(context.Background(), app, &app.SentryDSN, &app.SentryProxy))
}

type application struct {
	SentryDSN   string `required:"false" arg:"sentry-dsn"   env:"SENTRY_DSN"   usage:"SentryDSN"    display:"length"`
	SentryProxy string `required:"false" arg:"sentry-proxy" env:"SENTRY_PROXY" usage:"Sentry Proxy"`

	// Claude Code CLI configuration
	ClaudeConfigDir claudelib.ClaudeConfigDir `required:"false" arg:"claude-config-dir" env:"CLAUDE_CONFIG_DIR" usage:"Claude Code config directory"`

	// Agent directory (contains .claude/ with CLAUDE.md and commands)
	AgentDir claudelib.AgentDir `required:"false" arg:"agent-dir" env:"AGENT_DIR" usage:"Agent directory with .claude/ config" default:"agent"`

	// Model selection
	Model claudelib.ClaudeModel `required:"false" arg:"model" env:"MODEL" usage:"Claude model to use (sonnet, opus)" default:"sonnet"`

	// Branch for Kafka result delivery
	Branch base.Branch `required:"true" arg:"branch" env:"BRANCH" usage:"branch"`

	// Kafka delivery (optional — only active when TASK_ID is set)
	KafkaBrokers libkafka.Brokers `required:"false" arg:"kafka-brokers" env:"KAFKA_BROKERS" usage:"Comma separated list of Kafka brokers"`

	TaskID agentlib.TaskIdentifier `required:"false" arg:"task-id" env:"TASK_ID" usage:"Agent task identifier for publishing results back to task controller"`

	// Task content from agent pipeline
	TaskContent string `required:"true" arg:"task-content" env:"TASK_CONTENT" usage:"Raw task markdown from vault"`

	// GitHub token forwarded to the Claude CLI subprocess as GH_TOKEN for gh auth.
	GHToken string `required:"false" arg:"gh-token" env:"GH_TOKEN" usage:"GitHub token for gh CLI auth" display:"length"`
}

func (a *application) Run(ctx context.Context, sentryClient libsentry.Client) error {
	glog.V(2).Infof("agent-pr-reviewer started")

	deliverer, cleanup, err := a.createDeliverer(ctx)
	if err != nil {
		return errors.Wrap(ctx, err, "create deliverer")
	}
	defer cleanup()

	taskRunner := factory.CreateTaskRunner(
		a.ClaudeConfigDir,
		a.AgentDir,
		a.Model,
		a.GHToken,
		deliverer,
	)

	result, err := taskRunner.Run(ctx, a.TaskContent)
	if err != nil {
		return claudelib.PrintResult(ctx, claudelib.AgentResult{
			Status:  claudelib.AgentStatusFailed,
			Message: fmt.Sprintf("task runner failed: %v", err),
		})
	}

	return claudelib.PrintResult(ctx, *result)
}

func (a *application) createDeliverer(
	ctx context.Context,
) (claudelib.ResultDeliverer[claudelib.AgentResult], func(), error) {
	if a.TaskID != "" {
		if len(a.KafkaBrokers) == 0 {
			return nil, nil, errors.Errorf(ctx, "KAFKA_BROKERS must be set when TASK_ID is set")
		}
		syncProducer, err := factory.CreateSyncProducer(ctx, a.KafkaBrokers)
		if err != nil {
			return nil, nil, errors.Wrap(ctx, err, "create sync producer failed")
		}
		deliverer := factory.CreateKafkaResultDeliverer(
			syncProducer,
			a.Branch,
			a.TaskID,
			a.TaskContent,
			libtime.NewCurrentDateTime(),
		)
		return deliverer, func() {
			if err := syncProducer.Close(); err != nil {
				glog.Warningf("close sync producer failed: %v", err)
			}
		}, nil
	}
	glog.V(2).Infof("TASK_ID not set, skipping task result publishing")
	return claudelib.NewNoopResultDeliverer(), func() {}, nil
}

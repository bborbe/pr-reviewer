// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prompts

import (
	_ "embed"
	"strings"

	claudelib "github.com/bborbe/agent/lib/claude"
)

//go:embed execution_workflow.md
var executionWorkflow string

//go:embed execution_output-format.md
var executionOutputFormat string

// BuildExecutionInstructions assembles the execution-phase prompt from embedded
// modules, injecting the configured review mode into the slash-command invocation.
func BuildExecutionInstructions(reviewMode string) claudelib.Instructions {
	workflow := strings.ReplaceAll(executionWorkflow, "{{REVIEW_MODE}}", reviewMode)
	return claudelib.Instructions{
		{Name: "workflow", Content: workflow},
		{Name: "output-format", Content: executionOutputFormat},
	}
}

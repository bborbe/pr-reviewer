// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package review

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Reviewer runs a code review using the claude CLI.
//
//counterfeiter:generate -o ../../mocks/reviewer.go --fake-name Reviewer . Reviewer
type Reviewer interface {
	Review(ctx context.Context, worktreePath string, command string, model string) (string, error)
}

// NewDockerReviewer creates a Reviewer that invokes claude inside a Docker container.
func NewDockerReviewer(containerImage string) Reviewer {
	return &dockerReviewer{containerImage: containerImage}
}

type dockerReviewer struct {
	containerImage string
}

// Review runs claude inside a Docker container using the claude-yolo image.
// The container mounts the worktree, Claude config, and Go module cache.
func (r *dockerReviewer) Review(
	ctx context.Context,
	worktreePath string,
	command string,
	model string,
) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	// Build docker command
	// #nosec G204 -- paths from context, containerImage from config
	cmd := exec.CommandContext(
		ctx,
		"docker",
		"run",
		"--rm",
		"--cap-add=NET_ADMIN",
		"--cap-add=NET_RAW",
		"-w", "/workspace",
		"-v", fmt.Sprintf("%s:/workspace", worktreePath),
		"-v", fmt.Sprintf("%s/.claude-yolo:/home/node/.claude", home),
		"-v", fmt.Sprintf("%s/go/pkg:/home/node/go/pkg", home),
		r.containerImage,
		"claude",
		"--print",
		"--model", model,
		command,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("claude review failed: %s", strings.TrimSpace(stderr.String()))
	}

	return stdout.String(), nil
}

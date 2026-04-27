// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cursor

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/bborbe/errors"
	"github.com/golang/glog"
)

// DefaultCursorPath is the default path for cursor state persistence.
const DefaultCursorPath = "/data/cursor.json"

// State holds the watcher's persisted cursor.
type State struct {
	LastUpdatedAt time.Time         `json:"last_updated_at"`
	HeadSHAs      map[string]string `json:"head_shas"`
}

// Load reads cursor state from path.
// Returns cold-start state with startTime if the file is missing or corrupt.
func Load(ctx context.Context, path string, startTime time.Time) (State, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is config-controlled
	if err != nil {
		if os.IsNotExist(err) {
			glog.V(2).
				Infof("cursor file not found, using cold-start time=%s", startTime.Format(time.RFC3339))
			return State{LastUpdatedAt: startTime, HeadSHAs: make(map[string]string)}, nil
		}
		return State{}, errors.Wrapf(ctx, err, "read cursor file path=%s", path)
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		glog.Warningf("cursor file corrupt, using cold-start path=%s err=%v", path, err)
		return State{LastUpdatedAt: startTime, HeadSHAs: make(map[string]string)}, nil
	}
	if state.HeadSHAs == nil {
		state.HeadSHAs = make(map[string]string)
	}
	return state, nil
}

// Save persists cursor state to path atomically via a temp file + rename.
func Save(ctx context.Context, path string, state State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return errors.Wrapf(ctx, err, "marshal cursor state path=%s", path)
	}
	if err := os.WriteFile(path+".tmp", data, 0600); err != nil { // #nosec G306 -- intentional 0600
		return errors.Wrapf(ctx, err, "save cursor path=%s", path)
	}
	if err := os.Rename(path+".tmp", path); err != nil {
		return errors.Wrapf(ctx, err, "save cursor path=%s", path)
	}
	return nil
}

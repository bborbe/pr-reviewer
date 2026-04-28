// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import "strings"

// IsBotAuthor returns true if the PR author login matches any allowlist entry (exact match).
func IsBotAuthor(pr PullRequest, allowlist []string) bool {
	for _, entry := range allowlist {
		if pr.AuthorLogin == entry {
			return true
		}
	}
	return false
}

// ShouldSkipPR returns true if the PR should be filtered out because it is a draft or bot-authored.
func ShouldSkipPR(pr PullRequest, botAllowlist []string) bool {
	return pr.IsDraft || IsBotAuthor(pr, botAllowlist)
}

// ParseBotAllowlist splits a comma-separated allowlist string into a slice of trimmed, non-empty entries.
func ParseBotAllowlist(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

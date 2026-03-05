// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package verdict

import (
	"regexp"
	"strings"
)

// Verdict represents the review verdict type
type Verdict string

const (
	VerdictApprove        Verdict = "approve"
	VerdictRequestChanges Verdict = "request-changes"
	VerdictComment        Verdict = "comment"
)

// Result holds the verdict and reason
type Result struct {
	Verdict Verdict
	Reason  string
}

// Parse analyzes Claude review output and determines the appropriate verdict
func Parse(reviewText string) Result {
	if reviewText == "" {
		return Result{
			Verdict: VerdictComment,
			Reason:  "empty review text",
		}
	}

	// Check if review has expected sections
	hasReviewSections := hasExpectedReviewSections(reviewText)

	// Look for Must Fix section (case-insensitive, h2 or h3)
	mustFixPattern := regexp.MustCompile(`(?i)^##+ Must Fix`)
	lines := strings.Split(reviewText, "\n")

	mustFixIndex := -1
	for i, line := range lines {
		if mustFixPattern.MatchString(strings.TrimSpace(line)) {
			mustFixIndex = i
			break
		}
	}

	// No Must Fix section found
	if mustFixIndex == -1 {
		if hasReviewSections {
			return Result{
				Verdict: VerdictApprove,
				Reason:  "no must-fix section",
			}
		}
		return Result{
			Verdict: VerdictComment,
			Reason:  "unparseable review format",
		}
	}

	// Must Fix section found - check if it has content
	hasMustFixContent := checkMustFixContent(lines, mustFixIndex)

	if hasMustFixContent {
		return Result{
			Verdict: VerdictRequestChanges,
			Reason:  "must-fix items found",
		}
	}

	return Result{
		Verdict: VerdictApprove,
		Reason:  "no must-fix items",
	}
}

// hasExpectedReviewSections checks if the review has expected sections like Should Fix or Nice to Have
func hasExpectedReviewSections(reviewText string) bool {
	text := strings.ToLower(reviewText)
	return strings.Contains(text, "should fix") || strings.Contains(text, "nice to have")
}

// checkMustFixContent determines if Must Fix section has actual content (not just "None")
func checkMustFixContent(lines []string, mustFixIndex int) bool {
	nextHeadingPattern := regexp.MustCompile(`^##+ `)
	hasNonEmptyContent := false

	// Look at lines after the Must Fix heading
	for i := mustFixIndex + 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Stop at next heading
		if nextHeadingPattern.MatchString(line) {
			break
		}

		// Skip empty lines and horizontal rules
		if line == "" || isHorizontalRule(line) {
			continue
		}

		// Check if line is a "None" indicator (case-insensitive)
		if isNoneIndicator(line) {
			continue
		}

		// Found non-empty, non-"None" content
		hasNonEmptyContent = true
		break
	}

	return hasNonEmptyContent
}

// isHorizontalRule checks if a line is a markdown horizontal rule (---, ***, ___).
func isHorizontalRule(line string) bool {
	cleaned := strings.TrimSpace(line)
	if len(cleaned) < 3 {
		return false
	}
	for _, ch := range cleaned {
		if ch != '-' && ch != '*' && ch != '_' {
			return false
		}
	}
	return true
}

// isNoneIndicator checks if a line indicates "no issues" (e.g., "None", "*None*", "None identified.")
func isNoneIndicator(line string) bool {
	// Remove markdown formatting and punctuation
	cleaned := strings.Trim(line, "*_~` .")
	cleaned = strings.ToLower(strings.TrimSpace(cleaned))

	// Check for common "none" patterns
	if cleaned == "none" {
		return true
	}
	if strings.HasPrefix(cleaned, "none ") {
		return true
	}
	if strings.Contains(cleaned, "no issues") {
		return true
	}

	return false
}

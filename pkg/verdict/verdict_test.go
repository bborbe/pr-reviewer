// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package verdict_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/pr-reviewer/pkg/verdict"
)

var _ = Describe("Parse", func() {
	var (
		reviewText string
		result     verdict.Result
	)

	JustBeforeEach(func() {
		result = verdict.Parse(reviewText)
	})

	Context("empty review text", func() {
		BeforeEach(func() {
			reviewText = ""
		})

		It("returns VerdictComment", func() {
			Expect(result.Verdict).To(Equal(verdict.VerdictComment))
		})

		It("returns reason 'empty review text'", func() {
			Expect(result.Reason).To(Equal("empty review text"))
		})
	})

	Context("review with no recognizable sections", func() {
		BeforeEach(func() {
			reviewText = "This is just some random text without any sections."
		})

		It("returns VerdictComment", func() {
			Expect(result.Verdict).To(Equal(verdict.VerdictComment))
		})

		It("returns reason 'unparseable review format'", func() {
			Expect(result.Reason).To(Equal("unparseable review format"))
		})
	})

	Context("review with Must Fix section containing items", func() {
		BeforeEach(func() {
			reviewText = `### Must Fix (Critical)

- Security issue in authentication
- SQL injection vulnerability`
		})

		It("returns VerdictRequestChanges", func() {
			Expect(result.Verdict).To(Equal(verdict.VerdictRequestChanges))
		})

		It("returns reason 'must-fix items found'", func() {
			Expect(result.Reason).To(Equal("must-fix items found"))
		})
	})

	Context("review with Must Fix section saying *None*", func() {
		BeforeEach(func() {
			reviewText = `### Must Fix (Critical)

*None*

### Should Fix (Important)

- Add error handling`
		})

		It("returns VerdictApprove", func() {
			Expect(result.Verdict).To(Equal(verdict.VerdictApprove))
		})

		It("returns reason 'no must-fix items'", func() {
			Expect(result.Reason).To(Equal("no must-fix items"))
		})
	})

	Context("review with Must Fix section saying None identified", func() {
		BeforeEach(func() {
			reviewText = `### Must Fix (Critical)

None identified.

### Should Fix (Important)

- Improve error messages`
		})

		It("returns VerdictApprove", func() {
			Expect(result.Verdict).To(Equal(verdict.VerdictApprove))
		})

		It("returns reason 'no must-fix items'", func() {
			Expect(result.Reason).To(Equal("no must-fix items"))
		})
	})

	Context("review with Must Fix section saying none (lowercase)", func() {
		BeforeEach(func() {
			reviewText = `### Must Fix (Critical)

none

### Nice to Have (Optional)

- Add docstrings`
		})

		It("returns VerdictApprove", func() {
			Expect(result.Verdict).To(Equal(verdict.VerdictApprove))
		})

		It("returns reason 'no must-fix items'", func() {
			Expect(result.Reason).To(Equal("no must-fix items"))
		})
	})

	Context("review with Must Fix section that is empty", func() {
		BeforeEach(func() {
			reviewText = `### Must Fix (Critical)


### Should Fix (Important)

- Add tests`
		})

		It("returns VerdictApprove", func() {
			Expect(result.Verdict).To(Equal(verdict.VerdictApprove))
		})

		It("returns reason 'no must-fix items'", func() {
			Expect(result.Reason).To(Equal("no must-fix items"))
		})
	})

	Context("review with only Should Fix and Nice to Have sections", func() {
		BeforeEach(func() {
			reviewText = `### Should Fix (Important)

- Add error handling
- Improve logging

### Nice to Have (Optional)

- Add docstrings
- Refactor for clarity`
		})

		It("returns VerdictApprove", func() {
			Expect(result.Verdict).To(Equal(verdict.VerdictApprove))
		})

		It("returns reason 'no must-fix section'", func() {
			Expect(result.Reason).To(Equal("no must-fix section"))
		})
	})

	Context("review with only Should Fix section", func() {
		BeforeEach(func() {
			reviewText = `### Should Fix (Important)

- Add error handling`
		})

		It("returns VerdictApprove", func() {
			Expect(result.Verdict).To(Equal(verdict.VerdictApprove))
		})

		It("returns reason 'no must-fix section'", func() {
			Expect(result.Reason).To(Equal("no must-fix section"))
		})
	})

	Context("review with only Nice to Have section", func() {
		BeforeEach(func() {
			reviewText = `### Nice to Have (Optional)

- Add comments`
		})

		It("returns VerdictApprove", func() {
			Expect(result.Verdict).To(Equal(verdict.VerdictApprove))
		})

		It("returns reason 'no must-fix section'", func() {
			Expect(result.Reason).To(Equal("no must-fix section"))
		})
	})

	Context("review with h2 Must Fix instead of h3", func() {
		BeforeEach(func() {
			reviewText = `## Must Fix (Critical)

- Critical security flaw`
		})

		It("returns VerdictRequestChanges", func() {
			Expect(result.Verdict).To(Equal(verdict.VerdictRequestChanges))
		})

		It("returns reason 'must-fix items found'", func() {
			Expect(result.Reason).To(Equal("must-fix items found"))
		})
	})

	Context("review with h2 Must Fix saying None", func() {
		BeforeEach(func() {
			reviewText = `## Must Fix (Critical)

None

## Should Fix (Important)

- Add tests`
		})

		It("returns VerdictApprove", func() {
			Expect(result.Verdict).To(Equal(verdict.VerdictApprove))
		})

		It("returns reason 'no must-fix items'", func() {
			Expect(result.Reason).To(Equal("no must-fix items"))
		})
	})

	Context("review with case variations in Must Fix header", func() {
		BeforeEach(func() {
			reviewText = `### MUST FIX (Critical)

- Issue found`
		})

		It("returns VerdictRequestChanges", func() {
			Expect(result.Verdict).To(Equal(verdict.VerdictRequestChanges))
		})
	})

	Context("review with Must Fix containing 'No issues found' text", func() {
		BeforeEach(func() {
			reviewText = `### Must Fix (Critical)

No issues found.

### Should Fix (Important)

- Improve code structure`
		})

		It("returns VerdictApprove", func() {
			Expect(result.Verdict).To(Equal(verdict.VerdictApprove))
		})

		It("returns reason 'no must-fix items'", func() {
			Expect(result.Reason).To(Equal("no must-fix items"))
		})
	})

	Context("review with Must Fix at end of document", func() {
		BeforeEach(func() {
			reviewText = `### Should Fix (Important)

- Add tests

### Must Fix (Critical)

- Security vulnerability`
		})

		It("returns VerdictRequestChanges", func() {
			Expect(result.Verdict).To(Equal(verdict.VerdictRequestChanges))
		})

		It("returns reason 'must-fix items found'", func() {
			Expect(result.Reason).To(Equal("must-fix items found"))
		})
	})

	Context("complex review with multiple sections", func() {
		BeforeEach(func() {
			reviewText = `# Code Review

## Summary

This PR adds new features.

### Must Fix (Critical)

*None*

### Should Fix (Important)

- Add error handling in main.go:45
- Missing input validation

### Nice to Have (Optional)

- Add docstrings
- Refactor for clarity

## Conclusion

Overall good work!`
		})

		It("returns VerdictApprove", func() {
			Expect(result.Verdict).To(Equal(verdict.VerdictApprove))
		})

		It("returns reason 'no must-fix items'", func() {
			Expect(result.Reason).To(Equal("no must-fix items"))
		})
	})
})

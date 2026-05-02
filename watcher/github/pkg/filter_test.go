// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/code-reviewer/watcher/github/pkg"
)

var _ = Describe("Filter", func() {
	Describe("ShouldSkip", func() {
		DescribeTable("skipping rules",
			func(pr pkg.PullRequest, allowlist []string, expected bool) {
				Expect(pkg.ShouldSkipPR(pr, allowlist)).To(Equal(expected))
			},
			Entry("draft PR, empty allowlist → skipped",
				pkg.PullRequest{IsDraft: true, AuthorLogin: "alice"},
				[]string{},
				true,
			),
			Entry("non-draft PR, empty allowlist → not skipped",
				pkg.PullRequest{IsDraft: false, AuthorLogin: "alice"},
				[]string{},
				false,
			),
			Entry("non-draft PR, author in allowlist → skipped",
				pkg.PullRequest{IsDraft: false, AuthorLogin: "dependabot[bot]"},
				[]string{"dependabot[bot]"},
				true,
			),
			Entry("non-draft PR, author NOT in allowlist → not skipped",
				pkg.PullRequest{IsDraft: false, AuthorLogin: "alice"},
				[]string{"dependabot[bot]", "renovate[bot]"},
				false,
			),
			Entry("draft PR, author in allowlist → skipped (both conditions true)",
				pkg.PullRequest{IsDraft: true, AuthorLogin: "dependabot[bot]"},
				[]string{"dependabot[bot]"},
				true,
			),
			Entry("case sensitivity: Dependabot[bot] does NOT match dependabot[bot]",
				pkg.PullRequest{IsDraft: false, AuthorLogin: "dependabot[bot]"},
				[]string{"Dependabot[bot]"},
				false,
			),
		)
	})

	Describe("pkg.IsBotAuthor", func() {
		It("returns false for empty allowlist", func() {
			pr := pkg.PullRequest{AuthorLogin: "alice"}
			Expect(pkg.IsBotAuthor(pr, nil)).To(BeFalse())
		})

		It("returns true for exact match", func() {
			pr := pkg.PullRequest{AuthorLogin: "renovate[bot]"}
			Expect(
				pkg.IsBotAuthor(pr, []string{"dependabot[bot]", "renovate[bot]"}),
			).To(BeTrue())
		})

		It("returns false when no entry matches", func() {
			pr := pkg.PullRequest{AuthorLogin: "alice"}
			Expect(
				pkg.IsBotAuthor(pr, []string{"dependabot[bot]", "renovate[bot]"}),
			).To(BeFalse())
		})
	})

	Describe("ParseBotAllowlist", func() {
		It("returns nil for empty string", func() {
			Expect(pkg.ParseBotAllowlist("")).To(BeNil())
		})

		It("returns single entry", func() {
			Expect(pkg.ParseBotAllowlist("dependabot[bot]")).To(Equal([]string{"dependabot[bot]"}))
		})

		It("returns multiple comma-separated entries", func() {
			Expect(
				pkg.ParseBotAllowlist("dependabot[bot],renovate[bot]"),
			).To(Equal([]string{"dependabot[bot]", "renovate[bot]"}))
		})

		It("trims leading and trailing whitespace from entries", func() {
			Expect(
				pkg.ParseBotAllowlist(" dependabot[bot] , renovate[bot] "),
			).To(Equal([]string{"dependabot[bot]", "renovate[bot]"}))
		})

		It("filters out entries that are only whitespace after trimming", func() {
			Expect(
				pkg.ParseBotAllowlist("dependabot[bot],   ,renovate[bot]"),
			).To(Equal([]string{"dependabot[bot]", "renovate[bot]"}))
		})

		It("filters out trailing empty entry from trailing comma", func() {
			Expect(pkg.ParseBotAllowlist("dependabot[bot],")).To(Equal([]string{"dependabot[bot]"}))
		})
	})

	Describe("ParseTrustedAuthors", func() {
		It("returns nil for empty string", func() {
			Expect(pkg.ParseTrustedAuthors("")).To(BeNil())
		})

		It("returns single entry", func() {
			Expect(pkg.ParseTrustedAuthors("alice")).To(Equal([]string{"alice"}))
		})

		It("returns multiple comma-separated entries trimmed", func() {
			Expect(
				pkg.ParseTrustedAuthors("alice , bob"),
			).To(Equal([]string{"alice", "bob"}))
		})

		It("filters out whitespace-only entries", func() {
			Expect(
				pkg.ParseTrustedAuthors("alice,  ,bob"),
			).To(Equal([]string{"alice", "bob"}))
		})

		It("filters out trailing empty entry from trailing comma", func() {
			Expect(pkg.ParseTrustedAuthors("alice,")).To(Equal([]string{"alice"}))
		})
	})
})

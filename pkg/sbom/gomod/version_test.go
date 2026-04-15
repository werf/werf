package gomod

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("Version resolution", func() {
	Describe("ResolveVersionFromTags", func() {
		DescribeTable(
			"resolves version from tags",
			func(tags []string, commitMap map[string]string, targetCommit, expectedVersion string, expectedErrMatcher types.GomegaMatcher) {
				commitForTag := func(tag string) (string, error) {
					commit, ok := commitMap[tag]
					if !ok {
						return "", fmt.Errorf("tag %q not found", tag)
					}
					return commit, nil
				}

				version, err := ResolveVersionFromTags(tags, commitForTag, targetCommit)
				Expect(err).To(expectedErrMatcher)
				if err == nil {
					Expect(version).To(Equal(expectedVersion))
				}
			},

			Entry(
				"tag matches commit",
				[]string{"v1.2.3", "v1.2.4"},
				map[string]string{"v1.2.3": "abc1234", "v1.2.4": "def5678"},
				"abc1234",
				"v1.2.3",
				Succeed(),
			),
			Entry(
				"no tag matches",
				[]string{"v1.2.3", "v1.2.4"},
				map[string]string{"v1.2.3": "abc1234", "v1.2.4": "def5678"},
				"zzz9999",
				"v0.0.0-zzz9999",
				Succeed(),
			),
			Entry(
				"non-semver tags skipped",
				[]string{"latest", "v1.2.3", "release"},
				map[string]string{"latest": "abc1234", "v1.2.3": "def5678", "release": "abc1234"},
				"abc1234",
				"v0.0.0-abc1234",
				Succeed(),
			),
			Entry(
				"empty tags list",
				[]string{},
				map[string]string{},
				"abc1234",
				"v0.0.0-abc1234",
				Succeed(),
			),
			Entry(
				"commitForTag error propagated",
				[]string{"v1.2.3"},
				map[string]string{},
				"abc1234",
				"",
				MatchError(ContainSubstring("resolve tag \"v1.2.3\" commit")),
			),
			Entry(
				"short commit (< 7 chars)",
				[]string{},
				map[string]string{},
				"abc",
				"v0.0.0-abc",
				Succeed(),
			),
			Entry(
				"long commit truncated to 7 chars",
				[]string{},
				map[string]string{},
				"abcdefghijklmnop",
				"v0.0.0-abcdefg",
				Succeed(),
			),
		)
	})
})

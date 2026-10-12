package stage

import (
	"encoding/json"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/config"
)

var _ = ginkgo.Describe("Git sources metadata", func() {
	ginkgo.DescribeTable("sanitizes URLs and preserves JSON values",
		func(ctx ginkgo.SpecContext, origin, expected string, local bool) {
			repo := &gitSourcesRepoStub{local: local, headCommitHash: "resolved-sha", origin: origin}
			mapping := NewGitMapping()
			mapping.SetGitRepo(repo)
			mapping.Add, mapping.To = "src", "/path/with'quotes\"\n$(exit 1)"
			mapping.Tag = "tag'with\"quotes"
			data, err := GitSourcesJSON(ctx, []*GitMapping{mapping})
			gomega.Expect(err).To(gomega.Succeed())
			var records []map[string]interface{}
			gomega.Expect(json.Unmarshal([]byte(data), &records)).To(gomega.Succeed())
			gomega.Expect(records).To(gomega.HaveLen(1))
			gomega.Expect(records[0]).To(gomega.Equal(map[string]interface{}{
				"source": expected, "tag": mapping.Tag, "branch": nil, "commit": "resolved-sha",
				"commitDate": "0001-01-01T00:00:00Z", "add": mapping.Add, "to": mapping.To,
			}))
		},
		ginkgo.Entry("HTTPS", "https://user:secret@example.com/repo.git?token=secret#secret", "https://example.com/repo.git", false),
		ginkgo.Entry("SSH URL", "ssh://git:secret@example.com:2222/repo.git", "ssh://example.com:2222/repo.git", false),
		ginkgo.Entry("SCP URL", "git@example.com:repo.git", "example.com:repo.git", false),
		ginkgo.Entry("SCP path with at sign", "git@example.com:team/repo@version.git", "example.com:team/repo@version.git", false),
		ginkgo.Entry("file path with at sign", "/srv/repo@version.git", "/srv/repo@version.git", false),
		ginkgo.Entry("local", "must-not-be-read", ".", true),
	)

	ginkgo.It("keeps multiple mappings and explicit branches without guessing refs", func(ctx ginkgo.SpecContext) {
		repo := &gitSourcesRepoStub{headCommitHash: "resolved-sha", origin: "https://example.com/repo.git"}
		first, second := NewGitMapping(), NewGitMapping()
		first.SetGitRepo(repo)
		second.SetGitRepo(repo)
		first.To, second.To = "/one", "/two"
		first.Branch = "main"
		second.Commit = "pinned-sha"
		data, err := GitSourcesJSON(ctx, []*GitMapping{first, second, first})
		gomega.Expect(err).To(gomega.Succeed())
		var records []map[string]interface{}
		gomega.Expect(json.Unmarshal([]byte(data), &records)).To(gomega.Succeed())
		gomega.Expect(records).To(gomega.HaveLen(3))
		gomega.Expect(records).To(gomega.ContainElement(gomega.HaveKeyWithValue("branch", "main")))
		gomega.Expect(records).To(gomega.ContainElement(gomega.And(gomega.HaveKeyWithValue("commit", "pinned-sha"), gomega.HaveKeyWithValue("branch", gomega.BeNil()), gomega.HaveKeyWithValue("tag", gomega.BeNil()))))
		reordered, err := GitSourcesJSON(ctx, []*GitMapping{second, first, first})
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(reordered).To(gomega.Equal(data))
	})

	ginkgo.DescribeTable("preserves existing cache rules for source changes",
		func(ctx ginkgo.SpecContext, field string) {
			repo := &gitSourcesRepoStub{headCommitHash: "first", origin: "https://example.com/repo.git", checksum: "unchanged"}
			mapping := NewGitMapping()
			mapping.SetGitRepo(repo)
			stg := NewGitArchiveStage(&NewGitArchiveStageOptions{}, &BaseStageOptions{})
			stg.SetGitMappings([]*GitMapping{mapping})
			before, err := stg.GetDependencies(ctx, nil, nil, nil, nil, nil)
			gomega.Expect(err).To(gomega.Succeed())
			beforeInstall := GenerateBeforeInstallStage(ctx, &config.StapelImageBase{Shell: &config.Shell{BeforeInstall: []string{"echo \"$WERF_GIT_SOURCES\""}}}, &BaseStageOptions{})
			beforeInstall.SetGitMappings([]*GitMapping{mapping})
			beforeUser, err := beforeInstall.GetDependencies(ctx, nil, nil, nil, nil, nil)
			gomega.Expect(err).To(gomega.Succeed())

			switch field {
			case "commit":
				repo.headCommitHash = "second"
			case "date":
				repo.date = time.Unix(1234567890, 0)
			case "tag":
				mapping.Tag = "v2"
			case "branch":
				mapping.Branch = "release"
			}
			after, err := stg.GetDependencies(ctx, nil, nil, nil, nil, nil)
			gomega.Expect(err).To(gomega.Succeed())
			if field == "tag" || field == "branch" {
				gomega.Expect(after).NotTo(gomega.Equal(before))
			} else {
				gomega.Expect(after).To(gomega.Equal(before))
			}

			afterUser, err := beforeInstall.GetDependencies(ctx, nil, nil, nil, nil, nil)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(afterUser).To(gomega.Equal(beforeUser))
		},
		ginkgo.Entry("commit", "commit"), ginkgo.Entry("date", "date"), ginkgo.Entry("tag", "tag"), ginkgo.Entry("branch", "branch"),
	)
})

package true_git

import (
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
)

var _ = Describe("parseLsRemoteTagsOutput", func() {
	const (
		shaLight     = "1111111111111111111111111111111111111111"
		shaAnnotObj  = "2222222222222222222222222222222222222222"
		shaAnnotPeel = "3333333333333333333333333333333333333333"
	)

	DescribeTable("parses ls-remote --tags output",
		func(input string, expected map[string]RemoteTagRef) {
			got, err := parseLsRemoteTagsOutput(input)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(Equal(expected))
		},
		Entry("empty output", "", map[string]RemoteTagRef{}),
		Entry("lightweight tag only",
			shaLight+"\trefs/tags/light\n",
			map[string]RemoteTagRef{
				"light": {ObjectSHA: shaLight},
			},
		),
		Entry("annotated tag with peel line",
			shaAnnotObj+"\trefs/tags/annot\n"+shaAnnotPeel+"\trefs/tags/annot^{}\n",
			map[string]RemoteTagRef{
				"annot": {ObjectSHA: shaAnnotObj, PeeledSHA: shaAnnotPeel},
			},
		),
		Entry("peel line before object line",
			shaAnnotPeel+"\trefs/tags/annot^{}\n"+shaAnnotObj+"\trefs/tags/annot\n",
			map[string]RemoteTagRef{
				"annot": {ObjectSHA: shaAnnotObj, PeeledSHA: shaAnnotPeel},
			},
		),
		Entry("non-tag refs are skipped",
			shaLight+"\trefs/heads/main\n"+shaLight+"\trefs/tags/light\n",
			map[string]RemoteTagRef{
				"light": {ObjectSHA: shaLight},
			},
		),
	)

	It("returns error on malformed line", func() {
		_, err := parseLsRemoteTagsOutput("garbage\n")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("malformed"))
	})
})

var _ = Describe("shallow shell git helpers", func() {
	var (
		sourcePath string
		mirrorPath string
		headSHA    string
	)

	revParse := func(ctx SpecContext, dir string, args ...string) string {
		cmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: dir}, append([]string{"rev-parse"}, args...)...)
		Expect(cmd.Run(ctx)).To(Succeed())
		return strings.TrimSpace(cmd.OutBuf.String())
	}

	BeforeEach(func(ctx SpecContext) {
		sourcePath = filepath.Join(SuiteData.TestDirPath, "source")
		mirrorPath = filepath.Join(SuiteData.TestDirPath, "mirror")
		utils.MkdirAll(sourcePath)

		gitInSource := func(args ...string) {
			utils.RunSucceedCommand(ctx, sourcePath, "git", append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false"}, args...)...)
		}

		utils.RunSucceedCommand(ctx, sourcePath, "git", "-c", "init.defaultBranch=main", "init")
		gitInSource("checkout", "-b", "main")
		gitInSource("commit", "--allow-empty", "-m", "c1")
		gitInSource("commit", "--allow-empty", "-m", "c2")
		gitInSource("tag", "light")
		gitInSource("tag", "-a", "annot", "-m", "msg")
		gitInSource("-c", "advice.nestedTag=false", "tag", "-a", "nested", "-m", "msg", "annot")

		Expect(Init(ctx, Options{})).Should(Succeed())

		headSHA = utils.GetHeadCommit(ctx, sourcePath)
	})

	Describe("InitBareRepoWithOrigin", func() {
		It("creates bare repo with origin url set to source path", func(ctx SpecContext) {
			Expect(InitBareRepoWithOrigin(ctx, mirrorPath, sourcePath)).To(Succeed())

			cmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: mirrorPath}, "remote", "get-url", "origin")
			Expect(cmd.Run(ctx)).To(Succeed())
			Expect(strings.TrimSpace(cmd.OutBuf.String())).To(Equal(sourcePath))
		})
	})

	Describe("LsRemoteTags", func() {
		It("returns light, annotated and nested tags with correct peel", func(ctx SpecContext) {
			tags, err := LsRemoteTags(ctx, sourcePath, LsRemoteTagsOptions{})
			Expect(err).ToNot(HaveOccurred())

			Expect(tags).To(HaveKey("light"))
			Expect(tags["light"].ObjectSHA).To(Equal(headSHA))
			Expect(tags["light"].PeeledSHA).To(BeEmpty())

			Expect(tags).To(HaveKey("annot"))
			Expect(tags["annot"].ObjectSHA).ToNot(BeEmpty())
			Expect(tags["annot"].ObjectSHA).ToNot(Equal(tags["annot"].PeeledSHA))
			Expect(tags["annot"].PeeledSHA).To(Equal(headSHA))

			Expect(tags).To(HaveKey("nested"))
			Expect(tags["nested"].PeeledSHA).To(Equal(headSHA))
		})
	})

	Describe("ShallowFetch", func() {
		It("fetches a tag with depth=1", func(ctx SpecContext) {
			Expect(InitBareRepoWithOrigin(ctx, mirrorPath, sourcePath)).To(Succeed())

			Expect(ShallowFetch(ctx, mirrorPath, []string{"+refs/tags/annot:refs/tags/annot"}, ShallowFetchOptions{})).To(Succeed())

			Expect(revParse(ctx, mirrorPath, "annot^{commit}")).To(Equal(headSHA))
			Expect(revParse(ctx, mirrorPath, "--is-shallow-repository")).To(Equal("true"))

			cmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: mirrorPath}, "rev-list", "--count", "annot")
			Expect(cmd.Run(ctx)).To(Succeed())
			Expect(strings.TrimSpace(cmd.OutBuf.String())).To(Equal("1"))
		})

		It("fetches by commit SHA when SHA is an advertised tip", func(ctx SpecContext) {
			Expect(InitBareRepoWithOrigin(ctx, mirrorPath, sourcePath)).To(Succeed())

			Expect(ShallowFetch(ctx, mirrorPath, []string{"+" + headSHA + ":refs/werf/commits/" + headSHA}, ShallowFetchOptions{})).To(Succeed())

			cmd := NewGitCmd(ctx, &GitCmdOptions{RepoDir: mirrorPath}, "cat-file", "-t", headSHA)
			Expect(cmd.Run(ctx)).To(Succeed())
			Expect(strings.TrimSpace(cmd.OutBuf.String())).To(Equal("commit"))
		})

		It("returns error for a nonexistent tag refspec", func(ctx SpecContext) {
			Expect(InitBareRepoWithOrigin(ctx, mirrorPath, sourcePath)).To(Succeed())

			err := ShallowFetch(ctx, mirrorPath, []string{"+refs/tags/does-not-exist:refs/tags/does-not-exist"}, ShallowFetchOptions{})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("UpdateRef", func() {
		It("sets refs/tags/movable to head commit sha", func(ctx SpecContext) {
			Expect(InitBareRepoWithOrigin(ctx, mirrorPath, sourcePath)).To(Succeed())
			Expect(ShallowFetch(ctx, mirrorPath, []string{"+refs/tags/annot:refs/tags/annot"}, ShallowFetchOptions{})).To(Succeed())

			Expect(UpdateRef(ctx, mirrorPath, "refs/tags/movable", headSHA)).To(Succeed())

			Expect(revParse(ctx, mirrorPath, "refs/tags/movable")).To(Equal(headSHA))
		})
	})
})

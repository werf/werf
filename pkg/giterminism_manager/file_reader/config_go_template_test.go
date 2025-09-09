package file_reader_test

import (
	"context"
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"go.uber.org/mock/gomock"

	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/giterminism_manager/file_reader"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/test/mock"
)

var _ = Describe("Template file functions", func() {
	t := GinkgoT()

	var (
		reader            file_reader.FileReader
		sharedOptions     *MocksharedOptions
		giterminismConfig *MockgiterminismConfig
		fileSystemLayer   *MockfileSystem
		pathMatcher       *mock.MockPathMatcher
		gitRepo           *mock.MockGitRepo
		fileInfo          *MockFileInfo
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(t, gomock.WithOverridableExpectations())

		sharedOptions = NewMocksharedOptions(ctrl)
		giterminismConfig = NewMockgiterminismConfig(ctrl)
		fileSystemLayer = NewMockfileSystem(ctrl)
		pathMatcher = mock.NewMockPathMatcher(ctrl)
		gitRepo = mock.NewMockGitRepo(ctrl)
		fileInfo = NewMockFileInfo(ctrl)

		reader = file_reader.NewFileReader(sharedOptions)
		reader.SetGiterminismConfig(giterminismConfig)
		reader.SetFileSystemLayer(fileSystemLayer)
	})

	DescribeTable("ConfigGoTemplateFilesExists",
		func(ctx context.Context, setupMock MockFunc, expectation TestExpectationTemplateFileFunc) {
			ctx = logging.WithLogger(ctx)

			commit := "git head commit"
			relPath := "some.txt"

			sharedOptions.EXPECT().LocalGitRepo().Return(gitRepo).AnyTimes()

			sharedOptions.EXPECT().ProjectDir().Return(t.TempDir()).AnyTimes()
			sharedOptions.EXPECT().RelativeToGitProjectDir().Return(t.TempDir()).AnyTimes()
			sharedOptions.EXPECT().Dev().Return(false).AnyTimes()
			sharedOptions.EXPECT().HeadCommit(ctx).Return(commit).AnyTimes()
			sharedOptions.EXPECT().LooseGiterminism().Return(false).AnyTimes()

			giterminismConfig.EXPECT().UncommittedConfigGoTemplateRenderingFilePathMatcher().Return(pathMatcher)

			setupMock(ctx, commit, relPath)

			ok, err := reader.ConfigGoTemplateFilesExists(ctx, relPath)

			Expect(ok).To(expectation.OkMatcher)
			Expect(err).To(expectation.ErrMatcher)
		},
		Entry("should return false if the file does not exist in the Git repository (missing, UntrackedFilesError)",
			func(ctx context.Context, commit, relPath string) {
				pathMatcher.EXPECT().IsPathMatched(relPath).Return(false)

				gitRepo.EXPECT().ValidateStatusResult(ctx, gomock.Any()).Return(git_repo.UntrackedFilesFoundError{PathList: []string{"foo.txt"}})
			},
			TestExpectationTemplateFileFunc{
				OkMatcher:  BeFalse(),
				ErrMatcher: BeNil(),
			},
		),
		Entry("should return false if the file is not found in the project repository (missing, FileNotFoundInProjectRepositoryError)",
			func(ctx context.Context, commit, relPath string) {
				pathMatcher.EXPECT().IsPathMatched(relPath).Return(false)

				gitRepo.EXPECT().ValidateStatusResult(ctx, gomock.Any()).Return(nil)
				gitRepo.EXPECT().IsCommitTreeEntryExist(ctx, commit, gomock.Any()).Return(false, nil)
				gitRepo.EXPECT().IsCommitFileExist(ctx, commit, gomock.Any()).Return(false, nil)
			},
			TestExpectationTemplateFileFunc{
				OkMatcher:  BeFalse(),
				ErrMatcher: BeNil(),
			},
		),
		Entry("should return false if the file is not found in the project directory when --loose-giterminism is enabled (missing, FileNotFoundInProjectDirectoryError)",
			func(ctx context.Context, commit, relPath string) {
				sharedOptions.EXPECT().LooseGiterminism().Return(true).AnyTimes() // override

				fileSystemLayer.EXPECT().Lstat(gomock.Any()).Return(fileInfo, nil)
				fileInfo.EXPECT().Mode().Return(fs.ModePerm)
				fileSystemLayer.EXPECT().FileExists(gomock.Any()).Return(false, nil)
			},
			TestExpectationTemplateFileFunc{
				OkMatcher:  BeFalse(),
				ErrMatcher: BeNil(),
			},
		),
		Entry("should return false if the file is found and patch is not matched (not matched, FileNotAcceptedError)",
			func(ctx context.Context, commit, relPath string) {
				sharedOptions.EXPECT().LooseGiterminism().DoAndReturn(onFirstCallReturnTrueOtherwiseReturnFalse()).Times(2)

				pathMatcher.EXPECT().IsPathMatched(relPath).Return(false)
			},
			TestExpectationTemplateFileFunc{
				OkMatcher:  BeFalse(),
				ErrMatcher: BeNil(),
			},
		),
		Entry("should return false if the file does not exist in the Git repository (not committed, UncommittedFilesError)",
			func(ctx context.Context, commit, relPath string) {
				pathMatcher.EXPECT().IsPathMatched(relPath).Return(false)

				gitRepo.EXPECT().ValidateStatusResult(ctx, gomock.Any()).Return(git_repo.UncommittedFilesFoundError{PathList: []string{"foo.txt"}})
			},
			TestExpectationTemplateFileFunc{
				OkMatcher:  BeFalse(),
				ErrMatcher: BeNil(),
			},
		),
		Entry("should return true if the file present in the Git repository (committed)",
			func(ctx context.Context, commit, relPath string) {
				pathMatcher.EXPECT().IsPathMatched(relPath).Return(false)

				gitRepo.EXPECT().ValidateStatusResult(ctx, gomock.Any()).Return(nil)
				gitRepo.EXPECT().IsCommitTreeEntryExist(ctx, gomock.Any(), gomock.Any()).Return(true, nil)
				// gitRepo.EXPECT().IsCommitFileExist(ctx, gomock.Any(), gomock.Any()).Return(true, nil)
				gitRepo.EXPECT().ResolveAndCheckCommitFilePath(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(relPath, nil)

				fileSystemLayer.EXPECT().Lstat(gomock.Any()).Return(fileInfo, nil)
				fileInfo.EXPECT().Mode().Return(fs.ModePerm)
				fileSystemLayer.EXPECT().FileExists(gomock.Any()).Return(true, nil)
			},
			TestExpectationTemplateFileFunc{
				OkMatcher:  BeTrue(),
				ErrMatcher: BeNil(),
			},
		),
		Entry("should return true if uncommitted file exists in the Git repository and --loose-giterminism enabled for that file",
			func(ctx context.Context, commit, relPath string) {
				sharedOptions.EXPECT().LooseGiterminism().Return(true).AnyTimes() // override

				fileSystemLayer.EXPECT().Lstat(gomock.Any()).Return(fileInfo, nil).Times(2)
				fileInfo.EXPECT().Mode().Return(fs.ModePerm).Times(2)
				fileSystemLayer.EXPECT().FileExists(gomock.Any()).Return(true, nil).Times(2)
			},
			TestExpectationTemplateFileFunc{
				OkMatcher:  BeTrue(),
				ErrMatcher: BeNil(),
			},
		),
		Entry("should return true if the uncommitted file exists in the Git repository and --dev",
			func(ctx context.Context, commit, relPath string) {
				sharedOptions.EXPECT().Dev().Return(true).AnyTimes() // override

				pathMatcher.EXPECT().IsPathMatched(relPath).Return(false)

				gitRepo.EXPECT().IsCommitTreeEntryExist(ctx, gomock.Any(), gomock.Any()).Return(true, nil)
				gitRepo.EXPECT().ResolveAndCheckCommitFilePath(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(relPath, nil)

				fileSystemLayer.EXPECT().Lstat(gomock.Any()).Return(fileInfo, nil)
				fileInfo.EXPECT().Mode().Return(fs.ModePerm)
				fileSystemLayer.EXPECT().FileExists(gomock.Any()).Return(true, nil)
			},
			TestExpectationTemplateFileFunc{
				OkMatcher:  BeTrue(),
				ErrMatcher: BeNil(),
			},
		),
	)

	DescribeTable("ConfigGoTemplateFilesIsDir",
		func(ctx context.Context, setupMock MockFunc, expectation TestExpectationTemplateFileFunc) {
			ctx = logging.WithLogger(ctx)

			commit := "git head commit"
			relPath := "some-dir"

			sharedOptions.EXPECT().LocalGitRepo().Return(gitRepo).AnyTimes()

			sharedOptions.EXPECT().ProjectDir().Return(t.TempDir()).AnyTimes()
			sharedOptions.EXPECT().RelativeToGitProjectDir().Return(t.TempDir()).AnyTimes()
			sharedOptions.EXPECT().Dev().Return(false).AnyTimes()
			sharedOptions.EXPECT().HeadCommit(ctx).Return(commit).AnyTimes()
			sharedOptions.EXPECT().LooseGiterminism().Return(false).AnyTimes()

			giterminismConfig.EXPECT().UncommittedConfigGoTemplateRenderingFilePathMatcher().Return(pathMatcher)

			setupMock(ctx, commit, relPath)

			ok, err := reader.ConfigGoTemplateFilesIsDir(ctx, relPath)

			Expect(ok).To(expectation.OkMatcher)
			Expect(err).To(expectation.ErrMatcher)
		},
		Entry("should return false if dir does not exist in the Git repository (missing, UntrackedFilesError)",
			func(ctx context.Context, commit, relPath string) {
				pathMatcher.EXPECT().IsPathMatched(relPath).Return(false)

				gitRepo.EXPECT().ValidateStatusResult(ctx, gomock.Any()).Return(git_repo.UntrackedFilesFoundError{PathList: []string{"foo.txt"}})
			},
			TestExpectationTemplateFileFunc{
				OkMatcher:  BeFalse(),
				ErrMatcher: BeNil(),
			},
		),
		Entry("should return false if the file is not found in the project repository (missing, FileNotFoundInProjectRepositoryError)",
			func(ctx context.Context, commit, relPath string) {
				pathMatcher.EXPECT().IsPathMatched(relPath).Return(false)

				gitRepo.EXPECT().ValidateStatusResult(ctx, gomock.Any()).Return(nil)
				gitRepo.EXPECT().IsCommitTreeEntryExist(ctx, commit, gomock.Any()).Return(false, nil)
				gitRepo.EXPECT().IsCommitFileExist(ctx, commit, gomock.Any()).Return(false, nil)
			},
			TestExpectationTemplateFileFunc{
				OkMatcher:  BeFalse(),
				ErrMatcher: BeNil(),
			},
		),
		Entry("should return false if the file is not found in the project directory when --loose-giterminism is enabled (missing, FileNotFoundInProjectDirectoryError)",
			func(ctx context.Context, commit, relPath string) {
				sharedOptions.EXPECT().LooseGiterminism().Return(true).AnyTimes() // override

				fileSystemLayer.EXPECT().Lstat(gomock.Any()).Return(fileInfo, nil)
				fileInfo.EXPECT().Mode().Return(fs.ModePerm)
				fileSystemLayer.EXPECT().DirExists(gomock.Any()).Return(false, nil)
			},
			TestExpectationTemplateFileFunc{
				OkMatcher:  BeFalse(),
				ErrMatcher: BeNil(),
			},
		),
		Entry("should return false if the file is found and patch is not matched (not matched, FileNotAcceptedError)",
			func(ctx context.Context, commit, relPath string) {
				sharedOptions.EXPECT().LooseGiterminism().DoAndReturn(onFirstCallReturnTrueOtherwiseReturnFalse()).Times(2)

				pathMatcher.EXPECT().IsPathMatched(relPath).Return(false)
			},
			TestExpectationTemplateFileFunc{
				OkMatcher:  BeFalse(),
				ErrMatcher: BeNil(),
			},
		),
		Entry("should return false if dir does not exist in the Git repository (not committed, UncommittedFilesError)",
			func(ctx context.Context, commit, relPath string) {
				pathMatcher.EXPECT().IsPathMatched(relPath).Return(false)

				gitRepo.EXPECT().ValidateStatusResult(ctx, gomock.Any()).Return(git_repo.UncommittedFilesFoundError{PathList: []string{"foo.txt"}})
			},
			TestExpectationTemplateFileFunc{
				OkMatcher:  BeFalse(),
				ErrMatcher: BeNil(),
			},
		),
		Entry("should return false for a regular file (exist and committed)",
			func(ctx context.Context, commit, relPath string) {
				pathMatcher.EXPECT().IsPathMatched(relPath).Return(false)

				gitRepo.EXPECT().ValidateStatusResult(ctx, gomock.Any()).Return(nil)
				gitRepo.EXPECT().IsCommitTreeEntryExist(ctx, gomock.Any(), gomock.Any()).Return(true, nil)
				gitRepo.EXPECT().ResolveAndCheckCommitFilePath(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(relPath, nil)

				fileSystemLayer.EXPECT().Lstat(gomock.Any()).Return(fileInfo, nil)
				fileInfo.EXPECT().Mode().Return(fs.ModePerm)
				fileSystemLayer.EXPECT().DirExists(gomock.Any()).Return(false, nil)
			},
			TestExpectationTemplateFileFunc{
				OkMatcher:  BeFalse(),
				ErrMatcher: BeNil(),
			},
		),
		Entry("should return true if uncommitted dir exists in the Git repository and --loose-giterminism enabled for that dir",
			func(ctx context.Context, commit, relPath string) {
				sharedOptions.EXPECT().LooseGiterminism().Return(true).AnyTimes() // override

				fileSystemLayer.EXPECT().Lstat(gomock.Any()).Return(fileInfo, nil).Times(2)
				fileInfo.EXPECT().Mode().Return(fs.ModePerm).Times(2)
				fileSystemLayer.EXPECT().DirExists(gomock.Any()).Return(true, nil).Times(2)
			},
			TestExpectationTemplateFileFunc{
				OkMatcher:  BeTrue(),
				ErrMatcher: BeNil(),
			},
		),
		Entry("should return true if the uncommitted dir exists in the Git repository and --dev",
			func(ctx context.Context, commit, relPath string) {
				sharedOptions.EXPECT().Dev().Return(true).AnyTimes() // override

				pathMatcher.EXPECT().IsPathMatched(relPath).Return(false)

				gitRepo.EXPECT().IsCommitTreeEntryExist(ctx, gomock.Any(), gomock.Any()).Return(true, nil)
				gitRepo.EXPECT().ResolveAndCheckCommitFilePath(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(relPath, nil)

				fileSystemLayer.EXPECT().Lstat(gomock.Any()).Return(fileInfo, nil)
				fileInfo.EXPECT().Mode().Return(fs.ModePerm)
				fileSystemLayer.EXPECT().DirExists(gomock.Any()).Return(true, nil)
			},
			TestExpectationTemplateFileFunc{
				OkMatcher:  BeTrue(),
				ErrMatcher: BeNil(),
			},
		),
	)
})

type TestExpectationTemplateFileFunc struct {
	OkMatcher  types.GomegaMatcher
	ErrMatcher types.GomegaMatcher
}

type MockFunc func(ctx context.Context, commit, relPath string)

func onFirstCallReturnTrueOtherwiseReturnFalse() func() bool {
	callCount := 0
	return func() bool {
		callCount++
		return callCount == 1
	}
}

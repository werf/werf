package git_test

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"

	"github.com/alessio/shellescape"
	. "github.com/onsi/ginkgo/v2"

	"github.com/werf/werf/v2/test/pkg/utils"
	utilsDocker "github.com/werf/werf/v2/test/pkg/utils/docker"
)

var _ = Describe("cleanup empty directories with git patch apply", func() {
	var fixturesPathParts []string
	gitToPath := "/app"

	type removingEmptyDirectoriesEntry struct {
		dirToAdd        string
		shouldBeDeleted []string
		shouldBeSkipped []string
		skipOnWindows   bool
	}

	removingEmptyDirectoriesItBody := func(fixturePathFolder string) func(removingEmptyDirectoriesEntry) {
		return func(entry removingEmptyDirectoriesEntry) {
			if entry.skipOnWindows && runtime.GOOS == "windows" {
				Skip("skip on windows")
			}

			commonBeforeEach(utils.FixturePath(append(fixturesPathParts, fixturePathFolder)...))

			projectAddedFilePath := filepath.Join(entry.dirToAdd, "file")
			containerAddedDirPath := path.Join(gitToPath, entry.dirToAdd)

			By(fmt.Sprintf("Add file %s", shellescape.Quote(projectAddedFilePath)))
			createAndCommitFile(filepath.Join(SuiteData.TestDirPath, entry.dirToAdd), "file", 12)

			By("Build and cache source code in gitArchive stage")
			utils.RunSucceedCommand(
				SuiteData.TestDirPath,
				SuiteData.WerfBinPath,
				"build",
			)

			By(fmt.Sprintf("Check container directory %s exists", shellescape.Quote(containerAddedDirPath)))
			utilsDocker.CheckContainerDirectoryExists(SuiteData.WerfBinPath, SuiteData.TestDirPath, containerAddedDirPath)

			By(fmt.Sprintf("Remove file %s", shellescape.Quote(projectAddedFilePath)))

			utils.RunSucceedCommand(
				SuiteData.TestDirPath,
				"git",
				"rm", projectAddedFilePath,
			)

			utils.RunSucceedCommand(
				SuiteData.TestDirPath,
				"git",
				"commit", "-m", "Remove file "+projectAddedFilePath,
			)

			utils.RunSucceedCommand(
				SuiteData.TestDirPath,
				SuiteData.WerfBinPath,
				"build",
			)

			for _, relDirPath := range entry.shouldBeDeleted {
				containerDirPath := path.Join(gitToPath, relDirPath)
				By(fmt.Sprintf("Check container directory %s does not exist", shellescape.Quote(containerDirPath)))
				utilsDocker.CheckContainerDirectoryDoesNotExist(SuiteData.WerfBinPath, SuiteData.TestDirPath, containerDirPath)
			}

			for _, relDirPath := range entry.shouldBeSkipped {
				containerDirPath := path.Join(gitToPath, relDirPath)
				By(fmt.Sprintf("Check container directory %s exists", shellescape.Quote(containerDirPath)))
				utilsDocker.CheckContainerDirectoryExists(SuiteData.WerfBinPath, SuiteData.TestDirPath, containerDirPath)
			}
		}
	}

	BeforeEach(func() {
		fixturesPathParts = []string{"cleanup_empty_directories_with_git_patch_apply"}
	})

	DescribeTable("base",
		removingEmptyDirectoriesItBody("base"),
		Entry("should remove empty directory (dir)", removingEmptyDirectoriesEntry{
			dirToAdd:        "dir",
			shouldBeDeleted: []string{"dir"},
			shouldBeSkipped: []string{},
		}),
		Entry("should remove empty directories (dir/sub_dir)", removingEmptyDirectoriesEntry{
			dirToAdd:        "dir/sub_dir",
			shouldBeDeleted: []string{"dir/sub_dir", "dir"},
			shouldBeSkipped: []string{},
		}),
		Entry("should remove empty directories (dir/sub dir/sub dir with special ch@ra(c)ters? ())", removingEmptyDirectoriesEntry{
			dirToAdd:        "dir/sub dir/sub dir with special ch@ra(c)ters? ()",
			shouldBeDeleted: []string{"dir/sub dir/sub dir with special ch@ra(c)ters? ()", "dir/sub dir", "dir"},
			shouldBeSkipped: []string{},
			skipOnWindows:   true,
		}),
	)

	DescribeTable("processing directory created by user 'dir/dir_created_by_user'",
		removingEmptyDirectoriesItBody("skipping_user_directory"),
		Entry("should not remove directory (dir)", removingEmptyDirectoriesEntry{
			dirToAdd:        "dir",
			shouldBeDeleted: []string{},
			shouldBeSkipped: []string{"dir"},
		}),
		Entry("should remove only empty directory (dir/sub_dir)", removingEmptyDirectoriesEntry{
			dirToAdd:        "dir/sub_dir",
			shouldBeDeleted: []string{"dir/sub_dir"},
			shouldBeSkipped: []string{"dir"},
		}),
		Entry("should remove empty directories (dir/dir_created_by_user)", removingEmptyDirectoriesEntry{
			dirToAdd:        "dir/dir_created_by_user",
			shouldBeDeleted: []string{"dir/dir_created_by_user", "dir"},
			shouldBeSkipped: []string{},
		}),
	)
})

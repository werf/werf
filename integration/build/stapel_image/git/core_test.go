// +build integration

package git

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alessio/shellescape"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/integration/utils"
)

var _ = Describe("core", func() {
	var testDirPath string
	var fixturesPathParts []string
	gitToPath := "/app"

	BeforeEach(func() {
		testDirPath = tmpPath()
		fixturesPathParts = []string{"core"}
	})

	AfterEach(func() {
		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"stages", "purge", "-s", ":local", "--force",
		)
	})

	Describe("file lifecycle", func() {
		fileNameToAdd := "test"
		fileNameToAddAndModify := "test2"
		fileDataToAdd := []byte("test")
		fileDataToModify := []byte("test2")

		type fileLifecycleEntry struct {
			name   string
			data   []byte
			perm   os.FileMode
			delete bool
		}

		createFileFunc := func(filePath string, fileData []byte, filePerm os.FileMode) {
			utils.CreateFile(filePath, fileData)
			Ω(os.Chmod(filePath, filePerm)).Should(Succeed())
		}

		fileLifecycleEntryItBody := func(entry fileLifecycleEntry) {
			var commitMsg string

			filePath := filepath.Join(testDirPath, entry.name)
			if entry.delete {
				Ω(os.Remove(filePath)).Should(Succeed())
				commitMsg = "Delete file " + entry.name
			} else {
				createFileFunc(filePath, entry.data, entry.perm)
				commitMsg = "Add/Modify file " + entry.name
			}

			addAndCommitFile(testDirPath, entry.name, commitMsg)

			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"build",
			)

			var cmd []string
			dockerOptions := []string{"--rm"}

			if entry.delete {
				cmd = append(cmd, checkContainerFileCommand(path.Join(gitToPath, entry.name), false, false))
			} else {
				cmd = append(cmd, checkContainerFileCommand(path.Join(gitToPath, entry.name), false, true))
				cmd = append(cmd, fmt.Sprintf("diff <(stat -c %%a %s) <(echo %s)", path.Join(gitToPath, entry.name), strconv.FormatUint(uint64(entry.perm), 8)))
				cmd = append(cmd, fmt.Sprintf("diff %s %s", path.Join(gitToPath, entry.name), "/source"))

				dockerOptions = append(dockerOptions, fmt.Sprintf("-v %s:%s", filePath, "/source"))
			}

			utils.RunSucceedCommand(
				testDirPath,
				werfBinPath,
				"run", "--docker-options", strings.Join(dockerOptions, " "), "--", "bash", "-ec", strings.Join(cmd, " && "),
			)
		}

		BeforeEach(func() {
			fixturesPathParts = append(fixturesPathParts, "file_lifecycle")
			commonBeforeEach(testDirPath, fixturePath(fixturesPathParts...))
		})

		DescribeTable("processing file with archive apply",
			fileLifecycleEntryItBody,
			Entry("should add file (0755)", fileLifecycleEntry{
				name: fileNameToAdd,
				data: fileDataToAdd,
				perm: 0755,
			}),
			Entry("should add file (0644)", fileLifecycleEntry{
				name: fileNameToAdd,
				data: fileDataToAdd,
				perm: 0644,
			}),
		)

		Context("when gitArchive stage with file is built", func() {
			BeforeEach(func() {
				createFileFunc(filepath.Join(testDirPath, fileNameToAddAndModify), fileDataToAdd, 0755)
				addAndCommitFile(testDirPath, fileNameToAddAndModify, "Add file "+fileNameToAddAndModify)

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)
			})

			DescribeTable("processing file with patch apply",
				fileLifecycleEntryItBody,
				Entry("should add file (0755)", fileLifecycleEntry{
					name: fileNameToAdd,
					data: fileDataToAdd,
					perm: 0755,
				}),
				Entry("should add file (0644)", fileLifecycleEntry{
					name: fileNameToAdd,
					data: fileDataToAdd,
					perm: 0644,
				}),
				Entry("should modify file", fileLifecycleEntry{
					name: fileNameToAddAndModify,
					data: fileDataToModify,
					perm: 0755,
				}),
				Entry("should change file permission (0755->0644)", fileLifecycleEntry{
					name: fileNameToAddAndModify,
					data: fileDataToAdd,
					perm: 0644,
				}),
				Entry("should modify and change file permission (0755->0644)", fileLifecycleEntry{
					name: fileNameToAddAndModify,
					data: fileDataToModify,
					perm: 0644,
				}),
				Entry("should delete file", fileLifecycleEntry{
					name:   fileNameToAddAndModify,
					delete: true,
				}),
			)
		})

		Context("when file is symlink", func() {
			linkToAdd := "werf.yaml"
			linkToModify := "none"

			type symlinkFileLifecycleEntry struct {
				name   string
				link   string
				delete bool
			}

			symlinkFileLifecycleEntryItBody := func(entry symlinkFileLifecycleEntry) {
				var commitMsg string

				filePath := filepath.Join(testDirPath, entry.name)
				if entry.delete {
					Ω(os.Remove(filePath)).Should(Succeed())
					commitMsg = "Delete file " + entry.name
				} else {
					if _, err := os.Lstat(filePath); err == nil {
						Ω(os.Remove(filePath)).Should(Succeed())
					}

					Ω(os.Symlink(entry.link, filePath)).Should(Succeed())
					commitMsg = "Add/Modify file " + entry.name
				}

				addAndCommitFile(testDirPath, entry.name, commitMsg)

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)

				var cmd []string
				if entry.delete {
					cmd = append(cmd, checkContainerSymlinkFileCommand(path.Join(gitToPath, entry.name), false))
				} else {
					cmd = append(cmd, checkContainerSymlinkFileCommand(path.Join(gitToPath, entry.name), true))
					readlinkCmd := fmt.Sprintf("readlink %s", path.Join(gitToPath, entry.name))
					cmd = append(cmd, fmt.Sprintf("diff <(%s) <(echo %s)", readlinkCmd, entry.link))
				}

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"run", "--docker-options", "--rm", "--", "bash", "-ec", strings.Join(cmd, " && "),
				)
			}

			DescribeTable("processing symlink file with archive apply",
				symlinkFileLifecycleEntryItBody,
				Entry("should add symlink", symlinkFileLifecycleEntry{
					name: fileNameToAdd,
					link: linkToAdd,
				}),
			)

			Context("when gitArchive stage with file is built", func() {
				BeforeEach(func() {
					Ω(os.Symlink(linkToAdd, filepath.Join(testDirPath, fileNameToAddAndModify))).Should(Succeed())
					addAndCommitFile(testDirPath, fileNameToAddAndModify, "Add file "+fileNameToAddAndModify)

					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build",
					)
				})

				DescribeTable("processing symlink file with patch apply",
					symlinkFileLifecycleEntryItBody,
					Entry("should add symlink", symlinkFileLifecycleEntry{
						name: fileNameToAdd,
						link: linkToAdd,
					}),
					Entry("should modify file", symlinkFileLifecycleEntry{
						name: fileNameToAddAndModify,
						link: linkToModify,
					}),
					Entry("should delete file", symlinkFileLifecycleEntry{
						name:   fileNameToAddAndModify,
						delete: true,
					}))
			})
		})
	})

	Describe("removing empty directories with git patch apply", func() {
		type removingEmptyDirectoriesEntry struct {
			dirToAdd        string
			shouldBeDeleted []string
			shouldBeSkipped []string
		}

		removingEmptyDirectoriesItBody := func(fixturePathFolder string) func(removingEmptyDirectoriesEntry) {
			return func(entry removingEmptyDirectoriesEntry) {
				commonBeforeEach(testDirPath, fixturePath(append(fixturesPathParts, fixturePathFolder)...))

				projectAddedFilePath := filepath.Join(entry.dirToAdd, "file")
				containerAddedDirPath := path.Join(gitToPath, entry.dirToAdd)

				By(fmt.Sprintf("Add file %s", shellescape.Quote(projectAddedFilePath)))
				createAndCommitFile(filepath.Join(testDirPath, entry.dirToAdd), "file", 12)

				By("Build and cache source code in gitArchive stage")
				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)

				By(fmt.Sprintf("Check container directory %s exists", shellescape.Quote(containerAddedDirPath)))
				checkContainerDirectoryExists(testDirPath, containerAddedDirPath)

				By(fmt.Sprintf("Remove file %s", shellescape.Quote(projectAddedFilePath)))

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"rm", projectAddedFilePath,
				)

				utils.RunSucceedCommand(
					testDirPath,
					"git",
					"commit", "-m", "Remove file "+projectAddedFilePath,
				)

				utils.RunSucceedCommand(
					testDirPath,
					werfBinPath,
					"build",
				)

				for _, relDirPath := range entry.shouldBeDeleted {
					containerDirPath := path.Join(gitToPath, relDirPath)
					By(fmt.Sprintf("Check container directory %s does not exist", shellescape.Quote(containerDirPath)))
					checkContainerDirectoryDoesNotExist(testDirPath, containerDirPath)
				}

				for _, relDirPath := range entry.shouldBeSkipped {
					containerDirPath := path.Join(gitToPath, relDirPath)
					By(fmt.Sprintf("Check container directory %s exists", shellescape.Quote(containerDirPath)))
					checkContainerDirectoryExists(testDirPath, containerDirPath)
				}
			}
		}

		BeforeEach(func() {
			fixturesPathParts = append(fixturesPathParts, "removing_empty_directories")
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
})

func checkContainerDirectoryExists(projectPath, containerDirPath string) {
	checkContainerDirectory(projectPath, containerDirPath, true)
}

func checkContainerDirectoryDoesNotExist(projectPath, containerDirPath string) {
	checkContainerDirectory(projectPath, containerDirPath, false)
}

func checkContainerDirectory(projectPath, containerDirPath string, exist bool) {
	cmd := checkContainerFileCommand(containerDirPath, true, exist)

	utils.RunSucceedCommand(
		projectPath,
		werfBinPath,
		"run", "--docker-options", "--rm", "--", "bash", "-ec", cmd,
	)
}

func checkContainerFileCommand(containerDirPath string, directory bool, exist bool) string {
	var cmd string
	var flag string

	if directory {
		flag = "-d"
	} else {
		flag = "-f"
	}

	if exist {
		cmd = fmt.Sprintf("test %s %s", flag, shellescape.Quote(containerDirPath))
	} else {
		cmd = fmt.Sprintf("! test %s %s", flag, shellescape.Quote(containerDirPath))
	}

	return cmd
}

func checkContainerSymlinkFileCommand(containerDirPath string, exist bool) string {
	var cmd string

	if exist {
		cmd = fmt.Sprintf("test -h %s", shellescape.Quote(containerDirPath))
	} else {
		cmd = fmt.Sprintf("! test -h %s", shellescape.Quote(containerDirPath))
	}

	return cmd
}

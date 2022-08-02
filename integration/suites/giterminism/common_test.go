package giterminism_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

func CommonBeforeEach() {
	gitInit()
	utils.CopyIn(utils.FixturePath("default"), SuiteData.TestDirPath)
	gitAddAndCommit("werf-giterminism.yaml")
	gitAddAndCommit("werf.yaml")
}

func gitInit() {
	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		"git",
		"init",
	)

	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		"git",
		"commit", "--allow-empty", "-m", "Initial commit",
	)
}

func gitAddAndCommit(relPath string) {
	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		"git",
		"add", relPath,
	)

	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		"git",
		"commit", "-m", fmt.Sprint("Update ", relPath),
	)
}

func fileCreateOrAppend(relPath, content string) {
	path := filepath.Join(SuiteData.TestDirPath, relPath)

	Ω(os.MkdirAll(filepath.Dir(path), 0o777)).ShouldNot(HaveOccurred())

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	Ω(err).ShouldNot(HaveOccurred())

	_, err = f.WriteString(content)
	Ω(err).ShouldNot(HaveOccurred())

	Ω(f.Close()).ShouldNot(HaveOccurred())
}

func symlinkFileCreateOrModify(relPath, link string) {
	relPath = filepath.ToSlash(relPath)
	link = filepath.ToSlash(link)

	symlinkFileCreateOrModifyAndAdd(relPath, link)

	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		"git",
		"rm", "--cached", relPath,
	)
}

func symlinkFileCreateOrModifyAndAdd(relPath, link string) {
	relPath = filepath.ToSlash(relPath)
	link = filepath.ToSlash(link)

	hashBytes, _ := utils.RunCommandWithOptions(
		SuiteData.TestDirPath,
		"git",
		[]string{"hash-object", "-w", "--stdin"},
		utils.RunCommandOptions{
			ToStdin:       link,
			ShouldSucceed: true,
		},
	)

	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		"git",
		"update-index", "--add", "--cacheinfo", "120000", string(bytes.TrimSpace(hashBytes)), relPath,
	)

	utils.RunSucceedCommand(
		SuiteData.TestDirPath,
		"git",
		"checkout", relPath,
	)
}

func getLinkTo(linkFile, targetPath string) string {
	target, err := filepath.Rel(filepath.Dir(linkFile), targetPath)
	if err != nil {
		panic(err)
	}
	return filepath.ToSlash(target)
}

package giterminism_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
)

func CommonBeforeEach(ctx context.Context) {
	gitInit(ctx)
	utils.CopyIn(utils.FixturePath("default"), SuiteData.TestDirPath)
	gitAddAndCommit(ctx, "werf-giterminism.yaml")
	gitAddAndCommit(ctx, "werf.yaml")
}

func gitInit(ctx context.Context) {
	utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "init")

	utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "commit", "--allow-empty", "-m", "Initial commit")
}

func gitAddAndCommit(ctx context.Context, relPath string) {
	utils.RunSucceedCommand(
		ctx,
		SuiteData.TestDirPath,
		"git",
		"add", relPath,
	)

	utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "commit", "-m", fmt.Sprint("Update ", relPath))
}

func fileCreateOrAppend(relPath, content string) {
	path := filepath.Join(SuiteData.TestDirPath, relPath)

	Expect(os.MkdirAll(filepath.Dir(path), 0o777)).ShouldNot(HaveOccurred())

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	Expect(err).ShouldNot(HaveOccurred())

	_, err = f.WriteString(content)
	Expect(err).ShouldNot(HaveOccurred())

	Expect(f.Close()).ShouldNot(HaveOccurred())
}

func symlinkFileCreateOrModify(ctx context.Context, relPath, link string) {
	relPath = filepath.ToSlash(relPath)
	link = filepath.ToSlash(link)

	symlinkFileCreateOrModifyAndAdd(ctx, relPath, link)

	utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "rm", "--cached", relPath)
}

func symlinkFileCreateOrModifyAndAdd(ctx context.Context, relPath, link string) {
	relPath = filepath.ToSlash(relPath)
	link = filepath.ToSlash(link)

	hashBytes, _ := utils.RunCommandWithOptions(ctx, SuiteData.TestDirPath, "git", []string{"hash-object", "-w", "--stdin"}, utils.RunCommandOptions{
		ToStdin:       link,
		ShouldSucceed: true,
	})

	utils.RunSucceedCommand(ctx, SuiteData.TestDirPath, "git", "update-index", "--add", "--cacheinfo", "120000", string(bytes.TrimSpace(hashBytes)), relPath)

	utils.RunSucceedCommand(
		ctx,
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
